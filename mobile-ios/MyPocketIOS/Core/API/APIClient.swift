import Foundation

protocol HTTPTransport: Sendable {
    func data(for request: URLRequest) async throws -> (Data, URLResponse)
}

struct URLSessionTransport: HTTPTransport {
    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        try await URLSession.shared.data(for: request)
    }
}

final class APIClient: @unchecked Sendable {
    private let baseURL: URL
    private let transport: HTTPTransport
    private let sessionStore: SessionStore
    private lazy var refreshCoordinator = RefreshCoordinator { [weak self] in
        guard let self else { throw APIError.missingSession }
        return try await self.refreshSession()
    }

    init(
        baseURL: URL = EnvironmentConfig.apiBaseURL,
        transport: HTTPTransport = URLSessionTransport(),
        sessionStore: SessionStore = KeychainSessionStore()
    ) {
        self.baseURL = baseURL
        self.transport = transport
        self.sessionStore = sessionStore
    }

    func send<Response: Decodable>(_ request: APIRequest<Response>) async throws -> Response {
        do {
            return try await execute(request, sessionOverride: nil)
        } catch APIError.unauthorized {
            guard case .session = request.auth else { throw APIError.unauthorized }
            let refreshed = try await refreshCoordinator.refreshIfNeeded()
            return try await execute(request, sessionOverride: refreshed)
        }
    }

    func sendMultipart<Response: Decodable & Sendable>(
        path: String,
        fields: [String: String],
        files: [MultipartFile] = [],
        auth: APIAuth = .session,
        idempotencyKey: String? = nil
    ) async throws -> Response {
        do {
            return try await executeMultipart(path: path, fields: fields, files: files, auth: auth, idempotencyKey: idempotencyKey, sessionOverride: nil)
        } catch APIError.unauthorized {
            guard case .session = auth else { throw APIError.unauthorized }
            let refreshed = try await refreshCoordinator.refreshIfNeeded()
            return try await executeMultipart(path: path, fields: fields, files: files, auth: auth, idempotencyKey: idempotencyKey, sessionOverride: refreshed)
        }
    }

    private func executeMultipart<Response: Decodable & Sendable>(
        path: String,
        fields: [String: String],
        files: [MultipartFile],
        auth: APIAuth,
        idempotencyKey: String?,
        sessionOverride: AccessSession?
    ) async throws -> Response {
        let relativePath = path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        guard var components = URLComponents(url: baseURL, resolvingAgainstBaseURL: false) else { throw APIError.invalidURL }
        let basePath = components.path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        components.path = "/" + [basePath, relativePath].filter { !$0.isEmpty }.joined(separator: "/")
        guard let url = components.url else { throw APIError.invalidURL }
        let boundary = "Boundary-\(UUID().uuidString)"
        var request = URLRequest(url: url)
        request.httpMethod = HTTPMethod.post.rawValue
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        if let idempotencyKey { request.setValue(idempotencyKey, forHTTPHeaderField: "Idempotency-Key") }
        switch auth {
        case .none: break
        case .apiKey(let key): request.setValue("Bearer \(key)", forHTTPHeaderField: "Authorization")
        case .session:
            guard let session = sessionOverride ?? sessionStore.load() else { throw APIError.missingSession }
            request.setValue("Bearer \(session.token)", forHTTPHeaderField: "Authorization")
        }
        var payload = Data()
        for (name, value) in fields {
            payload.append(Data("--\(boundary)\r\n".utf8))
            payload.append(Data("Content-Disposition: form-data; name=\"\(name)\"\r\n\r\n\(value)\r\n".utf8))
        }
        for file in files {
            payload.append(Data("--\(boundary)\r\n".utf8))
            payload.append(Data("Content-Disposition: form-data; name=\"\(file.fieldName)\"; filename=\"\(file.fileName)\"\r\n".utf8))
            payload.append(Data("Content-Type: \(file.mimeType)\r\n\r\n".utf8))
            payload.append(file.data)
            payload.append(Data("\r\n".utf8))
        }
        payload.append(Data("--\(boundary)--\r\n".utf8))
        request.httpBody = payload
        do {
            let (data, response) = try await transport.data(for: request)
            guard let http = response as? HTTPURLResponse else { throw APIError.transport("Phản hồi không hợp lệ") }
            guard (200..<300).contains(http.statusCode) else {
                if http.statusCode == 401 { throw APIError.unauthorized }
                if let problem = try? JSONDecoder().decode(ProblemDetails.self, from: data) { throw APIError.problem(problem) }
                throw APIError.http(status: http.statusCode)
            }
            return try JSONDecoder.mypocket.decode(APIEnvelope<Response>.self, from: data).data
        } catch let error as APIError { throw error }
        catch { throw APIError.transport(error.localizedDescription) }
    }

    private func execute<Response: Decodable>(
        _ request: APIRequest<Response>,
        sessionOverride: AccessSession?
    ) async throws -> Response {
        let relativePath = request.path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        guard var components = URLComponents(url: baseURL, resolvingAgainstBaseURL: false) else {
            throw APIError.invalidURL
        }
        let basePath = components.path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        components.path = "/" + [basePath, relativePath].filter { !$0.isEmpty }.joined(separator: "/")
        guard let url = components.url else { throw APIError.invalidURL }
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = request.method.rawValue
        urlRequest.httpBody = request.body
        urlRequest.setValue("application/json", forHTTPHeaderField: "Accept")
        if request.body != nil {
            urlRequest.setValue("application/json", forHTTPHeaderField: "Content-Type")
        }
        if let idempotencyKey = request.idempotencyKey {
            urlRequest.setValue(idempotencyKey, forHTTPHeaderField: "Idempotency-Key")
        }

        switch request.auth {
        case .none:
            break
        case .apiKey(let key):
            urlRequest.setValue("Bearer \(key)", forHTTPHeaderField: "Authorization")
        case .session:
            guard let session = sessionOverride ?? sessionStore.load() else {
                throw APIError.missingSession
            }
            urlRequest.setValue("Bearer \(session.token)", forHTTPHeaderField: "Authorization")
        }

        do {
            let (data, response) = try await transport.data(for: urlRequest)
            guard let http = response as? HTTPURLResponse else {
                throw APIError.transport("Phản hồi không hợp lệ")
            }
            guard (200..<300).contains(http.statusCode) else {
                if http.statusCode == 401 { throw APIError.unauthorized }
                if let problem = try? JSONDecoder().decode(ProblemDetails.self, from: data) {
                    throw APIError.problem(problem)
                }
                throw APIError.http(status: http.statusCode)
            }
            do {
                return try JSONDecoder.mypocket.decode(APIEnvelope<Response>.self, from: data).data
            } catch {
                throw APIError.decoding(error.localizedDescription)
            }
        } catch let error as APIError {
            throw error
        } catch {
            throw APIError.transport(error.localizedDescription)
        }
    }

    private func refreshSession() async throws -> AccessSession {
        guard let current = sessionStore.load() else { throw APIError.missingSession }
        let requestBody = try JSONEncoder.mypocket.encode(RefreshRequest(refreshToken: current.refreshToken))
        let request = APIRequest<LoginOutput>(
            path: "/auth/refresh",
            method: .post,
            body: requestBody,
            auth: .none
        )
        let output = try await execute(request, sessionOverride: nil)
        let session = output.session
        try sessionStore.save(session)
        return session
    }
}

struct MultipartFile: Sendable {
    let fieldName: String
    let fileName: String
    let mimeType: String
    let data: Data
}

extension JSONDecoder {
    static var mypocket: JSONDecoder {
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        decoder.keyDecodingStrategy = .custom { keys in
            let key = keys.last?.stringValue ?? ""
            let parts = key.split(separator: "_")
            guard let first = parts.first else { return AnyCodingKey(stringValue: key)! }
            let converted = parts.dropFirst().reduce(String(first)) { result, part in
                switch part.lowercased() {
                case "id": return result + "ID"
                case "url": return result + "URL"
                case "bps": return result + "BPS"
                default: return result + part.prefix(1).uppercased() + part.dropFirst()
                }
            }
            return AnyCodingKey(stringValue: converted)!
        }
        return decoder
    }
}

private struct AnyCodingKey: CodingKey {
    let stringValue: String
    let intValue: Int? = nil

    init?(stringValue: String) { self.stringValue = stringValue }
    init?(intValue: Int) { return nil }
}

extension JSONEncoder {
    static var mypocket: JSONEncoder {
        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
        encoder.keyEncodingStrategy = .convertToSnakeCase
        return encoder
    }
}
