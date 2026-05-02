import SwiftUI

struct HabitRowView: View {
    let habit: Habit
    let isCheckedToday: Bool
    let streak: Int
    let onToggle: () -> Void

    var body: some View {
        HStack(spacing: AppSpacing.lg) {
            // Color tag
            RoundedRectangle(cornerRadius: 3, style: .continuous)
                .fill(habit.color)
                .frame(width: 4, height: 56)

            VStack(alignment: .leading, spacing: AppSpacing.xs) {
                Text(habit.name)
                    .font(AppFont.headline(size: 16))
                    .foregroundStyle(Color.appTextPrimary)
                    .lineLimit(1)

                if let description = habit.description, !description.isEmpty {
                    Text(description)
                        .font(AppFont.caption(size: 13))
                        .foregroundStyle(Color.appTextSecondary)
                        .lineLimit(1)
                }

                HStack(spacing: AppSpacing.xs) {
                    Image(systemName: "flame.fill")
                        .font(.system(size: 11, weight: .semibold))
                        .foregroundStyle(streak > 0 ? Color.appAccent : Color.appTextTertiary)
                    Text(streak == 0 ? "Start a streak" : "\(streak)-day streak")
                        .font(AppFont.caption(size: 12))
                        .foregroundStyle(streak > 0 ? Color.appTextPrimary : Color.appTextTertiary)
                }
                .padding(.top, 2)
            }

            Spacer(minLength: 0)

            CheckButton(isChecked: isCheckedToday, action: onToggle)
        }
        .padding(.vertical, AppSpacing.md)
        .padding(.horizontal, AppSpacing.lg)
        .background(
            RoundedRectangle(cornerRadius: AppRadius.lg, style: .continuous)
                .fill(Color.appSurface)
        )
        .overlay(
            RoundedRectangle(cornerRadius: AppRadius.lg, style: .continuous)
                .stroke(Color.appBorder, lineWidth: 1)
        )
    }
}

private struct CheckButton: View {
    let isChecked: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            ZStack {
                Circle()
                    .fill(isChecked ? Color.appAccent : Color.clear)
                    .frame(width: 36, height: 36)
                Circle()
                    .stroke(isChecked ? Color.appAccent : Color.appBorder, lineWidth: 1.5)
                    .frame(width: 36, height: 36)
                if isChecked {
                    Image(systemName: "checkmark")
                        .font(.system(size: 14, weight: .bold))
                        .foregroundStyle(.white)
                }
            }
        }
        .buttonStyle(PressableScaleStyle())
        .animation(.spring(response: 0.25, dampingFraction: 0.7), value: isChecked)
    }
}
