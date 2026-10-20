import SwiftUI

@main
struct MyPocketIOSApp: App {
    @UIApplicationDelegateAdaptor(PortraitOnlyApplicationDelegate.self)
    private var applicationDelegate

    var body: some Scene {
        WindowGroup {
            AppRootView()
        }
    }
}

/// Keeps the finance app in portrait so compact cards, bottom navigation and
/// keyboard-driven forms cannot be re-laid out into an unsupported landscape
/// presentation. This is intentionally app-wide; individual routes do not
/// opt into a different orientation.
final class PortraitOnlyApplicationDelegate: NSObject, UIApplicationDelegate {
    func application(
        _ application: UIApplication,
        supportedInterfaceOrientationsFor window: UIWindow?
    ) -> UIInterfaceOrientationMask {
        .portrait
    }
}
