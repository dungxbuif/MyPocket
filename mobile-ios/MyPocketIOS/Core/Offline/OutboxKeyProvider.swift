import Foundation
import CryptoKit
import Security

protocol OutboxKeyProviding: Sendable {
    func key(for namespace: String, createIfMissing: Bool) throws -> Data
}
struct KeychainOutboxKeyProvider: OutboxKeyProviding {
    let service: String
    func key(for namespace: String, createIfMissing: Bool) throws -> Data {
        let identity: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: namespace,
            kSecAttrSynchronizable as String: false
        ]
        var query = identity
        query[kSecReturnData as String] = true
        query[kSecMatchLimit as String] = kSecMatchLimitOne
        var result: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        if status == errSecSuccess {
            guard let bytes = result as? Data, bytes.count == 32 else { throw OutboxStoreError.keyUnavailable }
            return bytes
        }
        guard status == errSecItemNotFound, createIfMissing else { throw OutboxStoreError.keyUnavailable }
        let bytes = SymmetricKey(size: .bits256).withUnsafeBytes { Data($0) }
        var insertion = identity
        insertion[kSecAttrAccessible as String] = kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly
        insertion[kSecValueData as String] = bytes
        let added = SecItemAdd(insertion as CFDictionary, nil)
        if added == errSecDuplicateItem { return try key(for: namespace, createIfMissing: false) }
        guard added == errSecSuccess else { throw OutboxStoreError.keyUnavailable }
        return bytes
    }
}
