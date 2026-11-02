import SwiftUI

private struct ReportAssistantHandoffKey: EnvironmentKey {
    static var defaultValue: ((ReportAssistantHandoff) -> Void)? { nil }
}

extension EnvironmentValues {
    var reportAssistantHandoff: ((ReportAssistantHandoff) -> Void)? {
        get { self[ReportAssistantHandoffKey.self] }
        set { self[ReportAssistantHandoffKey.self] = newValue }
    }
}

struct ReportAnalysisView: View {
    let api: FinanceAPI
    let report: MonthlyReportModel
    let ownerID: String?
    let masked: Bool
    let walletLabel: String
    let onOpenEvidence: (String?, ReportInsightModel) -> Void
    @Environment(\.reportAssistantHandoff) private var onChatHandoff
    @State private var completed: ReportAnalysisModel?
    @State private var active: ReportAnalysisModel?
    @State private var freshness = "none"
    @State private var loadingLatest = false
    @State private var starting = false
    @State private var actionError: String?
    @State private var operation = UUID()
    private let cache = try? ReportAnalysisCache.makeDefault()

    private var cacheScope: String { report.month + "|" + (report.walletID ?? "all") + "|" + (report.timezone ?? "unknown") }
    private var loadKey: String { (ownerID ?? "") + "|" + cacheScope + "|" + (report.snapshotID ?? "") }

    var body: some View {
        MPSurfaceCard(style: .grouped) {
            VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                HStack {
                    MPText("Phân tích AI", style: .emphasis)
                    Spacer(minLength: 8)
                    MPStatusBadge(text: statusLabel, tone: statusTone)
                }
                if let completed {
                    if freshness == "stale" {
                        MPStatusMessage(text: "Dữ liệu hoặc cách phân tích đã thay đổi. Bài cũ vẫn được giữ để bạn đối chiếu; hãy cập nhật để xem nhận xét mới.", tone: .info)
                    } else if freshness == "unverified" {
                        MPStatusMessage(text: "Chưa kiểm tra được độ mới của bản đã lưu.", tone: .info)
                    }
                    MPReportAnalysisArticle(analysis: completed, masked: masked, onOpen: { item in
                        onOpenEvidence(completed.snapshotID, item)
                    }, onAsk: { item in handoff(completed, finding: item) })
                } else if loadingLatest {
                    MPLoadingView(label: "Đang tìm bản phân tích đã lưu…")
                } else {
                    MPText("Phân tích các số liệu đã được tổng hợp theo yêu cầu. Mở bản đã lưu không tạo phân tích mới.", style: .secondary)
                }
                if let active {
                    MPStatusMessage(text: active.status == "queued" ? "Phân tích đang chờ xử lý. Bạn có thể rời màn hình và quay lại sau." : "Đang cập nhật phân tích. Bài cũ vẫn đọc được trong lúc chờ.", tone: .info)
                }
                if let actionError { MPStatusMessage(text: actionError, tone: .danger) }
                HStack(spacing: MPMetrics.sectionSpacing) {
                    if let active {
                        MPButton("Hủy", style: .plain, fullWidth: false) { Task { await cancel(active.id) } }
                    } else if completed == nil || freshness != "current" {
                        MPButton(completed == nil ? "Phân tích sâu" : "Cập nhật phân tích", style: .secondary, isLoading: starting, fullWidth: false) { Task { await start() } }
                            .disabled(loadingLatest || starting || report.snapshotID == nil)
                    }
                    if let completed, completed.observations.isEmpty && completed.suggestions.isEmpty {
                        MPButton("Hỏi tiếp", style: .plain, fullWidth: false) { handoff(completed, finding: nil) }
                    }
                }
            }
        }
        .accessibilityIdentifier("report-analysis")
        .task(id: loadKey) { await restore() }
        .onDisappear { operation = UUID() }
    }

    private var statusLabel: String {
        if active != nil || starting { return active?.status == "queued" ? "Đang xếp hàng" : "Đang cập nhật" }
        if completed != nil { return freshness == "stale" ? "Bản đã lưu · Cần cập nhật" : "Bản đã lưu" }
        return "Theo yêu cầu"
    }

    private var statusTone: MPStatusBadge.Tone {
        if actionError != nil { return .danger }
        return completed != nil && freshness == "current" ? .success : .muted
    }

    private func isCurrent(_ token: UUID) -> Bool { operation == token && !Task.isCancelled }

    private func handoff(_ analysis: ReportAnalysisModel, finding: ReportInsightModel?) {
        guard let onChatHandoff,
              let context = ReportAssistantHandoff.make(analysis: analysis, finding: finding, month: report.month, walletLabel: walletLabel) else {
            actionError = "Bản cũ chưa có đủ ngữ cảnh để hỏi tiếp. Hãy tải lại báo cáo và cập nhật phân tích."
            return
        }
        onChatHandoff(context)
    }

    private func restore() async {
        let token = UUID()
        operation = token
        completed = nil; active = nil; freshness = "none"; actionError = nil; starting = false
        guard let snapshotID = report.snapshotID, !snapshotID.isEmpty else { loadingLatest = false; return }
        loadingLatest = true
        defer { if isCurrent(token) { loadingLatest = false } }
        if let ownerID, let cache, let cached = try? await cache.load(ownerID: ownerID, scopeKey: cacheScope) {
            guard isCurrent(token) else { return }
            completed = cached.0
            freshness = cached.0.hasCurrentProvenance ? "unverified" : "stale"
        }
        do {
            let latest = try await api.latestReportAnalysis(snapshotID: snapshotID)
            guard isCurrent(token) else { return }
            completed = latest.completed; active = latest.active
            freshness = latest.completed.map { $0.hasCurrentProvenance ? latest.freshness : "stale" } ?? latest.freshness
            loadingLatest = false
            if let completed { await save(completed, freshness: freshness) }
            guard isCurrent(token) else { return }
            if let active { await follow(active.id, token: token) }
        } catch {
            guard isCurrent(token) else { return }
            actionError = "Chưa kiểm tra được bản phân tích đã lưu. Báo cáo vẫn dùng được bình thường."
        }
    }

    private func start() async {
        guard active == nil, !starting, !loadingLatest, let snapshotID = report.snapshotID, !snapshotID.isEmpty else { return }
        let token = UUID()
        operation = token
        starting = true; actionError = nil
        defer { if isCurrent(token) { starting = false } }
        do {
            let accepted = try await api.requestReportAnalysis(ReportAnalysisRequestModel(snapshotID: snapshotID, clientRequestID: UUID().uuidString))
            guard isCurrent(token) else { return }
            active = accepted
            starting = false
            if accepted.status == "completed" {
                // POST intentionally returns only an acknowledgement, even when
                // the no-data result has already completed. Always fetch its body.
                let full = try await api.reportAnalysis(requestID: accepted.id)
                guard isCurrent(token) else { return }
                if full.status == "completed" { await finish(full, token: token) }
                else { active = full; await follow(full.id, token: token) }
            } else { await follow(accepted.id, token: token) }
        } catch {
            guard isCurrent(token) else { return }
            active = nil
            actionError = "Chưa tải được phân tích mới. Bản đã lưu vẫn được giữ; hãy tải lại báo cáo để kiểm tra yêu cầu."
        }
    }

    private func follow(_ analysisID: String, token: UUID) async {
        guard !analysisID.isEmpty else { return }
        var delay: UInt64 = 1_000_000_000
        while isCurrent(token) {
            do {
                try await Task.sleep(nanoseconds: delay)
                guard isCurrent(token) else { return }
                let result = try await api.reportAnalysis(requestID: analysisID)
                guard isCurrent(token) else { return }
                switch result.status.lowercased() {
                case "completed": await finish(result, token: token); return
                case "failed":
                    active = nil; actionError = "Máy chủ chưa tạo được bản phân tích mới. Bài cũ vẫn được giữ."; return
                case "cancelled": active = nil; return
                default: active = result
                }
                delay = min(delay * 2, 5_000_000_000)
            } catch {
                guard isCurrent(token) else { return }
                actionError = "Mất kết nối khi theo dõi. Hãy tải lại báo cáo để tiếp tục kiểm tra."; return
            }
        }
    }

    private func finish(_ result: ReportAnalysisModel, token: UUID) async {
        guard isCurrent(token) else { return }
        completed = result; active = nil; actionError = nil; freshness = "unverified"
        // Freshness is a server decision: an active job can have saved older
        // evidence while the report has already changed.
        if let snapshotID = report.snapshotID,
           let latest = try? await api.latestReportAnalysis(snapshotID: snapshotID) {
            guard isCurrent(token) else { return }
            if latest.completed?.id == result.id { freshness = latest.freshness }
        }
        guard isCurrent(token) else { return }
        if !result.hasCurrentProvenance { freshness = "stale" }
        await save(result, freshness: freshness)
    }

    private func save(_ analysis: ReportAnalysisModel, freshness: String) async {
        if let ownerID, let cache {
            try? await cache.save(ownerID: ownerID, scopeKey: cacheScope, freshness: freshness, analysis: analysis)
        }
    }

    private func cancel(_ analysisID: String) async {
        let token = operation
        do {
            try await api.cancelReportAnalysis(requestID: analysisID)
            guard isCurrent(token), active?.id == analysisID else { return }
            operation = UUID(); active = nil; starting = false
        } catch {
            guard isCurrent(token) else { return }
            actionError = "Không thể hủy phân tích đang chạy."
        }
    }
}
