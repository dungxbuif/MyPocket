import Foundation

enum APIError: Error, Equatable, Sendable {
    case invalidURL
    case transport(String)
    case unauthorized
    case http(status: Int)
    case problem(ProblemDetails)
    case decoding(String)
    case missingSession

    var userMessage: String {
        switch self {
        case .problem(let problem): return problem.detail ?? "Yêu cầu không hợp lệ."
        case .unauthorized: return "Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại."
        case .transport(let detail): return detail
        case .http(let status): return "Máy chủ trả về lỗi HTTP \(status)."
        case .decoding: return "Phản hồi máy chủ không đúng định dạng."
        case .invalidURL: return "Địa chỉ API không hợp lệ."
        case .missingSession: return "Chưa có phiên đăng nhập."
        }
    }
}

struct ProblemDetails: Codable, Equatable, Sendable {
    let title: String?
    let status: Int?
    let detail: String?
    let instance: String?
    let code: String?
    let requestID: String?

}
