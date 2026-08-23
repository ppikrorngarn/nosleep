import SwiftUI
import Observation

@MainActor
@Observable
final class NoSleepModel {
    var sleepState: SleepState = .unknown
    var isWorking: Bool = false
    var errorMessage: String = ""
    var needsSetup: Bool = false

    @ObservationIgnored private let client = SleepClient()

    func refreshStatus(showProgress: Bool = true) async {
        if showProgress {
            isWorking = true
            errorMessage = ""
        }
        await loadStatus()
        if showProgress { isWorking = false }
    }

    func runSetup() async {
        isWorking = true
        errorMessage = ""
        do {
            try await client.setup()
            await loadStatus()
        } catch {
            errorMessage = error.localizedDescription
        }
        isWorking = false
    }

    func turnOn() async {
        await setSleepDisabled(true)
    }

    func turnOff() async {
        await setSleepDisabled(false)
    }

    private func setSleepDisabled(_ disabled: Bool) async {
        needsSetup = await client.needsSetup()
        if needsSetup { return }

        isWorking = true
        errorMessage = ""

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
        isWorking = false
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
