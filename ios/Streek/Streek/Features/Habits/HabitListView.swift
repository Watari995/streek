import SwiftUI

struct HabitListView: View {
    @Environment(HabitStore.self) private var habitStore
    @Environment(CheckInStore.self) private var checkInStore
    @Environment(StatsStore.self) private var statsStore
    @Environment(PointStore.self) private var pointStore

    @State private var showingForm = false
    @State private var editingHabit: Habit?
    @State private var toggleErrorMessage: String?
    @State private var celebration: Celebration?

    /// Identifiable payload for the milestone celebration overlay.
    private struct Celebration: Identifiable {
        let id = UUID()
        let habitName: String
        let streak: Int
    }

    var body: some View {
        NavigationStack {
            content
                .navigationTitle("Habits")
                .toolbar {
                    ToolbarItem(placement: .primaryAction) {
                        Button {
                            showingForm = true
                        } label: {
                            Image(systemName: "plus")
                                .font(.system(size: 16, weight: .semibold))
                                .foregroundStyle(Color.appAccent)
                        }
                    }
                }
                .toolbarBackground(Color.appBackground, for: .navigationBar)
                .toolbarBackground(.visible, for: .navigationBar)
                .task {
                    if case .idle = habitStore.loadState {
                        await habitStore.loadHabits()
                    }
                    if case .idle = statsStore.loadState {
                        await statsStore.loadOverview()
                    }
                }
                .refreshable {
                    await habitStore.loadHabits()
                    await statsStore.loadOverview()
                }
                .sheet(isPresented: $showingForm) {
                    NavigationStack {
                        HabitFormView(mode: .create)
                    }
                }
                .sheet(item: $editingHabit) { habit in
                    NavigationStack {
                        HabitFormView(mode: .edit(habit))
                    }
                }
                .alert(
                    "Couldn't save check-in",
                    isPresented: Binding(
                        get: { toggleErrorMessage != nil },
                        set: { if !$0 { toggleErrorMessage = nil } }
                    ),
                    presenting: toggleErrorMessage
                ) { _ in
                    Button("OK", role: .cancel) {}
                } message: { message in
                    Text(message)
                }
                .overlay {
                    if let celebration {
                        MilestoneCelebrationView(
                            habitName: celebration.habitName,
                            streak: celebration.streak,
                            onDismiss: { self.celebration = nil }
                        )
                        .transition(.opacity)
                    }
                }
                .animation(.easeInOut(duration: 0.2), value: celebration?.id)
        }
        .preferredColorScheme(.dark)
    }

    // MARK: - Content branching

    @ViewBuilder
    private var content: some View {
        ZStack {
            Color.appBackground.ignoresSafeArea()

            if habitStore.habits.isEmpty {
                emptyOrLoadingState
            } else {
                habitListScroll
            }
        }
    }

    @ViewBuilder
    private var emptyOrLoadingState: some View {
        switch habitStore.loadState {
        case .loading:
            ProgressView()
                .progressViewStyle(.circular)
                .tint(Color.appAccent)
        case .failed(let message):
            EmptyStateView(
                systemImage: "exclamationmark.triangle",
                title: "Couldn't load habits",
                message: message,
                actionTitle: "Retry",
                action: {
                    Task { await habitStore.loadHabits() }
                }
            )
        case .idle, .loaded:
            EmptyStateView(
                systemImage: "flame",
                title: "No habits yet",
                message: "Tap + to add your first habit and start a streak.",
                actionTitle: "Add Habit",
                action: { showingForm = true }
            )
        }
    }

    private var habitListScroll: some View {
        ScrollView {
            LazyVStack(spacing: AppSpacing.md) {
                todayHeader

                ForEach(habitStore.habits) { habit in
                    HabitRowView(
                        habit: habit,
                        isCheckedToday: checkInStore.isChecked(habitId: habit.id),
                        streak: statsStore.streak(for: habit.id).current,
                        onToggle: { toggle(habit) }
                    )
                    .onTapGesture {
                        editingHabit = habit
                    }
                }
            }
            .padding(.horizontal, AppSpacing.lg)
            .padding(.bottom, AppSpacing.xxl)
        }
    }

    // MARK: - Today header

    private var todayHeader: some View {
        VStack(alignment: .leading, spacing: AppSpacing.xs) {
            Text(todayString)
                .font(AppFont.label())
                .foregroundStyle(Color.appTextSecondary)
                .textCase(.uppercase)
            HStack(spacing: AppSpacing.sm) {
                Text("\(completedCount) / \(habitStore.habits.count)")
                    .font(AppFont.title(size: 30))
                    .foregroundStyle(Color.appTextPrimary)
                Text("done")
                    .font(AppFont.body(size: 16))
                    .foregroundStyle(Color.appTextSecondary)
                    .padding(.bottom, 4)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, AppSpacing.xs)
        .padding(.top, AppSpacing.sm)
        .padding(.bottom, AppSpacing.md)
    }

    /// "Done today" count from CheckInStore (optimistic, instant UI feedback).
    /// StatsStore.doneToday lags by one network round-trip after a toggle, so
    /// we prefer the local count here.
    private var completedCount: Int {
        habitStore.habits.reduce(0) { count, habit in
            count + (checkInStore.isChecked(habitId: habit.id) ? 1 : 0)
        }
    }

    private var todayString: String {
        let f = DateFormatter()
        f.dateFormat = "EEEE, MMM d"
        return f.string(from: Date())
    }

    // MARK: - Actions

    /// Tapping the row's check button.
    /// 1. CheckInStore optimistically flips local state and dispatches the API.
    /// 2. After the API succeeds, refresh StatsStore so streaks update.
    private func toggle(_ habit: Habit) {
        let generator = UIImpactFeedbackGenerator(style: .medium)
        generator.impactOccurred()
        // Server-side streak before the toggle, used to detect a milestone
        // crossing once the overview re-syncs.
        let streakBefore = statsStore.streak(for: habit.id).current
        Task {
            do {
                try await checkInStore.toggle(habitId: habit.id)
                // Streaks must be re-synced before we can detect a milestone,
                // so await the overview; points can refresh in the background.
                async let _: Void = pointStore.loadBalance()
                await statsStore.loadOverview()

                let streakAfter = statsStore.streak(for: habit.id).current
                if let milestone = StreakMilestone.reached(
                    before: streakBefore,
                    after: streakAfter
                ) {
                    celebration = Celebration(
                        habitName: habit.name,
                        streak: milestone
                    )
                }
            } catch let APIError.server(_, _, message) {
                toggleErrorMessage = message
            } catch {
                toggleErrorMessage = error.localizedDescription
            }
        }
    }
}

#Preview {
    HabitListView()
        .environment(AuthStore())
        .environment(HabitStore())
        .environment(CheckInStore())
        .environment(StatsStore())
        .environment(PointStore())
}
