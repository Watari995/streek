import SwiftUI

struct EmptyStateView: View {
    let systemImage: String
    let title: String
    let message: String
    var actionTitle: String? = nil
    var action: (() -> Void)? = nil

    var body: some View {
        VStack(spacing: AppSpacing.lg) {
            Image(systemName: systemImage)
                .font(.system(size: 56, weight: .light))
                .foregroundStyle(Color.appTextTertiary)

            VStack(spacing: AppSpacing.sm) {
                Text(title)
                    .font(AppFont.headline(size: 18))
                    .foregroundStyle(Color.appTextPrimary)
                Text(message)
                    .font(AppFont.body(size: 14))
                    .foregroundStyle(Color.appTextSecondary)
                    .multilineTextAlignment(.center)
            }

            if let actionTitle, let action {
                PrimaryButton(title: actionTitle, action: action)
                    .frame(maxWidth: 240)
                    .padding(.top, AppSpacing.sm)
            }
        }
        .padding(.horizontal, AppSpacing.xxl)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}
