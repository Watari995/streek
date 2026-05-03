import Foundation

/// Single source for "today's date" in the format the backend expects (YYYY-MM-DD).
///
/// All check-in / stats requests need today in the *device's* calendar so the
/// user's local notion of "today" matches what's stored server-side. Using a
/// shared helper keeps every caller consistent (same time zone, same locale,
/// same format).
enum Today {
    static func string() -> String {
        let f = DateFormatter()
        f.dateFormat = "yyyy-MM-dd"
        f.calendar = Calendar(identifier: .gregorian)
        f.timeZone = .current
        f.locale = Locale(identifier: "en_US_POSIX")
        return f.string(from: Date())
    }
}
