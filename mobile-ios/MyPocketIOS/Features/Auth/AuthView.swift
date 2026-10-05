import SwiftUI

struct AuthView: View {
    @ObservedObject var model: AuthViewModel

    var body: some View {
        VStack(spacing: 18) {
            Spacer()
            MPSurfaceCard {
                VStack(alignment: .leading, spacing: 10) {
                    MPText("MyPocket", style: .title)
                    MPText("Đăng nhập để quản lý tài chính của bạn.", style: .secondary)
                }
            }
            if let errorMessage = model.errorMessage {
                MPStatusMessage(text: errorMessage, tone: .danger)
            }
            MPButton("Đăng nhập bằng Google", isLoading: model.isBusy) {
                Task { await model.signIn() }
            }
            Spacer()
        }
        .padding(MPMetrics.pagePadding)
        .background(MPColor.canvas.ignoresSafeArea())
        .accessibilityIdentifier("auth-screen")
    }
}
