import SwiftUI

struct BiometricUnlockView: View {
    @ObservedObject var model: AuthViewModel

    var body: some View {
        VStack(spacing: 16) {
            Spacer()
            MPSurfaceCard(style: AuthReferencePresentation.cardStyle) {
                VStack(alignment: .leading, spacing: AuthReferencePresentation.contentSpacing + 2) {
                    MPText("Mở khóa MyPocket", style: .title)
                    MPText("Phiên đăng nhập vẫn còn lưu an toàn. Xác thực Face ID hoặc mã khóa để tiếp tục.", style: .secondary)
                }
            }
            if let errorMessage = model.errorMessage { MPStatusMessage(text: errorMessage, tone: .danger) }
            MPButton("Xác thực để mở khóa") { Task { await model.unlock() } }
            MPButton("Đăng xuất", style: .secondary, isLoading: model.isBusy) { Task { await model.logout() } }
            Spacer()
        }
        .padding(MPMetrics.pagePadding)
        .background(MPColor.canvas.ignoresSafeArea())
    }
}
