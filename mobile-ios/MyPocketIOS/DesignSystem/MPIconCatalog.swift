import Foundation

/// Stable semantic roles for icons used by the native design system.
///
/// These identifiers are presentation-only. API `icon_key`/`system_key`
/// values remain persisted exactly as received; callers resolve those values
/// through `MPIconCatalog.icon(forLegacyKey:)` when they need a glyph.
enum MPIconID: String, CaseIterable, Hashable {
    case tag
    case home
    case ledger
    case assistant
    case account
    case feedback
    case wallet
    case group
    case bankLink
    case recurring
    case report
    case search
    case bell
    case add
    case send
    case food
    case coffee
    case snacks
    case meals
    case shopping
    case transport
    case health
    case bills
    case education
    case travel
    case savings
    case family
    case entertainment
    case gifts
    case insurance
    case investment
    case business
    case transfer
    case income
    case debt
    case credit
    case aggregate
    case adjustment
}

/// Presentation weight requested by the icon study. The shipped subset uses
/// Phosphor Regular vectors; `.duotone` remains a compatibility style until
/// a two-layer tintable subset is needed by a consumer.
enum MPIconStyle: String, CaseIterable, Hashable {
    case regular
    case duotone
}

/// Compatibility boundary between persisted/system icon keys and semantic
/// icon roles. This is deliberately data-only: it does not fetch SVGs at
/// runtime and it does not rename or migrate a persisted key.
enum MPIconCatalog {
    /// True because the semantic subset is bundled in the asset catalog. Any
    /// missing/unknown key still falls back to a platform symbol safely.
    static let usesVerifiedAssets = true

    /// Maps a legacy key, system-name alias, or known category family to a
    /// semantic role. Unknown keys intentionally become a visible tag.
    static func icon(forLegacyKey key: String) -> MPIconID {
        let normalizedKey = normalize(key)
        if let exact = exactAliases[normalizedKey] {
            return exact
        }
        if let family = familyAliases.first(where: { normalizedKey.hasPrefix($0.prefix) }) {
            return family.icon
        }
        return .tag
    }

    /// Returns the bundled Phosphor asset name for a semantic icon. The
    /// regular vectors are used for both styles for now; callers retain the
    /// style parameter so a later duotone layer does not change persisted
    /// icon keys or call sites.
    static func assetName(for icon: MPIconID, style: MPIconStyle = .regular) -> String {
        _ = style
        return bundledAssetNames[icon] ?? fallbackSystemNames[icon, default: "tag"]
    }

    static func isBundledAssetName(_ name: String) -> Bool {
        name.hasPrefix("mp-icon-")
    }

    /// Useful for pure compatibility tests and future catalog tooling. The
    /// returned values are aliases only; they are never written back to API
    /// models or persistence.
    static var supportedLegacyKeys: [String] {
        exactAliases.keys.sorted()
    }

    private static func normalize(_ key: String) -> String {
        key.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
    }

    private struct FamilyAlias {
        let prefix: String
        let icon: MPIconID
    }

    // Exact entries win over family prefixes so specific subcategories such
    // as expense_food_coffee keep their identity.
    private static let exactAliases: [String: MPIconID] = [
        // Generic fallback and common navigation/action aliases.
        "tag": .tag,
        "tag.fill": .tag,
        "tag.circle": .tag,
        "tag.circle.fill": .tag,
        "uncategorized": .tag,
        "home": .home,
        "house": .home,
        "house.fill": .family,
        "house.circle": .home,
        "house.circle.fill": .home,
        "ledger": .ledger,
        "list": .ledger,
        "list.bullet": .ledger,
        "list.bullet.rectangle": .ledger,
        "book": .ledger,
        "assistant": .assistant,
        "sparkles": .assistant,
        "wand.and.stars": .assistant,
        "bubble.left": .assistant,
        "bubble.left.fill": .assistant,
        "account": .account,
        "person": .account,
        "person.fill": .account,
        "person.circle": .account,
        "person.circle.fill": .account,
        "person.crop.circle": .account,
        "feedback": .feedback,
        "bubble.left.and.bubble.right": .feedback,
        "bubble.left.and.bubble.right.fill": .feedback,
        "wallet": .wallet,
        "wallet.fill": .wallet,
        "wallet.pass": .wallet,
        "wallet.pass.fill": .wallet,
        "wallet.bifold": .wallet,
        "wallet.bifold.fill": .wallet,
        "shippingbox": .group,
        "shippingbox.fill": .group,
        "group": .group,
        "collection": .group,
        "link": .bankLink,
        "link.circle": .bankLink,
        "link.circle.fill": .bankLink,
        "bank_link": .bankLink,
        "recurring": .recurring,
        "repeat": .recurring,
        "repeat.1": .recurring,
        "arrow.clockwise": .recurring,
        "report": .report,
        "chart.bar": .report,
        "chart.bar.fill": .report,
        "chart.pie": .report,
        "chart.pie.fill": .report,
        "search": .search,
        "magnifyingglass": .search,
        "bell": .bell,
        "bell.fill": .bell,
        "bell.badge": .bell,
        "bell.badge.fill": .bell,
        "add": .add,
        "plus": .add,
        "plus.circle": .add,
        "plus.circle.fill": .add,
        "send": .send,
        "paperplane": .send,
        "paperplane.fill": .send,

        // Category picker and transaction presentation aliases.
        "food": .food,
        "fork.knife": .food,
        "fork.knife.circle": .food,
        "fork.knife.circle.fill": .food,
        "coffee": .coffee,
        "cup.and.saucer": .coffee,
        "cup.and.saucer.fill": .coffee,
        "snacks": .snacks,
        "leaf": .snacks,
        "leaf.fill": .snacks,
        "meals": .meals,
        "meal": .meals,
        "bowl": .meals,
        "bowl.fill": .meals,
        "bills": .bills,
        "receipt": .bills,
        "receipt.fill": .bills,
        "doc.text": .bills,
        "doc.text.fill": .bills,
        "shopping": .shopping,
        "bag": .shopping,
        "bag.fill": .shopping,
        "cart": .shopping,
        "cart.fill": .shopping,
        "family": .family,
        "household": .family,
        "transport": .transport,
        "car": .transport,
        "car.fill": .transport,
        "bus.fill": .transport,
        "train.side.front.car": .transport,
        "health": .health,
        "heart": .health,
        "heart.fill": .health,
        "heart.text.square.fill": .health,
        "cross.case.fill": .health,
        "education": .education,
        "book.fill": .education,
        "graduationcap": .education,
        "graduationcap.fill": .education,
        "entertainment": .entertainment,
        "gamecontroller": .entertainment,
        "gamecontroller.fill": .entertainment,
        "music.note": .entertainment,
        "gifts": .gifts,
        "gift": .gifts,
        "gift.fill": .gifts,
        "insurance": .insurance,
        "checkmark.shield.fill": .insurance,
        "shield.checkered": .insurance,
        "investment": .investment,
        "chart.xyaxis.line": .investment,
        "chart.line.uptrend.xyaxis": .investment,
        "chart.bar.xaxis": .investment,
        "travel": .travel,
        "airplane": .travel,
        "airplane.circle.fill": .travel,
        "savings": .savings,
        "banknote": .savings,
        "banknote.fill": .savings,
        "business": .business,
        "briefcase": .business,
        "briefcase.fill": .business,
        "transfer": .transfer,
        "arrow.left.arrow.right": .transfer,
        "building.columns": .transfer,
        "building.columns.fill": .transfer,
        "income": .income,
        "income_root": .income,
        "income_salary": .income,
        "salary": .income,
        "income_interest": .income,
        "interest": .income,
        "income_gift": .gifts,
        "income_bonus": .gifts,
        "income_transfer_in": .transfer,
        "income_uncategorized": .tag,
        "debt": .debt,
        "hand.raised": .debt,
        "hand.raised.fill": .debt,
        "credit": .credit,
        "creditcard": .credit,
        "creditcard.fill": .credit,
        "globe": .aggregate,
        "aggregate": .aggregate,
        "adjustment": .adjustment,
        "arrow.up.and.down": .adjustment,

        // Existing category/system keys are explicit to keep picker and
        // ledger presentation compatible without changing API values.
        "expense_food": .food,
        "expense_food_snacks": .snacks,
        "expense_food_coffee": .coffee,
        "expense_food_meals": .meals,
        "expense_bills": .bills,
        "expense_shopping": .shopping,
        "expense_family": .family,
        "expense_transport": .transport,
        "expense_health": .health,
        "expense_education": .education,
        "expense_entertainment": .entertainment,
        "expense_gifts": .gifts,
        "expense_insurance": .insurance,
        "expense_investment": .investment,
        "expense_travel": .travel,
        "expense_savings": .savings,
        "expense_business": .business,
        "expense_transfer_out": .transfer,
        "expense_uncategorized": .tag,
        "debt_root": .debt
    ]

    private static let familyAliases: [FamilyAlias] = [
        FamilyAlias(prefix: "expense_food", icon: .food),
        FamilyAlias(prefix: "expense_bills", icon: .bills),
        FamilyAlias(prefix: "expense_shopping", icon: .shopping),
        FamilyAlias(prefix: "expense_family", icon: .family),
        FamilyAlias(prefix: "expense_transport", icon: .transport),
        FamilyAlias(prefix: "expense_health", icon: .health),
        FamilyAlias(prefix: "expense_education", icon: .education),
        FamilyAlias(prefix: "expense_entertainment", icon: .entertainment),
        FamilyAlias(prefix: "expense_gifts", icon: .gifts),
        FamilyAlias(prefix: "expense_insurance", icon: .insurance),
        FamilyAlias(prefix: "expense_investment", icon: .investment),
        FamilyAlias(prefix: "expense_travel", icon: .travel),
        FamilyAlias(prefix: "expense_savings", icon: .savings),
        FamilyAlias(prefix: "expense_business", icon: .business),
        FamilyAlias(prefix: "expense_transfer", icon: .transfer),
        FamilyAlias(prefix: "income_", icon: .income),
        FamilyAlias(prefix: "debt_", icon: .debt)
    ]

    /// Bundled Phosphor Regular asset names. Several semantic roles share a
    /// recognizable source glyph; this keeps category/icon identity stable
    /// without creating a second ad-hoc icon family.
    private static let bundledAssetNames: [MPIconID: String] = [
        .tag: "mp-icon-tag", .home: "mp-icon-house", .ledger: "mp-icon-list",
        .assistant: "mp-icon-sparkle", .account: "mp-icon-users-three", .feedback: "mp-icon-chat-circle",
        .wallet: "mp-icon-wallet", .group: "mp-icon-package", .bankLink: "mp-icon-link",
        .recurring: "mp-icon-repeat", .report: "mp-icon-chart-bar", .search: "mp-icon-magnifying-glass",
        .bell: "mp-icon-bell", .add: "mp-icon-plus", .send: "mp-icon-paper-plane-tilt",
        .food: "mp-icon-fork-knife", .coffee: "mp-icon-coffee", .snacks: "mp-icon-leaf",
        .meals: "mp-icon-bowl-food", .shopping: "mp-icon-shopping-bag", .transport: "mp-icon-car",
        .health: "mp-icon-heart", .bills: "mp-icon-receipt", .education: "mp-icon-graduation-cap",
        .travel: "mp-icon-airplane", .savings: "mp-icon-piggy-bank", .family: "mp-icon-house",
        .entertainment: "mp-icon-game-controller", .gifts: "mp-icon-gift", .insurance: "mp-icon-shield-check",
        .investment: "mp-icon-trend-up", .business: "mp-icon-briefcase", .transfer: "mp-icon-arrows-left-right",
        .income: "mp-icon-bank", .debt: "mp-icon-hand", .credit: "mp-icon-credit-card",
        .aggregate: "mp-icon-globe", .adjustment: "mp-icon-arrows-down-up"
    ]

    // Platform fallbacks remain for a damaged asset catalog or future roles.
    private static let fallbackSystemNames: [MPIconID: String] = [
        .tag: "tag",
        .home: "house",
        .ledger: "list.bullet.rectangle",
        .assistant: "sparkles",
        .account: "person.crop.circle",
        .feedback: "bubble.left.and.bubble.right",
        .wallet: "wallet.bifold",
        .group: "shippingbox",
        .bankLink: "link.circle",
        .recurring: "repeat",
        .report: "chart.bar",
        .search: "magnifyingglass",
        .bell: "bell",
        .add: "plus",
        .send: "paperplane",
        .food: "fork.knife",
        .coffee: "cup.and.saucer",
        .snacks: "leaf",
        .meals: "bowl",
        .shopping: "bag",
        .transport: "car",
        .health: "heart.text.square",
        .bills: "receipt",
        .education: "graduationcap",
        .travel: "airplane",
        .savings: "banknote",
        .family: "house",
        .entertainment: "gamecontroller",
        .gifts: "gift",
        .insurance: "checkmark.shield",
        .investment: "chart.xyaxis.line",
        .business: "briefcase",
        .transfer: "building.columns.fill",
        .income: "banknote",
        .debt: "hand.raised",
        .credit: "creditcard",
        .aggregate: "globe",
        .adjustment: "arrow.up.and.down"
    ]
}
