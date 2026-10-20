import SwiftUI
import UIKit
import UniformTypeIdentifiers

enum AssistantComposerPolicy {
    static func canSend(enabled: Bool, text: String, activeRunID: String?, isLocallySending: Bool = false) -> Bool {
        enabled
            && activeRunID == nil
            && !isLocallySending
            && !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }
}

enum AssistantHistoryLoadPolicy {
    static func isCurrent(
        requestToken: UUID,
        currentToken: UUID,
        responseConversationID: String,
        expectedConversationID: String
    ) -> Bool {
        requestToken == currentToken && responseConversationID == expectedConversationID
    }

    /// Accept metadata only for the conversation/generation that initiated the
    /// request. The view's request token covers legacy empty responses.
    static func canApply(
        responseConversationID: String,
        responseGeneration: Int64?,
        expectedConversationID: String,
        expectedGeneration: Int64?
    ) -> Bool {
        guard !responseConversationID.isEmpty,
              responseConversationID == expectedConversationID else { return false }
        guard let responseGeneration, let expectedGeneration else { return true }
        return responseGeneration == expectedGeneration
    }
}

enum AssistantAdvisorRecoveryPolicy {
    /// A transport/5xx failure may happen after the server accepted a request;
    /// replaying the same client request ID is therefore safe. Validation and
    /// authentication failures are known rejections and must not be replayed.
    static func shouldReplay(after error: APIError) -> Bool {
        switch error {
        case .transport, .decoding:
            return true
        case .http(let status):
            return status >= 500
        case .problem(let problem):
            guard let status = problem.status else { return true }
            return status >= 500
        case .invalidURL, .unauthorized, .missingSession, .accountMismatch:
            return false
        }
    }
}

enum AssistantHistoryExpiryPolicy {
    static func isExpired(_ error: Error) -> Bool {
        guard let error = error as? APIError else { return false }
        switch error {
        case .problem(let problem):
            return problem.status == 410 || problem.code == "advisor_history_expired"
        case .http(let status):
            // Keep the client safe if a proxy strips the problem envelope.
            return status == 410
        default:
            return false
        }
    }
}

enum AssistantPrivacyPolicy {
    static let maskedText = "Nội dung ẩn khi bật chế độ che số dư."

    static func text(_ value: String, masked: Bool) -> String {
        masked ? maskedText : value
    }
}

enum AssistantWalletProposal {
    /// A server-generated action identity is persisted with the assistant
    /// message. Reusing it across app restarts makes confirmation replay-safe;
    /// older messages without the field fall back to the mounted-card key.
    static func actionID(_ view: [String: JSONValue]) -> String? {
        guard let value = text(view["action_id"]), !value.isEmpty, value.utf8.count <= 128 else { return nil }
        return value
    }

    static func actionContext(_ view: [String: JSONValue]) -> (conversationID: String, generation: Int64)? {
        guard let conversationID = text(view["conversation_id"]),
              !conversationID.isEmpty, conversationID.utf8.count <= 128,
              case .number(let generation) = view["generation"],
              let exactGeneration = Int64(exactly: generation), exactGeneration > 0 else {
            return nil
        }
        return (conversationID, exactGeneration)
    }

    static func decode(_ view: [String: JSONValue]) -> WalletModelInput? {
        guard let name = text(view["name"]), !name.isEmpty, name.utf8.count <= 80,
              let type = text(view["type"]), ["basic", "goal", "credit"].contains(type) else { return nil }
        if let action = view["action"], !absent(action), text(action) != "create_wallet" { return nil }
        for key in ["opening_balance", "target_amount", "credit_limit", "last_statement_balance", "statement_day", "payment_due_day"] {
            if let value = view[key], !absent(value), money(value) == nil { return nil }
        }
        for key in ["description", "target_date"] {
            if let value = view[key], !absent(value), text(value) == nil { return nil }
        }
        let description = text(view["description"])
        guard (description?.utf8.count ?? 0) <= 512 else { return nil }
        var isInTotal = true
        if let value = view["is_in_total"], !absent(value) {
            guard case .bool(let flag) = value else { return nil }
            isInTotal = flag
        }
        let target = money(view["target_amount"])
        let credit = money(view["credit_limit"])
        if type == "goal", (target ?? 0) <= 0 { return nil }
        if type == "credit", (credit ?? 0) <= 0 { return nil }
        let date = text(view["target_date"])
        if type == "goal", let date, !date.isEmpty {
            guard let parsed = MPFormat.parseCalendarDate(date),
                  MPFormat.calendarDateString(parsed) == date else { return nil }
        }
        let statementDay = money(view["statement_day"])
        let dueDay = money(view["payment_due_day"])
        if type == "credit" {
            for day in [statementDay, dueDay].compactMap({ $0 }) where !(1...31).contains(day) { return nil }
        }
        return WalletModelInput(
            name: name, type: type, openingBalance: money(view["opening_balance"]) ?? 0,
            isInTotal: isInTotal, description: description?.isEmpty == false ? description : nil,
            targetAmount: type == "goal" ? target : nil,
            targetDate: type == "goal" && date?.isEmpty == false ? date : nil,
            creditLimit: type == "credit" ? credit : nil,
            lastStatementBalance: type == "credit" ? money(view["last_statement_balance"]) : nil,
            statementDay: type == "credit" ? statementDay.map(Int.init) : nil,
            paymentDueDay: type == "credit" ? dueDay.map(Int.init) : nil
        )
    }

    private static func absent(_ value: JSONValue?) -> Bool {
        guard let value else { return true }
        if case .null = value { return true }
        return false
    }

    private static func text(_ value: JSONValue?) -> String? {
        guard case .string(let text) = value else { return nil }
        return text.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private static func money(_ value: JSONValue?) -> Int64? {
        guard case .number(let number) = value, number >= 0, number <= 9_007_199_254_740_991 else { return nil }
        return Int64(exactly: number)
    }
}

enum AssistantWalletConfirmationState: Equatable {
    case idle
    case saving
    case succeeded
    case failed

    var canConfirm: Bool { self == .idle || self == .failed }

    mutating func begin() -> Bool {
        guard canConfirm else { return false }
        self = .saving
        return true
    }

    mutating func finish(success: Bool) {
        guard self == .saving else { return }
        self = success ? .succeeded : .failed
    }
}

enum AssistantJarProposal {
    static func actionID(_ view: [String: JSONValue]) -> String? {
        guard case .string(let value) = view["action_id"], !value.isEmpty, value.utf8.count <= 128 else { return nil }
        return value
    }

    static func actionContext(_ view: [String: JSONValue]) -> (conversationID: String, generation: Int64)? {
        AssistantWalletProposal.actionContext(view)
    }

    static func decode(_ view: [String: JSONValue]) -> (action: String, jarID: String?, month: String, name: String, mode: String, amount: Int64?, percent: Double?)? {
        guard let month = text(view["month"]), month.range(of: "^[0-9]{4}-[0-9]{2}$", options: .regularExpression) != nil,
              let name = text(view["name"]), !name.isEmpty, name.utf8.count <= 100,
              let mode = text(view["allocation_mode"]), ["none", "fixed", "percent"].contains(mode) else { return nil }
        let action = text(view["action"]) ?? "create_jar"
        guard action == "create_jar" || action == "update_jar" else { return nil }
        let jarID = text(view["jar_id"])
        if action == "update_jar", jarID == nil || jarID?.isEmpty == true { return nil }
        let amount = number(view["allocation_amount"]).flatMap(Int64.init(exactly:))
        let percent = number(view["allocation_percent"])
        switch mode {
        case "none": guard amount == nil && percent == nil else { return nil }
        case "fixed": guard let amount, amount > 0, percent == nil else { return nil }
        case "percent": guard let percent, percent >= 0, percent <= 100, amount == nil else { return nil }
        default: return nil
        }
        return (action, jarID, month, name, mode, amount, percent)
    }

    private static func text(_ value: JSONValue?) -> String? {
        guard case .string(let text) = value else { return nil }
        return text.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private static func number(_ value: JSONValue?) -> Double? {
        guard case .number(let number) = value, number.isFinite else { return nil }
        return number
    }
}

enum AssistantBudgetProposal {
    static func decode(_ view: [String: JSONValue]) -> (action: String, budgetID: String?, name: String, limitAmount: Int64, walletID: String?, categoryID: String?, startDate: String, endDate: String)? {
        let action = text(view["action"]) ?? "create_budget"
        guard action == "create_budget" || action == "update_budget",
              let name = text(view["name"]), !name.isEmpty, name.utf8.count <= 200,
              case .number(let limit) = view["limit_amount"], limit > 0, limit <= 9_007_199_254_740_991,
              let exactLimit = Int64(exactly: limit),
              let startDate = text(view["start_date"]), let start = MPFormat.parseCalendarDate(startDate), MPFormat.calendarDateString(start) == startDate,
              let endDate = text(view["end_date"]), let end = MPFormat.parseCalendarDate(endDate), MPFormat.calendarDateString(end) == endDate,
              end >= start else { return nil }
        let budgetID = text(view["budget_id"])
        if action == "update_budget", budgetID == nil || UUID(uuidString: budgetID!) == nil { return nil }
        let walletID = text(view["wallet_id"])
        let categoryID = text(view["category_id"])
        for id in [walletID, categoryID].compactMap({ $0 }) where UUID(uuidString: id) == nil { return nil }
        return (action, budgetID, name, exactLimit, walletID, categoryID, startDate, endDate)
    }

    private static func text(_ value: JSONValue?) -> String? {
        guard case .string(let text) = value else { return nil }
        return text.trimmingCharacters(in: .whitespacesAndNewlines)
    }
}

enum AssistantWalletConfirmationSummary {
    static func lines(for input: WalletModelInput, masked: Bool) -> [String] {
        var lines = ["Số dư ban đầu: \(MPFormat.money(input.openingBalance, masked: masked))"]
        if !input.isInTotal {
            lines.append("Không tính vào tổng tài sản")
        }
        if let description = input.description, !description.isEmpty {
            lines.append("Ghi chú: \(description)")
        }
        switch input.type {
        case "goal":
            if let targetAmount = input.targetAmount {
                lines.append("Mục tiêu: \(MPFormat.money(targetAmount, masked: masked))")
            }
            if let targetDate = input.targetDate, !targetDate.isEmpty {
                lines.append("Hạn mục tiêu: \(targetDate)")
            }
        case "credit":
            if let creditLimit = input.creditLimit {
                lines.append("Hạn mức: \(MPFormat.money(creditLimit, masked: masked))")
            }
            if let balance = input.lastStatementBalance {
                lines.append("Dư nợ sao kê: \(MPFormat.money(balance, masked: masked))")
            }
            if let statementDay = input.statementDay {
                lines.append("Sao kê ngày \(statementDay)")
            }
            if let paymentDueDay = input.paymentDueDay {
                lines.append("Đến hạn ngày \(paymentDueDay)")
            }
        default:
            break
        }
        return lines
    }
}

struct AssistantView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    let ownerID: String
    let masked: Bool
    @Environment(\.scenePhase) private var scenePhase
    @State private var capabilities: AdvisorCapabilitiesModel?
    @State private var overview: AdvisorOverviewModel?
    @State private var messages: [AdvisorMessageModel] = []
    @State private var historyMeta: AdvisorHistoryMetaModel?
    @State private var conversationID = ""
    @State private var localInput = ""
    private let draft: Binding<String>?
    @State private var sending = false
    @State private var historyLoading = true
    @State private var loadingOlder = false
    @State private var hasOlder = false
    @State private var activeRunID: String?
    @State private var error: String?
    @State private var walletActionMessage: String?
    @State private var overviewPresented = false
    @State private var keyboardVisible = false
    @State private var dismissKeyboardRequest = 0
    @State private var resetConfirmation = false
    @State private var resetting = false
    @State private var historyRequestToken = UUID()
    @State private var advisorPollToken = UUID()
    @State private var advisorSubmitRequestID: String?
    @State private var advisorSubmitUncertain = false
    @StateObject private var runObserver: AdvisorRunObserver

    private let suggestions = [
        "Tháng này tôi tiêu thế nào?",
        "Tìm các giao dịch trên 200.000đ",
        "So sánh chi tiêu với tháng trước"
    ]

    init(api: FinanceAPI, onBack: @escaping () -> Void, ownerID: String, masked: Bool = false, draft: Binding<String>? = nil) {
        self.api = api
        self.onBack = onBack
        self.ownerID = ownerID
        self.masked = masked
        self.draft = draft
        _runObserver = StateObject(wrappedValue: AdvisorRunObserver(fetchRun: { runID in
            try await api.advisorRun(id: runID)
        }))
        let key = Self.conversationKey(ownerID: ownerID)
        _conversationID = State(initialValue: UserDefaults.standard.string(forKey: key) ?? "")
    }

    private static func conversationKey(ownerID: String) -> String {
        "mypocket.advisor.conversation.\(ownerID)"
    }

    private var input: String {
        get { inputBinding.wrappedValue }
        nonmutating set { inputBinding.wrappedValue = newValue }
    }

    private var inputBinding: Binding<String> { draft ?? $localInput }

    /// Drop only the local conversation pointer after the server has expired
    /// the idle history. The composer is intentionally left untouched so a
    /// question that was being retried is never silently lost.
    private func clearExpiredConversation() {
        messages = []
        historyMeta = nil
        conversationID = ""
        hasOlder = false
        activeRunID = nil
        historyRequestToken = UUID()
        advisorPollToken = UUID()
        runObserver.reset()
        advisorSubmitRequestID = nil
        advisorSubmitUncertain = false
        UserDefaults.standard.removeObject(forKey: Self.conversationKey(ownerID: ownerID))
    }

    var body: some View {
        VStack(spacing: 0) {
            MPScreenHeader(title: "Finance Assistant", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: {
                if keyboardVisible {
                    HStack(spacing: 8) {
                        if overview != nil {
                            MPIconButton(systemName: "chart.bar", label: "Tổng quan tài chính") { overviewPresented = true }
                                .accessibilityIdentifier("assistant-header-overview")
                        }
                        MPIconButton(systemName: "keyboard.chevron.compact.down", label: "Đóng bàn phím") { dismissKeyboardRequest += 1 }
                            .accessibilityIdentifier("assistant-dismiss-keyboard")
                    }
                }
                if !keyboardVisible && !conversationID.isEmpty {
                    HStack(spacing: 4) {
                        MPIconButton(systemName: "arrow.clockwise", label: "Tải lại lịch sử") { Task { await reloadHistory() } }
                        MPOverflowMenu(label: "Tùy chọn hội thoại", actions: [
                            MPOverflowAction(title: "Chat mới") { resetConfirmation = true }
                        ])
                    }
                }
            })
                .padding(.horizontal, MPMetrics.pagePadding)
            if overview != nil && !keyboardVisible {
                MPChatOverview(income: overviewMoney("income"), expense: overviewMoney("expense")) {
                    overviewPresented = true
                }
                .padding(.horizontal, MPMetrics.pagePadding)
                .padding(.bottom, 8)
            }
            ScrollViewReader { proxy in
            ScrollView {
                LazyVStack(alignment: .leading, spacing: 16) {
                    if let capabilities, !capabilities.enabled { MPStatusMessage(text: "Finance Assistant chưa được bật.", tone: .info) }
                    if let walletActionMessage { MPStatusMessage(text: walletActionMessage, tone: .success) }
                    if !historyLoading && messages.isEmpty {
                            VStack(alignment: .leading, spacing: 10) {
                                MPText("Hỏi về tài chính của bạn", style: .title)
                                MPText("Khám phá chi tiêu, tìm giao dịch hoặc so sánh các kỳ.", style: .secondary)
                                ForEach(Array(suggestions.enumerated()), id: \.offset) { index, suggestion in
                                    MPButton(suggestion, style: .secondary) {
                                        input = suggestion
                                    }
                                    .accessibilityIdentifier("assistant-suggestion-\(index)")
                                }
                            }
                    }
                    if historyLoading {
                        MPLoadingView(label: "Đang tải lịch sử trợ lý...")
                    }
                    if hasOlder {
                        MPButton("Tải tin cũ", style: .secondary, isLoading: loadingOlder) { Task { await loadOlderMessages() } }
                    }
                    ForEach(messages) { message in
                        MPChatMessage(isUser: message.role == "user") {
                            VStack(alignment: .leading, spacing: 8) {
                                ForEach(Array(message.parts.enumerated()), id: \.offset) { _, part in
                                    if let text = part.text {
                                        if masked {
                                            MPText(AssistantPrivacyPolicy.maskedText, style: .secondary)
                                        } else if message.role == "assistant" {
                                            MPMarkdown(content: text)
                                        } else {
                                            MPText(text)
                                        }
                                    }
                                    else if let data = part.data {
                                        AssistantStructuredCard(type: part.type, data: data, masked: masked, onCreateWallet: createWalletFromAssistant, onConfirmWalletAction: confirmWalletAction, onConfirmJarAction: confirmJarAction, onConfirmBudgetAction: confirmBudgetAction)
                                    } else { MPText("Đã nhận thẻ dữ liệu tài chính.", style: .secondary) }
                                }
                            }
                        }
                        .id(message.id)
                    }
                    if sending || activeRunID != nil {
                        let label = runObserver.connectionState == .reconnecting
                            ? "Kết nối gián đoạn. Yêu cầu vẫn được theo dõi..."
                            : "Trợ lý đang phân tích..."
                        MPLoadingView(label: label)
                    }
                    if let error { MPStatusMessage(text: error, tone: .danger) }
                    if advisorSubmitUncertain {
                        MPStatusMessage(text: "Yêu cầu có thể đã được tiếp nhận. Giữ nguyên nội dung và dùng nút thử lại để tiếp tục; không tạo yêu cầu mới.", tone: .info)
                        MPButton("Thử lại yêu cầu", style: .secondary, isLoading: sending) {
                            Task { await retryUncertainAdvisorSubmit() }
                        }
                        .accessibilityIdentifier("assistant-retry-submit")
                    }
                    Color.clear.frame(height: 1).id("assistant-latest")
                }
                .padding(MPMetrics.pagePadding)
            }
            .accessibilityIdentifier("assistant-conversation")
            .scrollDismissesKeyboard(.interactively)
            .onChange(of: messages.last?.id) { _, _ in
                proxy.scrollTo("assistant-latest", anchor: .bottom)
            }
            }
            MPChatComposer(
                text: inputBinding,
                isBusy: sending || activeRunID != nil,
                isEnabled: capabilities?.enabled == true && !advisorSubmitUncertain,
                onSend: { Task { await send() } },
                onCancel: activeRunID.map { runID in { Task { await cancel(runID: runID) }; return () } },
                dismissKeyboardRequest: dismissKeyboardRequest
            )
            .padding(.horizontal, MPMetrics.pagePadding)
            .padding(.vertical, 8)
            .disabled(advisorSubmitUncertain)
            .background {
                GeometryReader { geometry in
                    Color.clear.preference(key: MPFloatingActionClearanceKey.self, value: geometry.size.height)
                }
            }
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .mpObserveKeyboardVisibility($keyboardVisible)
        .task { await loadInitialState() }
        .onChange(of: scenePhase) { _, phase in
            if phase == .inactive || phase == .background {
                runObserver.suspend()
            } else if phase == .active {
                if let runID = activeRunID {
                    let token = advisorPollToken
                    Task {
                        await runObserver.observe(ownerID: ownerID, runID: runID)
                        guard advisorPollToken == token, activeRunID == runID else { return }
                        await reloadHistory()
                        if advisorPollToken == token { activeRunID = nil }
                    }
                } else {
                    Task { await reloadHistory() }
                }
            }
        }
        .mpFormBottomSheet(isPresented: $overviewPresented, title: "Tổng quan tài chính", accessibilityIdentifier: "assistant-overview-details", onClose: { overviewPresented = false }) {
            if let overview {
                AssistantStructuredCard(
                    type: overview.viewKind,
                    data: .object(["view": overview.view ?? .null, "source": overview.source ?? .null]),
                    masked: masked,
                    onCreateWallet: createWalletFromAssistant,
                    onConfirmWalletAction: confirmWalletAction,
                    onConfirmJarAction: confirmJarAction,
                    onConfirmBudgetAction: confirmBudgetAction
                )
            }
        } footer: {
            MPButton("Đóng", style: .secondary) { overviewPresented = false }
        }
        .mpConfirmDialog(
            isPresented: $resetConfirmation,
            title: "Bắt đầu chat mới?",
            message: "Lịch sử trợ lý hiện tại sẽ bị xoá khỏi ngữ cảnh. Ví, giao dịch và báo cáo không bị ảnh hưởng.",
            confirmTitle: "Chat mới"
        ) {
            Task { await resetConversation() }
        }
    }

    private func overviewMoney(_ key: String) -> String {
        guard let overview, AssistantCardType.normalized(overview.viewKind) == "metric_group",
              case .object(let view) = overview.view,
              case .number(let value) = view[key], let exact = Int64(exactly: value) else {
            return masked ? MPFormat.maskedValue : "Chưa có dữ liệu"
        }
        return MPFormat.money(exact, masked: masked)
    }

    private func resetConversation() async {
        guard !resetting else { return }
        resetting = true
        defer { resetting = false }
        let pendingKeyName = "mypocket.advisor.reset-request.\(ownerID)"
        let requestID = UserDefaults.standard.string(forKey: pendingKeyName) ?? UUID().uuidString
        UserDefaults.standard.set(requestID, forKey: pendingKeyName)
        do {
            _ = try await api.resetAdvisorConversation(idempotencyKey: requestID)
            UserDefaults.standard.removeObject(forKey: pendingKeyName)
            clearExpiredConversation()
            error = "Đã bắt đầu cuộc trò chuyện mới."
        } catch let error as APIError {
            self.error = error.userMessage
        } catch {
            self.error = "Không thể bắt đầu cuộc trò chuyện mới."
        }
    }

    private func loadInitialState() async {
        historyLoading = true
        defer { historyLoading = false }
        do { capabilities = try await api.advisorCapabilities() }
        catch let error as APIError {
#if DEBUG
            print("[Assistant] capabilities failed: \(APIError.diagnosticDescription(error))")
#endif
            self.error = error.userMessage
        }
        catch {
#if DEBUG
            print("[Assistant] capabilities failed: \(APIError.diagnosticDescription(error))")
#endif
            self.error = "Không thể kết nối Finance Assistant."
        }
        do { overview = try await api.advisorOverview() }
        catch let error as APIError {
#if DEBUG
            print("[Assistant] overview failed: \(APIError.diagnosticDescription(error))")
#endif
            self.error = error.userMessage
        }
        catch {
#if DEBUG
            print("[Assistant] overview failed: \(APIError.diagnosticDescription(error))")
#endif
            self.error = "Không thể tải tổng quan trợ lý tài chính."
        }
        if !conversationID.isEmpty {
            await reloadHistory()
        }
    }

    private func reloadHistory() async {
        guard !conversationID.isEmpty else { return }
        let requestedConversationID = conversationID
        let requestToken = UUID()
        historyRequestToken = requestToken
        do {
            let page = try await api.advisorHistory(conversationID: requestedConversationID)
            guard AssistantHistoryLoadPolicy.isCurrent(
                requestToken: requestToken,
                currentToken: historyRequestToken,
                responseConversationID: conversationID,
                expectedConversationID: requestedConversationID
            ) else { return }
            if let meta = page.meta,
               !AssistantHistoryLoadPolicy.canApply(
                    responseConversationID: meta.conversationID,
                    responseGeneration: meta.generation,
                    expectedConversationID: requestedConversationID,
                    expectedGeneration: historyMeta?.generation
               ) {
                return
            }
            messages = page.messages
            historyMeta = page.meta
            hasOlder = messages.count == 50
            error = nil
            if let activeRun = page.meta?.activeRun,
               !["completed", "failed", "cancelled", "interrupted", "purged"].contains(activeRun.status),
               activeRunID != activeRun.id {
                resumeActiveRun(activeRun.id)
            }
        } catch let error as APIError {
            guard AssistantHistoryLoadPolicy.isCurrent(
                requestToken: requestToken,
                currentToken: historyRequestToken,
                responseConversationID: conversationID,
                expectedConversationID: requestedConversationID
            ) else { return }
            if AssistantHistoryExpiryPolicy.isExpired(error) {
                clearExpiredConversation()
                self.error = "Lịch sử trợ lý đã hết hạn. Bạn có thể gửi lại câu hỏi đang soạn."
            } else {
                self.error = error.userMessage
            }
        }
        catch {
            guard AssistantHistoryLoadPolicy.isCurrent(
                requestToken: requestToken,
                currentToken: historyRequestToken,
                responseConversationID: conversationID,
                expectedConversationID: requestedConversationID
            ) else { return }
            self.error = "Không thể tải lại lịch sử trợ lý."
        }
    }

    private func loadOlderMessages() async {
        guard !conversationID.isEmpty, let first = messages.first, !loadingOlder, hasOlder else { return }
        let requestedConversationID = conversationID
        let requestToken = UUID()
        historyRequestToken = requestToken
        loadingOlder = true
        defer { loadingOlder = false }
        do {
            let page = try await api.advisorHistory(conversationID: requestedConversationID, beforeSequence: first.seq)
            guard AssistantHistoryLoadPolicy.isCurrent(
                requestToken: requestToken,
                currentToken: historyRequestToken,
                responseConversationID: conversationID,
                expectedConversationID: requestedConversationID
            ) else { return }
            if let meta = page.meta,
               !AssistantHistoryLoadPolicy.canApply(
                    responseConversationID: meta.conversationID,
                    responseGeneration: meta.generation,
                    expectedConversationID: requestedConversationID,
                    expectedGeneration: historyMeta?.generation
               ) {
                return
            }
            messages = mergeMessages(messages, page.messages)
            if let incomingMeta = page.meta {
                if let currentGeneration = historyMeta?.generation {
                    if incomingMeta.generation >= currentGeneration { historyMeta = incomingMeta }
                } else {
                    historyMeta = incomingMeta
                }
            }
            hasOlder = page.messages.count == 50
            error = nil
        } catch let error as APIError {
            guard AssistantHistoryLoadPolicy.isCurrent(
                requestToken: requestToken,
                currentToken: historyRequestToken,
                responseConversationID: conversationID,
                expectedConversationID: requestedConversationID
            ) else { return }
            if AssistantHistoryExpiryPolicy.isExpired(error) {
                clearExpiredConversation()
                self.error = "Lịch sử trợ lý đã hết hạn. Bạn có thể gửi lại câu hỏi đang soạn."
            } else {
                self.error = error.userMessage
            }
        }
        catch {
            guard AssistantHistoryLoadPolicy.isCurrent(
                requestToken: requestToken,
                currentToken: historyRequestToken,
                responseConversationID: conversationID,
                expectedConversationID: requestedConversationID
            ) else { return }
            self.error = "Không thể tải tin cũ của trợ lý."
        }
    }

    private func mergeMessages(_ current: [AdvisorMessageModel], _ incoming: [AdvisorMessageModel]) -> [AdvisorMessageModel] {
        var byID = Dictionary(uniqueKeysWithValues: current.map { ($0.id, $0) })
        incoming.forEach { byID[$0.id] = $0 }
        return byID.values.sorted { $0.seq < $1.seq }
    }

    private func send() async {
        let text = input.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !advisorSubmitUncertain,
              AssistantComposerPolicy.canSend(
                  enabled: capabilities?.enabled == true,
                  text: text,
                  activeRunID: activeRunID,
                  isLocallySending: sending
              ) else { return }
        sending = true; error = nil; defer { sending = false }
        let clientRequestID = advisorSubmitRequestID ?? UUID().uuidString
        advisorSubmitRequestID = clientRequestID
        do {
            let result = try await api.beginAdvisorMessage(AdvisorMessageInputModel(clientRequestID: clientRequestID, text: text))
            advisorSubmitRequestID = nil
            advisorSubmitUncertain = false
            conversationID = result.conversationID
            UserDefaults.standard.set(result.conversationID, forKey: Self.conversationKey(ownerID: ownerID))
            activeRunID = result.runID
            let pollToken = UUID()
            advisorPollToken = pollToken
            input = ""
            if result.status == "queued" || result.status == "running" {
                await runObserver.observe(ownerID: ownerID, runID: result.runID)
                guard advisorPollToken == pollToken, activeRunID == result.runID else { return }
                if let run = runObserver.terminalResult {
                    if run.status == "purged" {
                        input = text
                        clearExpiredConversation()
                        self.error = "Lịch sử trợ lý đã hết hạn. Bạn có thể gửi lại câu hỏi đang soạn."
                    } else if run.status != "completed" {
                        self.error = "Yêu cầu trợ lý kết thúc với trạng thái \(MPSelectionLabels.display(for: run.status))."
                    }
                }
            }
            await reloadHistory()
            guard advisorPollToken == pollToken, activeRunID == result.runID else { return }
            activeRunID = nil
        } catch let apiError as APIError {
            if AssistantHistoryExpiryPolicy.isExpired(apiError) {
                input = text
                clearExpiredConversation()
                self.error = "Lịch sử trợ lý đã hết hạn. Bạn có thể gửi lại câu hỏi đang soạn."
            } else if advisorSubmitRequestID != nil && AssistantAdvisorRecoveryPolicy.shouldReplay(after: apiError) {
                advisorSubmitUncertain = true
                self.error = "Không thể xác nhận phản hồi của máy chủ. Giữ nguyên nội dung và tải lại yêu cầu trước khi gửi lại."
            } else {
                if advisorSubmitRequestID != nil { advisorSubmitRequestID = nil }
                self.error = apiError.userMessage
            }
        }
        catch {
            if advisorSubmitRequestID != nil {
                advisorSubmitUncertain = true
                self.error = "Không thể xác nhận máy chủ đã tiếp nhận yêu cầu. Giữ nguyên nội dung và thử lại yêu cầu này."
            } else {
                self.error = "Không thể xử lý yêu cầu trợ lý tài chính."
            }
        }
    }

    private func retryUncertainAdvisorSubmit() async {
        guard advisorSubmitUncertain, advisorSubmitRequestID != nil else { return }
        advisorSubmitUncertain = false
        await send()
    }

    private func cancel(runID: String) async {
        do {
            try await api.cancelAdvisorRun(id: runID)
            guard activeRunID == runID else { return }
            advisorPollToken = UUID()
            runObserver.reset()
            sending = false
            activeRunID = nil
            error = "Đã hủy yêu cầu."
        }
        catch { self.error = "Không thể hủy yêu cầu đang chạy." }
    }

    private func resumeActiveRun(_ runID: String) {
        guard activeRunID != runID else { return }
        activeRunID = runID
        let token = advisorPollToken
        Task {
            await runObserver.observe(ownerID: ownerID, runID: runID)
            guard advisorPollToken == token, activeRunID == runID else { return }
            await reloadHistory()
            guard advisorPollToken == token else { return }
            activeRunID = nil
        }
    }

    private func createWalletFromAssistant(_ input: WalletModelInput, idempotencyKey: String) async throws {
        _ = try await api.createWallet(input, idempotencyKey: idempotencyKey)
        walletActionMessage = "Đã tạo ví \(input.name) sau khi bạn xác nhận."
    }

    private func confirmWalletAction(_ actionID: String, conversationID: String, generation: Int64) async throws {
        _ = try await api.confirmAdvisorWalletAction(actionID: actionID, conversationID: conversationID, generation: generation)
        walletActionMessage = "Đã tạo ví sau khi bạn xác nhận."
    }

    private func confirmJarAction(_ actionID: String, conversationID: String, generation: Int64) async throws {
        _ = try await api.confirmAdvisorJarAction(actionID: actionID, conversationID: conversationID, generation: generation)
        walletActionMessage = "Đã tạo hũ sau khi bạn xác nhận."
    }

    private func confirmBudgetAction(_ actionID: String, conversationID: String, generation: Int64) async throws {
        _ = try await api.confirmAdvisorBudgetAction(actionID: actionID, conversationID: conversationID, generation: generation)
        walletActionMessage = "Đã tạo ngân sách sau khi bạn xác nhận."
    }
}

/// The stable card vocabulary shared with the web Finance Assistant.
///
/// Provider responses may use legacy aliases, so normalize before selecting a
/// native card. This keeps stored assistant messages renderable across app
/// versions without changing the backend/tool contract.
enum AssistantCardType {
    static func normalized(_ type: String) -> String {
        switch type {
        case "finance_summary": return "metric_group"
        case "spending_comparison": return "period_comparison"
        case "transaction_search": return "transaction_list"
        case "budget_progress": return "budget_status"
        case "wallet_creation_proposal": return "wallet_confirmation"
        case "jar_creation_proposal": return "jar_confirmation"
        case "jar_update_proposal": return "jar_confirmation"
        case "budget_creation_proposal", "budget_update_proposal": return "budget_confirmation"
        default: return type
        }
    }

    static func title(for type: String) -> String {
        if type == "jar_update_proposal" { return "Đề xuất cập nhật hũ" }
        if type == "budget_creation_proposal" { return "Đề xuất tạo ngân sách" }
        if type == "budget_update_proposal" { return "Đề xuất cập nhật ngân sách" }
        switch normalized(type) {
        case "metric_group": return "Tổng quan kỳ chọn"
        case "period_comparison": return "So sánh chi tiêu"
        case "transaction_list": return "Giao dịch"
        case "transaction_detail": return "Chi tiết giao dịch"
        case "wallet_balances": return "Số dư các ví"
        case "goal_progress": return "Tiến độ mục tiêu"
        case "budget_status": return "Tiến độ ngân sách"
        case "jar_progress": return "Tiến độ hũ chi tiêu"
        case "wallet_confirmation": return "Đề xuất tạo ví"
        case "jar_confirmation": return "Đề xuất tạo hũ"
        case "product_help": return "Hướng dẫn trong ứng dụng"
        default: return type.replacingOccurrences(of: "_", with: " ")
        }
    }

    /// Match the web card renderer's intentionally bounded visible prefixes.
    /// This does not change source totals or the backend response.
    static func visibleItems(_ items: [JSONValue], for type: String) -> [JSONValue] {
        switch normalized(type) {
        case "transaction_list": return Array(items.prefix(5))
        case "wallet_balances", "goal_progress", "budget_status", "jar_progress": return Array(items.prefix(8))
        default: return items
        }
    }

    static func transactionDetailSummary(_ view: [String: JSONValue]) -> String {
        let wallet = string(view["wallet_name"]) ?? "Ví"
        if let category = string(view["category_name"])?.trimmingCharacters(in: .whitespacesAndNewlines), !category.isEmpty {
            return "\(wallet) · \(category)"
        }
        return wallet
    }

    static func transactionDetailFlags(_ view: [String: JSONValue]) -> String {
        var flags = [bool(view["included_in_reports"]) == false ? "Không tính vào báo cáo" : "Tính vào báo cáo"]
        if let transferID = string(view["transfer_id"]), !transferID.isEmpty {
            flags.append("Chuyển ví")
        }
        return flags.joined(separator: " · ")
    }

    static func goalProgressPercent(current: Double?, target: Double?) -> Int? {
        guard let current, let target,
              monetaryValue(.number(current)) != nil,
              monetaryValue(.number(target)) != nil, target > 0 else { return nil }
        let ratio = min(max(current / target, 0), 1)
        return Int((ratio * 100).rounded())
    }

    /// Resolve the target used by both budget and jar progress cards. Jars
    /// use `allocation_amount`; budgets prefer `effective_limit_amount` including
    /// carryover, falling back to `limit_amount` (or the legacy
    /// calculated allocation fallback). A zero target is valid input and is
    /// rendered as 0% rather than silently borrowing another field.
    static func progressTarget(_ item: [String: JSONValue]) -> Double? {
        for key in ["allocation_amount", "effective_limit_amount", "limit_amount", "calculated_allocation"] {
            guard let amount = monetaryValue(item[key]), amount >= 0 else { continue }
            return Double(amount)
        }
        return nil
    }

    /// Match the finance wire contract; never truncate or fabricate a zero.
    static func monetaryValue(_ value: JSONValue?) -> Int64? {
        guard case .number(let amount) = value,
              amount.isFinite, abs(amount) <= 9_007_199_254_740_991,
              let exact = Int64(exactly: amount) else { return nil }
        return exact
    }

    static func money(_ value: JSONValue?, masked: Bool) -> String {
        if masked { return MPFormat.maskedValue }
        guard let amount = monetaryValue(value) else { return "Chưa có dữ liệu" }
        return MPFormat.money(amount)
    }

    static func fallbackText(_ value: JSONValue, masked: Bool) -> String {
        masked ? AssistantPrivacyPolicy.maskedText : value.prettyDescription
    }

    static func sourceLabel(_ value: JSONValue?) -> String {
        guard let value, case .object(let source) = value,
              case .string(let rawTimezone) = source["timezone"] else {
            return "Nguồn dữ liệu tài chính"
        }
        let timezone = rawTimezone.trimmingCharacters(in: .whitespacesAndNewlines)
        return timezone.isEmpty ? "Nguồn dữ liệu tài chính" : "Nguồn dữ liệu · \(timezone)"
    }

    struct ProductHelpSource: Equatable, Sendable {
        let id: String
        let title: String
        let excerpt: String
    }

    static func productHelpSources(_ view: [String: JSONValue]) -> [ProductHelpSource] {
        guard case .array(let values) = view["sources"] else { return [] }
        return values.prefix(5).compactMap { value in
            guard case .object(let source) = value,
                  let id = string(source["id"]), !id.isEmpty,
                  let title = string(source["title"]), !title.isEmpty,
                  let excerpt = string(source["excerpt"]), !excerpt.isEmpty else { return nil }
            return ProductHelpSource(id: String(id.prefix(128)), title: String(title.prefix(160)), excerpt: String(excerpt.prefix(480)))
        }
    }

    static func transactionCount(_ view: [String: JSONValue]) -> Int {
        if let value = view["total_count"], case .number(let count) = value,
           count >= 0, let exact = Int(exactly: count) {
            return exact
        }
        if let value = view["items"], case .array(let items) = value {
            return items.prefix(5).reduce(into: 0) { count, item in
                if case .object = item { count += 1 }
            }
        }
        return 0
    }

    private static func string(_ value: JSONValue?) -> String? {
        guard let value, case .string(let text) = value else { return nil }
        return text
    }

    private static func bool(_ value: JSONValue?) -> Bool? {
        guard let value, case .bool(let value) = value else { return nil }
        return value
    }
}

struct AssistantStructuredCard: View {
    @Environment(\.timeZone) private var timeZone
    let type: String
    let data: JSONValue
    let masked: Bool
    let onCreateWallet: ((WalletModelInput, String) async throws -> Void)?
    let onConfirmWalletAction: ((String, String, Int64) async throws -> Void)?
    let onConfirmJarAction: ((String, String, Int64) async throws -> Void)?
    let onConfirmBudgetAction: ((String, String, Int64) async throws -> Void)?
    @State private var walletConfirmationState: AssistantWalletConfirmationState = .idle
    @State private var walletConfirmationError: String?
    @State private var walletConfirmationRequestKey = UUID().uuidString

    init(type: String, data: JSONValue, masked: Bool = false, onCreateWallet: ((WalletModelInput, String) async throws -> Void)? = nil, onConfirmWalletAction: ((String, String, Int64) async throws -> Void)? = nil, onConfirmJarAction: ((String, String, Int64) async throws -> Void)? = nil, onConfirmBudgetAction: ((String, String, Int64) async throws -> Void)? = nil) {
        self.type = type
        self.data = data
        self.masked = masked
        self.onCreateWallet = onCreateWallet
        self.onConfirmWalletAction = onConfirmWalletAction
        self.onConfirmJarAction = onConfirmJarAction
        self.onConfirmBudgetAction = onConfirmBudgetAction
    }

    var body: some View {
        let normalizedType = AssistantCardType.normalized(type)
        MPAssistantCard(title: AssistantCardType.title(for: type)) {
            if let view = viewObject {
                switch normalizedType {
                case "metric_group": metricView(view)
                case "period_comparison": comparisonView(view)
                case "transaction_list": transactionListView(view)
                case "transaction_detail": transactionDetailView(view)
                case "wallet_balances", "goal_progress": walletListView(view, showsGoalProgress: normalizedType == "goal_progress")
                case "budget_status", "jar_progress": progressListView(view)
                case "wallet_confirmation": walletConfirmationView(view)
                case "jar_confirmation": jarConfirmationView(view)
                case "budget_confirmation": budgetConfirmationView(view)
                case "product_help": productHelpView(view)
                default: MPText(AssistantCardType.fallbackText(data, masked: masked), style: .secondary)
                }
            } else {
                MPText(AssistantCardType.fallbackText(data, masked: masked), style: .secondary)
            }
        }
    }

    private var viewObject: [String: JSONValue]? {
        guard case .object(let root) = data, let view = root["view"], case .object(let object) = view else { return nil }
        return object
    }

    @ViewBuilder private func metricView(_ view: [String: JSONValue]) -> some View {
        LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 8) {
            metric("Thu nhập", view["income"])
            metric("Chi tiêu", view["expense"])
            metric("Ròng", view["net"])
            if let count = number(view["count"]), let exact = Int64(exactly: count) { MPMetric(title: "Giao dịch", value: masked ? MPFormat.maskedValue : String(exact)) }
        }
        MPText(AssistantCardType.sourceLabel(sourceValue), style: .secondary)
    }

    @ViewBuilder private func comparisonView(_ view: [String: JSONValue]) -> some View {
        metric("Hiện tại", view["current_expense"])
        metric("Kỳ trước", view["previous_expense"])
        metric("Chênh lệch", view["delta"])
        MPText("Phần trăm thay đổi chỉ hiển thị khi kỳ trước có số liệu.", style: .secondary)
    }

    @ViewBuilder private func transactionListView(_ view: [String: JSONValue]) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Spacer()
                MPText("\(AssistantCardType.transactionCount(view)) dòng", style: .secondary)
            }
            if case .array(let values) = view["items"] {
                ForEach(Array(AssistantCardType.visibleItems(values, for: type).enumerated()), id: \.offset) { _, value in
                    if let item = object(value), let exactAmount = AssistantCardType.monetaryValue(item["amount"]) {
                        MPTransactionRow(
                            type: string(item["type"]) ?? "expense", rawAmount: exactAmount,
                            categoryName: string(item["category_name"]),
                            systemKey: string(item["category_system_key"]),
                            iconKey: string(item["category_icon_key"]), note: string(item["note"]),
                            occurredAt: string(item["occurred_at"]) ?? "", timeZone: timeZone,
                            masked: masked, transferID: string(item["transfer_id"]),
                            adjustmentDirection: string(item["adjustment_direction"]),
                            creditKind: string(item["credit_kind"])
                        )
                    }
                }
            } else {
                MPText(AssistantCardType.fallbackText(data, masked: masked), style: .secondary)
            }
            MPText(AssistantCardType.sourceLabel(sourceValue), style: .secondary)
        }
    }

    @ViewBuilder private func transactionDetailView(_ view: [String: JSONValue]) -> some View {
        MPText(AssistantCardType.transactionDetailSummary(view), style: .emphasis)
        if number(view["amount"]) != nil {
            MPText(money(view["amount"]), style: .emphasis)
        } else {
            MPText("Chưa có dữ liệu", style: .secondary)
        }
        MPText(AssistantCardType.transactionDetailFlags(view), style: .secondary)
    }

    @ViewBuilder private func productHelpView(_ view: [String: JSONValue]) -> some View {
        let sources = AssistantCardType.productHelpSources(view)
        if sources.isEmpty {
            MPText("Chưa tìm thấy hướng dẫn phù hợp.", style: .secondary)
        } else {
            VStack(alignment: .leading, spacing: 8) {
                if let version = string(view["version"]), !version.isEmpty {
                    MPText("Tài liệu v\(version)", style: .secondary)
                }
                ForEach(sources, id: \.id) { source in
                    VStack(alignment: .leading, spacing: 3) {
                        MPText(source.title, style: .emphasis)
                        MPText(source.excerpt, style: .secondary)
                        MPText(source.id, style: .secondary)
                    }
                }
            }
        }
    }

    @ViewBuilder private func walletListView(_ view: [String: JSONValue], showsGoalProgress: Bool = false) -> some View {
        if case .array(let values) = view["items"] {
            ForEach(Array(AssistantCardType.visibleItems(values, for: type).enumerated()), id: \.offset) { _, value in
                if let item = object(value) {
                    let name = string(item["name"]) ?? "Ví"
                    VStack(alignment: .leading, spacing: 4) {
                        MPWalletRow(name: name, type: string(item["type"]) ?? "", balance: money(item["current_balance"]))
                        if showsGoalProgress,
                           let percent = AssistantCardType.goalProgressPercent(current: number(item["current_balance"]), target: number(item["target_amount"])) {
                            MPProgressBar(progress: Double(percent) / 100, label: "Tiến độ \(name)")
                            MPText("\(percent)% mục tiêu", style: .secondary)
                        }
                    }
                }
            }
            MPText(AssistantCardType.sourceLabel(sourceValue), style: .secondary)
        } else { MPText(AssistantCardType.fallbackText(data, masked: masked), style: .secondary) }
    }

    @ViewBuilder private func progressListView(_ view: [String: JSONValue]) -> some View {
        if case .array(let values) = view["items"] {
            ForEach(Array(AssistantCardType.visibleItems(values, for: type).enumerated()), id: \.offset) { _, value in
                if let item = object(value) {
                    let spent = AssistantCardType.monetaryValue(item["spent"])
                    let target = AssistantCardType.progressTarget(item)
                    VStack(alignment: .leading, spacing: 4) {
                        MPText(string(item["name"]) ?? "Mục tiêu")
                        MPText(money(item["spent"]), style: .secondary)
                        if let target, let spent {
                            MPProgressBar(
                                progress: target > 0 ? Double(spent) / target : 0,
                                label: "/ \(MPFormat.money(Int64(target), masked: masked))"
                            )
                        }
                    }
                }
            }
            if AssistantCardType.normalized(type) == "budget_status" {
                MPText("Tổng: \(money(view["spent"])) / \(money(view["limit_amount"]))", style: .secondary)
            } else {
                MPText("Kỳ \(string(view["month"]) ?? "đang chọn")", style: .secondary)
            }
        } else { MPText(AssistantCardType.fallbackText(data, masked: masked), style: .secondary) }
    }

    private var sourceValue: JSONValue? {
        guard let root = rootObject else { return nil }
        return root["source"]
    }

    private var rootObject: [String: JSONValue]? {
        guard case .object(let root) = data else { return nil }
        return root
    }

    @ViewBuilder private func walletConfirmationView(_ view: [String: JSONValue]) -> some View {
        if let input = AssistantWalletProposal.decode(view) {
            MPText("\(input.name) · \(walletTypeLabel(input.type))")
            ForEach(Array(AssistantWalletConfirmationSummary.lines(for: input, masked: masked).enumerated()), id: \.offset) { _, line in
                MPText(line, style: .secondary)
            }
            if walletConfirmationState == .succeeded {
                MPStatusMessage(text: "Đã tạo ví sau khi bạn xác nhận.", tone: .success)
            } else {
                if let walletConfirmationError {
                    MPStatusMessage(text: walletConfirmationError, tone: .danger)
                }
                MPButton(
                    "Xác nhận tạo ví",
                    style: .secondary,
                    isLoading: walletConfirmationState == .saving,
                    isDisabled: (onCreateWallet == nil && onConfirmWalletAction == nil) || !walletConfirmationState.canConfirm
                ) {
                    confirmWallet(
                        input,
                        actionID: AssistantWalletProposal.actionID(view),
                        actionContext: AssistantWalletProposal.actionContext(view)
                    )
                }
                .accessibilityIdentifier("assistant-confirm-wallet")
            }
        } else {
            MPStatusMessage(text: "Đề xuất tạo ví chưa đủ dữ liệu để xác nhận.", tone: .danger)
        }
    }

    private func confirmWallet(_ input: WalletModelInput, actionID: String?, actionContext: (conversationID: String, generation: Int64)?) {
        guard walletConfirmationState.begin() else { return }
        if let actionID, let actionContext, let onConfirmWalletAction {
            runWalletConfirmation { try await onConfirmWalletAction(actionID, actionContext.conversationID, actionContext.generation) }
            return
        }
        guard let onCreateWallet else {
            walletConfirmationState.finish(success: false)
            walletConfirmationError = "Đề xuất không còn khả dụng để xác nhận."
            return
        }
        runWalletConfirmation { try await onCreateWallet(input, actionID ?? walletConfirmationRequestKey) }
    }

    @ViewBuilder private func budgetConfirmationView(_ view: [String: JSONValue]) -> some View {
        if let proposal = AssistantBudgetProposal.decode(view) {
            let isUpdate = proposal.action == "update_budget"
            MPText(proposal.name, style: .emphasis)
            MPText("Giới hạn: \(MPFormat.money(proposal.limitAmount, masked: masked))", style: .secondary)
            MPText("Thời gian: \(proposal.startDate) → \(proposal.endDate)", style: .secondary)
            if let walletID = proposal.walletID { MPText("Ví: \(walletID)", style: .secondary) }
            if let categoryID = proposal.categoryID { MPText("Danh mục: \(categoryID)", style: .secondary) }
            if walletConfirmationState == .succeeded {
                MPStatusMessage(text: isUpdate ? "Đã cập nhật ngân sách sau khi bạn xác nhận." : "Đã tạo ngân sách sau khi bạn xác nhận.", tone: .success)
            } else {
                if let walletConfirmationError { MPStatusMessage(text: walletConfirmationError, tone: .danger) }
                MPButton(isUpdate ? "Xác nhận cập nhật ngân sách" : "Xác nhận tạo ngân sách", style: .secondary, isLoading: walletConfirmationState == .saving, isDisabled: onConfirmBudgetAction == nil || !walletConfirmationState.canConfirm) {
                    guard walletConfirmationState.begin(), let actionID = AssistantJarProposal.actionID(view), let context = AssistantJarProposal.actionContext(view), let onConfirmBudgetAction else {
                        walletConfirmationError = "Đề xuất ngân sách chưa đủ dữ liệu để xác nhận."
                        return
                    }
                    walletConfirmationError = nil
                    Task {
                        do {
                            try await onConfirmBudgetAction(actionID, context.conversationID, context.generation)
                            walletConfirmationState.finish(success: true)
                        } catch let error as APIError {
                            walletConfirmationError = error.userMessage
                            walletConfirmationState.finish(success: false)
                        } catch {
                            walletConfirmationError = isUpdate ? "Không thể cập nhật ngân sách sau khi xác nhận. Bạn có thể thử lại." : "Không thể tạo ngân sách sau khi xác nhận. Bạn có thể thử lại."
                            walletConfirmationState.finish(success: false)
                        }
                    }
                }
                .accessibilityIdentifier("assistant-confirm-budget")
            }
        } else {
            MPStatusMessage(text: "Đề xuất tạo ngân sách chưa đủ dữ liệu để xác nhận.", tone: .danger)
        }
    }

    @ViewBuilder private func jarConfirmationView(_ view: [String: JSONValue]) -> some View {
        if let proposal = AssistantJarProposal.decode(view) {
            let isUpdate = proposal.action == "update_jar"
            MPText(proposal.name, style: .emphasis)
            MPText("Tháng \(proposal.month)", style: .secondary)
            switch proposal.mode {
            case "fixed": MPText("Phân bổ cố định: \(MPFormat.money(proposal.amount ?? 0, masked: masked))", style: .secondary)
            case "percent":
                let percentage = String(format: "%.2f", proposal.percent ?? 0)
                MPText("Phân bổ: \(percentage)% thu nhập", style: .secondary)
            default: MPText("Không đặt mức phân bổ", style: .secondary)
            }
            if walletConfirmationState == .succeeded {
                MPStatusMessage(text: isUpdate ? "Đã cập nhật hũ sau khi bạn xác nhận." : "Đã tạo hũ sau khi bạn xác nhận.", tone: .success)
            } else {
                if let walletConfirmationError { MPStatusMessage(text: walletConfirmationError, tone: .danger) }
                MPButton(isUpdate ? "Xác nhận cập nhật hũ" : "Xác nhận tạo hũ", style: .secondary, isLoading: walletConfirmationState == .saving, isDisabled: onConfirmJarAction == nil || !walletConfirmationState.canConfirm) {
                    guard walletConfirmationState.begin(), let actionID = AssistantJarProposal.actionID(view), let context = AssistantJarProposal.actionContext(view), let onConfirmJarAction else {
                        walletConfirmationError = "Đề xuất hũ chưa đủ dữ liệu để xác nhận."
                        return
                    }
                    walletConfirmationError = nil
                    Task {
                        do {
                            try await onConfirmJarAction(actionID, context.conversationID, context.generation)
                            walletConfirmationState.finish(success: true)
                        } catch let error as APIError {
                            walletConfirmationError = error.userMessage
                            walletConfirmationState.finish(success: false)
                        } catch {
                            walletConfirmationError = isUpdate ? "Không thể cập nhật hũ sau khi xác nhận. Bạn có thể thử lại." : "Không thể tạo hũ sau khi xác nhận. Bạn có thể thử lại."
                            walletConfirmationState.finish(success: false)
                        }
                    }
                }
                .accessibilityIdentifier("assistant-confirm-jar")
            }
        } else {
            MPStatusMessage(text: "Đề xuất tạo hũ chưa đủ dữ liệu để xác nhận.", tone: .danger)
        }
    }

    private func runWalletConfirmation(_ operation: @escaping () async throws -> Void) {
        walletConfirmationError = nil
        Task {
            do {
                try await operation()
                walletConfirmationState.finish(success: true)
            } catch let error as APIError {
                walletConfirmationError = error.userMessage
                walletConfirmationState.finish(success: false)
            } catch {
                walletConfirmationError = "Không thể tạo ví sau khi xác nhận. Bạn có thể thử lại."
                walletConfirmationState.finish(success: false)
            }
        }
    }

    private func walletTypeLabel(_ type: String) -> String {
        switch type {
        case "goal": return "Tiết kiệm"
        case "credit": return "Tín dụng"
        default: return "Thông thường"
        }
    }

    private func metric(_ title: String, _ value: JSONValue?) -> some View { MPMetric(title: title, value: money(value)) }
    private func object(_ value: JSONValue?) -> [String: JSONValue]? { guard let value, case .object(let object) = value else { return nil }; return object }
    private func string(_ value: JSONValue?) -> String? { guard let value, case .string(let text) = value else { return nil }; return text }
    private func number(_ value: JSONValue?) -> Double? { guard let value else { return nil }; if case .number(let number) = value { return number }; return nil }
    private func bool(_ value: JSONValue?) -> Bool? { guard let value, case .bool(let bool) = value else { return nil }; return bool }
    private func money(_ value: JSONValue?) -> String {
        AssistantCardType.money(value, masked: masked)
    }
}

enum AIEntryAttachmentPolicy {
    static func canSelect(filesConfigured: Bool?, isProcessing: Bool) -> Bool {
        filesConfigured == true && !isProcessing
    }
}

enum AIEntryProposalReview {
    static let supportedTypes = [
        MPSelectionOption(id: "expense", label: "Chi tiêu"),
        MPSelectionOption(id: "income", label: "Thu nhập")
    ]

    static func normalizedType(_ value: String) -> String? {
        value == "expense" || value == "income" ? value : nil
    }

    static func typeLabel(_ value: String) -> String {
        switch normalizedType(value) {
        case "expense": return "Chi tiêu"
        case "income": return "Thu nhập"
        default: return "Chưa xác định — chọn loại"
        }
    }

    static func canUseJar(type: String, walletType: String, category: CategoryModel? = nil) -> Bool {
        type == "expense"
            && category?.systemKey != "expense_transfer_out"
            && (walletType == "basic" || walletType == "goal")
    }

    static func jarIDForSave(type: String, walletType: String, category: CategoryModel? = nil, jarID: String?) -> String? {
        canUseJar(type: type, walletType: walletType, category: category) ? jarID : nil
    }

    static func activeJars(from summary: JarMonthSummaryModel) -> [JarIdentityModel] {
        let activeIDs = Set(summary.items.filter(\.active).map(\.jarID))
        return summary.jars.filter { activeIDs.contains($0.jarID) }
    }

    static func pending(_ proposals: [EntryProposalModel]) -> [EntryProposalModel] {
        proposals.filter { $0.status == "pending" }
    }

    @MainActor static func transactionRow(for proposal: EntryProposalModel, category: CategoryModel? = nil, timeZone: TimeZone, masked: Bool) -> MPTransactionRow {
        MPTransactionRow(
            type: proposal.draft.type,
            rawAmount: proposal.draft.amount,
            categoryName: category?.name,
            systemKey: category?.systemKey,
            iconKey: category?.iconKey,
            note: proposal.draft.note,
            occurredAt: proposal.draft.occurredAt,
            timeZone: timeZone,
            masked: masked
        )
    }
}

enum AIEntryRequestRecoveryPolicy {
    static func needsResultRefresh(after error: APIError) -> Bool {
        switch error {
        case .transport, .decoding:
            return true
        case .http(let status):
            return status >= 500
        case .problem(let problem):
            guard let status = problem.status else { return true }
            return status >= 500
        case .invalidURL, .unauthorized, .missingSession, .accountMismatch:
            return false
        }
    }

    static func isRequestNotFound(_ error: APIError) -> Bool {
        switch error {
        case .http(let status): return status == 404
        case .problem(let problem): return problem.status == 404
        default: return false
        }
    }

    static func canSubmit(isUncertain: Bool, isProcessing: Bool, isLocallyProcessing: Bool = false) -> Bool {
        !isUncertain && !isProcessing && !isLocallyProcessing
    }
}

enum AIEntryDecisionPolicy {
    static func canDecide(proposalID: String, deciding: Set<String>, savingAll: Bool) -> Bool {
        !savingAll && !deciding.contains(proposalID)
    }

    static func canSaveAll(savingAll: Bool, processing: Bool, deciding: Set<String>) -> Bool {
        !savingAll && !processing && deciding.isEmpty
    }
}

struct AIEntryView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @Environment(\.mpFinancialValuesMasked) private var masked
    @State private var capabilities: EntryCapabilitiesModel?
    @State private var text = ""
    @State private var process: EntryProcessModel?
    @State private var processing = false
    @State private var error: String?
    @State private var files: [MultipartFile] = []
    @State private var showFilePicker = false
    @State private var wallets: [WalletModel] = []
    @State private var categories: [CategoryModel] = []
    @State private var editProposal: EntryProposalModel?
    @State private var accountTimezone = TimeZone.current.identifier
    @State private var requestID: String?
    @State private var uncertain = false
    @State private var savingAll = false
    @State private var decidingProposalIDs: Set<String> = []

    var body: some View {
        MPEditorScaffold(isBusy: processing || savingAll, header: {
            MPScreenHeader(title: "Nhập giao dịch bằng AI", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: {
                    if !pendingProposals.isEmpty {
                        MPButton("Lưu tất cả", style: .secondary, isLoading: savingAll, isDisabled: !AIEntryDecisionPolicy.canSaveAll(savingAll: savingAll, processing: processing, deciding: decidingProposalIDs)) {
                            Task { await saveAll() }
                        }
                        .accessibilityIdentifier("ai-entry-save-all")
                    }
            })
        }) {
                MPText("AI chỉ đề xuất. Bạn phải duyệt từng giao dịch trước khi lưu.", style: .secondary)
                if let capabilities {
                    MPStatusMessage(text: capabilities.ocrConfigured ? "OCR đã sẵn sàng." : "OCR chưa được cấu hình.", tone: capabilities.ocrConfigured ? .success : .info)
                }
                MPTextEditor(title: "Mô tả", text: $text, placeholder: "Ví dụ: Chi 50.000 cho cà phê hôm nay")
                    .disabled(requestUnresolved)
                    .accessibilityIdentifier("ai-entry-description")
                MPDocumentPicker(title: "Đính kèm ảnh hoặc PDF", count: files.count, maxCount: 20, isPresented: $showFilePicker, onPhotoImport: importPhotos) { urls in
                    importFiles(urls)
                }
                .disabled(!AIEntryAttachmentPolicy.canSelect(filesConfigured: capabilities?.filesConfigured, isProcessing: requestUnresolved))
                ForEach(Array(files.enumerated()), id: \.offset) { index, file in
                    MPAttachmentThumb(name: file.fileName, data: file.data) { files.remove(at: index) }
                        .disabled(requestUnresolved)
                }
                if let error { MPStatusMessage(text: error, tone: .danger) }
                if uncertain {
                    MPStatusMessage(text: "Yêu cầu có thể đã được tiếp nhận. Tải lại kết quả để kiểm tra trước khi gửi nội dung khác. Không tự động gửi lại.", tone: .info)
                }
                if let process {
                    if let processError = process.error { MPStatusMessage(text: processError, tone: .danger) }
                    if let reply = process.reply { MPAssistantCard(title: "Phản hồi của AI") { MPMarkdown(content: reply) } }
                    if !process.processing && process.proposals.isEmpty && process.reply == nil && process.error == nil { MPEmptyView(title: "Chưa có đề xuất", detail: "Hãy thêm mô tả rõ hơn hoặc đính kèm chứng từ.") }
                    ForEach(process.proposals) { proposal in
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 8) {
                                AIEntryProposalReview.transactionRow(
                                    for: proposal,
                                    category: categories.first { $0.id == proposal.draft.categoryID },
                                    timeZone: TimeZone(identifier: accountTimezone) ?? .current,
                                    masked: masked
                                )
                                .accessibilityIdentifier("ai-entry-proposal-row-\(proposal.id)")
                                if AIEntryProposalReview.normalizedType(proposal.draft.type) == nil {
                                    MPStatusMessage(text: AIEntryProposalReview.typeLabel(proposal.draft.type), tone: .info)
                                }
                                if proposal.status == "pending" {
                                    HStack {
                                        MPButton("Sửa", style: .secondary, isDisabled: !AIEntryDecisionPolicy.canDecide(proposalID: proposal.id, deciding: decidingProposalIDs, savingAll: savingAll)) { editProposal = proposal }
                                        MPButton("Bỏ qua", style: .secondary, isLoading: decidingProposalIDs.contains(proposal.id), isDisabled: !AIEntryDecisionPolicy.canDecide(proposalID: proposal.id, deciding: decidingProposalIDs, savingAll: savingAll)) { Task { await decide(proposal, as: "reject") } }
                                        MPButton("Lưu giao dịch", isLoading: decidingProposalIDs.contains(proposal.id), isDisabled: !AIEntryDecisionPolicy.canDecide(proposalID: proposal.id, deciding: decidingProposalIDs, savingAll: savingAll)) { Task { await decide(proposal, as: "approve") } }
                                    }
                                } else { MPStatusMessage(text: proposal.status == "approved" ? "Đã lưu" : "Đã bỏ qua", tone: proposal.status == "approved" ? .success : .info) }
                            }
                        }
                    }
                }
                if (uncertain || process?.processing == true), !processing {
                    MPButton("Tải lại kết quả", style: .secondary, isDisabled: requestID == nil || savingAll) { Task { await refreshResult() } }
                        .accessibilityIdentifier("ai-entry-refresh-result")
                }
                if requestID != nil, !requestUnresolved, process?.proposals.isEmpty != false {
                    MPButton("Soạn nội dung mới", style: .secondary) { requestID = nil; process = nil; text = ""; files = [] }
                }
        } footer: {
            MPButton("Phân tích đề xuất", isLoading: processing, isDisabled: !AIEntryRequestRecoveryPolicy.canSubmit(isUncertain: uncertain, isProcessing: process?.processing == true, isLocallyProcessing: processing) || savingAll) { Task { await processEntry() } }
                .accessibilityIdentifier("ai-entry-process")
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task {
#if DEBUG
            if ProcessInfo.processInfo.arguments.contains("-UITestAIEntryProposal") {
                loadUITestProposal()
            } else if ProcessInfo.processInfo.arguments.contains("-UITestAIEntryAttachments") {
                loadUITestAttachments()
            } else if ProcessInfo.processInfo.arguments.contains("-UITestAIEntryUncertain") || ProcessInfo.processInfo.arguments.contains("-UITestAIEntryRecovery404") {
                loadUITestUncertain()
            } else {
                await loadCapabilities()
                await loadReferenceData()
            }
#else
            await loadCapabilities()
            await loadReferenceData()
#endif
        }
        .overlay {
            if let editProposal {
                EntryProposalEditor(api: api, proposal: editProposal, wallets: wallets, categories: categories, onSaved: { updated in
                    process?.proposals = process?.proposals.map { $0.id == updated.id ? updated : $0 } ?? []
                    self.editProposal = nil
                }, onCancel: { self.editProposal = nil })
                .background(MPColor.canvas.ignoresSafeArea())
            }
        }
    }

    private func loadCapabilities() async { do { capabilities = try await api.entryCapabilities() } catch { self.error = "Không thể tải trạng thái AI entry." } }
    private func loadReferenceData() async { do { async let walletList = api.wallets(); async let categoryList = api.categories(); async let profile = api.profile(); wallets = try await walletList; categories = try await categoryList; if let profile = try? await profile { accountTimezone = profile.timezone } } catch { self.error = "Không thể tải danh sách ví và danh mục." } }
    private func processEntry() async {
        guard AIEntryRequestRecoveryPolicy.canSubmit(isUncertain: uncertain, isProcessing: process?.processing == true, isLocallyProcessing: processing) else { return }
        guard !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || !files.isEmpty else { error = "Nhập mô tả hoặc đính kèm chứng từ trước khi phân tích."; return }
        guard capabilities?.aiConfigured != false else { error = "AI chưa được cấu hình trên máy chủ."; return }
        guard files.isEmpty || capabilities?.filesConfigured == true else { error = "OCR hoặc nơi lưu chứng từ chưa được cấu hình."; return }
        let id = UUID().uuidString
        processing = true; error = nil; defer { processing = false }
        requestID = id
        do {
            process = try await api.processEntry(requestID: id, text: text, timezone: accountTimezone, files: files)
            uncertain = false
            if process?.processing == true { await waitForResult(id: id) }
            if process?.processing != true {
                text = ""
                files = []
                requestID = nil
            }
        } catch let error as APIError {
            if AIEntryRequestRecoveryPolicy.needsResultRefresh(after: error) {
                uncertain = true
                self.error = error.userMessage
            } else {
                requestID = nil
                self.error = error.userMessage
            }
        } catch {
            // A non-API failure after submission has no reliable rejection
            // status; preserve the id and require a GET before any retry.
            uncertain = true
            self.error = "Không thể xác nhận máy chủ đã tiếp nhận yêu cầu."
        }
    }

    private func waitForResult(id: String) async {
        for _ in 0..<105 {
            do { try await Task.sleep(for: .seconds(2)) } catch { return }
            do {
                let next = try await api.entryRequest(id: id)
                process = next
                if !next.processing { return }
            } catch let error as APIError {
                self.error = error.userMessage
                uncertain = true
                return
            } catch {
                self.error = "Không thể tải trạng thái xử lý AI."
                uncertain = true
                return
            }
        }
        error = "Yêu cầu vẫn đang xử lý. Bạn có thể tải lại kết quả; không gửi lại nội dung."
        uncertain = true
    }

    private func refreshResult() async {
        guard let requestID, !processing else { return }
        do {
            let next = try await fetchEntryRequestForRecovery(id: requestID)
            process = next
            uncertain = false
            error = nil
            if !next.processing {
                text = ""
                files = []
                self.requestID = nil
            }
        } catch let error as APIError {
            if AIEntryRequestRecoveryPolicy.isRequestNotFound(error) {
                self.requestID = nil
                self.process = nil
                self.uncertain = false
                self.error = "Chưa tìm thấy yêu cầu đã tiếp nhận. Bạn có thể gửi lại nội dung đang soạn."
            } else {
                self.uncertain = true
                self.error = error.userMessage
            }
        } catch {
            uncertain = true
            self.error = "Không thể tải lại kết quả AI."
        }
    }

    private func fetchEntryRequestForRecovery(id: String) async throws -> EntryProcessModel {
#if DEBUG
        if ProcessInfo.processInfo.arguments.contains("-UITestAIEntryRecovery404") {
            throw APIError.http(status: 404)
        }
#endif
        return try await api.entryRequest(id: id)
    }
    private func decide(_ proposal: EntryProposalModel, as decision: String) async {
        guard AIEntryDecisionPolicy.canDecide(proposalID: proposal.id, deciding: decidingProposalIDs, savingAll: savingAll) else { return }
        decidingProposalIDs.insert(proposal.id)
        defer { decidingProposalIDs.remove(proposal.id) }
        do {
            let updated = try await api.decideEntryProposal(id: proposal.id, version: proposal.version, decision: decision)
            process?.proposals = process?.proposals.map { $0.id == updated.id ? updated : $0 } ?? []
        } catch { self.error = "Không thể cập nhật đề xuất. Tải lại để kiểm tra trạng thái." }
    }

    private var pendingProposals: [EntryProposalModel] {
        AIEntryProposalReview.pending(process?.proposals ?? [])
    }

    private var requestUnresolved: Bool {
        uncertain || processing || process?.processing == true
    }

    private func saveAll() async {
        guard AIEntryDecisionPolicy.canSaveAll(savingAll: savingAll, processing: processing, deciding: decidingProposalIDs) else { return }
        let pending = pendingProposals
        guard !pending.isEmpty else { return }
        savingAll = true
        error = nil
        defer { savingAll = false }
        for proposal in pending {
            do {
                let updated = try await api.decideEntryProposal(id: proposal.id, version: proposal.version, decision: "approve")
                process?.proposals = process?.proposals.map { $0.id == updated.id ? updated : $0 } ?? []
            } catch let apiError as APIError {
                self.error = apiError.userMessage
                return
            } catch {
                self.error = "Không thể lưu tất cả đề xuất. Tải lại để kiểm tra trạng thái."
                return
            }
        }
    }

    private func importFiles(_ urls: [URL]) {
        for url in urls {
            guard files.count < 20 else { break }
            let accessed = url.startAccessingSecurityScopedResource()
            defer { if accessed { url.stopAccessingSecurityScopedResource() } }
            guard let data = try? Data(contentsOf: url), data.count <= 5 * 1024 * 1024 else {
                error = "Mỗi tệp phải nhỏ hơn 5 MiB."
                continue
            }
            let extensionName = url.pathExtension.lowercased()
            let mime: String
            switch extensionName {
            case "jpg", "jpeg": mime = "image/jpeg"
            case "png": mime = "image/png"
            case "pdf": mime = "application/pdf"
            default:
                error = "Chỉ hỗ trợ JPEG, PNG hoặc PDF."
                continue
            }
            if files.contains(where: { $0.fileName == url.lastPathComponent }) {
                error = "Không thể chọn tệp trùng nhau."
                continue
            }
            files.append(MultipartFile(fieldName: "files", fileName: url.lastPathComponent, mimeType: mime, data: data))
        }
    }

    private func importPhotos(_ images: [MPPickedImage]) {
        guard !requestUnresolved else { return }
        for image in images {
            guard files.count < 20 else { break }
            if files.contains(where: { $0.data == image.data }) {
                error = "Ảnh này đã được đính kèm."
                continue
            }
            files.append(MultipartFile(fieldName: "files", fileName: image.name,
                                       mimeType: image.mimeType, data: image.data))
        }
    }

#if DEBUG
    private func loadUITestUncertain() {
        capabilities = EntryCapabilitiesModel(aiConfigured: true, ocrConfigured: true, filesConfigured: true)
        text = "Chi 50.000 cho cà phê hôm nay"
        requestID = "ui-entry-uncertain-request"
        uncertain = true
    }

    private func loadUITestProposal() {
        accountTimezone = "Asia/Ho_Chi_Minh"
        capabilities = EntryCapabilitiesModel(aiConfigured: true, ocrConfigured: true, filesConfigured: true)
        wallets = [WalletModel(id: "ui-entry-wallet", name: "Tiền mặt", type: "basic", currency: "VND", openingBalance: 0, currentBalance: 0, availableCredit: nil, isInTotal: true, description: nil, targetAmount: nil, targetDate: nil, creditLimit: nil, lastStatementBalance: nil, statementDay: nil, paymentDueDay: nil)]
        categories = []
        process = EntryProcessModel(
            id: "ui-entry-process",
            proposals: [EntryProposalModel(
                id: "ui-entry-proposal",
                processID: "ui-entry-process",
                version: 1,
                status: "pending",
                draft: EntryDraftModel(type: "unknown", amount: 50000, walletID: "ui-entry-wallet", categoryID: nil, jarID: nil, occurredAt: "2026-10-06T08:00:00+07:00", note: "Fixture proposal", includedInReports: true),
                questions: ["Chọn loại giao dịch trước khi lưu."],
                transactionID: nil
            )],
            processing: false,
            reply: nil,
            error: nil
        )
        requestID = "ui-entry-process"
    }

    private func loadUITestAttachments() {
        capabilities = EntryCapabilitiesModel(aiConfigured: true, ocrConfigured: true, filesConfigured: true)
        let tinyPNG = Data(base64Encoded: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=") ?? Data()
        files = (1...20).map { index in
            MultipartFile(
                fieldName: "files",
                fileName: String(format: "IMG_%02d.PNG", index),
                mimeType: "image/png",
                data: tinyPNG
            )
        }
    }
#endif
}

struct EntryProposalEditor: View {
    let api: FinanceAPI
    let proposal: EntryProposalModel
    let wallets: [WalletModel]
    let categories: [CategoryModel]
    let onSaved: (EntryProposalModel) -> Void
    let onCancel: () -> Void
    @State private var amount: String
    @State private var type: String
    @State private var walletID: String
    @State private var categoryID: String
    @State private var jarID: String
    @State private var includedInReports: Bool
    @State private var note: String
    @State private var date: Date
    @State private var jarSummary: JarMonthSummaryModel?
    @State private var jarsLoading = false
    @State private var jarLoadError: String?
    @State private var saving = false
    @State private var error: String?
    @Environment(\.timeZone) private var timeZone

    init(api: FinanceAPI, proposal: EntryProposalModel, wallets: [WalletModel], categories: [CategoryModel], onSaved: @escaping (EntryProposalModel) -> Void, onCancel: @escaping () -> Void) {
        self.api = api; self.proposal = proposal; self.wallets = wallets; self.categories = categories; self.onSaved = onSaved; self.onCancel = onCancel
        _amount = State(initialValue: String(proposal.draft.amount))
        _type = State(initialValue: AIEntryProposalReview.normalizedType(proposal.draft.type) ?? "")
        let initialWalletID = proposal.draft.walletID.isEmpty ? (TransactionsViewModel.ordinaryTransactionWallets(wallets).first?.id ?? "") : proposal.draft.walletID
        let initialType = AIEntryProposalReview.normalizedType(proposal.draft.type) ?? ""
        let initialWalletType = wallets.first(where: { $0.id == initialWalletID })?.type ?? "basic"
        _walletID = State(initialValue: initialWalletID)
        _categoryID = State(initialValue: TransactionEditorPolicy.requiredCategoryID(type: initialType, walletID: initialWalletID, walletType: initialWalletType, categories: categories, currentID: proposal.draft.categoryID ?? ""))
        _jarID = State(initialValue: proposal.draft.jarID ?? "")
        _includedInReports = State(initialValue: proposal.draft.includedInReports)
        _note = State(initialValue: proposal.draft.note)
        _date = State(initialValue: MPFormat.parseInstant(proposal.draft.occurredAt) ?? Date())
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Sửa đề xuất", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
                MPAmountField(title: "Số tiền", text: $amount)
                MPOptionSelector(title: "Loại giao dịch", options: AIEntryProposalReview.supportedTypes, selectionID: $type, emptyLabel: "Chưa xác định — chọn loại")
                MPOptionSelector(title: "Ví", options: TransactionsViewModel.ordinaryTransactionWallets(wallets).map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $walletID)
                MPOptionSelector(title: "Danh mục", options: categoryOptions, selectionID: $categoryID)
                if canUseJar {
                    MPOptionSelector(
                        title: "Hũ chi tiêu",
                        options: TransactionEditorPolicy.jarOptions(summary: jarSummary, selectedJarID: jarID, selectedJarName: nil),
                        selectionID: $jarID,
                        emptyLabel: jarsLoading ? "Đang tải hũ..." : "Không chọn"
                    )
                    if let jarLoadError { MPStatusMessage(text: jarLoadError, tone: .danger) }
                }
                MPBooleanSelector(title: "Tính vào báo cáo", value: $includedInReports, trueLabel: "Có", falseLabel: "Không")
                MPDateTimeField(title: "Ngày và giờ", date: $date, timeZone: timeZone.identifier)
                MPTextEditor(title: "Ghi chú", text: $note)
                if let error { MPStatusMessage(text: error, tone: .danger) }
                MPButton("Lưu đề xuất", isLoading: saving) { Task { await save() } }
                    .accessibilityIdentifier("ai-entry-proposal-save")
            }
            .padding(MPMetrics.pagePadding)
        }
        .accessibilityIdentifier("ai-entry-proposal-editor")
        .safeAreaInset(edge: .bottom) {
            MPBottomNavigationInset()
                .accessibilityIdentifier("ai-entry-proposal-editor-shell-inset")
        }
        .task {
            clearInvalidCategory()
            await loadJarsForDate()
        }
        .onChange(of: type) { _, _ in
            if !canUseJar { jarID = "" }
            clearInvalidCategory()
            Task { await loadJarsForDate() }
        }
        .onChange(of: walletID) { _, _ in
            if !canUseJar { jarID = "" }
            clearInvalidCategory()
            Task { await loadJarsForDate() }
        }
        .onChange(of: categoryID) { _, _ in
            if !canUseJar { jarID = "" }
        }
        .onChange(of: date) { _, _ in Task { await loadJarsForDate() } }
    }

    private func save() async {
        guard let normalizedType = AIEntryProposalReview.normalizedType(type) else { error = "Chọn loại giao dịch là thu nhập hoặc chi tiêu trước khi lưu."; return }
        guard let parsed = MPFormat.parsePositiveInteger(amount), let wallet = wallets.first(where: { $0.id == walletID }) else { error = "Chọn ví và nhập số tiền hợp lệ."; return }
        guard let selectedCategory = categories.first(where: { $0.id == categoryID }) else { error = "Hãy chọn nhóm cho giao dịch."; return }
        if !selectedCategory.applies(to: normalizedType, walletID: wallet.id, walletType: wallet.type) { error = "Danh mục không phù hợp với loại giao dịch hoặc ví đã chọn."; return }
        if !jarID.isEmpty {
            guard canUseJar else { error = "Hũ chỉ áp dụng cho khoản chi thường hoặc ví tiết kiệm."; return }
            guard TransactionEditorPolicy.hasActiveJar(jarID, summary: jarSummary) else { error = "Hũ đã gỡ khỏi tháng này. Hãy bỏ chọn hoặc chọn hũ đang hoạt động."; return }
        }
        saving = true; defer { saving = false }
        let draft = EntryDraftModel(type: normalizedType, amount: parsed, walletID: wallet.id, categoryID: selectedCategory.id, jarID: jarID.isEmpty ? nil : jarID, occurredAt: Self.timestamp(from: date, timeZone: timeZone), note: note.trimmingCharacters(in: .whitespacesAndNewlines), includedInReports: includedInReports)
        do { onSaved(try await api.editEntryProposal(id: proposal.id, version: proposal.version, draft: draft)) }
        catch { self.error = "Không thể cập nhật đề xuất." }
    }

    private var selectedWallet: WalletModel? { wallets.first(where: { $0.id == walletID }) }
    private var selectedCategory: CategoryModel? { categories.first(where: { $0.id == categoryID }) }
    private var canUseJar: Bool {
        AIEntryProposalReview.canUseJar(type: type, walletType: selectedWallet?.type ?? "", category: selectedCategory)
    }
    private var categoryOptions: [MPSelectionOption] {
        let valid = TransactionEditorPolicy.applicableCategories(type: type, walletID: selectedWallet?.id ?? "", walletType: selectedWallet?.type ?? "basic", categories: categories)
        var options = valid.map { MPSelectionOption(id: $0.id, label: $0.name) }
        if !categoryID.isEmpty, !options.contains(where: { $0.id == categoryID }), let current = categories.first(where: { $0.id == categoryID }) {
            options.append(MPSelectionOption(id: current.id, label: current.name))
        }
        return options
    }

    private func clearInvalidCategory() {
        guard let selectedWallet else { return }
        categoryID = TransactionEditorPolicy.requiredCategoryID(type: type, walletID: selectedWallet.id, walletType: selectedWallet.type, categories: categories, currentID: categoryID)
    }

    private func loadJarsForDate() async {
#if DEBUG
        if ProcessInfo.processInfo.arguments.contains("-UITestAIEntryProposal") {
            jarSummary = JarMonthSummaryModel(
                month: "2026-10",
                timezone: "Asia/Ho_Chi_Minh",
                actualIncome: 0,
                totalSpent: 0,
                unassignedSpent: 0,
                totalAllocated: 0,
                totalAllocationPercent: 0,
                overIncome: false,
                overOneHundredPercent: false,
                jars: [JarIdentityModel(jarID: "ui-entry-jar", name: "Ăn uống")],
                items: [JarItemModel(jarID: "ui-entry-jar", name: "Ăn uống", allocationMode: "amount", allocationAmount: nil, allocationPercent: nil, calculatedAllocation: nil, spent: 0, active: true)]
            )
            jarsLoading = false
            return
        }
#endif
        guard canUseJar else {
            jarSummary = nil
            jarLoadError = nil
            jarsLoading = false
            return
        }
        let month = String(MPFormat.calendarDateString(date, timeZone: timeZone).prefix(7))
        guard month != "0000-00", month != "" else { return }
        jarsLoading = true
        jarLoadError = nil
        defer { jarsLoading = false }
        do {
            jarSummary = try await api.jarMonth(month: month)
        } catch {
            jarSummary = nil
            jarLoadError = "Không thể tải hũ của tháng này."
        }
    }

    private static func timestamp(from date: Date, timeZone: TimeZone) -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "yyyy-MM-dd'T'HH:mm:ssXXXXX"
        formatter.locale = Locale(identifier: "en_US_POSIX")
        formatter.timeZone = timeZone
        return formatter.string(from: date)
    }
}

struct AccountView: View {
    @ObservedObject var auth: AuthViewModel
    let onBack: () -> Void
    @State private var profile: ProfileModel?
    @State private var keys: [APIKeyModel] = []
    @State private var timezone = TimeZone.current.identifier
    @State private var keyName = ""
    @State private var showKeyEditor = false
    @State private var createdSecret: String?
    @State private var createdKeyID: String?
    @State private var copyConfirmation: String?
    @State private var faceIDEnrollmentPresented = false
    @State private var faceIDPassword = ""
    @State private var forgetDeviceConfirmation = false
    @State private var error: String?

    var body: some View {
        VStack(spacing: 0) {
            // Keep navigation fixed while account settings scroll. This is
            // especially important on smaller phones where the API-key and
            // security sections extend beyond the first viewport.
            MPScreenHeader(title: "Tài khoản", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                .padding(.horizontal, MPMetrics.pagePadding)

            ScrollView {
                VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                if case .authenticated(let session) = auth.phase {
                    MPSurfaceCard(style: .reference) {
                        VStack(alignment: .leading, spacing: 8) {
                            MPText(session.user.name ?? "MyPocket")
                            if let email = session.user.email { MPText(email, style: .secondary) }
                            MPText("Timezone: \(session.user.timezone ?? TimeZone.current.identifier)", style: .secondary)
                        }
                    }
                }
                if let profile {
                    MPSurfaceCard(style: .reference) {
                        VStack(alignment: .leading, spacing: 10) {
                            MPText("Múi giờ tài khoản", style: .title)
                            MPTextField(title: "IANA timezone", text: $timezone, placeholder: profile.timezone)
                            MPButton("Lưu múi giờ", style: .secondary) { Task { await saveTimezone() } }
                        }
                    }
                }
                MPSurfaceCard(style: .reference) {
                    VStack(alignment: .leading, spacing: 10) {
                        HStack { MPText("API keys", style: .title); Spacer(); MPIconButton(systemName: "plus", label: "Tạo API key") { showKeyEditor = true }.accessibilityIdentifier("account-api-key-add") }
                        MPText("API key có toàn quyền đọc và thay đổi dữ liệu tài khoản của bạn. Chỉ chia sẻ với dịch vụ bạn tin tưởng.", style: .secondary)
                            .accessibilityIdentifier("account-api-key-full-access")
                        if keys.isEmpty { MPText("Chưa có API key.", style: .secondary) }
                        ForEach(keys) { key in
                            HStack {
                                VStack(alignment: .leading) { MPText(key.name); MPText("Toàn quyền tài khoản", style: .secondary) }
                                Spacer()
                                MPButton("Thu hồi", style: .destructive) { Task { await revoke(key) } }
                                    .accessibilityIdentifier("account-api-key-revoke-\(key.id)")
                            }
                        }
                        if showKeyEditor {
                            MPTextField(title: "Tên key", text: $keyName, placeholder: "Tích hợp báo cáo")
                            HStack { MPButton("Hủy", style: .secondary) { showKeyEditor = false }; MPButton("Tạo") { Task { await createKey() } }.accessibilityIdentifier("account-api-key-create") }
                        }
                        if let createdSecret {
                            MPSurfaceCard(style: .reference) {
                                VStack(alignment: .leading, spacing: 10) {
                                    MPText("Secret chỉ hiển thị một lần", style: .emphasis)
                                    MPText(createdSecret, style: .secondary)
                                        .accessibilityIdentifier("account-api-key-secret-value")
                                    MPButton("Sao chép API key", style: .secondary, fullWidth: false) {
                                        let didCopy = APIKeySecretClipboard.copy(createdSecret) { secret in
                                            UIPasteboard.general.setItems(
                                                [[UTType.plainText.identifier: secret]],
                                                options: [
                                                    .localOnly: true,
                                                    .expirationDate: Date(timeIntervalSinceNow: 5 * 60)
                                                ]
                                            )
                                        }
                                        copyConfirmation = didCopy ? "Đã sao chép API key." : nil
                                        if didCopy {
                                            UIAccessibility.post(notification: .announcement, argument: "Đã sao chép API key.")
                                        }
                                    }
                                    .accessibilityIdentifier("account-api-key-copy")
                                    if let copyConfirmation {
                                        MPStatusMessage(text: copyConfirmation, tone: .success)
                                            .accessibilityIdentifier("account-api-key-copy-confirmation")
                                    }
                                }
                            }
                        }
                    }
                }
                MPSurfaceCard(style: .reference) {
                    VStack(alignment: .leading, spacing: 10) {
                        MPText("Bảo mật", style: .title)
                        MPText("Đăng nhập lại bằng Face ID, kể cả sau khi đăng xuất. Chỉ dùng trên iPhone này.", style: .secondary)
                        MPButton(auth.canSignInWithBiometrics ? "Tắt đăng nhập Face ID" : "Bật đăng nhập Face ID", style: .secondary, isLoading: auth.isBusy) {
                            if auth.canSignInWithBiometrics { Task { await auth.disableBiometric() } }
                            else { faceIDPassword = ""; faceIDEnrollmentPresented = true }
                        }
                        .accessibilityIdentifier("account-face-id-toggle")
                        if auth.rememberedDevice != nil {
                            MPButton("Gỡ tài khoản khỏi thiết bị", style: .plain) { forgetDeviceConfirmation = true }
                                .accessibilityIdentifier("account-forget-device")
                        }
                    }
                }
                if let error = auth.errorMessage { MPStatusMessage(text: error, tone: .danger) }
                if let error { MPStatusMessage(text: error, tone: .danger) }
                MPButton("Đăng xuất", style: .destructive, isLoading: auth.isBusy) { Task { await auth.logout() } }
                    .accessibilityIdentifier("account-logout")
                }
                .padding(MPMetrics.pagePadding)
            }
            .background(MPColor.canvas.ignoresSafeArea())
        }
        .task { await loadAccount() }
        .mpInteractiveBackBlocked(showKeyEditor || faceIDEnrollmentPresented || auth.isBusy || (profile != nil && timezone != profile?.timezone))
        .mpFormBottomSheet(isPresented: $faceIDEnrollmentPresented, title: "Bật đăng nhập Face ID", accessibilityIdentifier: "face-id-enrollment", isDismissDisabled: auth.isBusy, onClose: { faceIDPassword = ""; faceIDEnrollmentPresented = false }) {
            MPText("Xác nhận mật khẩu để ghi nhớ tài khoản trên iPhone này.", style: .secondary)
            MPSecureField(title: "Mật khẩu", text: $faceIDPassword, placeholder: "Nhập mật khẩu", accessibilityIdentifier: "face-id-enrollment-password")
            if let message = auth.errorMessage { MPStatusMessage(text: message, tone: .danger) }
        } footer: {
            MPButton("Bật Face ID", isLoading: auth.isBusy, isDisabled: faceIDPassword.isEmpty) {
                Task {
                    if await auth.enrollBiometric(password: faceIDPassword, deviceName: UIDevice.current.model) {
                        faceIDPassword = ""
                        faceIDEnrollmentPresented = false
                    }
                }
            }
            .accessibilityIdentifier("face-id-enrollment-confirm")
        }
        .mpConfirmDialog(isPresented: $forgetDeviceConfirmation, title: "Gỡ tài khoản khỏi thiết bị?", message: "Face ID sẽ bị tắt và tài khoản được đăng xuất trên iPhone này.", confirmTitle: "Gỡ tài khoản") {
            Task { await auth.forgetDevice() }
        }
    }

    private func loadAccount() async {
        do { async let loadedProfile = auth.financeAPI.profile(); async let loadedKeys = auth.financeAPI.apiKeys(); profile = try await loadedProfile; keys = try await loadedKeys.filter { $0.revokedAt == nil }; timezone = profile?.timezone ?? timezone }
        catch { self.error = "Không thể tải thiết lập tài khoản." }
    }

    private func saveTimezone() async {
        do {
            profile = try await auth.financeAPI.updateTimezone(timezone)
            auth.updateAccountTimezone(profile?.timezone ?? timezone)
            error = nil
        }
        catch { self.error = "Múi giờ không hợp lệ hoặc không thể cập nhật." }
    }

    private func createKey() async {
        guard !keyName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { error = "Nhập tên API key."; return }
        do { let created = try await auth.financeAPI.createAPIKey(name: keyName); createdSecret = created.secret; createdKeyID = created.id; copyConfirmation = nil; keyName = ""; showKeyEditor = false; keys = try await auth.financeAPI.apiKeys().filter { $0.revokedAt == nil } }
        catch { self.error = "Không thể tạo API key." }
    }

    private func revoke(_ key: APIKeyModel) async {
        do {
            try await auth.financeAPI.revokeAPIKey(id: key.id)
            keys.removeAll { $0.id == key.id }
            if createdKeyID == key.id {
                createdKeyID = nil
                createdSecret = nil
                copyConfirmation = nil
            }
        }
        catch { self.error = "Không thể thu hồi API key." }
    }
}
