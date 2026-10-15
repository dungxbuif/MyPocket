import SwiftUI

enum MPTextStyle {
    case display
    case title
    case body
    case secondary
    case emphasis
    case action
    case danger
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
            .foregroundStyle(color)
    }

    private var font: Font {
        switch style {
        case .display:
            return MPFont.display
        case .title:
            return MPFont.headline
        case .body:
            return MPFont.body
        case .secondary:
            return MPFont.bodySmall
        case .emphasis:
            return MPFont.bodyEmphasis
        case .action, .danger:
            return MPFont.bodySmall
        }
    }

    private var color: Color {
        switch style {
        case .secondary: return MPColor.secondary
        case .action: return MPColor.action
        case .danger: return MPColor.danger
        default: return MPColor.ink
        }
    }
}

/// Safe, shared renderer for assistant prose. The model can only provide
/// Markdown text; this component turns the bounded subset into native SwiftUI
/// primitives and never evaluates HTML or view code from the model.
enum MPMarkdownBlock: Equatable {
    case heading(level: Int, text: String)
    case paragraph(String)
    case list(ordered: Bool, items: [String])
    case table(headers: [String], rows: [[String]])
    case code(String)
}

enum MPMarkdownParser {
    static func parse(_ content: String) -> [MPMarkdownBlock] {
        let lines = content.replacingOccurrences(of: "\r\n", with: "\n")
            .replacingOccurrences(of: "\r", with: "\n")
            .components(separatedBy: "\n")
        let fence = String(UnicodeScalar(96)!) + String(UnicodeScalar(96)!) + String(UnicodeScalar(96)!)
        var blocks: [MPMarkdownBlock] = []
        var index = 0

        while index < lines.count {
            let line = lines[index].trimmingCharacters(in: .whitespaces)
            if line.isEmpty { index += 1; continue }

            if line.hasPrefix(fence) {
                var code: [String] = []
                index += 1
                while index < lines.count && !lines[index].trimmingCharacters(in: .whitespaces).hasPrefix(fence) {
                    code.append(lines[index])
                    index += 1
                }
                if index < lines.count { index += 1 }
                blocks.append(.code(code.joined(separator: "\n")))
                continue
            }

            if let heading = headingParts(line) {
                blocks.append(.heading(level: min(3, heading.level), text: heading.text))
                index += 1
                continue
            }

            if line.hasPrefix("|"), index + 1 < lines.count, isTableSeparator(lines[index + 1]) {
                let headers = splitTableRow(line)
                var rows: [[String]] = []
                index += 2
                while index < lines.count && lines[index].trimmingCharacters(in: .whitespaces).hasPrefix("|") {
                    rows.append(splitTableRow(lines[index]))
                    index += 1
                }
                blocks.append(.table(headers: headers, rows: rows))
                continue
            }

            if let first = listPart(line) {
                var items = [first.item]
                let ordered = first.ordered
                index += 1
                while index < lines.count, let next = listPart(lines[index].trimmingCharacters(in: .whitespaces)), next.ordered == ordered {
                    items.append(next.item)
                    index += 1
                }
                blocks.append(.list(ordered: ordered, items: items))
                continue
            }

            var paragraph = [line]
            index += 1
            while index < lines.count {
                let next = lines[index].trimmingCharacters(in: .whitespaces)
                guard !next.isEmpty, headingParts(next) == nil, listPart(next) == nil, !next.hasPrefix("|") else { break }
                paragraph.append(next)
                index += 1
            }
            blocks.append(.paragraph(paragraph.joined(separator: " ")))
        }
        return blocks
    }

    private static func headingParts(_ line: String) -> (level: Int, text: String)? {
        let parts = line.split(separator: " ", maxSplits: 1).map(String.init)
        guard let marker = parts.first, marker.allSatisfy({ $0 == "#" }), (1...4).contains(marker.count), parts.count == 2 else { return nil }
        return (marker.count, parts[1])
    }

    private static func listPart(_ line: String) -> (ordered: Bool, item: String)? {
        if let match = line.range(of: #"^[-*]\s+"#, options: .regularExpression) {
            return (false, String(line[match.upperBound...]))
        }
        if let match = line.range(of: #"^\d+[.)]\s+"#, options: .regularExpression) {
            return (true, String(line[match.upperBound...]))
        }
        return nil
    }

    private static func isTableSeparator(_ line: String) -> Bool {
        splitTableRow(line).allSatisfy { cell in
            let value = cell.trimmingCharacters(in: .whitespaces)
            return value.range(of: #"^:?-{3,}:?$"#, options: .regularExpression) != nil
        }
    }

    private static func splitTableRow(_ line: String) -> [String] {
        var value = line.trimmingCharacters(in: .whitespaces)
        if value.hasPrefix("|") { value.removeFirst() }
        if value.hasSuffix("|") { value.removeLast() }
        return value.split(separator: "|", omittingEmptySubsequences: false).map { $0.trimmingCharacters(in: .whitespaces) }
    }
}

struct MPMarkdown: View {
    let content: String

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            ForEach(Array(MPMarkdownParser.parse(content).enumerated()), id: \.offset) { _, block in
                blockView(block)
            }
        }
    }

    @ViewBuilder
    private func blockView(_ block: MPMarkdownBlock) -> some View {
        switch block {
        case .heading(let level, let text):
            Text(markdownText(text))
                .font(level == 1 ? MPFont.headline : MPFont.subheadline)
                .foregroundStyle(MPColor.ink)
        case .paragraph(let text):
            Text(markdownText(text))
                .font(MPFont.body)
                .foregroundStyle(MPColor.ink)
                .fixedSize(horizontal: false, vertical: true)
        case .list(let ordered, let items):
            VStack(alignment: .leading, spacing: 4) {
                ForEach(Array(items.enumerated()), id: \.offset) { index, item in
                    HStack(alignment: .top, spacing: 6) {
                        Text(ordered ? "\(index + 1)." : "•")
                            .font(MPFont.bodyEmphasis)
                            .foregroundStyle(MPColor.secondary)
                        Text(markdownText(item))
                            .font(MPFont.body)
                            .foregroundStyle(MPColor.ink)
                            .fixedSize(horizontal: false, vertical: true)
                    }
                }
            }
        case .table(let headers, let rows):
            ScrollView(.horizontal, showsIndicators: false) {
                Grid(alignment: .leading, horizontalSpacing: 12, verticalSpacing: 0) {
                    GridRow {
                        ForEach(Array(headers.enumerated()), id: \.offset) { _, header in
                            Text(markdownText(header)).font(MPFont.labelSmall).foregroundStyle(MPColor.secondary)
                                .padding(.vertical, 6)
                        }
                    }
                    ForEach(Array(rows.enumerated()), id: \.offset) { _, row in
                        GridRow {
                            ForEach(0..<headers.count, id: \.self) { column in
                                Text(markdownText(column < row.count ? row[column] : ""))
                                    .font(MPFont.bodySmall)
                                    .foregroundStyle(MPColor.ink)
                                    .padding(.vertical, 6)
                            }
                        }
                        Divider().gridCellUnsizedAxes(.horizontal)
                    }
                }
                .padding(.horizontal, 8)
                .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
            }
        case .code(let text):
            ScrollView(.horizontal, showsIndicators: false) {
                Text(text)
                    .font(MPFont.bodySmall.monospaced())
                    .foregroundStyle(MPColor.ink)
                    .padding(8)
            }
            .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
        }
    }

    private func markdownText(_ value: String) -> AttributedString {
        (try? AttributedString(markdown: value)) ?? AttributedString(value)
    }
}
