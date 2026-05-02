import Foundation

struct User: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let email: String
    let createdAt: Date
}
