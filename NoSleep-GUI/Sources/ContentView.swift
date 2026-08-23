import SwiftUI

struct ContentView: View {
    @State private var model = NoSleepModel()
    
    var body: some View {
        VStack(spacing: 16) {
            Text("NoSleep · macOS")
                .font(.caption)
                .foregroundStyle(.secondary)
            
            if model.isWorking {
                ProgressView()
                    .progressViewStyle(CircularProgressViewStyle())
                    .padding(20)
                    .frame(height: 80)
            } else {
                StatusCard(state: model.sleepState)
            }
            
            if model.needsSetup {
                VStack(spacing: 8) {
                    Text("Setup Required")
                        .font(.headline)
                        .foregroundStyle(.red)
                    Text("The app needs permission to keep your Mac awake.")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                    
                    Button("Set Up Now") {
                        Task { await model.runSetup() }
                    }
                    .buttonStyle(.borderedProminent)
                    .tint(.red)
                    .disabled(model.isBusy)
                    .padding(.top, 4)
                }
                .padding()
                .background(Color.red.opacity(0.1))
                .cornerRadius(12)
            } else {
                Toggle(isOn: Binding(
                    get: { model.sleepState == .awake },
                    set: { newValue in
                        Task {
                            if newValue { await model.turnOn() } else { await model.turnOff() }
                        }
                    }
                )) {
                    Label("Keep awake", systemImage: "power")
                }
                .toggleStyle(.switch)
                .disabled(model.isBusy || model.sleepState == .unknown)
            }
            
            if model.sleepState == .awake {
                Text("⚠ Battery drain risk while disabled")
                    .foregroundStyle(.orange)
                    .font(.caption)
            } else {
                Text(" ")
                    .font(.caption)
            }
            
            if !model.errorMessage.isEmpty {
                Text(model.errorMessage)
                    .foregroundStyle(.red)
                    .font(.caption)
                    .multilineTextAlignment(.center)
            }
            
            Button("Refresh") {
                Task { await model.refreshStatus() }
            }
            .buttonStyle(.link)
            .font(.caption)
            .disabled(model.isBusy)
        }
        .padding(24)
        .frame(width: 320)
        .task {
            await model.refreshStatus()
            while !Task.isCancelled {
                try? await Task.sleep(nanoseconds: 5_000_000_000)
                await model.refreshStatus(showProgress: false)
            }
        }
    }
}

struct StatusCard: View {
    let state: SleepState
    
    var body: some View {
        RoundedRectangle(cornerRadius: 12)
            .fill(state == .awake ? Color.orange.opacity(0.15) : Color.gray.opacity(0.15))
            .frame(height: 80)
            .overlay(alignment: .center, content: {
                VStack(spacing: 6) {
                    Text("\(state.iconName)  \(state.displayName)")
                        .font(.title2.bold())
                    Text(state.descriptionText)
                        .foregroundStyle(.secondary)
                        .font(.callout)
                }
            })
    }
}
