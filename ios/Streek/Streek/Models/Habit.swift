import Foundation
import SwiftUI

struct Habit: Codable, Identifiable, Hashable, Sendable {
    let id: String
    let userId: String
    let name: String
    let description: String?
    let labelColor: String  // hex like "#FF0211"
    let createdAt: Date
    let updatedAt: Date

    // Convenience: SwiftUI Color from hex string
    var color: Color {
        Color(hexString: labelColor) ?? .appAccent
    }
}

extension Color {
    init?(hexString: String) {
        var s = hexString
        if s.hasPrefix("#") { s.removeFirst() }
        guard s.count == 6, let v = UInt32(s, radix: 16) else { return nil }
        self.init(hex: v)
    }
}
