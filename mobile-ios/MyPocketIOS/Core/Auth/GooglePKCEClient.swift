import AuthenticationServices
import CryptoKit
import Foundation
import Security
import UIKit

struct MobileGoogleExchangeRequest: Codable, Sendable {
    let code: String
    let codeVerifier: String
    let redirectURI: String
    let platform: String

    enum CodingKeys: String, CodingKey {
        case code
        case codeVerifier = "code_verifier"
        case redirectURI = "redirect_uri"
        case platform
    }
}

enum GooglePKCEError: Error, Equatable, Sendable {
    case invalidConfiguration
    case cancelled
    case callbackMissingCode
    case stateMismatch
}

struct PKCEChallenge: Sendable {
    let verifier: String
    let challenge: String
    let state: String

    static func make() -> PKCEChallenge {
        let verifier = randomURLSafeValue()
        let challengeData = SHA256.hash(data: Data(verifier.utf8))
        let challenge = Data(challengeData).base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
        return PKCEChallenge(verifier: verifier, challenge: challenge, state: randomURLSafeValue())
    }

    private static func randomURLSafeValue() -> String {
        var bytes = [UInt8](repeating: 0, count: 32)
        _ = SecRandomCopyBytes(kSecRandomDefault, bytes.count, &bytes)
        return Data(bytes).base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")
    }
}

@MainActor
final class GooglePKCEClient: NSObject, ASWebAuthenticationPresentationContextProviding {
    private let apiClient: APIClient
    private let clientID: String
    private let redirectURI: String
    private var browserSession: ASWebAuthenticationSession?

    init(
        apiClient: APIClient,
        clientID: String = Bundle.main.object(forInfoDictionaryKey: "MP_GOOGLE_CLIENT_ID") as? String ?? "",
        redirectURI: String = Bundle.main.object(forInfoDictionaryKey: "MP_GOOGLE_REDIRECT_URI") as? String ?? "mypocket://oauth/callback"
    ) {
        self.apiClient = apiClient
        self.clientID = clientID
        self.redirectURI = redirectURI
    }

    func signIn() async throws -> LoginOutput {
        guard !clientID.isEmpty, let callbackScheme = URL(string: redirectURI)?.scheme else {
            throw GooglePKCEError.invalidConfiguration
        }
        let challenge = PKCEChallenge.make()
        var components = URLComponents(string: "https://accounts.google.com/o/oauth2/v2/auth")
        components?.queryItems = [
            URLQueryItem(name: "client_id", value: clientID),
            URLQueryItem(name: "redirect_uri", value: redirectURI),
            URLQueryItem(name: "response_type", value: "code"),
            URLQueryItem(name: "scope", value: "openid email profile"),
            URLQueryItem(name: "code_challenge", value: challenge.challenge),
            URLQueryItem(name: "code_challenge_method", value: "S256"),
            URLQueryItem(name: "state", value: challenge.state),
        ]
        guard let authURL = components?.url else { throw GooglePKCEError.invalidConfiguration }

        let callbackURL: URL = try await withCheckedThrowingContinuation { continuation in
            let session = ASWebAuthenticationSession(url: authURL, callbackURLScheme: callbackScheme) { url, error in
                if let url {
                    continuation.resume(returning: url)
                } else if let error, (error as? ASWebAuthenticationSessionError)?.code == .canceledLogin {
                    continuation.resume(throwing: GooglePKCEError.cancelled)
                } else {
                    continuation.resume(throwing: error ?? GooglePKCEError.callbackMissingCode)
                }
            }
            session.presentationContextProvider = self
            session.prefersEphemeralWebBrowserSession = true
            browserSession = session
            session.start()
        }
        browserSession = nil

        guard let query = URLComponents(url: callbackURL, resolvingAgainstBaseURL: false)?.queryItems,
              let returnedState = query.first(where: { $0.name == "state" })?.value,
              returnedState == challenge.state else {
            throw GooglePKCEError.stateMismatch
        }
        guard let code = query.first(where: { $0.name == "code" })?.value, !code.isEmpty else {
            throw GooglePKCEError.callbackMissingCode
        }
        let body = try JSONEncoder.mypocket.encode(MobileGoogleExchangeRequest(code: code, codeVerifier: challenge.verifier, redirectURI: redirectURI, platform: "ios"))
        return try await apiClient.send(APIRequest<LoginOutput>(path: "/auth/mobile/google/exchange", method: .post, body: body, auth: .none))
    }

    func presentationAnchor(for session: ASWebAuthenticationSession) -> ASPresentationAnchor {
        UIApplication.shared.connectedScenes
            .compactMap { $0 as? UIWindowScene }
            .flatMap(\.windows)
            .first(where: \.isKeyWindow) ?? UIWindow()
    }
}
