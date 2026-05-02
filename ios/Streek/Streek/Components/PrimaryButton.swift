import SwiftUI

struct PrimaryButton: View {
    let title: String
    var systemImage: String? = nil
    var isLoading: Bool = false
    var isEnabled: Bool = true
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: AppSpacing.sm) {
                if isLoading {
                    ProgressView()
                        .progressViewStyle(.circular)
                        .tint(.white)
                } else if let systemImage {
                    Image(systemName: systemImage)
                }
                Text(title)
                    .font(AppFont.headline(size: 16))
            }
            .frame(maxWidth: .infinity)
            .frame(height: 52)
            .foregroundStyle(.white)
            .background(
                RoundedRectangle(cornerRadius: AppRadius.md, style: .continuous)
                    .fill(isEnabled && !isLoading ? Color.appAccent : Color.appAccent.opacity(0.4))
            )
        }
        .buttonStyle(PressableScaleStyle())
        .disabled(!isEnabled || isLoading)
        .animation(.easeOut(duration: 0.15), value: isLoading)
    }
}

struct SecondaryButton: View {
    let title: String
    var systemImage: String? = nil
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: AppSpacing.sm) {
                if let systemImage {
                    Image(systemName: systemImage)
                }
                Text(title)
                    .font(AppFont.headline(size: 16))
            }
            .frame(maxWidth: .infinity)
            .frame(height: 52)
            .foregroundStyle(Color.appTextPrimary)
            .background(
                RoundedRectangle(cornerRadius: AppRadius.md, style: .continuous)
                    .fill(Color.appSurfaceElevated)
            )
            .overlay(
                RoundedRectangle(cornerRadius: AppRadius.md, style: .continuous)
                    .stroke(Color.appBorder, lineWidth: 1)
            )
        }
        .buttonStyle(PressableScaleStyle())
    }
}

/// Subtle press-down scale animation, common on Mercari-like UIs.
struct PressableScaleStyle: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? 0.98 : 1.0)
            .opacity(configuration.isPressed ? 0.9 : 1.0)
            .animation(.easeOut(duration: 0.12), value: configuration.isPressed)
    }
}
