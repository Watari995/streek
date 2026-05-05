import Foundation

/// Server envelope for `GET /api/v1/points/balance`.
struct PointBalanceResponse: Codable, Sendable {
    let balance: Int
}
