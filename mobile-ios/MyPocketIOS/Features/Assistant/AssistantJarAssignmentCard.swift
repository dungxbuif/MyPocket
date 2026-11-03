import SwiftUI

/// Network seams for the `jar_assignment_confirmation` card. Nothing is applied
/// unless the user taps "Áp dụng"; edits go through the replace route first.
struct AssistantJarAssignmentActions {
    var loadJars: (String) async throws -> [MPSelectionOption]
    var replace: (String, JarAssignmentEditBody) async throws -> JSONValue
    var confirm: (String, String, Int64) async throws -> JarAssignmentConfirmModel
}

enum AssistantJarAssignmentPresentation {
    static let unassignID = ""

    /// Rows persisted on an already applied card (`assignment_results`).
    static func decodeResults(_ view: [String: JSONValue]) -> [JarAssignmentResultModel] {
        guard case .array(let items)? = view["assignment_results"] else { return [] }
        return items.compactMap { item in
            guard case .object(let o) = item, case .string(let id)? = o["transaction_id"], case .string(let status)? = o["status"] else { return nil }
            var detail: String?
            if case .string(let d)? = o["detail"], !d.isEmpty { detail = d }
            return JarAssignmentResultModel(transactionID: id, status: status, detail: detail)
        }
    }

    static func revision(_ view: [String: JSONValue]) -> Int64 {
        if case .number(let n)? = view["revision"], let v = Int64(exactly: n), v > 0 { return v }
        return 1
    }

    /// The new confirmable action id returned by the replace route.
    static func replacedActionID(_ value: JSONValue) -> String? {
        guard case .object(let root) = value else { return nil }
        if case .string(let id)? = root["action_id"], !id.isEmpty { return id }
        if case .object(let inner)? = root["view"], case .string(let id)? = inner["action_id"], !id.isEmpty { return id }
        return nil
    }

    /// Edit payload for the surviving proposed rows with the user's jar choices.
    static func editBody(rows: [JarAssignmentRow], jarByTransaction: [String: String], conversationID: String, generation: Int64, revision: Int64) -> JarAssignmentEditBody {
        JarAssignmentEditBody(
            conversationID: conversationID, generation: generation, revision: revision,
            patch: .init(assignments: rows.map { .init(transactionID: $0.transactionID, jarID: jarByTransaction[$0.transactionID] ?? $0.jarID) }))
    }
}

struct AssistantJarAssignmentCardBody: View {
    @Environment(\.timeZone) private var timeZone
    let view: [String: JSONValue]
    let masked: Bool
    let actions: AssistantJarAssignmentActions?

    @State private var jarChoice: [String: String] = [:]
    @State private var removed: Set<String> = []
    @State private var jars: [MPSelectionOption] = []
    @State private var results: [JarAssignmentResultModel]?
    @State private var summary: String?
    @State private var saving = false
    @State private var dismissed = false
    @State private var errorText: String?

    private var rows: [JarAssignmentRow] { JarAssignmentRow.decode(view) ?? [] }
    private var month: String { if case .string(let m)? = view["month"] { return m }; return "" }
    private var persistedResults: [JarAssignmentResultModel] { AssistantJarAssignmentPresentation.decodeResults(view) }
    private var isApplied: Bool { AssistantPersistedActionState.decode(view) == .applied }
    private var shownResults: [JarAssignmentResultModel]? { results ?? (isApplied && !persistedResults.isEmpty ? persistedResults : nil) }
    private var activeRows: [JarAssignmentRow] { rows.filter { !$0.rejected && !removed.contains($0.transactionID) } }

    var body: some View {
        if rows.isEmpty {
            MPStatusMessage(text: "Đề xuất gán hũ chưa đủ dữ liệu để xác nhận.", tone: .danger)
        } else {
            VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                ForEach(rows.filter { !removed.contains($0.transactionID) || shownResults != nil }) { row in rowView(row) }
                footer
            }
            .task(id: month) {
                guard jars.isEmpty, !month.isEmpty, let actions else { return }
                jars = (try? await actions.loadJars(month)) ?? []
            }
        }
    }

    @ViewBuilder private func rowView(_ row: JarAssignmentRow) -> some View {
        let result = shownResults?.first { $0.transactionID == row.transactionID }
        VStack(alignment: .leading, spacing: 6) {
            HStack(alignment: .firstTextBaseline) {
                MPText(row.note.isEmpty ? "Khoản chi" : row.note, style: .emphasis)
                Spacer(minLength: MPMetrics.rowSpacing)
                MPText(MPFormat.money(row.amount, masked: masked), style: .emphasis)
            }
            MPText([dateText(row.occurredAt), row.categoryName].filter { !$0.isEmpty }.joined(separator: " · "), style: .secondary)
            if row.rejected {
                MPStatusBadge(text: "Không đề xuất", tone: .muted)
                MPText(row.rejectReason.isEmpty ? row.reason : row.rejectReason, style: .secondary)
            } else {
                if !row.reason.isEmpty { MPText(row.reason, style: .secondary) }
                if let result {
                    let badge = result.badge
                    MPStatusBadge(text: badge.text, tone: badge.tone)
                    if let detail = result.detail { MPText(detail, style: .secondary) }
                    MPText("Hũ: \(chosenName(row))", style: .secondary)
                } else if !isApplied && !dismissed {
                    MPOptionSelector(title: "Hũ", options: [MPSelectionOption(id: AssistantJarAssignmentPresentation.unassignID, label: "Bỏ gán hũ")] + jarOptions(row),
                                     selectionID: Binding(get: { jarChoice[row.transactionID] ?? row.jarID }, set: { jarChoice[row.transactionID] = $0 }),
                                     accessibilityIdentifier: "assistant-jar-assignment-selector", isEnabled: !saving)
                    MPButton("Bỏ khoản này", style: .plain, isDisabled: saving, fullWidth: false) { removed.insert(row.transactionID) }
                } else {
                    MPText("Hũ đề xuất: \(chosenName(row))", style: .secondary)
                }
            }
        }
    }

    @ViewBuilder private var footer: some View {
        if let summary {
            MPStatusMessage(text: summary, tone: .success)
        } else if isApplied, shownResults == nil {
            MPStatusMessage(text: "Đã áp dụng đề xuất gán hũ.", tone: .success)
        } else if dismissed || AssistantPersistedActionState.decode(view) == .unavailable {
            MPStatusMessage(text: dismissed ? "Đã bỏ qua đề xuất." : "Đề xuất này không còn hiệu lực. Hãy yêu cầu trợ lý tạo đề xuất mới.", tone: .info)
        } else {
            if let errorText { MPStatusMessage(text: errorText, tone: .danger) }
            HStack(spacing: MPMetrics.sectionSpacing) {
                MPButton("Áp dụng", style: .primary, isLoading: saving, isDisabled: saving || activeRows.isEmpty || actions == nil, fullWidth: false) { Task { await apply() } }
                    .accessibilityIdentifier("assistant-confirm-jar-assignment")
                MPButton("Bỏ qua", style: .plain, isDisabled: saving, fullWidth: false) { dismissed = true }
            }
        }
    }

    private func jarOptions(_ row: JarAssignmentRow) -> [MPSelectionOption] {
        if row.jarID.isEmpty || jars.contains(where: { $0.id == row.jarID }) { return jars }
        return jars + [MPSelectionOption(id: row.jarID, label: row.jarName.isEmpty ? "Hũ đề xuất" : row.jarName)]
    }

    private func chosenName(_ row: JarAssignmentRow) -> String {
        let id = jarChoice[row.transactionID] ?? row.jarID
        if id.isEmpty { return "Bỏ gán hũ" }
        return jars.first { $0.id == id }?.label ?? (row.jarName.isEmpty ? "Hũ đã chọn" : row.jarName)
    }

    private func dateText(_ iso: String) -> String {
        guard let date = ISO8601DateFormatter().date(from: iso) else { return "" }
        let formatter = DateFormatter()
        formatter.locale = Locale(identifier: "vi_VN")
        formatter.timeZone = timeZone
        formatter.dateFormat = "dd/MM"
        return formatter.string(from: date)
    }

    private func apply() async {
        guard !saving, let actions, !activeRows.isEmpty,
              let actionID = AssistantJarProposal.actionID(view), let context = AssistantJarProposal.actionContext(view) else {
            errorText = "Đề xuất gán hũ chưa đủ dữ liệu để xác nhận."
            return
        }
        saving = true
        errorText = nil
        defer { saving = false }
        do {
            var confirmID = actionID
            let edited = !removed.isEmpty || activeRows.contains { (jarChoice[$0.transactionID] ?? $0.jarID) != $0.jarID }
            if edited {
                let body = AssistantJarAssignmentPresentation.editBody(rows: activeRows, jarByTransaction: jarChoice, conversationID: context.conversationID, generation: context.generation, revision: AssistantJarAssignmentPresentation.revision(view))
                let replaced = try await actions.replace(actionID, body)
                guard let newID = AssistantJarAssignmentPresentation.replacedActionID(replaced) else {
                    errorText = "Không thể lưu chỉnh sửa. Bạn có thể thử lại."
                    return
                }
                confirmID = newID
            }
            let response = try await actions.confirm(confirmID, context.conversationID, context.generation)
            results = response.results
            summary = response.summary
        } catch let error as APIError {
            errorText = error.userMessage
        } catch {
            errorText = "Không thể áp dụng đề xuất gán hũ. Bạn có thể thử lại."
        }
    }
}
