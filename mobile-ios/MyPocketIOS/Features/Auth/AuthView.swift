import SwiftUI

struct AuthView: View {
    @ObservedObject var model: AuthViewModel
    @State private var email = ""
    @State private var password = ""

    var body: some View {
        VStack(spacing: 18) {
            Spacer()
            MPSurfaceCard {
                VStack(alignment: .leading, spacing: 10) {
                    MPText("MyPocket", style: .title)
                    MPText("Đăng nhập bằng email để quản lý tài chính của bạn.", style: .secondary)
                }
            }
            MPSurfaceCard {
                VStack(spacing: 12) {
                    MPTextField(title: "Email", text: $email, placeholder: "you@example.com", accessibilityIdentifier: "auth-email")
                        .textInputAutocapitalization(.never)
                        .keyboardType(.emailAddress)
                    MPSecureField(title: "Mật khẩu", text: $password, placeholder: "Mật khẩu", accessibilityIdentifier: "auth-password")
                }
            }
            if let errorMessage = model.errorMessage {
                MPStatusMessage(text: errorMessage, tone: .danger)
            }
            MPButton("Đăng nhập", isLoading: model.isBusy, isDisabled: email.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || password.isEmpty) {
                Task { await model.signIn(email: email, password: password) }
            }
            Spacer()
        }
        .padding(MPMetrics.pagePadding)
        .background(MPColor.canvas.ignoresSafeArea())
        .accessibilityIdentifier("auth-screen")
    }
}
