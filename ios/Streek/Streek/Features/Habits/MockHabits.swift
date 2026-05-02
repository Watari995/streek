import Foundation

/// Mock data for the habit list while the backend handlers are not yet implemented.
/// Replace with real API calls once the backend is ready.
enum MockHabits {
    static let now = Date()

    static let all: [Habit] = [
        Habit(
            id: "01934567-89ab-cdef-0123-456789abcdef",
            userId: "user-1",
            name: "Morning Run",
            description: "5km outdoor run before breakfast",
            labelColor: "#FF0211",
            createdAt: now.addingTimeInterval(-86400 * 30),
            updatedAt: now
        ),
        Habit(
            id: "01934567-89ab-cdef-0123-456789abcde0",
            userId: "user-1",
            name: "Read 30 minutes",
            description: "Any book counts",
            labelColor: "#00C853",
            createdAt: now.addingTimeInterval(-86400 * 21),
            updatedAt: now
        ),
        Habit(
            id: "01934567-89ab-cdef-0123-456789abcde1",
            userId: "user-1",
            name: "Meditate",
            description: nil,
            labelColor: "#FF9500",
            createdAt: now.addingTimeInterval(-86400 * 14),
            updatedAt: now
        ),
        Habit(
            id: "01934567-89ab-cdef-0123-456789abcde2",
            userId: "user-1",
            name: "Stretch & mobility",
            description: "10 minutes of stretching",
            labelColor: "#5E5CE6",
            createdAt: now.addingTimeInterval(-86400 * 7),
            updatedAt: now
        ),
        Habit(
            id: "01934567-89ab-cdef-0123-456789abcde3",
            userId: "user-1",
            name: "No social media before noon",
            description: "Stay focused in the morning",
            labelColor: "#FF2D55",
            createdAt: now.addingTimeInterval(-86400 * 3),
            updatedAt: now
        ),
    ]

    /// Fake "checked today" map keyed by habit id.
    static var initialCheckedState: [String: Bool] {
        [
            all[0].id: true,
            all[1].id: true,
            all[2].id: false,
            all[3].id: false,
            all[4].id: true,
        ]
    }

    /// Fake streak counts keyed by habit id.
    static var streaks: [String: Int] {
        [
            all[0].id: 12,
            all[1].id: 5,
            all[2].id: 0,
            all[3].id: 2,
            all[4].id: 8,
        ]
    }
}
