import Foundation

enum OutboxStoreError: Error, Equatable { case storageUnavailable, keyUnavailable, corruptStore, unsupportedVersion, invalidAccount, invalidPayload, intentConflict, leaseLost, invalidState }
enum OutboxState: String, Codable, Sendable { case pending, sending, applied, retryWait, conflict, rejected, blockedAuth, cancelled }
enum OutboxPayload: Codable, Sendable {
    case transaction(TransactionInputModel)
    case transfer(TransferInputModel)

    var operation: String {
        switch self {
        case .transaction: "transaction.create"
        case .transfer: "transfer.create"
        }
    }

    func frozenForOutbox() -> OutboxPayload {
        switch self {
        case .transaction(var value):
            value.autoLinkTravel = false
            return .transaction(value)
        case .transfer: return self
        }
    }

    func canonicalData() throws -> Data {
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.sortedKeys]
        return try encoder.encode(self)
    }

    func hasSameContent(as other: OutboxPayload) throws -> Bool {
        func normalized(_ payload: OutboxPayload) throws -> OutboxPayload {
            func instant(_ value: String) throws -> String {
                let formatter = ISO8601DateFormatter()
                formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
                var date = formatter.date(from: value)
                if date == nil { formatter.formatOptions = [.withInternetDateTime]; date = formatter.date(from: value) }
                guard let date else { throw OutboxStoreError.invalidPayload }
                return String(date.timeIntervalSince1970)
            }
            switch payload {
            case .transaction(var input): input.occurredAt = try instant(input.occurredAt); return .transaction(input)
            case .transfer(var input): input.occurredAt = try instant(input.occurredAt); return .transfer(input)
            }
        }
        return try normalized(self).canonicalData() == normalized(other).canonicalData()
    }

    func validateNewIntent() throws {
        try validate()
        if case .transaction(let value) = self, value.categoryID == nil {
            throw OutboxStoreError.invalidPayload
        }
    }

    // Existing frozen intents retain their replay semantics. New drafts must use
    // validateNewIntent; tightening replay can strand legacy rows before HTTP.
    func validate() throws {
        let amount: Int64
        let timestamp: String
        switch self {
        case .transaction(let value):
            guard ["income", "expense"].contains(value.type), Self.identifier(value.walletID),
                  value.categoryID.map(Self.identifier) ?? true,
                  value.jarID.map(Self.identifier) ?? true,
                  value.travelEventID.map(Self.identifier) ?? true else { throw OutboxStoreError.invalidPayload }
            amount = value.amount
            timestamp = value.occurredAt
        case .transfer(let value):
            guard Self.identifier(value.sourceWalletID), Self.identifier(value.destinationWalletID),
                  value.sourceWalletID != value.destinationWalletID else { throw OutboxStoreError.invalidPayload }
            amount = value.amount
            timestamp = value.occurredAt
        }
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        var instant = formatter.date(from: timestamp)
        if instant == nil {
            formatter.formatOptions = [.withInternetDateTime]
            instant = formatter.date(from: timestamp)
        }
        guard amount > 0, instant != nil else { throw OutboxStoreError.invalidPayload }
    }

    private static func identifier(_ value: String) -> Bool {
        !value.isEmpty && value == value.trimmingCharacters(in: .whitespacesAndNewlines) && !value.contains("\0")
    }
}
struct OutboxMutation: Codable, Sendable {
    let id: UUID
    let accountID: String
    let idempotencyKey: String
    let payload: OutboxPayload
    let createdAt: Date
    var state: OutboxState
    var attemptCount: Int = 0
    var nextAttemptAt: Date?
    var leaseID: UUID?
    var leaseExpiresAt: Date?
    var lastErrorCode: String?
    var serverResourceIDs: [String] = []
    var supersedesID: UUID? = nil
}

enum OutboxOutcome: Sendable {
    case applied(resourceIDs: [String])
    case retry(at: Date, errorCode: String)
    case conflict(errorCode: String)
    case rejected(errorCode: String)
}

enum OutboxAcknowledgementError: Error, Equatable { case invalidReceipt, invalidResource, invalidLease }

enum OutboxAcknowledgement {
    static func validate(_ mutation: OutboxMutation, resources: [TransactionModel], meta: JSONValue?) throws -> [String] {
        guard mutation.state == .sending, mutation.leaseID != nil, mutation.leaseExpiresAt != nil else {
            throw OutboxAcknowledgementError.invalidLease
        }
        guard case .object(let metadata) = meta, case .object(let receipt) = metadata["mutation"],
              case .string(let key) = receipt["idempotency_key"], key == mutation.idempotencyKey,
              case .string(let operation) = receipt["operation"], operation == mutation.payload.operation,
              case .string("applied") = receipt["status"], case .bool = receipt["replayed"] else {
            throw OutboxAcknowledgementError.invalidReceipt
        }
        guard !resources.isEmpty, resources.allSatisfy({ !$0.id.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty && $0.ownerID == mutation.accountID }),
              Set(resources.map(\.id)).count == resources.count else { throw OutboxAcknowledgementError.invalidResource }
        switch mutation.payload {
        case .transaction(let input):
            guard input.autoLinkTravel == false, resources.count == 1, let row = resources.first,
                  row.walletID == input.walletID, row.type == input.type, row.amount == input.amount,
                  row.categoryID == input.categoryID, row.jarID == input.jarID,
                  row.travelEventID == input.travelEventID, row.includedInReports == input.includedInReports,
                  row.transferID == nil, sameInstant(row.occurredAt, input.occurredAt) else {
                throw OutboxAcknowledgementError.invalidResource
            }
            return [row.id]
        case .transfer(let input):
            guard resources.count == 2,
                  let source = resources.first(where: { $0.walletID == input.sourceWalletID }),
                  let destination = resources.first(where: { $0.walletID == input.destinationWalletID }),
                  source.id != destination.id, source.type == "expense", destination.type == "income",
                  let pair = source.transferID, !pair.isEmpty, destination.transferID == pair,
                  resources.allSatisfy({ $0.amount == input.amount && !$0.includedInReports && sameInstant($0.occurredAt, input.occurredAt) }) else {
                throw OutboxAcknowledgementError.invalidResource
            }
            return [source.id, destination.id]
        }
    }

    private static func sameInstant(_ lhs: String, _ rhs: String) -> Bool {
        func parse(_ value: String) -> Date? {
            let formatter = ISO8601DateFormatter()
            formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
            if let instant = formatter.date(from: value) { return instant }
            formatter.formatOptions = [.withInternetDateTime]
            return formatter.date(from: value)
        }
        guard let first = parse(lhs), let second = parse(rhs) else { return false }
        return first == second
    }
}
