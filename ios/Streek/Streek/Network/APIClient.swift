import Foundation

extension Notification.Name {
    /// Posted when the API returns 401, so the auth layer can clear local state.
    static let streekUnauthorized = Notification.Name("com.streek.unauthorized")
}

actor APIClient {
    static let shared = APIClient()

    private let baseURL: URL
    private let session: URLSession

    private let decoder: JSONDecoder = {
        let d = JSONDecoder()
        d.keyDecodingStrategy = .convertFromSnakeCase
        d.dateDecodingStrategy = .iso8601
        return d
    }()

    private let encoder: JSONEncoder = {
        let e = JSONEncoder()
        e.dateEncodingStrategy = .iso8601
        return e
    }()

    // Simulator は localhost、実機は Mac の LAN IP を使う。
    // 環境変数 STREEK_API_BASE_URL で上書き可能。
    private static let defaultBaseURL: URL = {
        if let override = ProcessInfo.processInfo.environment["STREEK_API_BASE_URL"],
           let url = URL(string: override) {
            return url
        }
        #if targetEnvironment(simulator)
        return URL(string: "http://localhost:8080")!
        #else
        return URL(string: "http://192.168.0.192:8080")!
        #endif
    }()

    init(baseURL: URL = APIClient.defaultBaseURL) {
        self.baseURL = baseURL
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 10
        config.timeoutIntervalForResource = 30
        self.session = URLSession(configuration: config)
    }

    // MARK: - Generic request

    func request<T: Decodable>(_ endpoint: Endpoint) async throws -> T {
        let data = try await performRequest(endpoint)
        if data.isEmpty, let empty = EmptyResponse() as? T {
            return empty
        }
        do {
            return try decoder.decode(T.self, from: data)
        } catch {
            throw APIError.decodingFailed(error)
        }
    }

    func requestVoid(_ endpoint: Endpoint) async throws {
        _ = try await performRequest(endpoint)
    }

    // MARK: - Internal

    private func performRequest(_ endpoint: Endpoint) async throws -> Data {
        guard let url = URL(string: endpoint.path, relativeTo: baseURL) else {
            throw APIError.invalidURL
        }

        var req = URLRequest(url: url)
        req.httpMethod = endpoint.method.rawValue
        req.setValue("application/json", forHTTPHeaderField: "Content-Type")
        req.setValue("application/json", forHTTPHeaderField: "Accept")

        if endpoint.requiresAuth, let token = KeychainHelper.load(.accessToken) {
            req.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        if let body = endpoint.body {
            req.httpBody = try encoder.encode(AnyEncodable(body))
        }

        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await session.data(for: req)
        } catch {
            throw APIError.requestFailed(error)
        }

        guard let http = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        // 2xx success
        if (200..<300).contains(http.statusCode) {
            return data
        }

        // 401: notify so the auth layer can clear local state
        if http.statusCode == 401 {
            await MainActor.run {
                NotificationCenter.default.post(name: .streekUnauthorized, object: nil)
            }
        }

        if let envelope = try? decoder.decode(APIErrorResponse.self, from: data) {
            throw APIError.server(
                status: http.statusCode,
                code: envelope.error.code,
                message: envelope.error.message
            )
        }
        throw APIError.server(
            status: http.statusCode,
            code: "UNKNOWN",
            message: "Request failed (\(http.statusCode))"
        )
    }
}

struct EmptyResponse: Codable, Sendable {}

private struct AnyEncodable: Encodable {
    private let _encode: (Encoder) throws -> Void
    init(_ wrapped: Encodable) {
        self._encode = wrapped.encode
    }
    func encode(to encoder: Encoder) throws {
        try _encode(encoder)
    }
}
