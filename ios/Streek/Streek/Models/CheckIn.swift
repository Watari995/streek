import Foundation

struct CheckIn: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let habitId: String
    let checkedDate: String  // YYYY-MM-DD
    let createdAt: Date
}
