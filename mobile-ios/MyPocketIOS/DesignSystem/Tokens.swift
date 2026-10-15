import SwiftUI
import UIKit

enum MPColor {
    // Keep these values aligned with app/src/ui/theme.css. SwiftUI's system
    // colors vary by appearance and do not preserve web parity.
    static let canvas = Color(red: 251.0 / 255.0, green: 249.0 / 255.0, blue: 249.0 / 255.0)
    static let surface = Color(red: 1.0, green: 1.0, blue: 1.0)
    static let ink = Color(red: 27.0 / 255.0, green: 28.0 / 255.0, blue: 28.0 / 255.0)
    static let secondary = Color(red: 85.0 / 255.0, green: 97.0 / 255.0, blue: 88.0 / 255.0)
    static let muted = Color(red: 111.0 / 255.0, green: 122.0 / 255.0, blue: 107.0 / 255.0)
    static let line = Color(red: 239.0 / 255.0, green: 237.0 / 255.0, blue: 237.0 / 255.0)
    static let row = Color(red: 245.0 / 255.0, green: 243.0 / 255.0, blue: 243.0 / 255.0)
    static let heading = Color(red: 27.0 / 255.0, green: 28.0 / 255.0, blue: 28.0 / 255.0)
    static let accent = Color(red: 76.0 / 255.0, green: 175.0 / 255.0, blue: 80.0 / 255.0)
    static let successSoft = Color(red: 217.0 / 255.0, green: 230.0 / 255.0, blue: 218.0 / 255.0)
    static let successLine = Color(red: 190.0 / 255.0, green: 202.0 / 255.0, blue: 185.0 / 255.0)
    static let dangerSoft = Color(red: 255.0 / 255.0, green: 235.0 / 255.0, blue: 238.0 / 255.0)
    static let warning = Color(red: 217.0 / 255.0, green: 119.0 / 255.0, blue: 6.0 / 255.0)
    static let warningSoft = Color(red: 255.0 / 255.0, green: 251.0 / 255.0, blue: 235.0 / 255.0)
    static let categoryTeal = Color(red: 13.0 / 255.0, green: 148.0 / 255.0, blue: 136.0 / 255.0)
    static let categoryTealSoft = Color(red: 240.0 / 255.0, green: 253.0 / 255.0, blue: 250.0 / 255.0)
    static let brand = Color(red: 76.0 / 255.0, green: 175.0 / 255.0, blue: 80.0 / 255.0)
    static let brandHover = Color(red: 0.0, green: 110.0 / 255.0, blue: 28.0 / 255.0)
    static let action = Color(red: 0.0, green: 110.0 / 255.0, blue: 28.0 / 255.0)
    static let actionHover = Color(red: 0.0, green: 83.0 / 255.0, blue: 19.0 / 255.0)
    static let danger = Color(red: 186.0 / 255.0, green: 26.0 / 255.0, blue: 26.0 / 255.0)
}

enum MPMetrics {
    static let controlHeight: CGFloat = 44
    static let hitTarget: CGFloat = 44
    static let cornerRadius: CGFloat = 12
    // Owner-selected compact action shape; circular icon/feedback controls
    // deliberately use Circle rather than this rectangular-button token.
    static let buttonCornerRadius: CGFloat = 8
    static let referenceCornerRadius: CGFloat = 24
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
    static let bottomNavigationHeight: CGFloat = 56
}

/// Reference typography shared by native base components. The Stitch export
/// names Manrope; the OFL-licensed variable font is bundled in the app
/// resources and remains behind this single base so screens cannot drift to
/// unrelated system faces.
enum MPFont {
    static let display = custom(size: 28, relativeTo: .largeTitle, weight: .bold).leading(.tight)
    static let headline = custom(size: 18, relativeTo: .title2, weight: .semibold)
    static let subheadline = custom(size: 16, relativeTo: .headline, weight: .semibold)
    static let body = custom(size: 15, relativeTo: .body, weight: .medium)
    static let bodyEmphasis = custom(size: 15, relativeTo: .body, weight: .semibold)
    static let bodySmall = custom(size: 13, relativeTo: .footnote, weight: .medium)
    static let label = custom(size: 14, relativeTo: .subheadline, weight: .semibold)
    static let labelSmall = custom(size: 12, relativeTo: .caption, weight: .medium)
    static let currency = custom(size: 20, relativeTo: .title, weight: .semibold).monospacedDigit()

    /// UIKit-backed fields use the same registered family as SwiftUI. The
    /// fallback keeps tests/preview hosts usable before UIAppFonts is loaded.
    static let uiBody: UIFont = UIFont(name: "Manrope", size: 16)
        ?? UIFont(name: "Manrope-Regular", size: 16)
        ?? UIFont.systemFont(ofSize: 16)

    private static func custom(size: CGFloat, relativeTo style: Font.TextStyle, weight: Font.Weight = .regular) -> Font {
        Font.custom("Manrope", size: size, relativeTo: style).weight(weight)
    }
}
