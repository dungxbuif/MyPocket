import SwiftUI

enum MPColor {
    static let canvas = Color(uiColor: .systemGroupedBackground)
    static let surface = Color(uiColor: .secondarySystemGroupedBackground)
    static let ink = Color(uiColor: .label)
    static let secondary = Color(uiColor: .secondaryLabel)
    static let brand = Color(red: 0.298, green: 0.686, blue: 0.216)
    static let action = Color(red: 0.020, green: 0.588, blue: 0.412)
    static let danger = Color(uiColor: .systemRed)
}

enum MPMetrics {
    static let controlHeight: CGFloat = 48
    static let hitTarget: CGFloat = 44
    static let cornerRadius: CGFloat = 12
    static let pagePadding: CGFloat = 20
}
