import Foundation

enum HTTPMethod: String {
    case get = "GET"
    case post = "POST"
    case put = "PUT"
    case delete = "DELETE"
}

struct Endpoint {
    let path: String
    let method: HTTPMethod
    let body: Encodable?
    let requiresAuth: Bool

    init(path: String, method: HTTPMethod, body: Encodable? = nil, requiresAuth: Bool = true) {
        self.path = path
        self.method = method
        self.body = body
        self.requiresAuth = requiresAuth
    }
}

// MARK: - Endpoint Definitions

extension Endpoint {
    // Auth
    static func register(email: String, password: String) -> Endpoint {
        struct Body: Encodable { let email: String; let password: String }
        return Endpoint(
            path: "/api/v1/auth/register",
            method: .post,
            body: Body(email: email, password: password),
            requiresAuth: false
        )
    }

    static func login(email: String, password: String) -> Endpoint {
        struct Body: Encodable { let email: String; let password: String }
        return Endpoint(
            path: "/api/v1/auth/login",
            method: .post,
            body: Body(email: email, password: password),
            requiresAuth: false
        )
    }

    // Habits
    static func listHabits() -> Endpoint {
        Endpoint(path: "/api/v1/habits", method: .get)
    }

    static func createHabit(name: String, description: String?, labelColor: String) -> Endpoint {
        struct Body: Encodable { let name: String; let description: String?; let label_color: String }
        return Endpoint(
            path: "/api/v1/habits",
            method: .post,
            body: Body(name: name, description: description, label_color: labelColor)
        )
    }

    static func updateHabit(id: String, name: String, description: String?, labelColor: String) -> Endpoint {
        struct Body: Encodable { let name: String; let description: String?; let label_color: String }
        return Endpoint(
            path: "/api/v1/habits/\(id)",
            method: .put,
            body: Body(name: name, description: description, label_color: labelColor)
        )
    }

    static func deleteHabit(id: String) -> Endpoint {
        Endpoint(path: "/api/v1/habits/\(id)", method: .delete)
    }

    // Stats
    static func getStatsOverview(today: String) -> Endpoint {
        Endpoint(path: "/api/v1/stats/overview?today=\(today)", method: .get)
    }

    // CheckIns
    static func checkIn(habitId: String, checkedDate: String) -> Endpoint {
        struct Body: Encodable { let checked_date: String }
        return Endpoint(
            path: "/api/v1/habits/\(habitId)/check",
            method: .post,
            body: Body(checked_date: checkedDate)
        )
    }

    static func undoCheckIn(habitId: String, checkedDate: String) -> Endpoint {
        struct Body: Encodable { let checked_date: String }
        return Endpoint(
            path: "/api/v1/habits/\(habitId)/check",
            method: .delete,
            body: Body(checked_date: checkedDate)
        )
    }
}
