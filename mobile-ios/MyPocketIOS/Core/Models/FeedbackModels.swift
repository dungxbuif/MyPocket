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
