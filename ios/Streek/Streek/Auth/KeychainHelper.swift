import Foundation
import Security

/// Lightweight Keychain helper for storing the access token and a small JSON-encoded
/// snapshot of the current user. Uses kSecClassGenericPassword.
enum KeychainHelper {
    private static let service = "com.streek.app"

    enum Key: String {
        case accessToken = "access_token"
        case userJSON = "user"
    }

    // MARK: - String

    @discardableResult
    static func save(_ value: String, for key: Key) -> Bool {
        guard let data = value.data(using: .utf8) else { return false }
        return saveData(data, for: key)
    }

    static func load(_ key: Key) -> String? {
        guard let data = loadData(key) else { return nil }
        return String(data: data, encoding: .utf8)
    }

    // MARK: - Data (JSON snapshots)

    @discardableResult
    static func saveData(_ data: Data, for key: Key) -> Bool {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key.rawValue
        ]
        SecItemDelete(query as CFDictionary)

        var attrs = query
        attrs[kSecValueData as String] = data
        attrs[kSecAttrAccessible as String] = kSecAttrAccessibleAfterFirstUnlock

        let status = SecItemAdd(attrs as CFDictionary, nil)
        return status == errSecSuccess
    }

    static func loadData(_ key: Key) -> Data? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key.rawValue,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]
        var item: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &item)
        guard status == errSecSuccess, let data = item as? Data else { return nil }
        return data
    }

    @discardableResult
    static func remove(_ key: Key) -> Bool {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: key.rawValue
        ]
        let status = SecItemDelete(query as CFDictionary)
        return status == errSecSuccess || status == errSecItemNotFound
    }

    static func wipe() {
        remove(.accessToken)
        remove(.userJSON)
    }
}
