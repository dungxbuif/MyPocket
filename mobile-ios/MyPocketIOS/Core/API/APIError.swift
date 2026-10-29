import Foundation

enum APIError: Error, Equatable, Sendable {
    case invalidURL
    case transport(String)
    case unauthorized
    case http(status: Int)
    case problem(ProblemDetails)
    case decoding(String)
    case missingSession
    case accountMismatch

    static func diagnosticDescription(_ error: Error) -> String {
        guard let error = error as? APIError else { return "unexpected" }
        switch error {
        case .invalidURL: return "invalidURL"
        case .transport: return "transport"
        case .unauthorized: return "unauthorized"
        case .http(let status): return "http \(status)"
        case .problem(let problem):
            return problem.status.map { "problem \($0)" } ?? "problem"
        case .decoding: return "decoding"
        case .missingSession: return "missingSession"
        case .accountMismatch: return "accountMismatch"
        }
    }

    var userMessage: String {
        switch self {
        case .problem(let problem): return problem.detail ?? "Yêu cầu không hợp lệ."
        case .unauthorized: return "Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại."
        case .transport(let detail): return detail
        case .http(let status): return "Máy chủ trả về lỗi HTTP \(status)."
        case .decoding: return "Phản hồi máy chủ không đúng định dạng."
        case .invalidURL: return "Địa chỉ API không hợp lệ."
        case .missingSession: return "Chưa có phiên đăng nhập."
        case .accountMismatch: return "Tài khoản đã thay đổi. Đăng nhập lại đúng tài khoản để tiếp tục đồng bộ."
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

    func withHTTPStatus(_ status: Int) -> ProblemDetails {
        ProblemDetails(title: title, status: status, detail: detail, instance: instance, code: code, requestID: requestID)
    }

}
