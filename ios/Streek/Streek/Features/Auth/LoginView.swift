import SwiftUI

struct LoginView: View {
    @Environment(AuthStore.self) private var auth
    @State private var email = ""
    @State private var password = ""
    @State private var isSubmitting = false
    @State private var errorMessage: String?
    @State private var showRegister = false

    @FocusState private var focusedField: Field?
    enum Field { case email, password }

    var body: some View {
        ZStack {
            Color.appBackground.ignoresSafeArea()

            ScrollView {
                VStack(alignment: .leading, spacing: AppSpacing.xxl) {
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
                            textContentType: .password,
                            isSecure: true,
                            submitLabel: .go,
                            onSubmit: submit
                        )
                    }

                    PrimaryButton(
                        title: "Log In",
                        isLoading: isSubmitting,
                        isEnabled: canSubmit,
                        action: submit
                    )

                    Divider()
                        .background(Color.appDivider)
                        .padding(.vertical, AppSpacing.sm)

                    VStack(spacing: AppSpacing.sm) {
                        Text("Don't have an account?")
                            .font(AppFont.body(size: 14))
                            .foregroundStyle(Color.appTextSecondary)
                        Button {
                            showRegister = true
                        } label: {
                            Text("Sign up")
                                .font(AppFont.headline(size: 15))
                                .foregroundStyle(Color.appAccent)
                        }
                    }
                    .frame(maxWidth: .infinity)
                }
                .padding(.horizontal, AppSpacing.xl)
                .padding(.top, AppSpacing.xxxl)
                .padding(.bottom, AppSpacing.xxl)
            }
            .scrollDismissesKeyboard(.interactively)
        }
        .sheet(isPresented: $showRegister) {
            NavigationStack {
                RegisterView()
            }
        }
        .preferredColorScheme(.dark)
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: AppSpacing.sm) {
            HStack(spacing: AppSpacing.md) {
                Image(systemName: "flame.fill")
                    .font(.system(size: 32, weight: .bold))
                    .foregroundStyle(Color.appAccent)
                Text("Streek")
                    .font(AppFont.title(size: 32))
                    .foregroundStyle(Color.appTextPrimary)
            }
            Text("Welcome back. Keep the streak going.")
                .font(AppFont.body(size: 15))
                .foregroundStyle(Color.appTextSecondary)
        }
    }

    private var canSubmit: Bool {
        !email.trimmingCharacters(in: .whitespaces).isEmpty &&
        password.count >= 8
    }

    private func submit() {
        guard canSubmit, !isSubmitting else { return }
        isSubmitting = true
        errorMessage = nil
        Task {
            do {
                try await auth.login(
                    email: email.trimmingCharacters(in: .whitespacesAndNewlines),
                    password: password
                )
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
    LoginView()
        .environment(AuthStore())
}
