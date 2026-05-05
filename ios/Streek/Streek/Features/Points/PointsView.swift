import SwiftUI

struct PointsView: View {
    @Environment(PointStore.self) private var pointStore
    @Environment(HabitStore.self) private var habitStore

    var body: some View {
        NavigationStack {
            ZStack {
                Color.appBackground.ignoresSafeArea()
                content
            }
            .navigationTitle("Points")
            .toolbarBackground(Color.appBackground, for: .navigationBar)
            .toolbarBackground(.visible, for: .navigationBar)
            .task {
                if case .idle = pointStore.balanceState {
                    await pointStore.loadBalance()
                }
                if case .idle = pointStore.historyState {
                    await pointStore.loadHistory()
                }
            }
            .refreshable {
                async let b: Void = pointStore.loadBalance()
                async let h: Void = pointStore.loadHistory()
                _ = await (b, h)
            }
        }
        .preferredColorScheme(.dark)
    }

    // MARK: - Content branching

    @ViewBuilder
    private var content: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: AppSpacing.lg) {
                balanceCard

                Text("Recent activity")
                    .font(AppFont.label())
                    .foregroundStyle(Color.appTextSecondary)
                    .textCase(.uppercase)
                    .padding(.horizontal, AppSpacing.xs)
                    .padding(.top, AppSpacing.sm)

                historySection
            }
            .padding(.horizontal, AppSpacing.lg)
            .padding(.top, AppSpacing.md)
            .padding(.bottom, AppSpacing.xxl)
        }
    }

    // MARK: - Balance

    private var balanceCard: some View {
        VStack(alignment: .leading, spacing: AppSpacing.sm) {
            Text("Current balance")
                .font(AppFont.caption(size: 12))
                .foregroundStyle(Color.appTextSecondary)
                .textCase(.uppercase)

            HStack(alignment: .firstTextBaseline, spacing: AppSpacing.sm) {
                Text(balanceDisplay)
                    .font(AppFont.title(size: 40))
                    .foregroundStyle(Color.appTextPrimary)
                    .contentTransition(.numericText())
                Text("pts")
                    .font(AppFont.headline(size: 16))
                    .foregroundStyle(Color.appTextSecondary)
            }

            if case .failed(let message) = pointStore.balanceState {
                Text(message)
                    .font(AppFont.caption())
                    .foregroundStyle(Color.appDanger)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(AppSpacing.lg)
        .background(
            RoundedRectangle(cornerRadius: AppRadius.lg, style: .continuous)
                .fill(Color.appSurface)
        )
        .overlay(
            RoundedRectangle(cornerRadius: AppRadius.lg, style: .continuous)
                .stroke(Color.appBorder, lineWidth: 1)
        )
    }

    private var balanceDisplay: String {
        switch pointStore.balanceState {
        case .idle, .loading:
            return "—"
        case .loaded, .failed:
            return "\(pointStore.balance)"
        }
    }

    // MARK: - History

    @ViewBuilder
    private var historySection: some View {
        switch pointStore.historyState {
        case .loading where pointStore.history.isEmpty:
            ProgressView()
                .progressViewStyle(.circular)
                .tint(Color.appAccent)
                .frame(maxWidth: .infinity)
                .padding(.vertical, AppSpacing.xxl)
        case .failed(let message) where pointStore.history.isEmpty:
            EmptyStateView(
                systemImage: "exclamationmark.triangle",
                title: "Couldn't load history",
                message: message,
                actionTitle: "Retry",
                action: {
                    Task { await pointStore.loadHistory() }
                }
            )
            .frame(minHeight: 200)
        case _ where pointStore.history.isEmpty:
            EmptyStateView(
                systemImage: "sparkles",
                title: "No points yet",
                message: "Check in on your habits to start earning points."
            )
            .frame(minHeight: 200)
        default:
            VStack(spacing: AppSpacing.sm) {
                ForEach(pointStore.history) { entry in
                    historyRow(for: entry)
                }
            }
        }
    }

    private func historyRow(for entry: PointHistoryEntry) -> some View {
        let habitName = habitName(for: entry.habitId)
        let isEarn = entry.type == "EARN"
        let amountColor: Color = isEarn ? .appSuccess : .appWarning
        let sign = isEarn ? "+" : "-"

        return HStack(spacing: AppSpacing.lg) {
            ZStack {
                Circle()
                    .fill(Color.appSurfaceElevated)
                    .frame(width: 36, height: 36)
                Image(systemName: isEarn ? "plus" : "minus")
                    .font(.system(size: 14, weight: .bold))
                    .foregroundStyle(amountColor)
            }
            VStack(alignment: .leading, spacing: 2) {
                Text(displayLabel(for: entry))
                    .font(AppFont.headline(size: 15))
                    .foregroundStyle(Color.appTextPrimary)
                if let habitName {
                    Text(habitName)
                        .font(AppFont.caption(size: 12))
                        .foregroundStyle(Color.appTextSecondary)
                } else {
                    Text(entry.createdAt.formatted(date: .abbreviated, time: .shortened))
                        .font(AppFont.caption(size: 12))
                        .foregroundStyle(Color.appTextSecondary)
                }
            }
            Spacer()
            Text("\(sign)\(entry.amount)")
                .font(AppFont.headline(size: 16))
                .foregroundStyle(amountColor)
                .monospacedDigit()
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

    private func habitName(for habitId: String?) -> String? {
        guard let habitId else { return nil }
        return habitStore.habits.first(where: { $0.id == habitId })?.name
    }

    private func displayLabel(for entry: PointHistoryEntry) -> String {
        switch entry.reason {
        case "checkIn", "CHECK_IN":
            return "Daily check-in"
        case "STREAK_BONUS":
            return "Streak bonus"
        case "REDEEM":
            return "Redeemed"
        default:
            return entry.reason
        }
    }
}

#Preview {
    PointsView()
        .environment(PointStore())
        .environment(HabitStore())
}
