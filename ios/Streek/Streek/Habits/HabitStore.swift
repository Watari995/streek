import Foundation
import Observation

/// Single source of truth for the user's habit list.
///
/// All Habit-related API calls live here — views only see `habits`, `loadState`,
/// and the methods below. APIClient is never touched directly from a view.
@MainActor
@Observable
final class HabitStore {
    enum LoadState: Equatable {
        case idle
        case loading
        case loaded
        case failed(String)
    }

    private(set) var habits: [Habit] = []
    private(set) var loadState: LoadState = .idle

    init() {}

    // MARK: - API

    /// API: `GET /api/v1/habits`
    ///
    /// Loads the authenticated user's habits and updates `habits` and `loadState`.
    /// Errors are captured into `loadState = .failed(...)` instead of being thrown,
    /// because this is typically called from a `.task` modifier.
    func loadHabits() async {
        loadState = .loading
        do {
            let response: HabitsListResponse = try await APIClient.shared.request(.listHabits())
            self.habits = response.habits
            self.loadState = .loaded
        } catch let APIError.server(_, _, message) {
            self.loadState = .failed(message)
        } catch {
            self.loadState = .failed(error.localizedDescription)
        }
    }

    /// API: `POST /api/v1/habits`
    ///
    /// Creates a habit and appends it to the local list.
    /// Errors propagate to the caller so the form can show inline feedback.
    @discardableResult
    func createHabit(name: String, description: String?, labelColor: String) async throws -> Habit {
        let created: Habit = try await APIClient.shared.request(
            .createHabit(name: name, description: description, labelColor: labelColor)
        )
        habits.append(created)
        return created
    }

    /// API: `PUT /api/v1/habits/:id`
    ///
    /// Updates a habit and replaces the corresponding entry in the local list.
    @discardableResult
    func updateHabit(id: String, name: String, description: String?, labelColor: String) async throws -> Habit {
        let updated: Habit = try await APIClient.shared.request(
            .updateHabit(id: id, name: name, description: description, labelColor: labelColor)
        )
        if let idx = habits.firstIndex(where: { $0.id == updated.id }) {
            habits[idx] = updated
        }
        return updated
    }

    /// API: `DELETE /api/v1/habits/:id`
    ///
    /// Deletes a habit on the server and removes it from the local list.
    func deleteHabit(id: String) async throws {
        try await APIClient.shared.requestVoid(.deleteHabit(id: id))
        habits.removeAll { $0.id == id }
    }

    // MARK: - Local state

    /// Clears local state. Call this on logout so the next user doesn't see
    /// the previous user's habits flicker through.
    func reset() {
        habits = []
        loadState = .idle
    }
}
