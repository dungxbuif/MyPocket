import SwiftUI

struct AppRootView: View {
    var body: some View {
        ZStack {
            MPColor.canvas
                .ignoresSafeArea()
            MPText("MyPocket", style: .title)
        }
        .accessibilityIdentifier("app-root")
    }
}

#Preview {
    AppRootView()
}
