import Foundation

enum APIError: Error, Equatable, Sendable {
    case invalidURL
    case transport(String)
    case unauthorized
    case http(status: Int)
    case problem(ProblemDetails)
    case decoding(String)
    case missingSession
}

struct ProblemDetails: Codable, Equatable, Sendable {
    let title: String?
    let status: Int?
    let detail: String?
    let instance: String?
    let code: String?
    let requestID: String?

}
