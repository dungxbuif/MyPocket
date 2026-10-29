import SwiftUI

struct AuthView: View {
    @ObservedObject var model: AuthViewModel
    @State private var email = ""
    @State private var password = ""

    var body: some View {
        ScrollView {
            VStack(spacing: 16) {
            Spacer()
            MPBrandMark(accessibilityIdentifier: "auth-brand-mark")
            MPText(AuthReferenceCopy.title, style: .display)
            MPText(AuthReferenceCopy.subtitle, style: .secondary)
                .multilineTextAlignment(.center)

            if let errorMessage = model.errorMessage {
                MPStatusMessage(text: errorMessage, tone: .danger)
            }

            if model.canSignInWithBiometrics {
                MPButton("Đăng nhập bằng Face ID", isLoading: model.isBusy) {
                    Task { await model.signInWithBiometrics() }
                }
                .accessibilityIdentifier("auth-face-id-login")
                MPText("Hoặc đăng nhập bằng mật khẩu", style: .secondary)
            }

            VStack(spacing: 12) {
                MPTextField(title: "Email", text: $email, placeholder: AuthReferenceCopy.emailPlaceholder, accessibilityIdentifier: "auth-email")
                    .textInputAutocapitalization(.never)
                    .keyboardType(.emailAddress)
                MPSecureField(title: "Mật khẩu", text: $password, placeholder: AuthReferenceCopy.passwordPlaceholder, accessibilityIdentifier: "auth-password")
            }

            MPButton("Đăng nhập", isLoading: model.isBusy, isDisabled: email.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || password.isEmpty) {
                Task { await model.signIn(email: email, password: password) }
            }
            Spacer()
            }
            .frame(maxWidth: 430)
            .frame(minHeight: 560)
            .padding(.horizontal, MPMetrics.pagePadding)
        }
        .scrollBounceBehavior(.basedOnSize)
        .background(MPColor.canvas.ignoresSafeArea())
        .onAppear {
            if email.isEmpty { email = model.rememberedDevice?.email ?? "" }
        }
    }
}
