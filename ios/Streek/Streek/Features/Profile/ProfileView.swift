import SwiftUI

struct ProfileView: View {
    @Environment(AuthStore.self) private var auth
    @State private var showLogoutConfirm = false

    var body: some View {
        NavigationStack {
            ZStack {
                Color.appBackground.ignoresSafeArea()

                ScrollView {
                    VStack(spacing: AppSpacing.xl) {
                        userCard
                        sectionList
                        Spacer(minLength: AppSpacing.xxxl)
                        logoutButton
                    }
                    .padding(.horizontal, AppSpacing.lg)
                    .padding(.top, AppSpacing.md)
                    .padding(.bottom, AppSpacing.xxl)
                }
            }
            .navigationTitle("Profile")
            .toolbarBackground(Color.appBackground, for: .navigationBar)
            .toolbarBackground(.visible, for: .navigationBar)
        }
        .preferredColorScheme(.dark)
        .alert("Log out?", isPresented: $showLogoutConfirm) {
            Button("Log Out", role: .destructive) { auth.logout() }
            Button("Cancel", role: .cancel) {}
        } message: {
            Text("You'll need to log in again to access your habits.")
        }
    }

    // MARK: - User card

    @ViewBuilder
    private var userCard: some View {
        if case .authenticated(let user) = auth.state {
            HStack(spacing: AppSpacing.lg) {
                ZStack {
                    Circle()
                        .fill(Color.appAccent.opacity(0.18))
                        .frame(width: 56, height: 56)
                    Text(initialFor(email: user.email))
                        .font(AppFont.title(size: 22))
                        .foregroundStyle(Color.appAccent)
                }
                VStack(alignment: .leading, spacing: 2) {
                    Text(user.email)
                        .font(AppFont.headline(size: 16))
                        .foregroundStyle(Color.appTextPrimary)
                        .lineLimit(1)
                    Text("Joined \(user.createdAt.formatted(date: .abbreviated, time: .omitted))")
                        .font(AppFont.caption(size: 12))
                        .foregroundStyle(Color.appTextSecondary)
                }
                Spacer()
            }
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
    }

    // MARK: - Section list

    private var sectionList: some View {
        VStack(spacing: 0) {
            row(icon: "bell", title: "Notifications", subtitle: "Coming soon")
            divider
            row(icon: "lock", title: "Change Password", subtitle: "Coming soon")
            divider
            row(icon: "questionmark.circle", title: "Help & Feedback", subtitle: "")
        }
        .background(
            RoundedRectangle(cornerRadius: AppRadius.lg, style: .continuous)
                .fill(Color.appSurface)
        )
        .overlay(
            RoundedRectangle(cornerRadius: AppRadius.lg, style: .continuous)
                .stroke(Color.appBorder, lineWidth: 1)
        )
    }

    private func row(icon: String, title: String, subtitle: String) -> some View {
        HStack(spacing: AppSpacing.lg) {
            Image(systemName: icon)
                .font(.system(size: 16, weight: .medium))
                .foregroundStyle(Color.appTextSecondary)
                .frame(width: 24)
            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(AppFont.body(size: 15))
                    .foregroundStyle(Color.appTextPrimary)
                if !subtitle.isEmpty {
                    Text(subtitle)
                        .font(AppFont.caption(size: 12))
                        .foregroundStyle(Color.appTextTertiary)
                }
            }
            Spacer()
            Image(systemName: "chevron.right")
                .font(.system(size: 12, weight: .semibold))
                .foregroundStyle(Color.appTextTertiary)
        }
        .padding(.horizontal, AppSpacing.lg)
        .padding(.vertical, AppSpacing.md)
    }

    private var divider: some View {
        Rectangle()
            .fill(Color.appDivider)
            .frame(height: 1)
            .padding(.leading, AppSpacing.xxl + AppSpacing.lg)
    }

    private var logoutButton: some View {
        SecondaryButton(title: "Log Out", systemImage: "rectangle.portrait.and.arrow.right") {
            showLogoutConfirm = true
        }
    }

    private func initialFor(email: String) -> String {
        guard let first = email.first else { return "?" }
        return String(first).uppercased()
    }
}

#Preview {
    let store = AuthStore()
    return ProfileView()
        .environment(store)
}
