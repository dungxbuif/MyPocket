import CryptoKit
import Foundation
import LocalAuthentication
import Security

struct DeviceSignInMetadata: Codable, Equatable, Sendable {
    let id: String
    let ownerID: String
    let email: String
    let expiresAt: String

    var isValid: Bool { (MPFormat.parseInstant(expiresAt) ?? .distantPast) > Date() }
}

struct DeviceSignInCredential: Codable, Sendable {
    let id: String
    let secret: String
    let expiresAt: String
    let ownerID: String
    let email: String

    var metadata: DeviceSignInMetadata { .init(id: id, ownerID: ownerID, email: email, expiresAt: expiresAt) }
}

struct DeviceEnrollmentResponse: Codable, Sendable {
    let id: String
    let secret: String
    let expiresAt: String
}

protocol DeviceSignInStore: Sendable {
    var metadata: DeviceSignInMetadata? { get }
    var isAvailable: Bool { get }
    func save(_ credential: DeviceSignInCredential) throws
    func unlock() async throws -> DeviceSignInCredential
    func clear()
}

/// Metadata may be read to render the login screen. The secret is only read
/// by an explicit biometric Keychain operation and cannot sync to other devices.
final class KeychainDeviceSignInStore: DeviceSignInStore, @unchecked Sendable {
    private let service: String
    private let defaults: UserDefaults
    private var metadataKey: String { service + ".metadata" }

    init(baseURL: URL = EnvironmentConfig.apiBaseURL, defaults: UserDefaults = .standard) {
        let namespace = SHA256.hash(data: Data(baseURL.absoluteString.utf8)).prefix(12).map { String(format: "%02x", $0) }.joined()
        service = "com.mypocket.ios.device-sign-in." + namespace
        self.defaults = defaults
    }

    var metadata: DeviceSignInMetadata? {
        guard let data = defaults.data(forKey: metadataKey) else { return nil }
        return try? JSONDecoder.mypocket.decode(DeviceSignInMetadata.self, from: data)
    }

    var isAvailable: Bool {
        LAContext().canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: nil)
    }

    func save(_ credential: DeviceSignInCredential) throws {
        var error: Unmanaged<CFError>?
        guard let control = SecAccessControlCreateWithFlags(nil, kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly, .biometryCurrentSet, &error) else {
            throw KeychainError.status(errSecAuthFailed)
        }
        let data = try JSONEncoder.mypocket.encode(credential)
        let metadata = try JSONEncoder.mypocket.encode(credential.metadata)
        clear()
        let item: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: "device-credential",
            kSecAttrAccessControl as String: control,
            kSecAttrSynchronizable as String: false,
            kSecValueData as String: data
        ]
        let status = SecItemAdd(item as CFDictionary, nil)
        guard status == errSecSuccess else { throw KeychainError.status(status) }
        defaults.set(metadata, forKey: metadataKey)
    }

    func unlock() async throws -> DeviceSignInCredential {
        let service = service
        return try await Task.detached(priority: .userInitiated) {
            let context = LAContext()
            context.localizedReason = "Đăng nhập MyPocket bằng Face ID"
            let query: [String: Any] = [
                kSecClass as String: kSecClassGenericPassword,
                kSecAttrService as String: service,
                kSecAttrAccount as String: "device-credential",
                kSecUseAuthenticationContext as String: context,
                kSecReturnData as String: true,
                kSecMatchLimit as String: kSecMatchLimitOne
            ]
            var result: CFTypeRef?
            let status = SecItemCopyMatching(query as CFDictionary, &result)
            guard status == errSecSuccess else { throw KeychainError.status(status) }
            guard let data = result as? Data else { throw KeychainError.status(errSecDecode) }
            return try JSONDecoder.mypocket.decode(DeviceSignInCredential.self, from: data)
        }.value
    }

    func clear() {
        let query: [String: Any] = [kSecClass as String: kSecClassGenericPassword, kSecAttrService as String: service, kSecAttrAccount as String: "device-credential"]
        SecItemDelete(query as CFDictionary)
        defaults.removeObject(forKey: metadataKey)
    }
}
