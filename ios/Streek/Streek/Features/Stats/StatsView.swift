import SwiftUI

struct StatsView: View {
    @Environment(HabitStore.self) private var habitStore

    // TODO: replace with `/api/v1/stats` endpoint once it exists.
    // Streaks are currently always zero because CheckIn handler is not wired up.
    private let streaks: [String: Int] = [:]

    var body: some View {
        NavigationStack {
            ZStack {
                Color.appBackground.ignoresSafeArea()

                if habitStore.habits.isEmpty {
                    EmptyStateView(
                        systemImage: "chart.bar",
                        title: "No stats yet",
                        message: "Create some habits to see your progress here."
                    )
                } else {
                    ScrollView {
                        VStack(alignment: .leading, spacing: AppSpacing.lg) {
                            summaryCard

                            Text("Per habit")
                                .font(AppFont.label())
                                .foregroundStyle(Color.appTextSecondary)
                                .textCase(.uppercase)
                                .padding(.horizontal, AppSpacing.xs)
                                .padding(.top, AppSpacing.sm)

                            VStack(spacing: AppSpacing.md) {
                                ForEach(habitStore.habits) { habit in
                                    statRow(for: habit)
                                }
                            }
                        }
                        .padding(.horizontal, AppSpacing.lg)
                        .padding(.top, AppSpacing.md)
                        .padding(.bottom, AppSpacing.xxl)
                    }
                }
            }
            .navigationTitle("Stats")
            .toolbarBackground(Color.appBackground, for: .navigationBar)
            .toolbarBackground(.visible, for: .navigationBar)
        }
        .preferredColorScheme(.dark)
    }

    // MARK: - Summary

    private var summaryCard: some View {
        let longest = streaks.values.max() ?? 0
        let active = habitStore.habits.count
        let totalDone = 0  // TODO: real "done today" count once CheckIn endpoint is wired up

        return HStack(spacing: AppSpacing.md) {
            statBlock(value: "\(longest)", label: "Longest streak")
            divider
            statBlock(value: "\(active)", label: "Active habits")
            divider
            statBlock(value: "\(totalDone)", label: "Done today")
        }
        .padding(AppSpacing.lg)
        .frame(maxWidth: .infinity)
        .background(
            RoundedRectangle(cornerRadius: AppRadius.lg, style: .continuous)
                .fill(Color.appSurface)
        )
        .overlay(
            RoundedRectangle(cornerRadius: AppRadius.lg, style: .continuous)
                .stroke(Color.appBorder, lineWidth: 1)
        )
    }

    private var divider: some View {
        Rectangle()
            .fill(Color.appBorder)
            .frame(width: 1, height: 36)
    }

    private func statBlock(value: String, label: String) -> some View {
        VStack(spacing: AppSpacing.xs) {
            Text(value)
                .font(AppFont.title(size: 22))
                .foregroundStyle(Color.appTextPrimary)
            Text(label)
                .font(AppFont.caption(size: 11))
                .foregroundStyle(Color.appTextSecondary)
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity)
    }

    // MARK: - Per-habit row

    private func statRow(for habit: Habit) -> some View {
        let streak = streaks[habit.id, default: 0]
        return HStack(spacing: AppSpacing.lg) {
            RoundedRectangle(cornerRadius: 3)
                .fill(habit.color)
                .frame(width: 4, height: 40)
            VStack(alignment: .leading, spacing: 2) {
                Text(habit.name)
                    .font(AppFont.headline(size: 15))
                    .foregroundStyle(Color.appTextPrimary)
                Text("Started \(habit.createdAt.formatted(date: .abbreviated, time: .omitted))")
                    .font(AppFont.caption(size: 12))
                    .foregroundStyle(Color.appTextSecondary)
            }
            Spacer()
            HStack(spacing: 4) {
                Image(systemName: "flame.fill")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundStyle(streak > 0 ? Color.appAccent : Color.appTextTertiary)
                Text("\(streak)")
                    .font(AppFont.headline(size: 16))
                    .foregroundStyle(streak > 0 ? Color.appTextPrimary : Color.appTextTertiary)
            }
        }
        .padding(.vertical, AppSpacing.md)
        .padding(.horizontal, AppSpacing.lg)
        .background(
            RoundedRectangle(cornerRadius: AppRadius.md, style: .continuous)
                .fill(Color.appSurface)
        )
        .overlay(
            RoundedRectangle(cornerRadius: AppRadius.md, style: .continuous)
                .stroke(Color.appBorder, lineWidth: 1)
        )
    }
}

#Preview {
    StatsView()
        .environment(HabitStore())
}
