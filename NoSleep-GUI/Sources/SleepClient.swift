import Foundation

final class SleepClient {
    enum ClientError: LocalizedError {
        case scriptNotFound
        case timeout
        case executionFailed(exitCode: Int32, stderr: String)
        case invalidOutput
        
        var errorDescription: String? {
            switch self {
            case .scriptNotFound: return "Internal script not found in bundle."
            case .timeout: return "Command timed out."
            case .executionFailed(let code, let err): return "Execution failed (\(code)): \(err)"
            case .invalidOutput: return "Failed to parse script output."
            }
        }
    }

    private func runScript(args: [String]) async throws -> Data {
        guard let scriptURL = Bundle.main.url(forResource: "nosleep", withExtension: "sh") else {
            throw ClientError.scriptNotFound
        }

        let process = Process()
        // Run via bash to avoid chmod +x bundle issues
        process.executableURL = URL(fileURLWithPath: "/bin/bash")
        process.arguments = [scriptURL.path] + args

        let outPipe = Pipe()
        let errPipe = Pipe()
        process.standardOutput = outPipe
        process.standardError = errPipe

        return try await withCheckedThrowingContinuation { continuation in
            var isResumed = false
            let lock = NSLock()
            
            func resume(with result: Result<Data, Error>) {
                lock.lock()
                defer { lock.unlock() }
                if !isResumed {
                    isResumed = true
                    continuation.resume(with: result)
                }
            }

            process.terminationHandler = { p in
                let errData = errPipe.fileHandleForReading.readDataToEndOfFile()
                let stderr = String(data: errData, encoding: .utf8)?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
                
                if p.terminationStatus != 0 {
                    resume(with: .failure(ClientError.executionFailed(exitCode: p.terminationStatus, stderr: stderr)))
                } else {
                    let outData = outPipe.fileHandleForReading.readDataToEndOfFile()
                    resume(with: .success(outData))
                }
            }

            do {
                try process.run()
                
                // 5-second timeout matching Go client
                DispatchQueue.global().asyncAfter(deadline: .now() + 5.0) {
                    if process.isRunning {
                        process.terminate()
                        resume(with: .failure(ClientError.timeout))
                    }
                }
            } catch {
                resume(with: .failure(error))
            }
        }
    }

    func status() async throws -> SleepState {
        let data = try await runScript(args: ["status", "--json"])
        let resp = try JSONDecoder().decode(StatusResponse.self, from: data)
        return resp.parsedState
    }

    func turnOn() async throws {
        let data = try await runScript(args: ["on", "--json"])
        let resp = try JSONDecoder().decode(ActionResponse.self, from: data)
        if !resp.ok { throw ClientError.executionFailed(exitCode: 1, stderr: "Script reported failure via JSON") }
    }

    func turnOff() async throws {
        let data = try await runScript(args: ["off", "--json"])
        let resp = try JSONDecoder().decode(ActionResponse.self, from: data)
        if !resp.ok { throw ClientError.executionFailed(exitCode: 1, stderr: "Script reported failure via JSON") }
    }

    func setup() async throws {
        guard let scriptURL = Bundle.main.url(forResource: "nosleep", withExtension: "sh") else {
            throw ClientError.scriptNotFound
        }
        let user = NSUserName()
        
        // Pass USER explicitly so the bash script writes the sudoers rule for the GUI user, not root
        let scriptSource = """
        do shell script "USER=" & quoted form of "\(user)" & " /bin/bash " & quoted form of "\(scriptURL.path)" & " setup" with administrator privileges
        """
        
        if let appleScript = NSAppleScript(source: scriptSource) {
            var errorInfo: NSDictionary?
            appleScript.executeAndReturnError(&errorInfo)
            if let err = errorInfo {
                throw ClientError.executionFailed(exitCode: -1, stderr: "Setup failed: \(err)")
            }
        } else {
            throw ClientError.executionFailed(exitCode: -1, stderr: "Failed to compile AppleScript")
        }
    }

    func needsSetup() -> Bool {
        return !FileManager.default.fileExists(atPath: "/etc/sudoers.d/nosleep")
    }
}
