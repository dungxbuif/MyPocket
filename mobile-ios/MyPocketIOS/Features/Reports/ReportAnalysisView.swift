import SwiftUI

struct ReportAnalysisView: View {
    let api: FinanceAPI
    let report: MonthlyReportModel
    let ownerID: String?
    let masked: Bool
    let onChatHandoff: (String) -> Void
    @State private var completed: ReportAnalysisModel?
    @State private var active: ReportAnalysisModel?
    @State private var freshness = "none"
    @State private var loadingLatest = false
    @State private var actionError: String?
    private let cache = try? ReportAnalysisCache.makeDefault()

    private var cacheScope: String { report.month + "|" + (report.walletID ?? "all") + "|" + (report.timezone ?? "unknown") }

    var body: some View {
        MPSurfaceCard(style: .grouped) {
            VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                HStack {
                    MPText("Phân tích AI", style: .emphasis)
                    Spacer(minLength: 8)
                    MPStatusBadge(text: statusLabel, tone: statusTone)
                }

                if let completed {
                    if freshness == "stale" { MPStatusMessage(text: "Có dữ liệu mới. Bài cũ vẫn được giữ để bạn đọc và đối chiếu.", tone: .info) }
                    MPReportAnalysisArticle(analysis: completed, masked: masked)
                } else if loadingLatest {
                    MPLoadingView(label: "Đang tìm bản phân tích đã lưu…")
                } else {
                    MPText("AI giải thích các số liệu đã được tổng hợp. Bài được tạo theo yêu cầu và lưu để lần sau mở lại.", style: .secondary)
                }

                if let active {
                    MPStatusMessage(text: active.status == "queued" ? "Phân tích đang chờ xử lý. Bạn có thể rời màn hình và quay lại sau." : "Đang cập nhật phân tích. Bài cũ vẫn đọc được trong lúc chờ.", tone: .info)
                }
                if let actionError { MPStatusMessage(text: actionError, tone: .danger) }

                HStack(spacing: 10) {
                    if let active {
                        MPButton("Hủy", style: .plain, fullWidth: false) { Task { await cancel(active.id) } }
                    } else if completed == nil || freshness == "stale" {
                        MPButton(completed == nil ? "Phân tích sâu" : "Cập nhật phân tích", style: .secondary, fullWidth: false) { Task { await start() } }
                    }
                    if completed != nil {
                        MPButton("Hỏi tiếp", style: .plain, fullWidth: false) {
                            onChatHandoff("Tôi muốn hỏi tiếp về bài phân tích báo cáo tháng \(report.month).")
                        }
                    }
                }
            }
        }
        .accessibilityIdentifier("report-analysis")
        .task(id: report.snapshotID) { await restore() }
    }

    private var statusLabel: String {
        if active != nil { return active?.status == "queued" ? "Đang xếp hàng" : "Đang cập nhật" }
        if completed != nil { return freshness == "stale" ? "Có dữ liệu mới" : "Bản đã lưu" }
        return "Theo yêu cầu"
    }

    private var statusTone: MPStatusBadge.Tone {
        if actionError != nil { return .danger }
        return completed != nil && freshness == "current" ? .success : .muted
    }

    private func restore() async {
        guard let snapshotID = report.snapshotID, !snapshotID.isEmpty else { return }
        loadingLatest = completed == nil
        actionError = nil
        defer { loadingLatest = false }
        if completed == nil, let ownerID, let cache, let cached = try? await cache.load(ownerID: ownerID, scopeKey: cacheScope) {
            completed = cached.0; freshness = cached.1
        }
        do {
            let latest = try await api.latestReportAnalysis(snapshotID: snapshotID)
            try Task.checkCancellation()
            completed = latest.completed
            active = latest.active
            freshness = latest.freshness
            if let completed, let ownerID, let cache { try? await cache.save(ownerID: ownerID, scopeKey: cacheScope, freshness: freshness, analysis: completed) }
            if let active { await follow(active.id) }
        } catch is CancellationError { }
        catch { actionError = "Chưa kiểm tra được bản phân tích đã lưu. Báo cáo vẫn dùng được bình thường." }
    }

    private func start() async {
        guard active == nil, let snapshotID = report.snapshotID, !snapshotID.isEmpty else { return }
        actionError = nil
        let requestID = UUID().uuidString
        do {
            let accepted = try await api.requestReportAnalysis(ReportAnalysisRequestModel(snapshotID: snapshotID, clientRequestID: requestID))
            active = accepted
            if accepted.status == "completed" {
                completed = accepted; active = nil; freshness = "current"
            } else {
                await follow(accepted.id)
            }
        } catch { actionError = "Chưa thể bắt đầu phân tích. Bài đã lưu, nếu có, vẫn được giữ nguyên." }
    }

    private func follow(_ analysisID: String) async {
        guard !analysisID.isEmpty else { return }
        var delay: UInt64 = 1_000_000_000
        while !Task.isCancelled {
            do {
                try await Task.sleep(nanoseconds: delay)
                let result = try await api.reportAnalysis(requestID: analysisID)
                switch result.status.lowercased() {
                case "completed":
                    completed = result; active = nil; freshness = "current"; actionError = nil
                    if let ownerID, let cache { try? await cache.save(ownerID: ownerID, scopeKey: cacheScope, freshness: freshness, analysis: result) }
                    return
                case "failed":
                    active = nil; actionError = "Máy chủ chưa tạo được bản phân tích mới. Bài cũ vẫn được giữ."; return
                case "cancelled": active = nil; return
                default: active = result
                }
                delay = min(delay * 2, 5_000_000_000)
            } catch is CancellationError { return }
            catch { actionError = "Mất kết nối khi theo dõi. Bạn có thể rời màn hình và quay lại sau."; return }
        }
    }

    private func cancel(_ analysisID: String) async {
        do { try await api.cancelReportAnalysis(requestID: analysisID); active = nil }
        catch { actionError = "Không thể hủy phân tích đang chạy." }
    }
}
