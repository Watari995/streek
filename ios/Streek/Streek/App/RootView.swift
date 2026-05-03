import SwiftUI

struct RootView: View {
    @Environment(AuthStore.self) private var auth
    @Environment(CheckInStore.self) private var checkInStore

    var body: some View {
        Group {
            switch auth.state {
            case .loading:
                LoadingScreen()
            case .unauthenticated:
                LoginView()
                    .transition(.opacity)
            case .authenticated:
                MainTabView()
                    .transition(.opacity)
            }
        }
        .animation(.easeInOut(duration: 0.25), value: auth.state)
        .task {
            // Load any persisted session and today's check-in snapshot.
            await auth.bootstrap()
            checkInStore.bootstrap()
        }
        .preferredColorScheme(.dark)
    }
}

private struct LoadingScreen: View {
    var body: some View {
        ZStack {
            Color.appBackground.ignoresSafeArea()
            VStack(spacing: AppSpacing.lg) {
                Image(systemName: "flame.fill")
                    .font(.system(size: 56, weight: .bold))
                    .foregroundStyle(Color.appAccent)
                ProgressView()
                    .progressViewStyle(.circular)
                    .tint(Color.appTextSecondary)
            }
        }
    }
}

