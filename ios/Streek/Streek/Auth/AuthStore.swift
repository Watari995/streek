import Foundation
import Observation

@MainActor
@Observable
final class AuthStore {
    enum State: Equatable {
        case loading           // Reading Keychain on launch
        case unauthenticated
        case authenticated(User)
    }

    private(set) var state: State = .loading

    @ObservationIgnored
    private var unauthorizedObserver: NSObjectProtocol?

    @ObservationIgnored
    private let decoder: JSONDecoder = {
        let d = JSONDecoder()
        d.keyDecodingStrategy = .convertFromSnakeCase
        d.dateDecodingStrategy = .iso8601
        return d
    }()

    @ObservationIgnored
    private let encoder: JSONEncoder = {
        let e = JSONEncoder()
        e.keyEncodingStrategy = .convertToSnakeCase
        e.dateEncodingStrategy = .iso8601
        return e
    }()

    init() {
        // Observe 401s posted by APIClient and clear local state.
        unauthorizedObserver = NotificationCenter.default.addObserver(
            forName: .streekUnauthorized,
            object: nil,
            queue: .main
        ) { [weak self] _ in
            guard let self else { return }
            // Already on the main queue thanks to `queue: .main`; this hop just
            // re-enters the MainActor-isolated context for the compiler.
            MainActor.assumeIsolated {
                self.logout()
            }
        }
    }

    deinit {
        if let observer = unauthorizedObserver {
            NotificationCenter.default.removeObserver(observer)
        }
    }

    // MARK: - Lifecycle

    /// Loads any persisted session from Keychain. Call once at app launch.
    func bootstrap() async {
        if let userData = KeychainHelper.loadData(.userJSON),
           KeychainHelper.load(.accessToken) != nil,
           let user = try? decoder.decode(User.self, from: userData) {
            self.state = .authenticated(user)
        } else {
            self.state = .unauthenticated
        }
    }

    // MARK: - Auth flows

    func login(email: String, password: String) async throws {
        let response: AuthResponse = try await APIClient.shared.request(
            .login(email: email, password: password)
        )
        persist(response)
    }

    func register(email: String, password: String) async throws {
        let response: AuthResponse = try await APIClient.shared.request(
            .register(email: email, password: password)
        )
        persist(response)
    }

    func logout() {
        KeychainHelper.wipe()
        self.state = .unauthenticated
    }

    // MARK: - Internal

    private func persist(_ response: AuthResponse) {
        KeychainHelper.save(response.accessToken, for: .accessToken)
        if let data = try? encoder.encode(response.user) {
            KeychainHelper.saveData(data, for: .userJSON)
        }
        self.state = .authenticated(response.user)
    }
}
