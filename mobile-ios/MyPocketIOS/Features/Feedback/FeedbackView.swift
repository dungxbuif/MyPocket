import SwiftUI

@MainActor
final class FeedbackViewModel: ObservableObject {
    @Published private(set) var items: [FeedbackModel] = []
    @Published private(set) var changelog: [ChangelogModel] = []
    @Published private(set) var isLoading = false
    @Published var errorMessage: String?
    @Published var title = ""
    @Published var description = ""
    @Published var type = "bug"
    @Published var isSending = false

    private let api: FinanceAPI
    init(api: FinanceAPI) { self.api = api }

    func load() async {
        guard !isLoading else { return }
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }
        do {
            async let feedback = api.feedback()
            async let changelog = api.changelog()
            items = try await feedback
            self.changelog = try await changelog
        } catch { errorMessage = "Không thể tải phản hồi và changelog." }
    }

    func submit() async -> Bool {
        guard !title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty, !description.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            errorMessage = "Nhập tiêu đề và mô tả trước khi gửi."
            return false
        }
        isSending = true
        defer { isSending = false }
        do {
            let item = try await api.createFeedback(FeedbackInputModel(type: type, title: title, description: description))
            items.insert(item, at: 0)
            title = ""
            description = ""
            return true
        } catch {
            errorMessage = "Không thể gửi phản hồi."
            return false
        }
    }
}

struct FeedbackOverlay: View {
    let api: FinanceAPI
    let onClose: () -> Void
    @StateObject private var model: FeedbackViewModel

    init(api: FinanceAPI, onClose: @escaping () -> Void) {
        self.api = api
        self.onClose = onClose
        _model = StateObject(wrappedValue: FeedbackViewModel(api: api))
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            MPScreenHeader(title: "Phản hồi", leading: { EmptyView() }, trailing: { MPIconButton(systemName: "xmark", label: "Đóng", action: onClose) })
            MPWalletSelector(title: "Loại", options: ["bug", "feature", "improvement"], selection: $model.type)
            MPTextField(title: "Tiêu đề", text: $model.title, placeholder: "Mô tả ngắn")
            MPTextEditor(title: "Chi tiết", text: $model.description, placeholder: "Điều gì đã xảy ra?")
            if let errorMessage = model.errorMessage { MPStatusMessage(text: errorMessage, tone: .danger) }
            MPButton("Gửi phản hồi", isLoading: model.isSending) {
                Task { if await model.submit() { onClose() } }
            }
        }
        .padding(16)
        .background(MPColor.surface, in: RoundedRectangle(cornerRadius: MPMetrics.cornerRadius))
        .shadow(radius: 12)
        .accessibilityIdentifier("feedback-overlay")
    }
}

struct FeedbackView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @StateObject private var model: FeedbackViewModel

    init(api: FinanceAPI, onBack: @escaping () -> Void) {
        self.api = api
        self.onBack = onBack
        _model = StateObject(wrappedValue: FeedbackViewModel(api: api))
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Phản hồi & changelog", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                MPText("Gửi lỗi hoặc đề xuất; trạng thái sẽ được cập nhật theo changelog.", style: .secondary)
                MPWalletSelector(title: "Loại", options: ["bug", "feature", "improvement"], selection: $model.type)
                MPTextField(title: "Tiêu đề", text: $model.title, placeholder: "Mô tả ngắn")
                MPTextEditor(title: "Chi tiết", text: $model.description, placeholder: "Điều gì đã xảy ra?")
                if let errorMessage = model.errorMessage { MPStatusMessage(text: errorMessage, tone: .danger) }
                MPButton("Gửi phản hồi", isLoading: model.isSending) { Task { _ = await model.submit() } }
                MPAssistantCard(title: "Phản hồi của bạn") {
                    if model.items.isEmpty { MPEmptyView(title: "Chưa có phản hồi", detail: "Các phản hồi bạn gửi sẽ hiển thị tại đây.") }
                    ForEach(model.items) { item in
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 6) {
                                MPText(item.title)
                                MPText("\(item.type) · \(item.status)", style: .secondary)
                                if let fixedAt = item.fixedAt { MPText("Đã xử lý: \(MPFormat.shortDate(fixedAt))", style: .secondary) }
                            }
                        }
                    }
                }
                MPAssistantCard(title: "Changelog") {
                    ForEach(model.changelog) { item in
                        MPText("v\(item.version) — \(item.title)")
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await model.load() }
        .accessibilityIdentifier("feedback-screen")
    }
}
