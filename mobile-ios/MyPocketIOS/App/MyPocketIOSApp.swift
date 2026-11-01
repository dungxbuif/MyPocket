import SwiftUI
import UIKit
import UserNotifications

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
    func application(_ application: UIApplication, didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil) -> Bool {
        UNUserNotificationCenter.current().delegate = PushNotificationCoordinator.shared
        return true
    }

    func application(_ application: UIApplication, didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
        PushNotificationCoordinator.shared.received(deviceToken: deviceToken)
    }

    func application(_ application: UIApplication, didFailToRegisterForRemoteNotificationsWithError error: Error) {
        PushNotificationCoordinator.shared.registrationFailed(error)
    }

    func application(
        _ application: UIApplication,
        supportedInterfaceOrientationsFor window: UIWindow?
    ) -> UIInterfaceOrientationMask {
        .portrait
    }
}

@MainActor
final class PushNotificationCoordinator: NSObject, UNUserNotificationCenterDelegate {
    static let shared = PushNotificationCoordinator()

    private let registrationIDKey = "mypocket.push.registration.id"
    private let installationIDKey = "mypocket.push.installation.id"
    private var api: FinanceAPI?
    private var pendingToken: String?

    func activate(api: FinanceAPI) async {
#if DEBUG
        if ProcessInfo.processInfo.arguments.contains(where: { $0.hasPrefix("-UITest") }) { return }
#endif
        self.api = api
        let center = UNUserNotificationCenter.current()
        let settings = await center.notificationSettings()
        var authorized = settings.authorizationStatus == .authorized || settings.authorizationStatus == .provisional
        if settings.authorizationStatus == .notDetermined {
            authorized = (try? await center.requestAuthorization(options: [.alert, .badge, .sound])) == true
        }
        guard authorized else { return }
        UIApplication.shared.registerForRemoteNotifications()
        if let pendingToken { await persist(token: pendingToken) }
    }

    func received(deviceToken: Data) {
        let token = deviceToken.map { String(format: "%02x", $0) }.joined()
        pendingToken = token
        Task { await persist(token: token) }
    }

    func registrationFailed(_ error: Error) {
#if DEBUG
        print("[Push] APNs registration failed: \(error.localizedDescription)")
#endif
    }

    func revoke(api: FinanceAPI) async {
        if let id = UserDefaults.standard.string(forKey: registrationIDKey) {
            try? await api.revokePushToken(id: id)
        }
        UserDefaults.standard.removeObject(forKey: registrationIDKey)
        pendingToken = nil
        self.api = nil
        UIApplication.shared.unregisterForRemoteNotifications()
    }

    private func persist(token: String) async {
        guard let api else { return }
        let defaults = UserDefaults.standard
        let installationID: String
        if let existing = defaults.string(forKey: installationIDKey) {
            installationID = existing
        } else {
            installationID = UUID().uuidString
            defaults.set(installationID, forKey: installationIDKey)
        }
#if DEBUG
        let environment = "sandbox"
#else
        let environment = "production"
#endif
        do {
            let registration = try await api.registerPushToken(
                deviceID: installationID,
                deviceName: UIDevice.current.name,
                token: token,
                environment: environment,
                idempotencyKey: "push-\(installationID)-\(environment)"
            )
            defaults.set(registration.id, forKey: registrationIDKey)
        } catch {
#if DEBUG
            print("[Push] Token persistence failed: \(APIError.diagnosticDescription(error))")
#endif
        }
    }

    nonisolated func userNotificationCenter(_ center: UNUserNotificationCenter, willPresent notification: UNNotification) async -> UNNotificationPresentationOptions {
        [.banner, .sound, .badge]
    }
}
