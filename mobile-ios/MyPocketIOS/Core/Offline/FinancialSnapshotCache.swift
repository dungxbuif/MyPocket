import Foundation
import CryptoKit

struct FinancialSnapshot: Codable, Sendable {
    var version = 1
    let ownerID: String
    let fetchedAt: Date
    let transactions: [TransactionModel]
    let wallets: [WalletModel]
    let categories: [CategoryModel]
    let timezone: String
    /// Optional server-authoritative report payloads. Keeping these optional
    /// preserves compatibility with existing version-1 ledger snapshots.
    var monthSummary: MonthSummaryModel? = nil
    var insider: InsiderSummaryModel? = nil
}

actor FinancialSnapshotCache {
    static func makeDefault(serverURL: URL = EnvironmentConfig.apiBaseURL) throws -> FinancialSnapshotCache {
        let support = try FileManager.default.url(for: .applicationSupportDirectory, in: .userDomainMask, appropriateFor: nil, create: true)
        return try FinancialSnapshotCache(directory: support.appendingPathComponent("MyPocket/financial-cache", isDirectory: true),
            serverURL: serverURL, keys: KeychainOutboxKeyProvider(service: "com.mypocket.financial-cache.keys"))
    }
    private let directory: URL
    private let server: String
    private let keys: any OutboxKeyProviding

    init(directory: URL, serverURL: URL, keys: any OutboxKeyProviding) throws {
        guard let scheme = serverURL.scheme, ["https", "http"].contains(scheme), serverURL.host != nil,
              serverURL.user == nil, serverURL.password == nil, serverURL.query == nil, serverURL.fragment == nil else {
            throw OutboxStoreError.storageUnavailable
        }
        self.directory = directory
        self.server = serverURL.absoluteString.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        self.keys = keys
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true,
            attributes: [.protectionKey: FileProtectionType.completeUntilFirstUserAuthentication])
        var protectedDirectory = directory
        var attributes = URLResourceValues(); attributes.isExcludedFromBackup = true
        try protectedDirectory.setResourceValues(attributes)
    }

    func save(_ snapshot: FinancialSnapshot) throws {
        try validate(snapshot, ownerID: snapshot.ownerID)
        let namespace = try namespace(snapshot.ownerID)
        let file = location(namespace)
        let exists = FileManager.default.fileExists(atPath: file.path)
        // A missing/wrong key or corrupt existing snapshot must never trigger overwrite.
        var snapshotToSave = snapshot
        if exists, let existing = try load(ownerID: snapshot.ownerID) {
            // Ledger refreshes must not erase the last verified report payload
            // merely because they did not request reports themselves.
            snapshotToSave.monthSummary = snapshot.monthSummary ?? existing.monthSummary
            snapshotToSave.insider = snapshot.insider ?? existing.insider
        }
        let secret = try key(namespace, create: !exists)
        let plaintext = try JSONEncoder().encode(snapshotToSave)
        let box = try AES.GCM.seal(plaintext, using: secret, authenticating: Data(namespace.utf8))
        guard let sealed = box.combined else { throw OutboxStoreError.storageUnavailable }
        try sealed.write(to: file, options: [.atomic, .completeFileProtectionUntilFirstUserAuthentication])
    }

    func load(ownerID: String) throws -> FinancialSnapshot? {
        let namespace = try namespace(ownerID)
        let file = location(namespace)
        guard FileManager.default.fileExists(atPath: file.path) else { return nil }
        let secret = try key(namespace, create: false)
        let snapshot: FinancialSnapshot
        do {
            let box = try AES.GCM.SealedBox(combined: Data(contentsOf: file))
            let plaintext = try AES.GCM.open(box, using: secret, authenticating: Data(namespace.utf8))
            snapshot = try JSONDecoder().decode(FinancialSnapshot.self, from: plaintext)
        } catch { throw OutboxStoreError.corruptStore }
        try validate(snapshot, ownerID: ownerID)
        return snapshot
    }

    private func validate(_ snapshot: FinancialSnapshot, ownerID: String) throws {
        guard snapshot.version == 1 else { throw OutboxStoreError.unsupportedVersion }
        _ = try namespace(ownerID)
        guard snapshot.ownerID == ownerID, snapshot.fetchedAt.timeIntervalSince1970.isFinite,
              TimeZone(identifier: snapshot.timezone) != nil,
              snapshot.transactions.allSatisfy({ $0.ownerID == ownerID }) else { throw OutboxStoreError.invalidPayload }
    }

    private func namespace(_ ownerID: String) throws -> String {
        guard !ownerID.isEmpty, ownerID == ownerID.trimmingCharacters(in: .whitespacesAndNewlines), !ownerID.contains("\0") else {
            throw OutboxStoreError.invalidAccount
        }
        return SHA256.hash(data: Data(("financial-cache:1|" + server + "|" + ownerID).utf8)).map { String(format: "%02x", $0) }.joined()
    }

    private func location(_ namespace: String) -> URL { directory.appendingPathComponent(namespace + ".aes") }

    private func key(_ namespace: String, create: Bool) throws -> SymmetricKey {
        let bytes = try keys.key(for: namespace, createIfMissing: create)
        guard bytes.count == 32 else { throw OutboxStoreError.keyUnavailable }
        return SymmetricKey(data: bytes)
    }

    static func allowsFallback(_ error: Error) -> Bool {
        guard let api = error as? APIError else { return false }
        switch api {
        case .transport: return true
        case .http(let status): return status == 408 || status == 429 || (500...599).contains(status)
        case .problem(let problem):
            guard let status = problem.status else { return false }
            return status == 408 || status == 429 || (500...599).contains(status)
        default: return false
        }
    }
}
