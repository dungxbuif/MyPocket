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

struct EntryAttachmentDiagnosticModel: Codable, Identifiable, Sendable {
    let position: Int
    let filename: String
    let stage: String
    let status: String
    let errorCode: String?

    var id: String { "\(position)-\(stage)" }
}

struct EntryProcessModel: Codable, Identifiable, Sendable {
    let id: String
    var proposals: [EntryProposalModel]
    var processing: Bool
    var reply: String?
    var error: String?
    var attachmentDiagnostics: [EntryAttachmentDiagnosticModel] = []

    private enum CodingKeys: String, CodingKey {
        case id, proposals, processing, reply, error, attachmentDiagnostics
    }

    init(
        id: String,
        proposals: [EntryProposalModel],
        processing: Bool,
        reply: String?,
        error: String?,
        attachmentDiagnostics: [EntryAttachmentDiagnosticModel] = []
    ) {
        self.id = id
        self.proposals = proposals
        self.processing = processing
        self.reply = reply
        self.error = error
        self.attachmentDiagnostics = attachmentDiagnostics
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        id = try container.decode(String.self, forKey: .id)
        proposals = try container.decode([EntryProposalModel].self, forKey: .proposals)
        processing = try container.decode(Bool.self, forKey: .processing)
        reply = try container.decodeIfPresent(String.self, forKey: .reply)
        error = try container.decodeIfPresent(String.self, forKey: .error)
        attachmentDiagnostics = try container.decodeIfPresent([EntryAttachmentDiagnosticModel].self, forKey: .attachmentDiagnostics) ?? []
    }
}

struct AdvisorPartModel: Decodable, Sendable {
    var type: String
    var text: String?
    var data: JSONValue?

    var entryBatch: AdvisorEntryBatchModel? {
        guard type == "entry_batch" else { return nil }
        return AdvisorEntryBatchModel(json: data)
    }

    var entryBatchEnvelope: AdvisorEntryBatchEnvelopeModel? {
        guard type == "entry_batch" else { return nil }
        return AdvisorEntryBatchEnvelopeModel(json: data)
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

    private enum CodingKeys: String, CodingKey {
        case task
        case id, ownerID, conversationID, generation, clientRequestID, sessionID, status, stage, batchID, batchVersion, errorCode, createdAt, updatedAt, completedAt
    }

    private struct Payload: Decodable {
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

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        let payload: Payload
        if container.contains(.task) {
            payload = try container.decode(Payload.self, forKey: .task)
        } else {
            payload = try Payload(from: decoder)
        }
        id = payload.id
        ownerID = payload.ownerID
        conversationID = payload.conversationID
        generation = payload.generation
        clientRequestID = payload.clientRequestID
        sessionID = payload.sessionID
        status = payload.status
        stage = payload.stage
        batchID = payload.batchID
        batchVersion = payload.batchVersion
        errorCode = payload.errorCode
        createdAt = payload.createdAt
        updatedAt = payload.updatedAt
        completedAt = payload.completedAt
    }
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

/// Additive native contract for a typed assistant entry batch. Older advisor
/// responses simply omit this payload and remain renderable as read-only chat.
struct AdvisorEntryBatchItemModel: Codable, Identifiable, Sendable {
    var id: String
    var version: Int
    var status: String
    var draft: EntryDraftModel?
    var proposalID: String?
    var error: String?
    var sourceLabel: String?
}

struct AdvisorEntryBatchEnvelopeModel: Codable, Sendable {
    var taskID: String
    var batchID: String
    var version: Int64
    var items: [AdvisorEntryBatchItemModel]
    var summary: String?

    init(taskID: String, batchID: String, version: Int64, items: [AdvisorEntryBatchItemModel], summary: String? = nil) {
        self.taskID = taskID
        self.batchID = batchID
        self.version = version
        self.items = items
        self.summary = summary
    }

    init?(json: JSONValue?) {
        guard case .object(let object) = json,
              case .string(let taskID) = object["task_id"], !taskID.isEmpty,
              case .string(let batchID) = object["batch_id"], !batchID.isEmpty,
              case .number(let version) = object["version"], let exactVersion = Int64(exactly: version),
              case .array(let rawItems) = object["items"] else { return nil }
        var decodedItems: [AdvisorEntryBatchItemModel] = []
        for rawItem in rawItems {
            guard case .object(let item) = rawItem,
                  case .string(let id) = item["id"],
                  case .string(let status) = item["status"] else { return nil }
            var draft: EntryDraftModel?
            if case .object(let rawDraft) = item["draft"],
               case .string(let type) = rawDraft["type"],
               case .number(let amount) = rawDraft["amount"],
               let exactAmount = Int64(exactly: amount),
               case .string(let walletID) = rawDraft["wallet_id"] {
                let categoryID = rawDraft["category_id"].flatMap { value in if case .string(let value) = value { return value }; return nil }
                let jarID = rawDraft["jar_id"].flatMap { value in if case .string(let value) = value { return value }; return nil }
                let occurredAt = rawDraft["occurred_at"].flatMap { value in if case .string(let value) = value { return value }; return nil } ?? ""
                let note = rawDraft["note"].flatMap { value in if case .string(let value) = value { return value }; return nil } ?? ""
                let included = rawDraft["included_in_reports"].flatMap { value in if case .bool(let value) = value { return value }; return nil } ?? true
                draft = EntryDraftModel(type: type, amount: exactAmount, walletID: walletID, categoryID: categoryID, jarID: jarID, occurredAt: occurredAt, note: note, includedInReports: included)
            }
            let proposalID = item["proposal_id"].flatMap { value in if case .string(let value) = value { return value }; return nil }
            let error = item["error"].flatMap { value in if case .string(let value) = value { return value }; return nil }
            let sourceLabel = item["source_label"].flatMap { value in if case .string(let value) = value { return value }; return nil }
            let version = item["version"].flatMap { value in if case .number(let value) = value { return Int(exactly: value) }; return nil } ?? 1
            decodedItems.append(AdvisorEntryBatchItemModel(id: id, version: version, status: status, draft: draft, proposalID: proposalID, error: error, sourceLabel: sourceLabel))
        }
        self.init(taskID: taskID, batchID: batchID, version: exactVersion, items: decodedItems, summary: object["summary"].flatMap { value in if case .string(let value) = value { return value }; return nil })
    }
}

enum AssistantEntryRowState: String, Codable, CaseIterable, Sendable {
    case pending
    case invalid
    case saving
    case applied
    case rejected
    case conflict

    var label: String {
        switch self {
        case .pending: return "Chờ duyệt"
        case .invalid: return "Cần bổ sung"
        case .saving: return "Đang lưu"
        case .applied: return "Đã ghi"
        case .rejected: return "Đã bỏ qua"
        case .conflict: return "Có xung đột"
        }
    }
}

struct AdvisorEntryInputModel: Codable, Sendable {
    var clientRequestID: String
    var conversationID: String
    var generation: Int64
    var text: String
    var timezone: String?
    var attachmentIDs: [String]?
}

struct AdvisorEntryDecisionModel: Codable, Sendable {
    var taskID: String
    var batchID: String
    var batchVersion: Int64
    var itemID: String
    var itemVersion: Int
    var decision: String
    var clientRequestID: String

    enum CodingKeys: String, CodingKey {
        case taskID = "task_id"
        case batchID = "batch_id"
        case batchVersion = "batch_version"
        case itemID = "item_id"
        case itemVersion = "item_version"
        case decision
        case clientRequestID = "client_request_id"
    }
}
