import Foundation

/// Stable navigation targets shared by native deep links and in-app navigation.
/// Resource identifiers are hints only; feature services still scope all reads
/// and writes by the authenticated session owner.
enum AppRoute: Equatable, Hashable, Sendable {
    case overview
    case transactions(walletID: String?)
    case newTransaction(walletID: String?)
    case wallets
    case wallet(id: String)
    case budgets
    case categories
    case jars
    case months
    case recurring
    case travel
    case reports
    case portfolio
    case credit
    case notifications
    case assistant
    case aiEntry
    case account
    case feedback

    var requiresAuthentication: Bool { true }

    var ownerScopedResourceID: String? {
        if case .wallet(let id) = self { return id }
        return nil
    }
}

enum AppRouteParser {
    static let universalLinkHost = "money.dungxbuif.com"
    static let scheme = "mypocket"

    static func parse(_ url: URL) -> AppRoute? {
        guard url.user == nil, url.password == nil, url.fragment == nil else { return nil }
        guard let path = routePath(for: url) else { return nil }
        guard let components = URLComponents(url: url, resolvingAgainstBaseURL: false) else { return nil }
        if path == ["wallets"], components.percentEncodedPath.hasSuffix("/") { return nil }

        let queryItems = components.queryItems ?? []
        let allowedQueryNames: Set<String> = ["wallet_id"]
        guard queryItems.allSatisfy({ allowedQueryNames.contains($0.name.lowercased()) }) else { return nil }
        let walletID = queryItems.first(where: { $0.name.lowercased() == "wallet_id" })?.value
        guard queryItems.filter({ $0.name.lowercased() == "wallet_id" }).count <= 1 else { return nil }
        guard walletID?.isEmpty != true else { return nil }

        switch path {
        case ["overview"]:
            return .overview
        case ["transactions"]:
            return .transactions(walletID: walletID)
        case ["transactions", "new"]:
            return .newTransaction(walletID: walletID)
        case ["wallets"]:
            guard walletID == nil else { return nil }
            return .wallets
        case ["budgets"]:
            guard walletID == nil else { return nil }
            return .budgets
        case ["categories"]:
            guard walletID == nil else { return nil }
            return .categories
        case ["jars"]:
            guard walletID == nil else { return nil }
            return .jars
        case ["months"]:
            guard walletID == nil else { return nil }
            return .months
        case ["recurring"]:
            guard walletID == nil else { return nil }
            return .recurring
        case ["travel"]:
            guard walletID == nil else { return nil }
            return .travel
        case ["reports"]:
            guard walletID == nil else { return nil }
            return .reports
        case ["portfolio"]:
            guard walletID == nil else { return nil }
            return .portfolio
        case ["credit"]:
            guard walletID == nil else { return nil }
            return .credit
        case ["notifications"]:
            guard walletID == nil else { return nil }
            return .notifications
        case ["assistant"]:
            guard walletID == nil else { return nil }
            return .assistant
        case ["ai-entry"]:
            guard walletID == nil else { return nil }
            return .aiEntry
        case ["account"]:
            guard walletID == nil else { return nil }
            return .account
        case ["feedback"]:
            guard walletID == nil else { return nil }
            return .feedback
        default:
            guard path.count == 2, path[0] == "wallets", !path[1].isEmpty, walletID == nil else { return nil }
            return .wallet(id: path[1])
        }
    }

    private static func routePath(for url: URL) -> [String]? {
        guard let scheme = url.scheme?.lowercased() else { return nil }
        switch scheme {
        case Self.scheme:
            guard url.host == nil || url.host?.lowercased() == "app" else { return nil }
            let parts = pathComponents(url.path)
            if url.host?.lowercased() == "app" { return parts }
            guard parts.first == "app" else { return nil }
            return Array(parts.dropFirst())
        case "https":
            guard url.host?.lowercased() == Self.universalLinkHost else { return nil }
            let parts = pathComponents(url.path)
            guard parts.first == "app" else { return nil }
            return Array(parts.dropFirst())
        default:
            return nil
        }
    }

    private static func pathComponents(_ path: String) -> [String] {
        path.split(separator: "/", omittingEmptySubsequences: true).map(String.init)
    }
}
