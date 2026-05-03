import SwiftUI

@main
struct StreekApp: App {
    @State private var auth = AuthStore()
    @State private var habitStore = HabitStore()
    @State private var checkInStore = CheckInStore()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environment(auth)
                .environment(habitStore)
                .environment(checkInStore)
        }
    }
}
