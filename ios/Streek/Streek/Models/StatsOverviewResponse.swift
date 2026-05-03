import Foundation

/// Per-habit stats from `GET /api/v1/stats/overview`. JSON snake_case keys are
/// converted to camelCase by the APIClient's decoder.
struct HabitOverview: Codable, Identifiable, Hashable, Sendable {
    let habitId: String
    let habitName: String
    let labelColor: String
    let currentStreak: Int
    let longestStreak: Int
    let checkedToday: Bool

    var id: String { habitId }
}

/// Server envelope for `GET /api/v1/stats/overview`.
struct StatsOverviewResponse: Codable, Sendable {
    let habits: [HabitOverview]
    let longestStreak: Int
    let doneToday: Int
}
