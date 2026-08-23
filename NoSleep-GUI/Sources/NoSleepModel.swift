import SwiftUI
import Observation

@MainActor
@Observable
final class NoSleepModel {
    var sleepState: SleepState = .unknown
    // Any command in flight, background poll included; the controls key off this
    var isBusy: Bool = false
    // Only a user-initiated command, which is what the progress spinner stands in for
    var isWorking: Bool = false
    var errorMessage: String = ""
    var needsSetup: Bool = false

    @ObservationIgnored private let client = SleepClient()

    func refreshStatus(showProgress: Bool = true) async {
        guard beginOperation(showProgress: showProgress) else { return }
        defer { endOperation() }

        await loadStatus()
    }

    func runSetup() async {
        guard beginOperation(showProgress: true) else { return }
        defer { endOperation() }

        do {
            try await client.setup()
            await loadStatus()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func turnOn() async {
        await setSleepDisabled(true)
    }

    func turnOff() async {
        await setSleepDisabled(false)
    }

    private func setSleepDisabled(_ disabled: Bool) async {
        guard beginOperation(showProgress: true) else { return }
        defer { endOperation() }

        needsSetup = await client.needsSetup()
        if needsSetup { return }

        do {
            if disabled { try await client.turnOn() } else { try await client.turnOff() }
            // Reflect the command that just succeeded, so a failed re-read cannot leave the
            // switch showing the opposite of what the system is actually doing
            sleepState = disabled ? .awake : .normal
        } catch {
            errorMessage = error.localizedDescription
            sleepState = .unknown
        }

        do {
            sleepState = try await client.status()
        } catch {
            if errorMessage.isEmpty {
                errorMessage = "Sleep setting applied, but reading the status back failed: \(error.localizedDescription)"
            }
        }
    }

    // One command at a time: whoever passes this guard owns the flags until it ends
    private func beginOperation(showProgress: Bool) -> Bool {
        if isBusy { return false }
        isBusy = true
        isWorking = showProgress
        if showProgress { errorMessage = "" }
        return true
    }

    private func endOperation() {
        isWorking = false
        isBusy = false
    }

    private func loadStatus() async {
        needsSetup = await client.needsSetup()

        do {
            sleepState = try await client.status()
            errorMessage = "" // clear on successful sync
        } catch {
            sleepState = .unknown
            errorMessage = error.localizedDescription
        }
    }
}
