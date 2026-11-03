import Foundation
import CryptoKit

/// Encrypted, owner-scoped store of raw successful GET responses for screens that must paint
/// from local data first (reports, budgets, jars, months, portfolio). It never answers on the
/// network's behalf: callers show it immediately and always refresh in the background.
actor MPReadThroughCache {
    /// Server-derived reads that are safe to show as "last known" until refreshed.
    static func allows(path: String) -> Bool {
        let prefixes = ["/reports/monthly?", "/reports/savings", "/reports/insider", "/budgets", "/budget-series",
                        "/jars", "/months/", "/portfolio/summary", "/portfolio/assets", "/recurring", "/travel/events"]
        return prefixes.contains { path.hasPrefix($0) }
    }

    static func makeDefault(serverURL: URL = EnvironmentConfig.apiBaseURL) -> MPReadThroughCache? {
        guard let support = try? FileManager.default.url(for: .applicationSupportDirectory, in: .userDomainMask, appropriateFor: nil, create: true) else { return nil }
        return try? MPReadThroughCache(directory: support.appendingPathComponent("MyPocket/read-cache", isDirectory: true),
            serverURL: serverURL, keys: KeychainOutboxKeyProvider(service: "com.mypocket.read-cache.keys"))
    }

    private let directory: URL
    private let server: String
    private let keys: any OutboxKeyProviding

    init(directory: URL, serverURL: URL, keys: any OutboxKeyProviding) throws {
        guard let scheme = serverURL.scheme, ["https", "http"].contains(scheme), serverURL.host != nil else { throw OutboxStoreError.storageUnavailable }
        self.directory = directory
        self.server = serverURL.absoluteString.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        self.keys = keys
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true,
            attributes: [.protectionKey: FileProtectionType.completeUntilFirstUserAuthentication])
        var protectedDirectory = directory
        var attributes = URLResourceValues(); attributes.isExcludedFromBackup = true
        try protectedDirectory.setResourceValues(attributes)
    }

    func save(_ data: Data, ownerID: String, key: String) {
        guard !ownerID.isEmpty, let secret = try? ownerKey(ownerID, create: true) else { return }
        let file = location(ownerID: ownerID, key: key)
        guard let sealed = try? AES.GCM.seal(data, using: secret, authenticating: Data(file.lastPathComponent.utf8)).combined else { return }
        try? sealed.write(to: file, options: [.atomic, .completeFileProtectionUntilFirstUserAuthentication])
    }

    func load(ownerID: String, key: String) -> Data? {
        guard !ownerID.isEmpty, let secret = try? ownerKey(ownerID, create: false) else { return nil }
        let file = location(ownerID: ownerID, key: key)
        guard let combined = try? Data(contentsOf: file), let box = try? AES.GCM.SealedBox(combined: combined) else { return nil }
        return try? AES.GCM.open(box, using: secret, authenticating: Data(file.lastPathComponent.utf8))
    }

    private func digest(_ text: String) -> String {
        SHA256.hash(data: Data(text.utf8)).map { String(format: "%02x", $0) }.joined()
    }

    private func ownerKey(_ ownerID: String, create: Bool) throws -> SymmetricKey {
        let bytes = try keys.key(for: digest("read-cache:1|" + server + "|" + ownerID), createIfMissing: create)
        guard bytes.count == 32 else { throw OutboxStoreError.keyUnavailable }
        return SymmetricKey(data: bytes)
    }

    private func location(ownerID: String, key: String) -> URL {
        directory.appendingPathComponent(digest("read-cache:1|" + server + "|" + ownerID + "|" + key) + ".aes")
    }
}
