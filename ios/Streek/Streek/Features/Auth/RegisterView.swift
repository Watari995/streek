import SwiftUI

struct RegisterView: View {
    @Environment(AuthStore.self) private var auth
    @Environment(\.dismiss) private var dismiss
    @State private var email = ""
    @State private var password = ""
    @State private var confirmPassword = ""
    @State private var isSubmitting = false
    @State private var errorMessage: String?

    @FocusState private var focusedField: Field?
    enum Field { case email, password, confirm }

    var body: some View {
        ZStack {
            Color.appBackground.ignoresSafeArea()

            ScrollView {
                VStack(alignment: .leading, spacing: AppSpacing.xl) {
                    header

                    if let errorMessage {
                        ErrorBanner(message: errorMessage)
                    }

                    VStack(spacing: AppSpacing.lg) {
                        StreekTextField(
                            label: "Email",
                            text: $email,
                            placeholder: "you@example.com",
                            keyboardType: .emailAddress,
                            textContentType: .emailAddress,
                            submitLabel: .next,
                            onSubmit: { focusedField = .password }
                        )

                        StreekTextField(
                            label: "Password",
                            text: $password,
                            placeholder: "8+ characters",
                            textContentType: .newPassword,
                            isSecure: true,
                            submitLabel: .next,
                            onSubmit: { focusedField = .confirm }
                        )

                        StreekTextField(
                            label: "Confirm Password",
                            text: $confirmPassword,
                            placeholder: "Re-enter password",
                            textContentType: .newPassword,
                            isSecure: true,
                            submitLabel: .go,
                            onSubmit: submit
                        )

                        if !password.isEmpty && !confirmPassword.isEmpty && password != confirmPassword {
                            Text("Passwords don't match.")
                                .font(AppFont.caption())
                                .foregroundStyle(Color.appDanger)
                        }
                    }

                    PrimaryButton(
                        title: "Create Account",
                        isLoading: isSubmitting,
                        isEnabled: canSubmit,
                        action: submit
                    )
                    .padding(.top, AppSpacing.sm)
                }
                .padding(.horizontal, AppSpacing.xl)
                .padding(.top, AppSpacing.lg)
                .padding(.bottom, AppSpacing.xxl)
            }
            .scrollDismissesKeyboard(.interactively)
        }
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button("Cancel") { dismiss() }
                    .foregroundStyle(Color.appTextSecondary)
            }
        }
        .toolbarBackground(Color.appBackground, for: .navigationBar)
        .toolbarBackground(.visible, for: .navigationBar)
        .preferredColorScheme(.dark)
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: AppSpacing.sm) {
            Text("Sign Up")
                .font(AppFont.title(size: 28))
                .foregroundStyle(Color.appTextPrimary)
            Text("Build habits one day at a time.")
                .font(AppFont.body(size: 15))
                .foregroundStyle(Color.appTextSecondary)
        }
    }

    private var canSubmit: Bool {
        !email.trimmingCharacters(in: .whitespaces).isEmpty &&
        password.count >= 8 &&
        password == confirmPassword
    }

    private func submit() {
        guard canSubmit, !isSubmitting else { return }
        isSubmitting = true
        errorMessage = nil
        Task {
            do {
                try await auth.register(
                    email: email.trimmingCharacters(in: .whitespacesAndNewlines),
                    password: password
                )
                dismiss()
            } catch let APIError.server(_, _, message) {
                errorMessage = message
            } catch {
                errorMessage = error.localizedDescription
            }
            isSubmitting = false
        }
    }
}

#Preview {
    NavigationStack {
        RegisterView()
            .environment(AuthStore())
    }
}
