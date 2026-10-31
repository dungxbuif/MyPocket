import SwiftUI

struct AssistantActionCard: View {
    let title: String
    let summary: String
    let state: AssistantEntryRowState
    let onPrimary: () -> Void
    let onSecondary: () -> Void

    var body: some View {
        MPSurfaceCard(style: .reference) {
            VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                HStack(alignment: .firstTextBaseline) {
                    MPText(title, style: .emphasis)
                    Spacer(minLength: MPMetrics.rowSpacing)
                    MPStatusBadge(text: state.label, tone: state == .applied ? .success : state == .invalid || state == .conflict ? .danger : .muted)
                }
                MPText(summary, style: .secondary)
                if state != .applied && state != .rejected {
                    HStack(spacing: MPMetrics.sectionSpacing) {
                        MPButton("Ghi giao dịch", style: .secondary, fullWidth: false, action: onPrimary)
                        MPButton("Bỏ qua", style: .plain, fullWidth: false, action: onSecondary)
                    }
                }
            }
        }
    }
}
