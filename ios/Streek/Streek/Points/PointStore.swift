import Foundation
import Observation

/// Single source of truth for the user's point balance and ledger history.
///
/// Read-only from a domain perspective: points are awarded server-side as a
/// side effect of check-ins (handled by the `EarnPointsOnCheckIn` event handler
/// in the backend). After a successful check-in, callers should invoke
/// `loadBalance()` to re-sync.
///
/// All API calls live here — views call `loadBalance()`, `loadHistory()`, or
/// inspect the published properties. APIClient is never touched directly from
/// a view.
@MainActor
@Observable
final class PointStore {
    enum LoadState: Equatable {
        case idle
        case loading
        case loaded
        case failed(String)
    }

    private(set) var balance: Int = 0
    private(set) var history: [PointHistoryEntry] = []
    private(set) var balanceState: LoadState = .idle
    private(set) var historyState: LoadState = .idle

    init() {}

    // MARK: - API

    /// API: `GET /api/v1/points/balance`
    func loadBalance() async {
        balanceState = .loading
        do {
            let response: PointBalanceResponse = try await APIClient.shared.request(
                .getPointBalance()
            )
            self.balance = response.balance
            self.balanceState = .loaded
        } catch let APIError.server(_, _, message) {
            self.balanceState = .failed(message)
        } catch {
            self.balanceState = .failed(error.localizedDescription)
        }
    }

    /// API: `GET /api/v1/points/history`
    func loadHistory() async {
        historyState = .loading
        do {
            let response: PointHistoryResponse = try await APIClient.shared.request(
                .getPointHistory()
            )
            self.history = response.entries
            self.historyState = .loaded
        } catch let APIError.server(_, _, message) {
            self.historyState = .failed(message)
        } catch {
            self.historyState = .failed(error.localizedDescription)
        }
    }

    // MARK: - Local state

    /// Clears local state (call on logout).
    func reset() {
        balance = 0
        history = []
        balanceState = .idle
        historyState = .idle
    }
}
