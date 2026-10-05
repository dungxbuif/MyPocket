import SwiftUI

enum MPTextStyle {
    case title
    case body
    case secondary
}

struct MPText: View {
    private let value: String
    private let style: MPTextStyle

    init(_ value: String, style: MPTextStyle = .body) {
        self.value = value
        self.style = style
    }

    var body: some View {
        Text(value)
            .font(font)
            .foregroundStyle(MPColor.ink)
    }

    private var font: Font {
        switch style {
        case .title:
            return .title2.weight(.bold)
        case .body:
            return .body
        case .secondary:
            return .subheadline
        }
    }
}
