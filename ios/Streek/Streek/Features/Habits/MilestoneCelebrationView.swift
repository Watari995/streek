import SwiftUI

/// Streak milestone parity with the backend.
///
/// The server fires an email when a habit's current streak hits one of these
/// values (see `backend/internal/domain/service/streak_service.go`
/// `streakMilestones`). The app can't receive that email, so we mirror the
/// same thresholds locally and celebrate in-app instead — no extra API needed,
/// the value comes from the existing stats overview `current_streak`.
enum StreakMilestone {
    /// Must stay in sync with the backend `streakMilestones` slice.
    static let thresholds: [Int] = [10, 20, 30, 40, 50, 60, 70, 80, 90]

    /// Returns the milestone the streak just reached, if a check-in moved the
    /// streak from `before` up to `after` and `after` lands exactly on a
    /// threshold. Returns `nil` for undo (`after <= before`) or non-milestone
    /// values, so relaunches and toggles don't re-trigger the celebration.
    static func reached(before: Int, after: Int) -> Int? {
        guard after > before else { return nil }
        return thresholds.contains(after) ? after : nil
    }
}

/// Full-screen, auto-dismissing celebration shown when a habit crosses a
/// streak milestone. Tap anywhere (or wait) to dismiss.
struct MilestoneCelebrationView: View {
    let habitName: String
    let streak: Int
    let onDismiss: () -> Void

    @State private var appeared = false

    var body: some View {
        ZStack {
            Color.black.opacity(0.6)
                .ignoresSafeArea()
                .onTapGesture { onDismiss() }

            VStack(spacing: AppSpacing.lg) {
                Image(systemName: "flame.fill")
                    .font(.system(size: 64, weight: .bold))
                    .foregroundStyle(Color.appAccent)
                    .scaleEffect(appeared ? 1 : 0.4)
                    .rotationEffect(.degrees(appeared ? 0 : -20))

                Text("\(streak)-day streak!")
                    .font(AppFont.title(size: 32))
                    .foregroundStyle(Color.appTextPrimary)

                Text(habitName)
                    .font(AppFont.headline(size: 16))
                    .foregroundStyle(Color.appTextSecondary)
                    .multilineTextAlignment(.center)

                Text("Keep it up — you're on fire.")
                    .font(AppFont.body(size: 14))
                    .foregroundStyle(Color.appTextSecondary)
                    .padding(.top, AppSpacing.xs)
            }
            .padding(AppSpacing.xxl)
            .frame(maxWidth: 320)
            .background(
                RoundedRectangle(cornerRadius: AppRadius.lg, style: .continuous)
                    .fill(Color.appSurface)
            )
            .overlay(
                RoundedRectangle(cornerRadius: AppRadius.lg, style: .continuous)
                    .stroke(Color.appAccent.opacity(0.4), lineWidth: 1)
            )
            .scaleEffect(appeared ? 1 : 0.7)
            .opacity(appeared ? 1 : 0)
        }
        .onAppear {
            withAnimation(.spring(response: 0.45, dampingFraction: 0.6)) {
                appeared = true
            }
            let generator = UINotificationFeedbackGenerator()
            generator.notificationOccurred(.success)
            // Auto-dismiss so it never blocks the user.
            Task {
                try? await Task.sleep(for: .seconds(2.8))
                onDismiss()
            }
        }
    }
}

#Preview {
    MilestoneCelebrationView(habitName: "Morning run", streak: 30, onDismiss: {})
}
