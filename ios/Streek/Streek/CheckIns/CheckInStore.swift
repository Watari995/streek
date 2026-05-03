import Foundation
import Observation

/// Single source of truth for "which habits are checked in today" in the UI.
///
/// All check-in API calls live here — views only call `toggle(habitId:)` /
/// `isChecked(habitId:)`. APIClient is never touched directly from a view.
///
/// **State model**
/// The backend currently exposes only `POST` and `DELETE` for `/check`, so
/// there's no way to fetch "what habits are already checked today" from the
/// server. Until a `GET /api/v1/habits/:id/check-ins` endpoint exists, we
/// keep a local cache that:
///   - is keyed by habit id and stores today's "checked" flag
///   - is persisted to `UserDefaults` so it survives app relaunch
///   - resets automatically when the calendar day rolls over (the next
///     load notices the stored date is no longer today and starts fresh)
@MainActor
@Observable
final class CheckInStore {

    // MARK: - Public state

    /// The set of habit IDs that are checked in for today.
    private(set) var checkedToday: Set<String> = []

    /// `true` while a mutation is in flight for this habit. Drives the spinner
    /// inside the row's check button to prevent rapid double-tap.
    private(set) var inFlight: Set<String> = []

    // MARK: - Persistence

    private static let storageKey = "com.streek.checkInStore.v1"

    private struct Snapshot: Codable {
        let date: String          // YYYY-MM-DD that the snapshot was taken on
        let habitIDs: [String]
    }

    init() {}

    /// Loads any persisted snapshot from `UserDefaults`. Discards it if the
    /// snapshot is from a previous calendar day. Call on app launch.
    func bootstrap() {
        guard
            let data = UserDefaults.standard.data(forKey: Self.storageKey),
            let snapshot = try? JSONDecoder().decode(Snapshot.self, from: data)
        else {
            return
        }
        if snapshot.date == Self.todayString() {
            self.checkedToday = Set(snapshot.habitIDs)
        } else {
            // Day rolled over — drop the stale snapshot.
            UserDefaults.standard.removeObject(forKey: Self.storageKey)
        }
    }

    private func persist() {
        let snapshot = Snapshot(date: Self.todayString(), habitIDs: Array(checkedToday))
        if let data = try? JSONEncoder().encode(snapshot) {
            UserDefaults.standard.set(data, forKey: Self.storageKey)
        }
    }

    // MARK: - Queries

    func isChecked(habitId: String) -> Bool {
        checkedToday.contains(habitId)
    }

    func isInFlight(habitId: String) -> Bool {
        inFlight.contains(habitId)
    }

    // MARK: - API

    /// Toggles "checked today" for a habit. Optimistically updates the UI,
    /// then dispatches the matching API call. If the call fails, the optimistic
    /// state is reverted and the error is rethrown to the caller.
    func toggle(habitId: String) async throws {
        if isChecked(habitId: habitId) {
            try await undo(habitId: habitId)
        } else {
            try await check(habitId: habitId)
        }
    }

    /// API: `POST /api/v1/habits/:id/check`
    func check(habitId: String) async throws {
        guard !inFlight.contains(habitId) else { return }
        let dateStr = Self.todayString()

        // Optimistic update
        checkedToday.insert(habitId)
        persist()
        inFlight.insert(habitId)

        do {
            try await APIClient.shared.requestVoid(
                .checkIn(habitId: habitId, checkedDate: dateStr)
            )
            inFlight.remove(habitId)
        } catch {
            // Roll back
            checkedToday.remove(habitId)
            persist()
            inFlight.remove(habitId)
            throw error
        }
    }

    /// API: `DELETE /api/v1/habits/:id/check`
    func undo(habitId: String) async throws {
        guard !inFlight.contains(habitId) else { return }
        let dateStr = Self.todayString()

        // Optimistic update
        checkedToday.remove(habitId)
        persist()
        inFlight.insert(habitId)

        do {
            try await APIClient.shared.requestVoid(
                .undoCheckIn(habitId: habitId, checkedDate: dateStr)
            )
            inFlight.remove(habitId)
        } catch {
            // Roll back
            checkedToday.insert(habitId)
            persist()
            inFlight.remove(habitId)
            throw error
        }
    }

    // MARK: - Local state

    /// Clears local state. Call this on logout so the next user doesn't see
    /// the previous user's check-ins flicker through.
    func reset() {
        checkedToday = []
        inFlight = []
        UserDefaults.standard.removeObject(forKey: Self.storageKey)
    }

    // MARK: - Helpers

    /// Today's date in `YYYY-MM-DD` form, in the device's current calendar.
    /// Matches the backend's `DateString` value-object format.
    static func todayString() -> String {
        let f = DateFormatter()
        f.dateFormat = "yyyy-MM-dd"
        f.calendar = Calendar(identifier: .gregorian)
        f.timeZone = .current
        f.locale = Locale(identifier: "en_US_POSIX")
        return f.string(from: Date())
    }
}
