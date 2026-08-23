import Foundation

enum SleepState: String, CaseIterable, Codable {
    case normal = "normal"
    case awake  = "awake"
    case unknown = "unknown"

    var displayName: String {
        switch self {
        case .awake:   return "AWAKE"
        case .normal:  return "SLEEPING"
        case .unknown: return "UNKNOWN"
        }
    }

    var iconName: String {
        switch self {
        case .awake:   return "☕"
        case .normal:  return "😴"
        case .unknown: return "❓"
        }
    }

    var descriptionText: String {
        switch self {
        case .awake:   return "Your Mac will not sleep"
        case .normal:  return "Your Mac can sleep normally"
        case .unknown: return "Checking status..."
        }
    }
}

struct StatusResponse: Codable {
    let state: String
    let disablesleep: Int

    var parsedState: SleepState {
        return SleepState(rawValue: state) ?? .unknown
    }
}

struct ActionResponse: Codable {
    let ok: Bool
    let action: String
    let user: String?
}
