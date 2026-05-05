import SwiftUI

@main
struct StreekApp: App {
    @State private var auth = AuthStore()
    @State private var habitStore = HabitStore()
    @State private var checkInStore = CheckInStore()
    @State private var statsStore = StatsStore()
    @State private var pointStore = PointStore()

    var body: some Scene {
        WindowGroup {
            RootView()
                .environment(auth)
                .environment(habitStore)
                .environment(checkInStore)
                .environment(statsStore)
                .environment(pointStore)
        }
    }
}
