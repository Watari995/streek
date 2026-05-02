import SwiftUI

struct ErrorBanner: View {
    let message: String

    var body: some View {
        HStack(alignment: .top, spacing: AppSpacing.md) {
            Image(systemName: "exclamationmark.triangle.fill")
                .foregroundStyle(Color.appDanger)
            Text(message)
                .font(AppFont.body(size: 14))
                .foregroundStyle(Color.appTextPrimary)
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(AppSpacing.md)
        .background(
            RoundedRectangle(cornerRadius: AppRadius.md, style: .continuous)
                .fill(Color.appDanger.opacity(0.12))
        )
        .overlay(
            RoundedRectangle(cornerRadius: AppRadius.md, style: .continuous)
                .stroke(Color.appDanger.opacity(0.3), lineWidth: 1)
        )
        .transition(.move(edge: .top).combined(with: .opacity))
    }
}
