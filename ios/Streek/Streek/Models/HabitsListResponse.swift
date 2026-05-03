import Foundation

/// Server envelope for `GET /api/v1/habits`.
struct HabitsListResponse: Codable, Sendable {
    let habits: [Habit]
}
