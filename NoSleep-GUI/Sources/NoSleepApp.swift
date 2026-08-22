import SwiftUI

@main
struct NoSleepApp: App {
    var body: some Scene {
        WindowGroup("NoSleep") {
            ContentView()
        }
        .windowResizability(.contentSize)
    }
}
