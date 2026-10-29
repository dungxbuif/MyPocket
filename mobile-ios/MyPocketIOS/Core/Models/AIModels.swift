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
