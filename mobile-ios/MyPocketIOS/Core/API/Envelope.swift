import Foundation

enum JSONValue: Decodable, Sendable {
    case string(String)
    case number(Double)
    case bool(Bool)
    case object([String: JSONValue])
    case array([JSONValue])
    case null

    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        if container.decodeNil() {
            self = .null
        } else if let value = try? container.decode(Bool.self) {
            self = .bool(value)
        } else if let value = try? container.decode(Double.self) {
            self = .number(value)
        } else if let value = try? container.decode(String.self) {
            self = .string(value)
        } else if let value = try? container.decode([String: JSONValue].self) {
            self = .object(value)
        } else if let value = try? container.decode([JSONValue].self) {
            self = .array(value)
        } else {
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Unsupported JSON value")
        }
    }
}

struct APIEnvelope<Payload: Decodable>: Decodable {
    let data: Payload
    let meta: JSONValue?
}

enum HTTPMethod: String, Sendable {
    case get = "GET"
    case post = "POST"
    case patch = "PATCH"
    case put = "PUT"
    case delete = "DELETE"
}

enum APIAuth: Sendable {
    case session
    case none
    case apiKey(String)
}

struct APIRequest<Response: Decodable & Sendable>: Sendable {
    let path: String
    let method: HTTPMethod
    let body: Data?
    let auth: APIAuth
    let idempotencyKey: String?

    init(
        path: String,
        method: HTTPMethod = .get,
        body: Data? = nil,
        auth: APIAuth = .session,
        idempotencyKey: String? = nil
    ) {
        self.path = path
        self.method = method
        self.body = body
        self.auth = auth
        self.idempotencyKey = idempotencyKey
    }
}
