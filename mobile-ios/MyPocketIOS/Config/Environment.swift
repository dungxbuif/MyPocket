import Foundation

enum EnvironmentConfig {
    static let apiBaseURL: URL = {
        let value = Bundle.main.object(forInfoDictionaryKey: "MP_API_BASE_URL") as? String
            ?? "http://127.0.0.1:8080/api/v1"
        return URL(string: value) ?? URL(string: "http://127.0.0.1:8080/api/v1")!
    }()
}
