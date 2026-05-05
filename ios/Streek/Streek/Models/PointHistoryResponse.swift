import Foundation

/// Single ledger entry from `GET /api/v1/points/history`.
struct PointHistoryEntry: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let habitId: String?
    let type: String   // "EARN" or "SPEND"
    let amount: Int    // always positive; sign derived from `type`
    let reason: String
    let idempotencyKey: String
    let createdAt: Date

    /// Signed amount for display: EARN is +N, SPEND is -N.
    var signedAmount: Int { type == "SPEND" ? -amount : amount }
}

/// Server envelope for `GET /api/v1/points/history`.
struct PointHistoryResponse: Codable, Sendable {
    let entries: [PointHistoryEntry]
}
