import SwiftUI

struct StatsView: View {
    @Environment(HabitStore.self) private var habitStore
    @Environment(StatsStore.self) private var statsStore

    var body: some View {
        NavigationStack {
            ZStack {
                Color.appBackground.ignoresSafeArea()
                content
            }
            .navigationTitle("Stats")
            .toolbarBackground(Color.appBackground, for: .navigationBar)
            .toolbarBackground(.visible, for: .navigationBar)
            .task {
                if case .idle = statsStore.loadState {
                    await statsStore.loadOverview()
                }
            }
            .refreshable {
                await statsStore.loadOverview()
            }
        }
        .preferredColorScheme(.dark)
    }

    // MARK: - Content branching

    @ViewBuilder
    private var content: some View {
        if statsStore.habitOverviews.isEmpty {
            emptyOrLoadingState
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
                        ForEach(statsStore.habitOverviews) { habitStat in
                            statRow(for: habitStat)
                        }
                    }
                }
                .padding(.horizontal, AppSpacing.lg)
                .padding(.top, AppSpacing.md)
                .padding(.bottom, AppSpacing.xxl)
            }
        }
    }

    @ViewBuilder
    private var emptyOrLoadingState: some View {
        switch statsStore.loadState {
        case .loading:
            ProgressView()
                .progressViewStyle(.circular)
                .tint(Color.appAccent)
        case .failed(let message):
            EmptyStateView(
                systemImage: "exclamationmark.triangle",
                title: "Couldn't load stats",
                message: message,
                actionTitle: "Retry",
                action: {
                    Task { await statsStore.loadOverview() }
                }
            )
        case .idle, .loaded:
            EmptyStateView(
                systemImage: "chart.bar",
                title: "No stats yet",
                message: "Create some habits to see your progress here."
            )
        }
    }

    // MARK: - Summary

    private var summaryCard: some View {
        HStack(spacing: AppSpacing.md) {
            statBlock(value: "\(statsStore.longestStreak)", label: "Longest streak")
            divider
            statBlock(value: "\(statsStore.habitOverviews.count)", label: "Active habits")
            divider
            statBlock(value: "\(statsStore.doneToday)", label: "Done today")
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

    private func statRow(for habitStat: HabitOverview) -> some View {
        let labelColor = Color(hexString: habitStat.labelColor) ?? .appAccent
        let createdAt = habitStore.habits.first(where: { $0.id == habitStat.habitId })?.createdAt
        let streak = habitStat.currentStreak

        return HStack(spacing: AppSpacing.lg) {
            RoundedRectangle(cornerRadius: 3)
                .fill(labelColor)
                .frame(width: 4, height: 40)
            VStack(alignment: .leading, spacing: 2) {
                Text(habitStat.habitName)
                    .font(AppFont.headline(size: 15))
                    .foregroundStyle(Color.appTextPrimary)
                if let createdAt {
                    Text("Started \(createdAt.formatted(date: .abbreviated, time: .omitted))")
                        .font(AppFont.caption(size: 12))
                        .foregroundStyle(Color.appTextSecondary)
                } else {
                    Text("Best streak: \(habitStat.longestStreak)")
                        .font(AppFont.caption(size: 12))
                        .foregroundStyle(Color.appTextSecondary)
                }
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
        .environment(StatsStore())
}
