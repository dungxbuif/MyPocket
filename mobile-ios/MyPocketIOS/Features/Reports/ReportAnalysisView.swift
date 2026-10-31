import SwiftUI

struct ReportAnalysisView: View {
    let api: FinanceAPI
    let report: MonthlyReportModel
    let ownerID: String?
    let masked: Bool
    let onChatHandoff: (String) -> Void
    @State private var state: ReportAnalysisState = .idle

    var body: some View {
        MPSurfaceCard(style: .grouped) {
            VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                HStack {
                    MPText("Phân tích sâu", style: .emphasis)
                    Spacer(minLength: MPMetrics.rowSpacing)
                    MPStatusBadge(text: statusLabel, tone: statusTone)
                }
                switch state {
                case .completed(_, let revision, let article):
                    if let revision, let reportRevision = report.dataRevision, revision != reportRevision {
                        MPStatusMessage(text: "Bài phân tích thuộc phiên bản cũ; số liệu mới vẫn là nguồn chính.", tone: .info)
                    }
                    MPText(masked ? AssistantPrivacyPolicy.maskedText : article, style: .body)
                case .failed(let message):
                    MPStatusMessage(text: message, tone: .danger)
                case .expired:
                    MPStatusMessage(text: "Bài phân tích đã hết hạn. Hãy yêu cầu lại.", tone: .info)
                case .queued, .running:
                    MPLoadingView(label: "Đang tạo phân tích…")
                case .cancelled:
                    MPText("Đã hủy phân tích.", style: .secondary)
                case .idle:
                    MPText("Phân tích được tạo theo yêu cầu, không tự động chạy khi mở báo cáo.", style: .secondary)
                }
                HStack(spacing: MPMetrics.sectionSpacing) {
                    if isWorking {
                        MPButton("Hủy", style: .plain, fullWidth: false) { Task { await cancel() } }
                    } else {
                        MPButton(actionTitle, style: .secondary, fullWidth: false) { Task { await start() } }
                    }
                    MPButton("Hỏi trợ lý", style: .plain, fullWidth: false) {
                        onChatHandoff("Hãy phân tích báo cáo tháng \(report.month) dựa trên các số liệu đang hiển thị.")
                    }
                }
            }
        }
        .accessibilityIdentifier("report-analysis")
    }

    private var isWorking: Bool {
        switch state { case .queued, .running: return true; default: return false }
    }

    private var actionTitle: String {
        switch state { case .completed, .failed, .expired: return "Tạo lại phân tích"; default: return "Phân tích sâu" }
    }

    private var statusLabel: String {
        switch state {
        case .idle: return "Theo yêu cầu"
        case .queued: return "Đang xếp hàng"
        case .running: return "Đang chạy"
        case .completed: return "Sẵn sàng"
        case .failed: return "Lỗi"
        case .cancelled: return "Đã hủy"
        case .expired: return "Hết hạn"
        }
    }

    private var statusTone: MPStatusBadge.Tone {
        switch state {
        case .completed: return .success
        case .failed: return .danger
        case .queued, .running: return .muted
        default: return .muted
        }
    }

    private func start() async {
        guard !isWorking else { return }
        let requestID = UUID().uuidString
        state = .queued(requestID: requestID)
        do {
            guard let snapshotID = report.snapshotID, !snapshotID.isEmpty else {
                state = .failed(message: "Báo cáo hiện tại chưa có bản chụp để phân tích.")
                return
            }
            let accepted = try await api.requestReportAnalysis(ReportAnalysisRequestModel(snapshotID: snapshotID, clientRequestID: requestID))
            let analysisID = accepted.id.isEmpty ? requestID : accepted.id
            state = .running(requestID: analysisID)
            for _ in 0..<30 {
                try await Task.sleep(for: .seconds(1))
                let result = try await api.reportAnalysis(requestID: analysisID)
                switch result.status.lowercased() {
                case "completed":
                    state = .completed(requestID: analysisID, revision: result.revision, article: result.renderedArticle.isEmpty ? "Chưa có bài phân tích." : result.renderedArticle)
                    return
                case "expired": state = .expired; return
                case "cancelled": state = .cancelled; return
                case "failed": state = .failed(message: "Máy chủ chưa tạo được phân tích."); return
                default: state = .running(requestID: analysisID)
                }
            }
            state = .failed(message: "Phân tích đang mất nhiều thời gian hơn dự kiến. Bạn có thể tải lại sau.")
        } catch {
            state = .failed(message: "Chưa thể tạo phân tích sâu trên máy chủ hiện tại.")
        }
    }

    private func cancel() async {
        let requestID: String?
        switch state {
        case .queued(let id), .running(let id): requestID = id
        default: requestID = nil
        }
        guard let requestID else { return }
        do {
            try await api.cancelReportAnalysis(requestID: requestID)
            state = .cancelled
        } catch {
            state = .failed(message: "Không thể hủy phân tích đang chạy.")
        }
    }
}
