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
    var onSessionInvalidated: (@Sendable () -> Void)?
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
        try await sendEnvelope(request).data
    }

    func sendEnvelope<Response: Decodable>(_ request: APIRequest<Response>) async throws -> APIEnvelope<Response> {
        let requestID = UUID().uuidString
        do {
            return try await executeEnvelope(request, requestID: requestID, sessionOverride: nil)
        } catch APIError.unauthorized {
            guard case .session = request.auth else { throw APIError.unauthorized }
            try validateExpectedOwner(request)
            let refreshed = try await refreshCoordinator.refreshIfNeeded()
            return try await executeEnvelope(request, requestID: requestID, sessionOverride: refreshed)
        }
    }

    func sendMultipart<Response: Decodable & Sendable>(
        path: String,
        fields: [String: String],
        files: [MultipartFile] = [],
        auth: APIAuth = .session,
        idempotencyKey: String? = nil
    ) async throws -> Response {
        let requestID = UUID().uuidString
        do {
            return try await executeMultipart(path: path, fields: fields, files: files, auth: auth, idempotencyKey: idempotencyKey, requestID: requestID, sessionOverride: nil)
        } catch APIError.unauthorized {
            guard case .session = auth else { throw APIError.unauthorized }
            let refreshed = try await refreshCoordinator.refreshIfNeeded()
            return try await executeMultipart(path: path, fields: fields, files: files, auth: auth, idempotencyKey: idempotencyKey, requestID: requestID, sessionOverride: refreshed)
        }
    }

    private func executeMultipart<Response: Decodable & Sendable>(
        path: String,
        fields: [String: String],
        files: [MultipartFile],
        auth: APIAuth,
        idempotencyKey: String?,
        requestID: String,
        sessionOverride: AccessSession?
    ) async throws -> Response {
        let url = try makeURL(path: path)
        let boundary = "Boundary-\(UUID().uuidString)"
        var request = URLRequest(url: url)
        request.httpMethod = HTTPMethod.post.rawValue
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue(requestID, forHTTPHeaderField: "X-Request-ID")
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
                if let problem = try? JSONDecoder().decode(ProblemDetails.self, from: data) { throw APIError.problem(problem.withHTTPStatus(http.statusCode)) }
                throw APIError.http(status: http.statusCode)
            }
            if http.statusCode == 204 {
                guard let empty = EmptyResponse() as? Response else {
                    throw APIError.decoding("Phản hồi 204 không phù hợp với kiểu dữ liệu yêu cầu")
                }
                return empty
            }
            return try JSONDecoder.mypocket.decode(APIEnvelope<Response>.self, from: data).data
        } catch let error as APIError { throw error }
        catch { throw APIError.transport(error.localizedDescription) }
    }

    private func execute<Response: Decodable>(
        _ request: APIRequest<Response>,
        requestID: String,
        sessionOverride: AccessSession?
    ) async throws -> Response {
        try await executeEnvelope(request, requestID: requestID, sessionOverride: sessionOverride).data
    }

    private func executeEnvelope<Response: Decodable>(
        _ request: APIRequest<Response>,
        requestID: String,
        sessionOverride: AccessSession?
    ) async throws -> APIEnvelope<Response> {
        try validateExpectedOwner(request)
        let url = try makeURL(path: request.path)
        var urlRequest = URLRequest(url: url)
        urlRequest.httpMethod = request.method.rawValue
        urlRequest.httpBody = request.body
        urlRequest.setValue("application/json", forHTTPHeaderField: "Accept")
        urlRequest.setValue(requestID, forHTTPHeaderField: "X-Request-ID")
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
            if let expectedOwnerID = request.expectedOwnerID, session.user.id != expectedOwnerID { throw APIError.accountMismatch }
            urlRequest.setValue("Bearer \(session.token)", forHTTPHeaderField: "Authorization")
        }

        do {
#if DEBUG
            print("[API] -> \(request.method.rawValue) [\(requestID)]")
#endif
            let (data, response) = try await transport.data(for: urlRequest)
            try validateExpectedOwner(request)
            guard let http = response as? HTTPURLResponse else {
                throw APIError.transport("Phản hồi không hợp lệ")
            }
#if DEBUG
            // URLs and error bodies may contain financial data or credentials.
            print("[API] <- \(request.method.rawValue) \(http.statusCode) [\(requestID)]")
#endif
            guard (200..<300).contains(http.statusCode) else {
                if http.statusCode == 401 { throw APIError.unauthorized }
                if let problem = try? JSONDecoder().decode(ProblemDetails.self, from: data) {
                    throw APIError.problem(problem.withHTTPStatus(http.statusCode))
                }
                throw APIError.http(status: http.statusCode)
            }
            if http.statusCode == 204 {
                guard let empty = EmptyResponse() as? Response else {
                    throw APIError.decoding("Phản hồi 204 không phù hợp với kiểu dữ liệu yêu cầu")
                }
                return APIEnvelope(data: empty, meta: nil)
            }
            do {
                return try JSONDecoder.mypocket.decode(APIEnvelope<Response>.self, from: data)
            } catch {
                throw APIError.decoding(error.localizedDescription)
            }
        } catch let error as APIError {
            throw error
        } catch {
            throw APIError.transport(error.localizedDescription)
        }
    }

    private func validateExpectedOwner<Response>(_ request: APIRequest<Response>) throws {
        guard let owner = request.expectedOwnerID else { return }
        guard case .session = request.auth, !owner.isEmpty, sessionStore.load()?.user.id == owner else { throw APIError.accountMismatch }
    }

    private func makeURL(path: String) throws -> URL {
        let pathParts = path.split(separator: "?", maxSplits: 1, omittingEmptySubsequences: false)
        let relativePath = String(pathParts[0]).trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        guard var components = URLComponents(url: baseURL, resolvingAgainstBaseURL: false) else {
            throw APIError.invalidURL
        }
        let basePath = components.path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        components.path = "/" + [basePath, relativePath].filter { !$0.isEmpty }.joined(separator: "/")
        if pathParts.count == 2, !pathParts[1].isEmpty {
            components.percentEncodedQuery = String(pathParts[1])
        }
        guard let url = components.url else { throw APIError.invalidURL }
        return url
    }

    private func refreshSession(invalidateOnUnauthorized: Bool = true) async throws -> AccessSession {
        guard let current = sessionStore.load() else { throw APIError.missingSession }
        let requestBody = try JSONEncoder.mypocket.encode(RefreshRequest(refreshToken: current.refreshToken))
        let request = APIRequest<LoginOutput>(
            path: "/auth/refresh",
            method: .post,
            body: requestBody,
            auth: .none
        )
        let output: LoginOutput
        do {
            output = try await execute(request, requestID: UUID().uuidString, sessionOverride: nil)
        } catch APIError.unauthorized {
            try validateRefreshIdentity(current)
            if invalidateOnUnauthorized {
                guard sessionStore.clear(ifCredentialsMatch: current) else { throw APIError.accountMismatch }
                onSessionInvalidated?()
            }
            throw APIError.unauthorized
        }
        let session = output.session
        try validateRefreshIdentity(current)
        guard session.user.id == current.user.id else { throw APIError.accountMismatch }
        guard try sessionStore.replace(session, ifCredentialsMatch: current) else { throw APIError.accountMismatch }
        return session
    }

    private func validateRefreshIdentity(_ original: AccessSession) throws {
        guard sessionStore.load()?.hasSameCredentials(as: original) == true else {
            throw APIError.accountMismatch
        }
    }

    func refreshSessionForRestore() async throws -> AccessSession {
        try await refreshSession(invalidateOnUnauthorized: false)
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
                case "ids": return result + "IDs"
                case "url": return result + "URL"
                case "bps": return result + "BPS"
                case "pnl": return result + "PnL"
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
        // Stable wire bytes make retries and idempotency diagnostics reproducible.
        encoder.outputFormatting = [.sortedKeys]
        return encoder
    }
}
