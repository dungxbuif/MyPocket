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

private struct CachedReportAnalysis: Codable {
    let version: Int
    let ownerID: String
    let scopeKey: String
    let freshness: String
    let id: String
    let snapshotID: String?
    let status: String
    let revision: Int64?
    let summary: String?
    let observations: [ReportInsightModel]
    let suggestions: [ReportInsightModel]
    let article: String?
    let retainedUntil: String?

    init(ownerID: String, scopeKey: String, freshness: String, analysis: ReportAnalysisModel) {
        version = 1; self.ownerID = ownerID; self.scopeKey = scopeKey; self.freshness = freshness
        id = analysis.id; snapshotID = analysis.snapshotID; status = analysis.status; revision = analysis.revision
        summary = analysis.summary; observations = analysis.observations; suggestions = analysis.suggestions
        article = analysis.article; retainedUntil = analysis.retainedUntil
    }

    var model: ReportAnalysisModel {
        ReportAnalysisModel(id: id, snapshotID: snapshotID, status: status, revision: revision, summary: summary, observations: observations, suggestions: suggestions, article: article, retainedUntil: retainedUntil)
    }
}

actor ReportAnalysisCache {
    static func makeDefault(serverURL: URL = EnvironmentConfig.apiBaseURL) throws -> ReportAnalysisCache {
        let support = try FileManager.default.url(for: .applicationSupportDirectory, in: .userDomainMask, appropriateFor: nil, create: true)
        return try ReportAnalysisCache(directory: support.appendingPathComponent("MyPocket/report-analysis-cache", isDirectory: true), serverURL: serverURL, keys: KeychainOutboxKeyProvider(service: "com.mypocket.report-analysis-cache.keys"))
    }

    private let directory: URL
    private let server: String
    private let keys: any OutboxKeyProviding

    init(directory: URL, serverURL: URL, keys: any OutboxKeyProviding) throws {
        guard let scheme = serverURL.scheme, ["https", "http"].contains(scheme), serverURL.host != nil else { throw OutboxStoreError.storageUnavailable }
        self.directory = directory; server = serverURL.absoluteString.trimmingCharacters(in: CharacterSet(charactersIn: "/")); self.keys = keys
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true, attributes: [.protectionKey: FileProtectionType.completeUntilFirstUserAuthentication])
        var protectedDirectory = directory; var attributes = URLResourceValues(); attributes.isExcludedFromBackup = true; try protectedDirectory.setResourceValues(attributes)
    }

    func save(ownerID: String, scopeKey: String, freshness: String, analysis: ReportAnalysisModel) throws {
        let namespace = try namespace(ownerID: ownerID, scopeKey: scopeKey)
        let secret = SymmetricKey(data: try keys.key(for: namespace, createIfMissing: true))
        let payload = try JSONEncoder().encode(CachedReportAnalysis(ownerID: ownerID, scopeKey: scopeKey, freshness: freshness, analysis: analysis))
        guard let sealed = try AES.GCM.seal(payload, using: secret, authenticating: Data(namespace.utf8)).combined else { throw OutboxStoreError.storageUnavailable }
        try sealed.write(to: directory.appendingPathComponent(namespace + ".aes"), options: [.atomic, .completeFileProtectionUntilFirstUserAuthentication])
    }

    func load(ownerID: String, scopeKey: String, now: Date = Date()) throws -> (ReportAnalysisModel, String)? {
        let namespace = try namespace(ownerID: ownerID, scopeKey: scopeKey)
        let file = directory.appendingPathComponent(namespace + ".aes")
        guard FileManager.default.fileExists(atPath: file.path) else { return nil }
        let secret = SymmetricKey(data: try keys.key(for: namespace, createIfMissing: false))
        let box = try AES.GCM.SealedBox(combined: Data(contentsOf: file))
        let payload = try AES.GCM.open(box, using: secret, authenticating: Data(namespace.utf8))
        let cached = try JSONDecoder().decode(CachedReportAnalysis.self, from: payload)
        guard cached.version == 1, cached.ownerID == ownerID, cached.scopeKey == scopeKey else { throw OutboxStoreError.invalidPayload }
        if let retained = cached.retainedUntil {
            guard let expiry = Self.parseISO8601(retained), expiry > now else { return nil }
        }
        return (cached.model, cached.freshness)
    }

    private static func parseISO8601(_ value: String) -> Date? {
        let fractional = ISO8601DateFormatter()
        fractional.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let date = fractional.date(from: value) { return date }
        return ISO8601DateFormatter().date(from: value)
    }

    private func namespace(ownerID: String, scopeKey: String) throws -> String {
        guard !ownerID.isEmpty, !scopeKey.isEmpty, !ownerID.contains("\0"), !scopeKey.contains("\0") else { throw OutboxStoreError.invalidAccount }
        return SHA256.hash(data: Data(("report-analysis-cache:1|" + server + "|" + ownerID + "|" + scopeKey).utf8)).map { String(format: "%02x", $0) }.joined()
    }
}
