import SwiftUI

struct HabitListView: View {
    @Environment(AuthStore.self) private var auth
    @State private var habits: [Habit] = MockHabits.all
    @State private var checkedState: [String: Bool] = MockHabits.initialCheckedState
    @State private var streaks: [String: Int] = MockHabits.streaks
    @State private var showingForm = false
    @State private var editingHabit: Habit?

    var body: some View {
        NavigationStack {
            ZStack {
                Color.appBackground.ignoresSafeArea()

                if habits.isEmpty {
                    EmptyStateView(
                        systemImage: "flame",
                        title: "No habits yet",
                        message: "Tap + to add your first habit and start a streak.",
                        actionTitle: "Add Habit",
                        action: { showingForm = true }
                    )
                } else {
                    ScrollView {
                        LazyVStack(spacing: AppSpacing.md) {
                            todayHeader

                            ForEach(habits) { habit in
                                HabitRowView(
                                    habit: habit,
                                    isCheckedToday: checkedState[habit.id, default: false],
                                    streak: streaks[habit.id, default: 0],
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
            }
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
            .sheet(isPresented: $showingForm) {
                NavigationStack {
                    HabitFormView(mode: .create) { newHabit in
                        habits.append(newHabit)
                        checkedState[newHabit.id] = false
                        streaks[newHabit.id] = 0
                    }
                }
            }
            .sheet(item: $editingHabit) { habit in
                NavigationStack {
                    HabitFormView(mode: .edit(habit)) { updated in
                        if let idx = habits.firstIndex(where: { $0.id == updated.id }) {
                            habits[idx] = updated
                        }
                    } onDelete: { id in
                        habits.removeAll { $0.id == id }
                        checkedState.removeValue(forKey: id)
                        streaks.removeValue(forKey: id)
                    }
                }
            }
        }
        .preferredColorScheme(.dark)
    }

    // MARK: - Today header

    private var todayHeader: some View {
        VStack(alignment: .leading, spacing: AppSpacing.xs) {
            Text(todayString)
                .font(AppFont.label())
                .foregroundStyle(Color.appTextSecondary)
                .textCase(.uppercase)
            HStack(spacing: AppSpacing.sm) {
                Text("\(completedCount) / \(habits.count)")
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

    private var completedCount: Int {
        checkedState.values.filter { $0 }.count
    }

    private var todayString: String {
        let f = DateFormatter()
        f.dateFormat = "EEEE, MMM d"
        return f.string(from: Date())
    }

    // MARK: - Actions

    private func toggle(_ habit: Habit) {
        let wasChecked = checkedState[habit.id, default: false]
        checkedState[habit.id] = !wasChecked
        let current = streaks[habit.id, default: 0]
        streaks[habit.id] = wasChecked ? max(0, current - 1) : current + 1

        // Haptic feedback
        let generator = UIImpactFeedbackGenerator(style: .medium)
        generator.impactOccurred()
    }
}

#Preview {
    HabitListView()
        .environment(AuthStore())
}
