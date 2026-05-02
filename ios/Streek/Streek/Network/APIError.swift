import Foundation

/// Server-shaped error response: {"error": {"code": "...", "message": "..."}}
struct APIErrorBody: Codable, Sendable {
    let code: String
    let message: String
}

struct APIErrorResponse: Codable, Sendable {
    let error: APIErrorBody
}

enum APIError: Error, LocalizedError {
    case invalidURL
    case requestFailed(Error)
    case invalidResponse
    case decodingFailed(Error)
    case server(status: Int, code: String, message: String)
    case unauthorized

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Invalid URL"
        case .requestFailed(let err):
            return "Network error: \(err.localizedDescription)"
        case .invalidResponse:
            return "Invalid server response"
        case .decodingFailed:
            return "Failed to decode response"
        case .server(_, _, let message):
            return message
        case .unauthorized:
            return "Session expired. Please log in again."
        }
    }
}
