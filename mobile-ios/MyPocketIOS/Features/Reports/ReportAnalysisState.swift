import Foundation

enum ReportAnalysisState: Equatable, Sendable {
    case idle
    case queued(requestID: String)
    case running(requestID: String)
    case completed(requestID: String, revision: String?, article: String)
    case failed(message: String)
    case cancelled
    case expired
}
