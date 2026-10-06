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

    var session: AccessSession {
        AccessSession(token: token, expiresAt: expiresAt, refreshToken: refreshToken, refreshExpiresAt: refreshExpiresAt, user: user)
    }
}

struct LoginRequest: Codable, Sendable {
    let email: String
    let password: String
}

struct RefreshRequest: Codable, Sendable {
    let refreshToken: String

}

struct AccessSession: Codable, Equatable, Sendable {
    let token: String
    let expiresAt: String
    let refreshToken: String
    let refreshExpiresAt: String
    let user: AuthUser

    var accessTokenExpiry: Date? { Self.parseDate(expiresAt) }
    var refreshTokenExpiry: Date? { Self.parseDate(refreshExpiresAt) }

    func hasValidAccessToken(at date: Date = Date()) -> Bool {
        guard let expiry = accessTokenExpiry else { return false }
        return expiry > date
    }

    func hasValidRefreshToken(at date: Date = Date()) -> Bool {
        guard let expiry = refreshTokenExpiry else { return false }
        return expiry > date
    }

    private static func parseDate(_ value: String) -> Date? {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter.date(from: value) ?? {
            formatter.formatOptions = [.withInternetDateTime]
            return formatter.date(from: value)
        }()
    }
}

protocol SessionStore: Sendable {
    func load() -> AccessSession?
    func save(_ session: AccessSession) throws
    func clear()
}
