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

struct AdvisorCapabilitiesModel: Codable, Sendable {
    var enabled: Bool
    var write: Bool
    var partsVersion: Int
    var tools: [String]
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

struct AdvisorRunModel: Codable, Sendable {
    var id: String
    var conversationID: String
    var status: String
    var errorCode: String?
}
