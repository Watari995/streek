import SwiftUI

struct StreekTextField: View {
    let label: String
    @Binding var text: String
    var placeholder: String = ""
    var keyboardType: UIKeyboardType = .default
    var textContentType: UITextContentType? = nil
    var autocapitalization: TextInputAutocapitalization = .never
    var isSecure: Bool = false
    var submitLabel: SubmitLabel = .next
    var onSubmit: (() -> Void)? = nil

    @FocusState private var isFocused: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: AppSpacing.sm) {
            Text(label)
                .font(AppFont.label())
                .foregroundStyle(Color.appTextSecondary)

            HStack {
                inputField
                    .focused($isFocused)
                    .keyboardType(keyboardType)
                    .textInputAutocapitalization(autocapitalization)
                    .autocorrectionDisabled()
                    .submitLabel(submitLabel)
                    .onSubmit { onSubmit?() }
                    .ifLet(textContentType) { view, value in
                        view.textContentType(value)
                    }
            }
            .font(AppFont.body(size: 16))
            .foregroundStyle(Color.appTextPrimary)
            .padding(.horizontal, AppSpacing.lg)
            .frame(height: 52)
            .background(
                RoundedRectangle(cornerRadius: AppRadius.md, style: .continuous)
                    .fill(Color.appSurface)
            )
            .overlay(
                RoundedRectangle(cornerRadius: AppRadius.md, style: .continuous)
                    .stroke(isFocused ? Color.appAccent : Color.appBorder, lineWidth: 1)
            )
            .animation(.easeOut(duration: 0.15), value: isFocused)
        }
    }

    @ViewBuilder
    private var inputField: some View {
        if isSecure {
            SecureField(placeholder, text: $text)
        } else {
            TextField(placeholder, text: $text)
        }
    }
}

// Helpers for conditional view modifiers
extension View {
    @ViewBuilder
    func ifLet<Value, Transform: View>(_ value: Value?,
                                       transform: (Self, Value) -> Transform) -> some View {
        if let value {
            transform(self, value)
        } else {
            self
        }
    }
}
