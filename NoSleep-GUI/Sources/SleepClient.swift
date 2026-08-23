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

    private struct ProcessResult {
        let exitCode: Int32
        let stdout: Data
        let stderr: String
    }

    // A nil timeout means "wait as long as it takes" — the setup prompt waits on the user.
    private func runProcess(executable: String, arguments: [String], timeout: TimeInterval?) async throws -> ProcessResult {
        let process = Process()
        process.executableURL = URL(fileURLWithPath: executable)
        process.arguments = arguments

        let outPipe = Pipe()
        let errPipe = Pipe()
        process.standardOutput = outPipe
        process.standardError = errPipe

        return try await withCheckedThrowingContinuation { continuation in
            var isResumed = false
            let lock = NSLock()

            func resume(with result: Result<ProcessResult, Error>) {
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
                let outData = outPipe.fileHandleForReading.readDataToEndOfFile()
                resume(with: .success(ProcessResult(exitCode: p.terminationStatus, stdout: outData, stderr: stderr)))
            }

            do {
                try process.run()

                if let timeout {
                    DispatchQueue.global().asyncAfter(deadline: .now() + timeout) {
                        if process.isRunning {
                            process.terminate()
                            resume(with: .failure(ClientError.timeout))
                        }
                    }
                }
            } catch {
                resume(with: .failure(error))
            }
        }
    }

    private func runScript(args: [String]) async throws -> Data {
        guard let scriptURL = Bundle.main.url(forResource: "nosleep", withExtension: "sh") else {
            throw ClientError.scriptNotFound
        }

        // Run via bash to avoid chmod +x bundle issues, with the 5-second timeout matching the Go client
        let result = try await runProcess(
            executable: "/bin/bash",
            arguments: [scriptURL.path] + args,
            timeout: 5.0
        )
        guard result.exitCode == 0 else {
            throw ClientError.executionFailed(exitCode: result.exitCode, stderr: result.stderr)
        }
        return result.stdout
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

    // The user name and script path stay out of the script source: osascript hands them over as
    // `on run argv` items, so quotes or backslashes in either cannot break out of the AppleScript
    // literal. USER is passed on so the bash script writes the sudoers rule for the GUI user, not root.
    private static let setupScriptSource = """
    on run argv
        do shell script "USER=" & quoted form of (item 1 of argv) & " /bin/bash " & quoted form of (item 2 of argv) & " setup" with administrator privileges
    end run
    """

    func setup() async throws {
        guard let scriptURL = Bundle.main.url(forResource: "nosleep", withExtension: "sh") else {
            throw ClientError.scriptNotFound
        }

        // "--" stops osascript from reading a value that begins with "-" as one of its own options.
        // No timeout: the administrator prompt waits on the user.
        let result = try await runProcess(
            executable: "/usr/bin/osascript",
            arguments: ["-e", Self.setupScriptSource, "--", NSUserName(), scriptURL.path],
            timeout: nil
        )
        guard result.exitCode == 0 else {
            throw ClientError.executionFailed(exitCode: result.exitCode, stderr: Self.readableScriptError(result.stderr))
        }
    }

    // osascript prefixes failures with a source location, e.g. "2:44: execution error: User canceled. (-128)"
    private static func readableScriptError(_ stderr: String) -> String {
        let message = stderr.replacingOccurrences(
            of: "^[0-9]+:[0-9]+: (execution|syntax) error: ",
            with: "",
            options: [.regularExpression]
        )
        return message.isEmpty ? "Setup was not completed." : message
    }

    func needsSetup() -> Bool {
        return !FileManager.default.fileExists(atPath: "/etc/sudoers.d/nosleep")
    }
}
