import Foundation

enum ReportAnalysisState: Equatable, Sendable {
    case idle
    case queued(requestID: String)
    case running(requestID: String)
    case completed(requestID: String, revision: Int64?, article: String)
    case failed(message: String)
    case cancelled
    case expired
}
