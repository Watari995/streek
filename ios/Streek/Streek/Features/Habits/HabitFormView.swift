import SwiftUI

struct HabitFormView: View {
    enum Mode {
        case create
        case edit(Habit)

        var title: String {
            switch self {
            case .create: return "New Habit"
            case .edit: return "Edit Habit"
            }
        }
        var isEdit: Bool {
            if case .edit = self { return true }
            return false
        }
    }

    @Environment(\.dismiss) private var dismiss
    @Environment(HabitStore.self) private var habitStore

    let mode: Mode

    @State private var name: String
    @State private var description: String
    @State private var selectedColor: String
    @State private var isSubmitting = false
    @State private var errorMessage: String?
    @State private var showDeleteConfirm = false

    private static let colorPalette: [String] = [
        "#FF0211", "#FF9500", "#FFCC00",
        "#00C853", "#00BCD4", "#5E5CE6",
        "#FF2D55", "#AF52DE", "#34C759"
    ]

    init(mode: Mode) {
        self.mode = mode
        switch mode {
        case .create:
            _name = State(initialValue: "")
            _description = State(initialValue: "")
            _selectedColor = State(initialValue: Self.colorPalette[0])
        case .edit(let habit):
            _name = State(initialValue: habit.name)
            _description = State(initialValue: habit.description ?? "")
            _selectedColor = State(initialValue: habit.labelColor)
        }
    }

    var body: some View {
        ZStack {
            Color.appBackground.ignoresSafeArea()

            ScrollView {
                VStack(alignment: .leading, spacing: AppSpacing.xl) {
                    if let errorMessage {
                        ErrorBanner(message: errorMessage)
                    }

                    StreekTextField(
                        label: "Name",
                        text: $name,
                        placeholder: "e.g. Morning Run",
                        autocapitalization: .sentences,
                        submitLabel: .next
                    )

                    StreekTextField(
                        label: "Description (optional)",
                        text: $description,
                        placeholder: "Short note about this habit",
                        autocapitalization: .sentences,
                        submitLabel: .done
                    )

                    VStack(alignment: .leading, spacing: AppSpacing.md) {
                        Text("Color")
                            .font(AppFont.label())
                            .foregroundStyle(Color.appTextSecondary)
                        colorGrid
                    }

                    PrimaryButton(
                        title: mode.isEdit ? "Save Changes" : "Add Habit",
                        isLoading: isSubmitting,
                        isEnabled: canSave,
                        action: save
                    )

                    if mode.isEdit {
                        Button(role: .destructive) {
                            showDeleteConfirm = true
                        } label: {
                            HStack(spacing: AppSpacing.sm) {
                                Image(systemName: "trash")
                                Text("Delete Habit")
                                    .font(AppFont.headline(size: 15))
                            }
                            .frame(maxWidth: .infinity)
                            .frame(height: 48)
                            .foregroundStyle(Color.appDanger)
                        }
                        .padding(.top, AppSpacing.md)
                        .disabled(isSubmitting)
                    }
                }
                .padding(.horizontal, AppSpacing.xl)
                .padding(.top, AppSpacing.lg)
                .padding(.bottom, AppSpacing.xxl)
            }
            .scrollDismissesKeyboard(.interactively)
        }
        .navigationTitle(mode.title)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") { dismiss() }
                    .foregroundStyle(Color.appTextSecondary)
                    .disabled(isSubmitting)
            }
        }
        .toolbarBackground(Color.appBackground, for: .navigationBar)
        .toolbarBackground(.visible, for: .navigationBar)
        .preferredColorScheme(.dark)
        .alert("Delete this habit?", isPresented: $showDeleteConfirm) {
            Button("Delete", role: .destructive) { delete() }
            Button("Cancel", role: .cancel) {}
        } message: {
            Text("This will remove the habit and all its check-ins.")
        }
    }

    // MARK: - Color grid

    private var colorGrid: some View {
        let columns = Array(repeating: GridItem(.flexible(), spacing: AppSpacing.md), count: 5)
        return LazyVGrid(columns: columns, spacing: AppSpacing.md) {
            ForEach(Self.colorPalette, id: \.self) { hex in
                Button {
                    selectedColor = hex
                    let g = UISelectionFeedbackGenerator()
                    g.selectionChanged()
                } label: {
                    Circle()
                        .fill(Color(hexString: hex) ?? .appAccent)
                        .frame(height: 40)
                        .overlay(
                            Circle()
                                .stroke(Color.white,
                                        lineWidth: selectedColor == hex ? 2.5 : 0)
                        )
                        .padding(2)
                }
                .buttonStyle(PressableScaleStyle())
            }
        }
    }

    // MARK: - Actions

    private var canSave: Bool {
        !name.trimmingCharacters(in: .whitespaces).isEmpty
    }

    /// Persists the form via HabitStore. The store is the only place that
    /// knows about `APIClient`; this view just calls a method.
    private func save() {
        guard canSave, !isSubmitting else { return }
        let trimmedName = name.trimmingCharacters(in: .whitespacesAndNewlines)
        let trimmedDesc = description.trimmingCharacters(in: .whitespacesAndNewlines)
        let descValue: String? = trimmedDesc.isEmpty ? nil : trimmedDesc

        isSubmitting = true
        errorMessage = nil
        Task {
            do {
                switch mode {
                case .create:
                    _ = try await habitStore.createHabit(
                        name: trimmedName,
                        description: descValue,
                        labelColor: selectedColor
                    )
                case .edit(let habit):
                    _ = try await habitStore.updateHabit(
                        id: habit.id,
                        name: trimmedName,
                        description: descValue,
                        labelColor: selectedColor
                    )
                }
                dismiss()
            } catch let APIError.server(_, _, message) {
                errorMessage = message
                isSubmitting = false
            } catch {
                errorMessage = error.localizedDescription
                isSubmitting = false
            }
        }
    }

    private func delete() {
        guard case .edit(let habit) = mode, !isSubmitting else { return }
        isSubmitting = true
        errorMessage = nil
        Task {
            do {
                try await habitStore.deleteHabit(id: habit.id)
                dismiss()
            } catch let APIError.server(_, _, message) {
                errorMessage = message
                isSubmitting = false
            } catch {
                errorMessage = error.localizedDescription
                isSubmitting = false
            }
        }
    }
}

#Preview {
    NavigationStack {
        HabitFormView(mode: .create)
            .environment(HabitStore())
    }
}
