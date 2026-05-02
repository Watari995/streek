import SwiftUI

// MARK: - Color Palette (Mercari-inspired dark)

extension Color {
    // Backgrounds
    static let appBackground = Color(hex: 0x0F0F0F)
    static let appSurface = Color(hex: 0x1A1A1A)
    static let appSurfaceElevated = Color(hex: 0x262626)
    static let appSurfacePressed = Color(hex: 0x2E2E2E)

    // Text
    static let appTextPrimary = Color(hex: 0xFFFFFF)
    static let appTextSecondary = Color(hex: 0x9A9A9A)
    static let appTextTertiary = Color(hex: 0x666666)

    // Accents
    static let appAccent = Color(hex: 0xFF0211)
    static let appAccentPressed = Color(hex: 0xCC010E)

    // Semantic
    static let appSuccess = Color(hex: 0x00C853)
    static let appWarning = Color(hex: 0xFF9500)
    static let appDanger = Color(hex: 0xFF3B30)

    // Borders / dividers
    static let appBorder = Color(hex: 0x2A2A2A)
    static let appDivider = Color(hex: 0x1F1F1F)

    init(hex: UInt32, opacity: Double = 1.0) {
        let r = Double((hex >> 16) & 0xFF) / 255.0
        let g = Double((hex >> 8) & 0xFF) / 255.0
        let b = Double(hex & 0xFF) / 255.0
        self.init(.sRGB, red: r, green: g, blue: b, opacity: opacity)
    }
}

// MARK: - Typography

enum AppFont {
    static func title(size: CGFloat = 28) -> Font { .system(size: size, weight: .bold, design: .default) }
    static func headline(size: CGFloat = 18) -> Font { .system(size: size, weight: .semibold, design: .default) }
    static func body(size: CGFloat = 15) -> Font { .system(size: size, weight: .regular, design: .default) }
    static func label(size: CGFloat = 13) -> Font { .system(size: size, weight: .medium, design: .default) }
    static func caption(size: CGFloat = 12) -> Font { .system(size: size, weight: .regular, design: .default) }
}

// MARK: - Spacing

enum AppSpacing {
    static let xs: CGFloat = 4
    static let sm: CGFloat = 8
    static let md: CGFloat = 12
    static let lg: CGFloat = 16
    static let xl: CGFloat = 20
    static let xxl: CGFloat = 28
    static let xxxl: CGFloat = 40
}

// MARK: - Corner Radius

enum AppRadius {
    static let sm: CGFloat = 8
    static let md: CGFloat = 12
    static let lg: CGFloat = 16
    static let pill: CGFloat = 999
}
