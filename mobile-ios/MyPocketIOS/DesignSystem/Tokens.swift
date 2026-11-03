import SwiftUI
import UIKit

enum MPColor {
    // Canonical native roles adapt explicitly so SwiftUI and UIKit-backed
    // controls retain the same visual hierarchy in both appearances.
    static let pageCanvas = adaptive(
        light: (246, 245, 250),
        dark: (18, 18, 20)
    )
    static let groupedSurface = adaptive(
        light: (255, 255, 255),
        dark: (28, 28, 30)
    )
    static let separator = adaptive(
        light: (229, 226, 234),
        dark: (58, 58, 60)
    )
    static let primaryText = adaptive(
        light: (27, 28, 28),
        dark: (242, 242, 247)
    )
    static let secondaryText = adaptive(
        light: (85, 97, 88),
        dark: (199, 199, 204)
    )
    static let mutedText = adaptive(
        light: (111, 122, 107),
        dark: (142, 142, 147)
    )
    static let rowFill = adaptive(
        light: (245, 243, 243),
        dark: (44, 44, 46)
    )
    static let primaryAccent = adaptive(
        light: (76, 175, 80),
        dark: (105, 209, 125)
    )
    static let income = adaptive(
        light: (0, 110, 28),
        dark: (117, 214, 138)
    )
    static let expense = adaptive(
        light: (186, 26, 26),
        dark: (255, 180, 171)
    )

    // Compatibility aliases let feature screens migrate independently.
    static let canvas = pageCanvas
    static let surface = groupedSurface
    static let ink = primaryText
    static let secondary = secondaryText
    static let muted = mutedText
    static let line = separator
    static let row = rowFill
    static let heading = primaryText
    static let accent = primaryAccent
    static let successSoft = adaptive(light: (217, 230, 218), dark: (30, 52, 36))
    static let successLine = adaptive(light: (190, 202, 185), dark: (58, 84, 64))
    static let dangerSoft = adaptive(light: (255, 235, 238), dark: (64, 30, 34))
    static let warning = Color(red: 217.0 / 255.0, green: 119.0 / 255.0, blue: 6.0 / 255.0)
    static let warningSoft = adaptive(light: (255, 251, 235), dark: (60, 46, 20))
    static let categoryTeal = Color(red: 13.0 / 255.0, green: 148.0 / 255.0, blue: 136.0 / 255.0)
    static let categoryTealSoft = Color(red: 240.0 / 255.0, green: 253.0 / 255.0, blue: 250.0 / 255.0)
    static let brand = primaryAccent
    static let brandHover = Color(red: 0.0, green: 110.0 / 255.0, blue: 28.0 / 255.0)
    static let action = income
    /// Text/icon colour on filled action/danger surfaces; dark text keeps contrast on the lighter dark-mode accents.
    static let onAction = adaptive(light: (255, 255, 255), dark: (8, 40, 14))
    static let actionHover = Color(red: 0.0, green: 83.0 / 255.0, blue: 19.0 / 255.0)
    static let danger = expense

    private static func adaptive(light: (Int, Int, Int), dark: (Int, Int, Int)) -> Color {
        Color(UIColor { traits in
            let components = traits.userInterfaceStyle == .dark ? dark : light
            return UIColor(
                red: CGFloat(components.0) / 255.0,
                green: CGFloat(components.1) / 255.0,
                blue: CGFloat(components.2) / 255.0,
                alpha: 1
            )
        })
    }
}

enum MPMetrics {
    static let controlHeight: CGFloat = 44
    static let hitTarget: CGFloat = 44
    static let cornerRadius: CGFloat = 12
    // Owner-selected compact action shape; circular icon/feedback controls
    // deliberately use Circle rather than this rectangular-button token.
    static let buttonCornerRadius: CGFloat = 8
    static let referenceCornerRadius: CGFloat = 24
    static let groupedSurfaceCornerRadius: CGFloat = 24
    static let groupedRowMinHeight: CGFloat = 44
    static let groupedRowHorizontalPadding: CGFloat = 16
    // Keep grouped rows compact while preserving the shared 44pt hit target.
    // Screens should use this base value instead of inventing local row rhythm.
    static let groupedRowVerticalPadding: CGFloat = 6
    static let groupedSeparatorLeadingInset: CGFloat = 56
    static let inputHorizontalPadding: CGFloat = 12
    static let inputBorderWidth: CGFloat = 1
    static let inputFocusWidth: CGFloat = 2
    // Stitch/web container margin is 16px. Keep this shared inset at the
    // logical iOS equivalent so every screen aligns without local overrides.
    static let pagePadding: CGFloat = 16
    static let cardPadding: CGFloat = 8
    static let cardHorizontalPadding: CGFloat = 12
    static let cardVerticalPadding: CGFloat = 8
    static let sectionSpacing: CGFloat = 8
    static let rowSpacing: CGFloat = 4
    static let transactionRowVerticalPadding: CGFloat = 6
    static let bottomNavigationHeight: CGFloat = 56
    static let balanceTextSize: CGFloat = 34
}

/// Reference typography shared by native base components. Use the iOS system
/// face so native text metrics, Vietnamese diacritics, Dynamic Type, and the
/// platform chrome match the device reference. Keeping the decision here
/// prevents individual screens from drifting to unrelated font families.
enum MPFont {
    static let display = semantic(.title, weight: .semibold).leading(.tight)
    static let balance = semantic(.largeTitle, weight: .semibold).monospacedDigit()
    static let headline = semantic(.headline, weight: .semibold)
    static let section = semantic(.subheadline, weight: .semibold)
    static let subheadline = semantic(.subheadline, weight: .semibold)
    // Owner preference: keep the compact type scale, but give prose one
    // additional weight step for clearer rendering on device.
    static let body = semantic(.subheadline, weight: .medium)
    static let chatBody = semantic(.subheadline, weight: .medium)
    static let chatCaption = Font.system(.footnote, design: .default).weight(.medium)
    static let bodyEmphasis = semantic(.subheadline, weight: .medium)
    static let bodySmall = semantic(.footnote, weight: .medium)
    static let label = semantic(.caption, weight: .semibold)
    static let labelSmall = semantic(.caption, weight: .medium)
    /// Compact money rows use a distinct role so transaction and report
    /// amounts do not inherit prose sizing or introduce local font choices.
    static let amount = semantic(.callout, weight: .medium).monospacedDigit()
    static let currency = semantic(.title3, weight: .medium).monospacedDigit()

    /// UIKit-backed fields use the same system family as SwiftUI and preserve
    /// the device's native text metrics.
    static let uiBody: UIFont = UIFont.systemFont(ofSize: 16, weight: .medium)

    private static func semantic(_ style: Font.TextStyle, weight: Font.Weight = .regular) -> Font {
        // Use native semantic styles so Dynamic Type remains automatic. The
        // hierarchy is expressed by the selected style and weight, rather
        // than a fixed point size that would stop scaling with the system.
        return Font.system(style, design: .default).weight(weight)
    }
}
