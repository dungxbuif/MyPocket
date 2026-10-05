import SwiftUI

struct BiometricUnlockView: View {
    @ObservedObject var model: AuthViewModel

    var body: some View {
        VStack(spacing: 16) {
            MPText("Bảo mật phiên", style: .title)
            MPText("Bật Face ID để mở nhanh phiên đang còn hiệu lực.", style: .secondary)
            MPButton(model.biometricEnabled ? "Đã bật Face ID" : "Bật Face ID", style: .secondary) {
                Task { await model.toggleBiometric() }
            }
        }
        .padding(MPMetrics.pagePadding)
    }
}
