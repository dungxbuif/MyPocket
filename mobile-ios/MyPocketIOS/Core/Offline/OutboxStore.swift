import Foundation
import CryptoKit

actor OutboxStore {
    private let database: OutboxDatabase
    private let serverNamespace: String
    private let keys: any OutboxKeyProviding

    init(directory: URL, serverURL: URL, keys: any OutboxKeyProviding) throws {
        guard let scheme = serverURL.scheme, ["http", "https"].contains(scheme), serverURL.host != nil,
              serverURL.user == nil, serverURL.password == nil, serverURL.query == nil, serverURL.fragment == nil else {
            throw OutboxStoreError.storageUnavailable
        }
        let base = serverURL.absoluteString.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        let namespace = Self.hash(Data(base.utf8))
        var location = directory.appendingPathComponent(namespace, isDirectory: true)
        try FileManager.default.createDirectory(at: location, withIntermediateDirectories: true,
                                              attributes: [.protectionKey: FileProtectionType.completeUntilFirstUserAuthentication])
        var attributes = URLResourceValues()
        attributes.isExcludedFromBackup = true
        try location.setResourceValues(attributes)
        database = try OutboxDatabase(url: location.appendingPathComponent("outbox.sqlite"))
        serverNamespace = namespace
        self.keys = keys
    }

    func enqueue(accountID: String, payload: OutboxPayload, key: UUID, now: Date = Date()) throws -> OutboxMutation {
        let namespace = try namespace(accountID: accountID)
        let payload = payload.frozenForOutbox()
        try payload.validateNewIntent()
        let intentHash = Self.hash(Data((payload.operation + "|" + key.uuidString).utf8))
        return try database.transaction {
            let rows = try database.rows(namespace: namespace)
            let secret = try encryptionKey(namespace: namespace, createIfMissing: rows.isEmpty)
            // Validate all existing records before adding another. A lost/wrong key
            // must never replace a previously persisted financial intent.
            let existing = try rows.map { try open($0, namespace: namespace, accountID: accountID, key: secret) }
            if let index = rows.firstIndex(where: { $0.intentHash == intentHash }) {
                guard try existing[index].payload.canonicalData() == payload.canonicalData() else { throw OutboxStoreError.intentConflict }
                return existing[index]
            }
            let mutation = OutboxMutation(id: UUID(), accountID: accountID, idempotencyKey: key.uuidString, payload: payload, createdAt: now, state: .pending)
            let plaintext = try JSONEncoder().encode(mutation)
            let box = try AES.GCM.seal(plaintext, using: secret, authenticating: associatedData(namespace: namespace, id: mutation.id.uuidString))
            guard let sealed = box.combined else { throw OutboxStoreError.storageUnavailable }
            try database.insert(id: mutation.id.uuidString, namespace: namespace, intentHash: intentHash, sealed: sealed)
            return mutation
        }
    }

    func items(accountID: String) throws -> [OutboxMutation] {
        let namespace = try namespace(accountID: accountID)
        let rows = try database.rows(namespace: namespace)
        guard !rows.isEmpty else { return [] }
        let key = try encryptionKey(namespace: namespace, createIfMissing: false)
        return try rows.map { try open($0, namespace: namespace, accountID: accountID, key: key) }
    }

    func pendingCount(accountID: String) throws -> Int {
        try items(accountID: accountID).filter { $0.state != .applied && $0.state != .cancelled }.count
    }

    func retryNow(accountID: String, id: UUID, now: Date = Date()) throws {
        guard now.timeIntervalSince1970.isFinite else { throw OutboxStoreError.invalidPayload }
        try mutate(accountID: accountID) { items in
            guard let index = items.firstIndex(where: { $0.id == id }), items[index].state == .retryWait else { throw OutboxStoreError.invalidState }
            items[index].nextAttemptAt = now
        }
    }
    func discard(accountID: String, id: UUID) throws {
        try mutate(accountID: accountID) { items in
            guard let index = items.firstIndex(where: { $0.id == id }),
                  [.pending, .retryWait, .conflict, .rejected, .blockedAuth, .cancelled].contains(items[index].state) else { throw OutboxStoreError.invalidState }
            items[index].state = .cancelled
            items[index].nextAttemptAt = nil
            items[index].leaseID = nil
            items[index].leaseExpiresAt = nil
        }
    }
    func replaceRejected(accountID: String, id: UUID, payload: OutboxPayload, key: UUID, now: Date = Date()) throws -> OutboxMutation {
        let namespace = try namespace(accountID: accountID)
        let payload = payload.frozenForOutbox()
        try payload.validateNewIntent()
        guard now.timeIntervalSince1970.isFinite else { throw OutboxStoreError.invalidPayload }
        let intentHash = Self.hash(Data((payload.operation + "|" + key.uuidString).utf8))
        return try database.transaction {
            let rows = try database.rows(namespace: namespace)
            guard !rows.isEmpty else { throw OutboxStoreError.invalidState }
            let secret = try encryptionKey(namespace: namespace, createIfMissing: false)
            var items = try rows.map { try open($0, namespace: namespace, accountID: accountID, key: secret) }
            guard let parent = items.firstIndex(where: { $0.id == id }) else { throw OutboxStoreError.invalidState }
            if let index = rows.firstIndex(where: { $0.intentHash == intentHash }) {
                guard items[index].supersedesID == id, try items[index].payload.canonicalData() == payload.canonicalData() else {
                    throw OutboxStoreError.intentConflict
                }
                return items[index]
            }
            guard [.conflict, .rejected].contains(items[parent].state) else { throw OutboxStoreError.invalidState }
            guard items[parent].payload.operation == payload.operation,
                  try !items[parent].payload.hasSameContent(as: payload) else { throw OutboxStoreError.invalidPayload }
            items[parent].state = .cancelled
            items[parent].nextAttemptAt = nil
            items[parent].leaseID = nil
            items[parent].leaseExpiresAt = nil
            var child = OutboxMutation(id: UUID(), accountID: accountID, idempotencyKey: key.uuidString, payload: payload, createdAt: now, state: .pending)
            child.supersedesID = id
            func seal(_ mutation: OutboxMutation) throws -> Data {
                let box = try AES.GCM.seal(JSONEncoder().encode(mutation), using: secret,
                    authenticating: associatedData(namespace: namespace, id: mutation.id.uuidString))
                guard let sealed = box.combined else { throw OutboxStoreError.storageUnavailable }
                return sealed
            }
            try database.update(id: rows[parent].id, namespace: namespace, sealed: seal(items[parent]))
            try database.insert(id: child.id.uuidString, namespace: namespace, intentHash: intentHash, sealed: seal(child))
            return child
        }
    }

    func leaseNext(accountID: String, now: Date = Date(), leaseDuration: TimeInterval = 60) throws -> OutboxMutation? {
        guard leaseDuration.isFinite, leaseDuration > 0, now.timeIntervalSince1970.isFinite else { throw OutboxStoreError.invalidPayload }
        return try mutate(accountID: accountID) { items in
            guard let index = items.firstIndex(where: { [.pending, .sending, .retryWait, .blockedAuth].contains($0.state) }) else { return nil }
            if items[index].state == .blockedAuth { return nil }
            if items[index].state == .sending {
                guard let expires = items[index].leaseExpiresAt else { throw OutboxStoreError.corruptStore }
                if expires > now { return nil }
            }
            if items[index].state == .retryWait {
                guard let retryAt = items[index].nextAttemptAt else { throw OutboxStoreError.corruptStore }
                if retryAt > now { return nil }
            }
            guard items[index].attemptCount < Int.max else { throw OutboxStoreError.corruptStore }
            items[index].state = .sending
            items[index].attemptCount += 1
            items[index].leaseID = UUID()
            items[index].leaseExpiresAt = now.addingTimeInterval(leaseDuration)
            items[index].nextAttemptAt = nil
            return items[index]
        }
    }

    func markOutcome(accountID: String, id: UUID, leaseID: UUID, outcome: OutboxOutcome) throws {
        try mutate(accountID: accountID) { items in
            guard let index = items.firstIndex(where: { $0.id == id }), items[index].state == .sending,
                  items[index].leaseID == leaseID else { throw OutboxStoreError.leaseLost }
            switch outcome {
            case .applied(let ids):
                guard !ids.isEmpty, ids.allSatisfy({ !$0.isEmpty && $0 == $0.trimmingCharacters(in: .whitespacesAndNewlines) }) else { throw OutboxStoreError.invalidPayload }
                items[index].state = .applied
                items[index].serverResourceIDs = ids
                items[index].lastErrorCode = nil
            case .retry(let at, let code):
                guard at.timeIntervalSince1970.isFinite else { throw OutboxStoreError.invalidPayload }
                try Self.validateErrorCode(code)
                items[index].state = .retryWait
                items[index].nextAttemptAt = at
                items[index].lastErrorCode = code
            case .conflict(let code):
                try Self.validateErrorCode(code)
                items[index].state = .conflict
                items[index].lastErrorCode = code
            case .rejected(let code):
                try Self.validateErrorCode(code)
                items[index].state = .rejected
                items[index].lastErrorCode = code
            }
            items[index].leaseID = nil
            items[index].leaseExpiresAt = nil
        }
    }

    func resumeLeases(accountID: String, now: Date = Date()) throws {
        try mutate(accountID: accountID) { items in
            for index in items.indices where items[index].state == .sending {
                guard let expires = items[index].leaseExpiresAt else { throw OutboxStoreError.corruptStore }
                if expires <= now {
                    items[index].state = .pending
                    items[index].leaseID = nil
                    items[index].leaseExpiresAt = nil
                }
            }
        }
    }

    func blockForAuth(accountID: String) throws {
        try mutate(accountID: accountID) { items in
            for index in items.indices where [.pending, .sending, .retryWait].contains(items[index].state) {
                items[index].state = .blockedAuth
                items[index].leaseID = nil
                items[index].leaseExpiresAt = nil
                items[index].nextAttemptAt = nil
            }
        }
    }

    func resumeAuth(accountID: String) throws {
        try mutate(accountID: accountID) { items in
            for index in items.indices where items[index].state == .blockedAuth { items[index].state = .pending }
        }
    }

    /// Encrypt and commit every changed record before returning any lease/result.
    private func mutate<T>(accountID: String, body: (inout [OutboxMutation]) throws -> T) throws -> T {
        let namespace = try namespace(accountID: accountID)
        return try database.transaction {
            let rows = try database.rows(namespace: namespace)
            if rows.isEmpty {
                var empty: [OutboxMutation] = []
                return try body(&empty)
            }
            let secret = try encryptionKey(namespace: namespace, createIfMissing: false)
            var items = try rows.map { try open($0, namespace: namespace, accountID: accountID, key: secret) }
            let encoder = JSONEncoder()
            encoder.outputFormatting = [.sortedKeys]
            let before = try items.map { try encoder.encode($0) }
            let result = try body(&items)
            for index in items.indices {
                let data = try encoder.encode(items[index])
                if before[index] != data {
                    let box = try AES.GCM.seal(data, using: secret, authenticating: associatedData(namespace: namespace, id: rows[index].id))
                    guard let sealed = box.combined else { throw OutboxStoreError.storageUnavailable }
                    try database.update(id: rows[index].id, namespace: namespace, sealed: sealed)
                }
            }
            return result
        }
    }

    private static func validateErrorCode(_ code: String) throws {
        guard !code.isEmpty, code.utf8.count <= 128, code.unicodeScalars.allSatisfy({
            (65...90).contains($0.value) || (97...122).contains($0.value) || (48...57).contains($0.value) || $0.value == 95
        }) else { throw OutboxStoreError.invalidPayload }
    }

    private func namespace(accountID: String) throws -> String {
        guard !accountID.isEmpty, accountID == accountID.trimmingCharacters(in: .whitespacesAndNewlines), !accountID.contains("\0") else { throw OutboxStoreError.invalidAccount }
        return Self.hash(Data((serverNamespace + "|" + accountID).utf8))
    }

    private func encryptionKey(namespace: String, createIfMissing: Bool) throws -> SymmetricKey {
        let bytes = try keys.key(for: namespace, createIfMissing: createIfMissing)
        guard bytes.count == 32 else { throw OutboxStoreError.keyUnavailable }
        return SymmetricKey(data: bytes)
    }

    private func open(_ row: OutboxDatabase.Row, namespace: String, accountID: String, key: SymmetricKey) throws -> OutboxMutation {
        do {
            let box = try AES.GCM.SealedBox(combined: row.sealed)
            let data = try AES.GCM.open(box, using: key, authenticating: associatedData(namespace: namespace, id: row.id))
            let mutation = try JSONDecoder().decode(OutboxMutation.self, from: data)
            guard mutation.id.uuidString == row.id, mutation.accountID == accountID,
                  Self.hash(Data((mutation.payload.operation + "|" + mutation.idempotencyKey).utf8)) == row.intentHash else { throw OutboxStoreError.corruptStore }
            return mutation
        } catch { throw OutboxStoreError.corruptStore }
    }

    private func associatedData(namespace: String, id: String) -> Data { Data(("outbox:1|" + namespace + "|" + id).utf8) }
    private static func hash(_ data: Data) -> String { SHA256.hash(data: data).map { String(format: "%02x", $0) }.joined() }
}
