import Foundation

struct EntryCapabilitiesModel: Codable, Sendable {
    var aiConfigured: Bool
    var ocrConfigured: Bool
    var filesConfigured: Bool
}

struct EntryDraftModel: Codable, Sendable {
    var type: String
    var amount: Int64
    var walletID: String
    var categoryID: String?
    var jarID: String?
    var occurredAt: String
    var note: String
    var includedInReports: Bool
}

struct EntryProposalModel: Codable, Identifiable, Sendable {
    let id: String
    var processID: String
    var version: Int
    var status: String
    var draft: EntryDraftModel
    var questions: [String]
    var transactionID: String?
}

struct EntryProcessModel: Codable, Identifiable, Sendable {
    let id: String
    var proposals: [EntryProposalModel]
    var processing: Bool
    var reply: String?
    var error: String?
}

struct AdvisorPartModel: Decodable, Sendable {
    var type: String
    var text: String?
    var data: JSONValue?

    var entryBatch: AdvisorEntryBatchModel? {
        guard type == "entry_batch" else { return nil }
        return AdvisorEntryBatchModel(json: data)
    }
}

struct AdvisorEntryBatchModel: Sendable {
    var taskID: String
    var batchID: String
    var version: Int64
    var itemIDs: [String]
    var itemCount: Int
    var summary: String

    init?(json: JSONValue?) {
        guard case .object(let object) = json,
              case .string(let taskID) = object["task_id"], !taskID.isEmpty,
              case .string(let batchID) = object["batch_id"], !batchID.isEmpty,
              case .number(let version) = object["version"], let exactVersion = Int64(exactly: version), exactVersion > 0,
              case .array(let rawIDs) = object["item_ids"],
              case .number(let count) = object["item_count"], let exactCount = Int(exactly: count), exactCount >= 0,
              case .string(let summary) = object["summary"] else { return nil }
        var ids: [String] = []
        for rawID in rawIDs {
            guard case .string(let id) = rawID else { return nil }
            ids.append(id)
        }
        guard ids.count == exactCount else { return nil }
        self.taskID = taskID
        self.batchID = batchID
        self.version = exactVersion
        self.itemIDs = ids
        self.itemCount = exactCount
        self.summary = summary
    }
}

struct AdvisorEntryTaskModel: Decodable, Sendable {
    var id: String
    var ownerID: String
    var conversationID: String
    var generation: Int64
    var clientRequestID: String
    var sessionID: String
    var status: String
    var stage: String?
    var batchID: String?
    var batchVersion: Int64?
    var errorCode: String?
    var createdAt: String
    var updatedAt: String
    var completedAt: String?
}

struct AdvisorMessageModel: Decodable, Identifiable, Sendable {
    let id: String
    var conversationID: String
    var seq: Int
    var role: String
    var parts: [AdvisorPartModel]
    var createdAt: String
}

struct AdvisorHistoryMetaModel: Sendable, Equatable {
    var conversationID: String
    var generation: Int64
    var lastUserMessageAt: String?
    var expiresAt: String?
    var serverTime: String
    var activeRun: AdvisorActiveRunModel?

    init?(json: JSONValue?) {
        guard case .object(let object) = json,
              case .string(let conversationID) = object["conversation_id"], !conversationID.isEmpty,
              case .number(let generation) = object["generation"], let exactGeneration = Int64(exactly: generation),
              case .string(let serverTime) = object["server_time"], !serverTime.isEmpty else { return nil }
        self.conversationID = conversationID
        self.generation = exactGeneration
        if case .string(let value) = object["last_user_message_at"] { self.lastUserMessageAt = value } else { self.lastUserMessageAt = nil }
        if case .string(let value) = object["expires_at"] { self.expiresAt = value } else { self.expiresAt = nil }
        self.serverTime = serverTime
        self.activeRun = AdvisorActiveRunModel(json: object["active_run"])
    }
}

struct AdvisorActiveRunModel: Sendable, Equatable {
    var id: String
    var status: String
    var createdAt: String?
    var deadlineAt: String?

    init?(json: JSONValue?) {
        guard case .object(let object) = json,
              case .string(let id) = object["id"], !id.isEmpty,
              case .string(let status) = object["status"], !status.isEmpty else { return nil }
        self.id = id
        self.status = status
        if case .string(let value) = object["created_at"] { self.createdAt = value } else { self.createdAt = nil }
        if case .string(let value) = object["deadline_at"] { self.deadlineAt = value } else { self.deadlineAt = nil }
    }
}

struct AdvisorHistoryPageModel: Sendable {
    var messages: [AdvisorMessageModel]
    var meta: AdvisorHistoryMetaModel?
}

struct AdvisorCapabilitiesModel: Codable, Sendable {
    var enabled: Bool
    var write: Bool
    var partsVersion: Int
    var tools: [String]
    var runTimeoutSeconds: Int? = nil
    var groundingVersion: Int? = nil
    var entryIntegration: Bool? = nil
    var supportedPartKinds: [String]? = nil
    var actionKinds: [String]? = nil
    var attachmentFormats: [String]? = nil
    var attachmentMaxCount: Int? = nil
    var attachmentMaxBytes: Int64? = nil
}

struct AdvisorOverviewModel: Decodable, Sendable {
    var viewKind: String
    var view: JSONValue?
    var source: JSONValue?
    var status: String
}

struct AdvisorSubmitModel: Codable, Sendable {
    var runID: String
    var conversationID: String
    var text: String
    var status: String
}

struct AdvisorAsyncSubmitModel: Codable, Sendable {
    var runID: String
    var conversationID: String
    var status: String
}

struct AdvisorResetModel: Codable, Sendable {
    var conversationID: String
    var generation: Int64
    var serverTime: String
    var replayed: Bool
}

struct AdvisorRunModel: Codable, Sendable {
    var id: String
    var conversationID: String
    var status: String
    var errorCode: String?
    var ownerID: String? = nil
    var serverTime: String? = nil
    var deadlineAt: String? = nil
    var retryAfterMs: Int? = nil
    var createdAt: String? = nil
}
