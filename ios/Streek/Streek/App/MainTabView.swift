import SwiftUI

struct MainTabView: View {
    @State private var selection: Tab = .habits

    enum Tab: Hashable {
        case habits, stats, points, profile
    }

    var body: some View {
        TabView(selection: $selection) {
            HabitListView()
                .tabItem {
                    Label("Habits", systemImage: "flame.fill")
                }
                .tag(Tab.habits)

            StatsView()
                .tabItem {
                    Label("Stats", systemImage: "chart.bar.fill")
                }
                .tag(Tab.stats)

            PointsView()
                .tabItem {
                    Label("Points", systemImage: "sparkles")
                }
                .tag(Tab.points)

            ProfileView()
                .tabItem {
                    Label("Profile", systemImage: "person.fill")
                }
                .tag(Tab.profile)
        }
        .tint(Color.appAccent)
        .onAppear {
            // Style UIKit-backed tab bar to match the dark theme.
            let appearance = UITabBarAppearance()
            appearance.configureWithOpaqueBackground()
            appearance.backgroundColor = UIColor(Color.appBackground)
            appearance.shadowColor = UIColor(Color.appBorder)

            // Unselected items
            let unselectedColor = UIColor(Color.appTextTertiary)
            appearance.stackedLayoutAppearance.normal.iconColor = unselectedColor
            appearance.stackedLayoutAppearance.normal.titleTextAttributes = [
                .foregroundColor: unselectedColor
            ]

            UITabBar.appearance().standardAppearance = appearance
            UITabBar.appearance().scrollEdgeAppearance = appearance
        }
    }
}

#Preview {
    MainTabView()
        .environment(AuthStore())
}
