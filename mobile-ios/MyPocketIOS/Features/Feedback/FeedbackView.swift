import SwiftUI

enum FeedbackPresentation {
    static func statusLabel(for status: String) -> String {
        switch status {
        case "open": return "Đã gửi"
        case "triaged": return "Đã phân loại"
        case "in_progress": return "Đang xử lý"
        case "fixed": return "Đã xử lý"
        case "rejected": return "Không thực hiện"
        default: return status
        }
    }

    static func changelogVersion(for feedback: FeedbackModel, in changelog: [ChangelogModel]) -> String? {
        guard feedback.status == "fixed", let changelogID = feedback.changelogID else { return nil }
        return changelog.first(where: { $0.id == changelogID })?.version
    }

    static func shouldShowChangelog(for feedback: FeedbackModel, in changelog: [ChangelogModel]) -> Bool {
        changelogVersion(for: feedback, in: changelog) != nil
    }
}

/// Presentation-only contract for the owner-facing feedback inbox. Resolution
/// remains an explicit owner confirmation and is independent from backend
/// development status/changelog state.
enum FeedbackReferencePresentation {
    static let surfaceStyle: MPSurfaceCardStyle = .reference
    static let dividerSpacing: CGFloat = 0
    static let resolutionHint = "Tick khi bạn xác nhận đã hết lỗi. Bỏ tick nếu vấn đề xuất hiện lại."

    static func tabID(forResolved resolved: Bool) -> String {
        resolved ? "resolved" : "unresolved"
    }

    static func tabTitle(forResolved resolved: Bool, count: Int) -> String {
        resolved ? "Đã giải quyết (\(count))" : "Chưa giải quyết (\(count))"
    }
}

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
    @Published private(set) var updatingIDs: Set<String> = []

    private let api: FinanceAPI
    private var mutationRevision = 0
    init(api: FinanceAPI) { self.api = api }

    func setResolved(_ item: FeedbackModel, resolved: Bool) async {
        // Invalidate any in-flight list snapshot before PATCH. The response
        // from an older GET must not restore the row to its previous tab.
        guard !updatingIDs.contains(item.id) else { return }
        updatingIDs.insert(item.id)
        mutationRevision &+= 1
        errorMessage = nil
        // A refresh may have started after the mutation began but before the
        // PATCH completed. Bump the generation again when the request settles
        // so a stale GET that completes afterwards cannot overwrite the
        // mutation response (success or failure).
        defer {
            mutationRevision &+= 1
            updatingIDs.remove(item.id)
        }
        do {
            let updated = try await api.setFeedbackResolution(id: item.id, resolved: resolved)
            if let index = items.firstIndex(where: { $0.id == updated.id }) { items[index] = updated }
        } catch {
            errorMessage = "Không lưu được trạng thái. Phản hồi vẫn giữ nguyên; hãy thử lại."
        }
    }

    func load() async {
        guard !isLoading else { return }
        isLoading = true
        let snapshotRevision = mutationRevision
        errorMessage = nil
        defer { isLoading = false }
        do {
            async let feedback = api.feedback()
            async let changelog = api.changelog()
            let loadedItems = try await feedback
            let loadedChangelog = try await changelog
            guard snapshotRevision == mutationRevision else { return }
            items = loadedItems
            self.changelog = loadedChangelog
        } catch { errorMessage = "Không thể tải phản hồi và changelog." }
    }

    func submit(screenshot: Data? = nil) async -> Bool {
        guard let input = FeedbackDraftValidator.normalized(type: type, title: title, description: description) else {
            errorMessage = "Nhập tiêu đề và mô tả trước khi gửi."
            return false
        }
        isSending = true
        defer { isSending = false }
        do {
            let item = try await api.createFeedback(input, screenshot: screenshot)
            items.insert(item, at: 0)
            title = ""
            description = ""
            return true
        } catch let error where screenshot != nil && Self.shouldFallbackToTextOnly(error) {
            // Local development may intentionally run without private object
            // storage. Preserve the user's feedback instead of turning the
            // optional screenshot into a hard failure; production still keeps
            // the screenshot whenever storage is configured.
            do {
                let item = try await api.createFeedback(input)
                items.insert(item, at: 0)
                title = ""
                description = ""
                return true
            } catch {
                errorMessage = "Không thể gửi phản hồi."
                return false
            }
        } catch {
            errorMessage = "Không thể gửi phản hồi."
            return false
        }
    }

    static func shouldFallbackToTextOnly(_ error: Error) -> Bool {
        guard let apiError = error as? APIError, case .problem(let problem) = apiError else { return false }
        return problem.code == "FEEDBACK_SCREENSHOT_INVALID" || problem.code == "FEEDBACK_SCREENSHOT_STORAGE_UNAVAILABLE"
    }
}

struct FeedbackOverlay: View {
    let api: FinanceAPI
    let screenshot: Data?
    let onClose: () -> Void
    @StateObject private var model: FeedbackViewModel
    @State private var attachment: Data?

    init(api: FinanceAPI, screenshot: Data? = nil, onClose: @escaping () -> Void) {
        self.api = api
        self.screenshot = screenshot
        self.onClose = onClose
        _model = StateObject(wrappedValue: FeedbackViewModel(api: api))
        _attachment = State(initialValue: screenshot)
    }

    var body: some View {
        MPEditorScaffold(
            title: "Phản hồi",
            isBusy: model.isSending,
            onCancel: onClose
        ) {
            MPWalletSelector(title: "Loại", options: ["bug", "feature", "improvement"], selection: $model.type)
            MPTextField(title: "Tiêu đề", text: $model.title, placeholder: "Mô tả ngắn")
            MPTextEditor(title: "Chi tiết", text: $model.description, placeholder: "Điều gì đã xảy ra?")
            MPImageAttachmentPicker(data: $attachment)
            if let errorMessage = model.errorMessage { MPStatusMessage(text: errorMessage, tone: .danger) }
        } footer: {
            MPButton("Gửi phản hồi", isLoading: model.isSending) {
                Task { if await model.submit(screenshot: attachment) { onClose() } }
            }
        }
        .accessibilityIdentifier("feedback-overlay")
    }
}

private struct FeedbackItemRow: View {
    let item: FeedbackModel
    let changelog: [ChangelogModel]
    let isUpdating: Bool
    let onToggle: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            MPCheckboxRow(
                title: item.title,
                isSelected: item.isResolved,
                accessibilityIdentifier: "feedback-resolution-\(item.id)",
                action: onToggle
            )
            .disabled(isUpdating)
            MPText(item.description, style: .secondary)
                .fixedSize(horizontal: false, vertical: true)
            HStack(spacing: 8) {
                MPText(MPSelectionLabels.display(for: item.type), style: .secondary)
                MPFeedbackStatus(
                    statusLabel: FeedbackPresentation.statusLabel(for: item.status),
                    isFixed: item.status == "fixed",
                    changelogVersion: FeedbackPresentation.changelogVersion(for: item, in: changelog)
                )
                Spacer(minLength: 0)
            }
            if isUpdating { MPText("Đang lưu…", style: .secondary) }
            if let fixedAt = item.fixedAt { MPText("Đã xử lý: \(MPFormat.shortDate(fixedAt))", style: .secondary) }
            if let resolvedAt = item.resolvedAt { MPText("Bạn đã xác nhận: \(MPFormat.shortDate(resolvedAt))", style: .secondary) }
        }
    }
}

struct FeedbackView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @StateObject private var model: FeedbackViewModel
    @State private var selection = "unresolved"
    @State private var createPresented = false
    @State private var showChangelog = false

    init(api: FinanceAPI, onBack: @escaping () -> Void) {
        self.api = api
        self.onBack = onBack
        _model = StateObject(wrappedValue: FeedbackViewModel(api: api))
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                MPScreenHeader(title: "Phản hồi", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: {
                    MPIconButton(systemName: "plus", label: "Gửi phản hồi mới") { createPresented = true }
                })
                MPSegmentedControl(options: [
                    (id: "unresolved", title: "Chưa giải quyết (\(model.items.filter { !$0.isResolved }.count))"),
                    (id: "resolved", title: "Đã giải quyết (\(model.items.filter { $0.isResolved }.count))")
                ], selection: $selection, accessibilityIdentifier: "feedback-tabs")
                MPText(FeedbackReferencePresentation.resolutionHint, style: .secondary)
                if let errorMessage = model.errorMessage { MPStatusMessage(text: errorMessage, tone: .danger) }
                if model.isLoading && model.items.isEmpty { MPLoadingView(label: "Đang tải phản hồi…") }
                else if filteredItems.isEmpty {
                    MPEmptyView(title: "Không có phản hồi", detail: selection == "resolved" ? "Các phản hồi bạn đã tick sẽ xuất hiện ở đây." : "Không còn phản hồi chưa giải quyết trong danh sách.")
                }
                    if !filteredItems.isEmpty {
                        MPSurfaceCard(style: FeedbackReferencePresentation.surfaceStyle) {
                            VStack(alignment: .leading, spacing: FeedbackReferencePresentation.dividerSpacing) {
                                ForEach(Array(filteredItems.enumerated()), id: \.element.id) { index, item in
                                    if index > 0 { MPDivider() }
                                    FeedbackItemRow(
                                        item: item,
                                        changelog: model.changelog,
                                        isUpdating: model.isLoading || model.updatingIDs.contains(item.id)
                                    ) {
                                        Task { await model.setResolved(item, resolved: !item.isResolved) }
                                    }
                                    .accessibilityIdentifier("feedback-item-\(item.id)")
                                }
                            }
                        }
                    }
                MPButton(showChangelog ? "Ẩn changelog" : "Xem changelog", style: .plain) { showChangelog.toggle() }
                if showChangelog {
                MPAssistantCard(title: "Changelog") {
                    ForEach(model.changelog) { item in
                        VStack(alignment: .leading, spacing: 4) {
                            MPText("v\(item.version) — \(item.title)", style: .emphasis)
                            MPText(item.description, style: .secondary)
                        }
                    }
                }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await model.load() }
        .refreshable { await model.load() }
        .mpFullScreenCover(isPresented: $createPresented) {
            FeedbackOverlay(api: api) {
                createPresented = false
                Task { await model.load() }
            }
        }
    }

    private var filteredItems: [FeedbackModel] {
        model.items.filter { $0.isResolved == (selection == "resolved") }
    }
}
