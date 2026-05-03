import Foundation
import Observation

/// Single source of truth for the user's stats overview, fetched from the
/// server. Hosts streak counts and the server-side "done today" count.
///
/// This store is read-only from a domain perspective — mutations to check-in
/// state happen via `CheckInStore`. After a successful toggle, callers should
/// invoke `loadOverview()` here to re-sync streaks.
///
/// All API calls live here — views call `streak(for:)`, `isCheckedToday(_:)`,
/// `loadOverview()`. APIClient is never touched directly from a view.
@MainActor
@Observable
final class StatsStore {
    enum LoadState: Equatable {
        case idle
        case loading
        case loaded
        case failed(String)
    }

    private(set) var overview: StatsOverviewResponse?
    private(set) var loadState: LoadState = .idle

    init() {}

    // MARK: - API

    /// API: `GET /api/v1/stats/overview?today=YYYY-MM-DD`
    ///
    /// Loads the authenticated user's overview using the device's local "today"
    /// for streak grace-period semantics. Errors are surfaced via `loadState`.
    func loadOverview() async {
        loadState = .loading
        do {
            let response: StatsOverviewResponse = try await APIClient.shared.request(
                .getStatsOverview(today: Today.string())
            )
            self.overview = response
            self.loadState = .loaded
        } catch let APIError.server(_, _, message) {
            self.loadState = .failed(message)
        } catch {
            self.loadState = .failed(error.localizedDescription)
        }
    }

    // MARK: - Queries

    /// Returns (current, longest) streak for a habit. (0, 0) if not loaded.
    func streak(for habitId: String) -> (current: Int, longest: Int) {
        guard let entry = overview?.habits.first(where: { $0.habitId == habitId }) else {
            return (0, 0)
        }
        return (entry.currentStreak, entry.longestStreak)
    }

    /// Returns whether the server reports this habit checked today. May lag
    /// a CheckInStore optimistic update by one network round-trip.
    func isCheckedToday(_ habitId: String) -> Bool {
        overview?.habits.first(where: { $0.habitId == habitId })?.checkedToday ?? false
    }

    /// Aggregated metrics. Default to zero when overview not yet loaded.
    var longestStreak: Int { overview?.longestStreak ?? 0 }
    var doneToday: Int { overview?.doneToday ?? 0 }
    var habitOverviews: [HabitOverview] { overview?.habits ?? [] }

    // MARK: - Local state

    /// Clears local state (call on logout).
    func reset() {
        overview = nil
        loadState = .idle
    }
}
