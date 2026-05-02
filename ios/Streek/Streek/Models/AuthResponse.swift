import Foundation

struct AuthResponse: Codable, Sendable {
    let accessToken: String
    let user: User
}
