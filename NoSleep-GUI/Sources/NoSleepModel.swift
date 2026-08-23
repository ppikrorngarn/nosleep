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
        needsSetup = await client.needsSetup()
        
        do {
            sleepState = try await client.status()
            errorMessage = "" // clear on successful sync
        } catch {
            errorMessage = error.localizedDescription
        }
        if showProgress { isWorking = false }
    }
    
    func runSetup() async {
        isWorking = true
        errorMessage = ""
        do {
            try await client.setup()
            await refreshStatus()
        } catch {
            errorMessage = error.localizedDescription
            isWorking = false
        }
    }
    
    func turnOn() async {
        needsSetup = await client.needsSetup()
        if needsSetup { return }
        
        isWorking = true
        errorMessage = ""
        
        do {
            try await client.turnOn()
            sleepState = try await client.status()
        } catch {
            errorMessage = error.localizedDescription
        }
        isWorking = false
    }
    
    func turnOff() async {
        needsSetup = await client.needsSetup()
        if needsSetup { return }
        
        isWorking = true
        errorMessage = ""
        
        do {
            try await client.turnOff()
            sleepState = try await client.status()
        } catch {
            errorMessage = error.localizedDescription
        }
        isWorking = false
    }
}
