import Foundation

struct FeedbackModel: Codable, Identifiable, Sendable {
    let id: String
    var type: String
    var title: String
    var description: String
    var status: String
    var fixedAt: String?
    var changelogID: String?
    var createdAt: String
    var updatedAt: String
    var screenshotAvailable: Bool
}

struct ChangelogModel: Codable, Identifiable, Sendable {
    let id: String
    var version: String
    var title: String
    var description: String
    var publishedAt: String
}

struct FeedbackInputModel: Codable, Sendable {
    var type: String
    var title: String
    var description: String
}

enum FeedbackDraftValidator {
    static let allowedTypes: Set<String> = ["bug", "feature", "improvement"]

    static func normalized(type: String, title: String, description: String) -> FeedbackInputModel? {
        let normalizedTitle = title.trimmingCharacters(in: .whitespacesAndNewlines)
        let normalizedDescription = description.trimmingCharacters(in: .whitespacesAndNewlines)
        guard allowedTypes.contains(type), !normalizedTitle.isEmpty, !normalizedDescription.isEmpty else { return nil }
        return FeedbackInputModel(type: type, title: normalizedTitle, description: normalizedDescription)
    }
}
