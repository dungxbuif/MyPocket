import Foundation

struct AuthUser: Codable, Equatable, Sendable {
    let id: String
    let name: String?
    let email: String?
    let timezone: String?
}

struct LoginOutput: Codable, Equatable, Sendable {
    let token: String
    let expiresAt: String
    let refreshToken: String
    let refreshExpiresAt: String
    let user: AuthUser

    enum CodingKeys: String, CodingKey {
        case token
        case expiresAt = "expires_at"
        case refreshToken = "refresh_token"
        case refreshExpiresAt = "refresh_expires_at"
        case user
    }

    var session: AccessSession {
        AccessSession(token: token, expiresAt: expiresAt, refreshToken: refreshToken, refreshExpiresAt: refreshExpiresAt, user: user)
    }
}

struct RefreshRequest: Codable, Sendable {
    let refreshToken: String

    enum CodingKeys: String, CodingKey {
        case refreshToken = "refresh_token"
    }
}

struct AccessSession: Codable, Equatable, Sendable {
    let token: String
    let expiresAt: String
    let refreshToken: String
    let refreshExpiresAt: String
    let user: AuthUser
}

protocol SessionStore: Sendable {
    func load() -> AccessSession?
    func save(_ session: AccessSession) throws
    func clear()
}
