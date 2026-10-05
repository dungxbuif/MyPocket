import SwiftUI

struct AssistantView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    let ownerID: String
    @State private var capabilities: AdvisorCapabilitiesModel?
    @State private var overview: AdvisorOverviewModel?
    @State private var messages: [AdvisorMessageModel] = []
    @State private var conversationID = ""
    @State private var input = ""
    @State private var sending = false
    @State private var historyLoading = true
    @State private var loadingOlder = false
    @State private var hasOlder = false
    @State private var activeRunID: String?
    @State private var error: String?

    init(api: FinanceAPI, onBack: @escaping () -> Void, ownerID: String) {
        self.api = api
        self.onBack = onBack
        self.ownerID = ownerID
        let key = Self.conversationKey(ownerID: ownerID)
        _conversationID = State(initialValue: UserDefaults.standard.string(forKey: key) ?? "")
    }

    private static func conversationKey(ownerID: String) -> String {
        "mypocket.advisor.conversation.\(ownerID)"
    }

    var body: some View {
        VStack(spacing: 0) {
            MPScreenHeader(title: "Finance Assistant", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: {
                if !conversationID.isEmpty {
                    MPIconButton(systemName: "arrow.clockwise", label: "Tải lại lịch sử") { Task { await reloadHistory() } }
                }
            })
                .padding(.horizontal, MPMetrics.pagePadding)
            ScrollView {
                VStack(alignment: .leading, spacing: 12) {
                    if let capabilities, !capabilities.enabled { MPStatusMessage(text: "Finance Assistant chưa được bật.", tone: .info) }
                    if let overview {
                        AssistantStructuredCard(
                            type: overview.viewKind,
                            data: .object(["view": overview.view ?? .null, "source": overview.source ?? .null])
                        )
                    }
                    if historyLoading {
                        MPLoadingView(label: "Đang tải lịch sử trợ lý...")
                    } else if messages.isEmpty {
                        MPEmptyView(title: "Hỏi về tài chính của bạn", detail: "Ví dụ: Tháng này tôi chi ăn uống bao nhiêu?")
                    }
                    if hasOlder {
                        MPButton("Tải tin cũ", style: .secondary, isLoading: loadingOlder) { Task { await loadOlderMessages() } }
                    }
                    ForEach(messages) { message in
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 8) {
                                MPText(message.role == "user" ? "Bạn" : "Trợ lý", style: .secondary)
                                ForEach(Array(message.parts.enumerated()), id: \.offset) { _, part in
                                    if let text = part.text { MPText(text) }
                                    else if let data = part.data {
                                        AssistantStructuredCard(type: part.type, data: data)
                                    } else { MPText("Đã nhận thẻ dữ liệu tài chính.", style: .secondary) }
                                }
                            }
                        }
                    }
                    if let error { MPStatusMessage(text: error, tone: .danger) }
                }
                .padding(MPMetrics.pagePadding)
            }
            VStack(spacing: 8) {
                MPTextEditor(title: "", text: $input, placeholder: "Hỏi về tiền của bạn...")
                    .frame(minHeight: 90)
                HStack {
                    MPButton("Gửi", isLoading: sending) { Task { await send() } }
                    if let activeRunID {
                        MPButton("Hủy", style: .destructive) { Task { await cancel(runID: activeRunID) } }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
            .background(.bar)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await loadInitialState() }
        .accessibilityIdentifier("assistant-screen")
    }

    private func loadInitialState() async {
        historyLoading = true
        defer { historyLoading = false }
        do { capabilities = try await api.advisorCapabilities() }
        catch let error as APIError { self.error = error.userMessage }
        catch { self.error = "Không thể kết nối Finance Assistant." }
        do { overview = try await api.advisorOverview() }
        catch let error as APIError { self.error = error.userMessage }
        catch { self.error = "Không thể tải tổng quan trợ lý tài chính." }
        if !conversationID.isEmpty {
            await reloadHistory()
        }
    }

    private func reloadHistory() async {
        guard !conversationID.isEmpty else { return }
        do {
            messages = try await api.advisorMessages(conversationID: conversationID)
            hasOlder = messages.count == 50
            error = nil
        } catch let error as APIError { self.error = error.userMessage }
        catch { self.error = "Không thể tải lại lịch sử trợ lý." }
    }

    private func loadOlderMessages() async {
        guard !conversationID.isEmpty, let first = messages.first, !loadingOlder, hasOlder else { return }
        loadingOlder = true
        defer { loadingOlder = false }
        do {
            let older = try await api.advisorMessages(conversationID: conversationID, beforeSequence: first.seq)
            messages = mergeMessages(messages, older)
            hasOlder = older.count == 50
            error = nil
        } catch let error as APIError { self.error = error.userMessage }
        catch { self.error = "Không thể tải tin cũ của trợ lý." }
    }

    private func mergeMessages(_ current: [AdvisorMessageModel], _ incoming: [AdvisorMessageModel]) -> [AdvisorMessageModel] {
        var byID = Dictionary(uniqueKeysWithValues: current.map { ($0.id, $0) })
        incoming.forEach { byID[$0.id] = $0 }
        return byID.values.sorted { $0.seq < $1.seq }
    }

    private func send() async {
        let text = input.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty, activeRunID == nil else { return }
        sending = true; error = nil; defer { sending = false }
        do {
            let result = try await api.submitAdvisorMessage(AdvisorMessageInputModel(clientRequestID: UUID().uuidString, text: text))
            conversationID = result.conversationID
            UserDefaults.standard.set(result.conversationID, forKey: Self.conversationKey(ownerID: ownerID))
            activeRunID = result.runID
            input = ""
            if result.status == "queued" || result.status == "running" {
                var completed = false
                for _ in 0..<60 {
                    try await Task.sleep(for: .milliseconds(500))
                    let run = try await api.advisorRun(id: result.runID)
                    if ["completed", "failed", "cancelled", "interrupted", "purged"].contains(run.status) {
                        completed = true
                        if run.status != "completed" { self.error = "Yêu cầu trợ lý kết thúc với trạng thái \(run.status)." }
                        break
                    }
                }
                guard completed else {
                    self.error = "Yêu cầu vẫn đang xử lý. Bạn có thể bấm Hủy trước khi gửi yêu cầu khác."
                    return
                }
            }
            await reloadHistory()
            activeRunID = nil
        } catch let error as APIError { self.error = error.userMessage }
        catch { self.error = "Không thể xử lý yêu cầu trợ lý tài chính." }
    }

    private func cancel(runID: String) async {
        do { try await api.cancelAdvisorRun(id: runID); sending = false; activeRunID = nil; error = "Đã hủy yêu cầu." }
        catch { self.error = "Không thể hủy yêu cầu đang chạy." }
    }
}

struct AssistantStructuredCard: View {
    let type: String
    let data: JSONValue

    var body: some View {
        MPAssistantCard(title: type.replacingOccurrences(of: "_", with: " ")) {
            if let view = viewObject {
                switch type {
                case "metric_group": metricView(view)
                case "period_comparison": comparisonView(view)
                case "transaction_list": transactionListView(view)
                case "wallet_balances", "goal_progress": walletListView(view)
                case "budget_status", "jar_progress": progressListView(view)
                default: MPText(data.prettyDescription, style: .secondary)
                }
            } else {
                MPText(data.prettyDescription, style: .secondary)
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
            if let count = number(view["count"]) { MPMetricCard(title: "Giao dịch", value: String(Int(count))) }
        }
    }

    @ViewBuilder private func comparisonView(_ view: [String: JSONValue]) -> some View {
        metric("Hiện tại", view["current_expense"])
        metric("Kỳ trước", view["previous_expense"])
        metric("Chênh lệch", view["delta"])
    }

    @ViewBuilder private func transactionListView(_ view: [String: JSONValue]) -> some View {
        if case .array(let values) = view["items"] {
            ForEach(Array(values.enumerated()), id: \.offset) { _, value in
                if let item = object(value) {
                    MPTransactionRow(title: string(item["note"]) ?? "Giao dịch", subtitle: string(item["occurred_at"]) ?? "", amount: money(item["amount"]), isIncome: string(item["type"]) == "income")
                }
            }
        } else { MPText(data.prettyDescription, style: .secondary) }
    }

    @ViewBuilder private func walletListView(_ view: [String: JSONValue]) -> some View {
        if case .array(let values) = view["items"] {
            ForEach(Array(values.enumerated()), id: \.offset) { _, value in
                if let item = object(value) {
                    MPWalletRow(name: string(item["name"]) ?? "Ví", type: string(item["type"]) ?? "", balance: money(item["current_balance"]))
                }
            }
        } else { MPText(data.prettyDescription, style: .secondary) }
    }

    @ViewBuilder private func progressListView(_ view: [String: JSONValue]) -> some View {
        if case .array(let values) = view["items"] {
            ForEach(Array(values.enumerated()), id: \.offset) { _, value in
                if let item = object(value) {
                    let spent = number(item["spent"]) ?? 0
                    let limit = number(item["limit_amount"] ?? item["calculated_allocation"]) ?? 1
                    VStack(alignment: .leading, spacing: 4) { MPText(string(item["name"]) ?? "Mục tiêu"); MPProgressBar(progress: spent / max(1, limit), label: "\(money(item["spent"])) / \(money(item["limit_amount"] ?? item["calculated_allocation"]))") }
                }
            }
        } else { MPText(data.prettyDescription, style: .secondary) }
    }

    private func metric(_ title: String, _ value: JSONValue?) -> some View { MPMetricCard(title: title, value: money(value)) }
    private func object(_ value: JSONValue?) -> [String: JSONValue]? { guard let value, case .object(let object) = value else { return nil }; return object }
    private func string(_ value: JSONValue?) -> String? { guard let value, case .string(let text) = value else { return nil }; return text }
    private func number(_ value: JSONValue?) -> Double? { guard let value else { return nil }; if case .number(let number) = value { return number }; return nil }
    private func money(_ value: JSONValue?) -> String { MPFormat.money(Int64(number(value) ?? 0)) }
}

enum AIEntryAttachmentPolicy {
    static func canSelect(filesConfigured: Bool?, isProcessing: Bool) -> Bool {
        filesConfigured == true && !isProcessing
    }
}

struct AIEntryView: View {
    let api: FinanceAPI
    let onBack: () -> Void
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

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Nhập giao dịch bằng AI", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                MPText("AI chỉ đề xuất. Bạn phải duyệt từng giao dịch trước khi lưu.", style: .secondary)
                if let capabilities {
                    MPStatusMessage(text: capabilities.ocrConfigured ? "OCR đã sẵn sàng." : "OCR chưa được cấu hình.", tone: capabilities.ocrConfigured ? .success : .info)
                }
                MPTextEditor(title: "Mô tả", text: $text, placeholder: "Ví dụ: Chi 50.000 cho cà phê hôm nay")
                MPDocumentPicker(title: "Đính kèm ảnh hoặc PDF", count: files.count, maxCount: 20, isPresented: $showFilePicker) { urls in
                    importFiles(urls)
                }
                .disabled(!AIEntryAttachmentPolicy.canSelect(filesConfigured: capabilities?.filesConfigured, isProcessing: processing))
                ForEach(Array(files.enumerated()), id: \.offset) { index, file in
                    MPAttachmentThumb(name: file.fileName, data: file.data) { files.remove(at: index) }
                }
                MPButton("Phân tích đề xuất", isLoading: processing, isDisabled: process?.processing == true) { Task { await processEntry() } }
                if let error { MPStatusMessage(text: error, tone: .danger) }
                if let process {
                    if let processError = process.error { MPStatusMessage(text: processError, tone: .danger) }
                    if process.processing && !processing {
                        MPButton("Tải lại kết quả", style: .secondary) { Task { await refreshResult() } }
                    }
                    if let reply = process.reply { MPAssistantCard(title: "Phản hồi của AI") { MPText(reply) } }
                    if !process.processing && process.proposals.isEmpty && process.reply == nil && process.error == nil { MPEmptyView(title: "Chưa có đề xuất", detail: "Hãy thêm mô tả rõ hơn hoặc đính kèm chứng từ.") }
                    ForEach(process.proposals) { proposal in
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 8) {
                                MPText(proposal.draft.type == "income" ? "Thu nhập" : "Chi tiêu")
                                MPText("\(MPFormat.money(proposal.draft.amount)) · \(MPFormat.shortDate(proposal.draft.occurredAt, timeZone: TimeZone(identifier: accountTimezone) ?? .current))", style: .secondary)
                                if !proposal.draft.note.isEmpty { MPText(proposal.draft.note, style: .secondary) }
                                if proposal.status == "pending" {
                                    HStack { MPButton("Sửa", style: .secondary) { editProposal = proposal }; MPButton("Bỏ qua", style: .secondary) { Task { await decide(proposal, as: "reject") } }; MPButton("Lưu giao dịch") { Task { await decide(proposal, as: "approve") } } }
                                } else { MPStatusMessage(text: proposal.status == "approved" ? "Đã lưu" : "Đã bỏ qua", tone: proposal.status == "approved" ? .success : .info) }
                            }
                        }
                    }
                }
                if requestID != nil, !processing, process?.processing != true, process?.proposals.isEmpty != false {
                    MPButton("Soạn nội dung mới", style: .secondary) { requestID = nil; process = nil; text = ""; files = [] }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await loadCapabilities(); await loadReferenceData() }
        .overlay {
            if let editProposal {
                EntryProposalEditor(api: api, proposal: editProposal, wallets: wallets, categories: categories, onSaved: { updated in
                    process?.proposals = process?.proposals.map { $0.id == updated.id ? updated : $0 } ?? []
                    self.editProposal = nil
                }, onCancel: { self.editProposal = nil })
                .background(MPColor.canvas.ignoresSafeArea())
            }
        }
        .accessibilityIdentifier("ai-entry-screen")
    }

    private func loadCapabilities() async { do { capabilities = try await api.entryCapabilities() } catch { self.error = "Không thể tải trạng thái AI entry." } }
    private func loadReferenceData() async { do { async let walletList = api.wallets(); async let categoryList = api.categories(); async let profile = api.profile(); wallets = try await walletList; categories = try await categoryList; if let profile = try? await profile { accountTimezone = profile.timezone } } catch { self.error = "Không thể tải danh sách ví và danh mục." } }
    private func processEntry() async {
        guard !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || !files.isEmpty else { error = "Nhập mô tả hoặc đính kèm chứng từ trước khi phân tích."; return }
        guard capabilities?.aiConfigured != false else { error = "AI chưa được cấu hình trên máy chủ."; return }
        guard files.isEmpty || capabilities?.filesConfigured == true else { error = "OCR hoặc nơi lưu chứng từ chưa được cấu hình."; return }
        let id = UUID().uuidString
        processing = true; error = nil; defer { processing = false }
        requestID = id
        do {
            process = try await api.processEntry(requestID: id, text: text, timezone: accountTimezone, files: files)
            if process?.processing == true { await waitForResult(id: id) }
            if process?.processing != true {
                text = ""
                files = []
                requestID = nil
            }
        } catch let error as APIError {
            self.error = "\(error.userMessage) Không có giao dịch nào được lưu."
        } catch { self.error = "Không thể xử lý yêu cầu. Không có giao dịch nào được lưu." }
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
                return
            } catch {
                self.error = "Không thể tải trạng thái xử lý AI."
                return
            }
        }
        error = "Yêu cầu vẫn đang xử lý. Bạn có thể tải lại kết quả; không gửi lại nội dung."
    }

    private func refreshResult() async {
        guard let requestID, !processing else { return }
        do {
            let next = try await api.entryRequest(id: requestID)
            process = next
            error = nil
            if !next.processing {
                text = ""
                files = []
                self.requestID = nil
            }
        } catch let error as APIError {
            self.error = error.userMessage
        } catch {
            self.error = "Không thể tải lại kết quả AI."
        }
    }
    private func decide(_ proposal: EntryProposalModel, as decision: String) async {
        do { let updated = try await api.decideEntryProposal(id: proposal.id, version: proposal.version, decision: decision); process?.proposals = process?.proposals.map { $0.id == updated.id ? updated : $0 } ?? [] }
        catch { self.error = "Không thể cập nhật đề xuất. Tải lại để kiểm tra trạng thái." }
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
}

struct EntryProposalEditor: View {
    let api: FinanceAPI
    let proposal: EntryProposalModel
    let wallets: [WalletModel]
    let categories: [CategoryModel]
    let onSaved: (EntryProposalModel) -> Void
    let onCancel: () -> Void
    @State private var amount: String
    @State private var walletID: String
    @State private var categoryID: String
    @State private var note: String
    @State private var date: Date
    @State private var saving = false
    @State private var error: String?
    @Environment(\.timeZone) private var timeZone

    init(api: FinanceAPI, proposal: EntryProposalModel, wallets: [WalletModel], categories: [CategoryModel], onSaved: @escaping (EntryProposalModel) -> Void, onCancel: @escaping () -> Void) {
        self.api = api; self.proposal = proposal; self.wallets = wallets; self.categories = categories; self.onSaved = onSaved; self.onCancel = onCancel
        _amount = State(initialValue: String(proposal.draft.amount))
        _walletID = State(initialValue: proposal.draft.walletID.isEmpty ? (TransactionsViewModel.ordinaryTransactionWallets(wallets).first?.id ?? "") : proposal.draft.walletID)
        _categoryID = State(initialValue: proposal.draft.categoryID ?? "")
        _note = State(initialValue: proposal.draft.note)
        _date = State(initialValue: MPFormat.parseInstant(proposal.draft.occurredAt) ?? Date())
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Sửa đề xuất", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
                MPAmountField(title: "Số tiền", text: $amount)
                MPOptionSelector(title: "Ví", options: TransactionsViewModel.ordinaryTransactionWallets(wallets).map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $walletID)
                MPOptionSelector(title: "Danh mục", options: categoryOptions, selectionID: $categoryID, emptyLabel: "Không chọn")
                MPDateField(title: "Ngày giao dịch", date: $date)
                MPTextEditor(title: "Ghi chú", text: $note)
                if let error { MPStatusMessage(text: error, tone: .danger) }
                MPButton("Lưu đề xuất", isLoading: saving) { Task { await save() } }
            }
            .padding(MPMetrics.pagePadding)
        }
        .onChange(of: walletID) { _, _ in clearInvalidCategory() }
    }

    private func save() async {
        guard let parsed = MPFormat.parsePositiveInteger(amount), let wallet = wallets.first(where: { $0.id == walletID }) else { error = "Chọn ví và nhập số tiền hợp lệ."; return }
        let selectedCategory = categories.first(where: { $0.id == categoryID })
        if let selectedCategory, !selectedCategory.applies(to: proposal.draft.type, walletID: wallet.id, walletType: wallet.type) { error = "Danh mục không phù hợp với loại giao dịch hoặc ví đã chọn."; return }
        saving = true; defer { saving = false }
        let draft = EntryDraftModel(type: proposal.draft.type, amount: parsed, walletID: wallet.id, categoryID: selectedCategory?.id, jarID: proposal.draft.jarID, occurredAt: MPFormat.instantString(date), note: note.trimmingCharacters(in: .whitespacesAndNewlines), includedInReports: proposal.draft.includedInReports)
        do { onSaved(try await api.editEntryProposal(id: proposal.id, version: proposal.version, draft: draft)) }
        catch { self.error = "Không thể cập nhật đề xuất." }
    }

    private var selectedWallet: WalletModel? { wallets.first(where: { $0.id == walletID }) }
    private var categoryOptions: [MPSelectionOption] {
        let valid = categories.filter { $0.applies(to: proposal.draft.type, walletID: selectedWallet?.id ?? "", walletType: selectedWallet?.type ?? "basic") }
        var options = [MPSelectionOption(id: "", label: "Không chọn")]
        options += valid.map { MPSelectionOption(id: $0.id, label: $0.name) }
        if !categoryID.isEmpty, !options.contains(where: { $0.id == categoryID }), let current = categories.first(where: { $0.id == categoryID }) {
            options.append(MPSelectionOption(id: current.id, label: current.name))
        }
        return options
    }

    private func clearInvalidCategory() {
        guard let selectedCategory = categories.first(where: { $0.id == categoryID }), let selectedWallet else { return }
        if !selectedCategory.applies(to: proposal.draft.type, walletID: selectedWallet.id, walletType: selectedWallet.type) {
            categoryID = ""
        }
    }
}

struct AccountView: View {
    @ObservedObject var auth: AuthViewModel
    let onBack: () -> Void
    @State private var profile: ProfileModel?
    @State private var keys: [APIKeyModel] = []
    @State private var timezone = TimeZone.current.identifier
    @State private var keyName = ""
    @State private var keyScope = "finance:read"
    @State private var showKeyEditor = false
    @State private var createdSecret: String?
    @State private var error: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Tài khoản", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                if case .authenticated(let session) = auth.phase {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 8) {
                            MPText(session.user.name ?? "MyPocket")
                            if let email = session.user.email { MPText(email, style: .secondary) }
                            MPText("Timezone: \(session.user.timezone ?? TimeZone.current.identifier)", style: .secondary)
                        }
                    }
                }
                if let profile {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 10) {
                            MPText("Múi giờ tài khoản", style: .title)
                            MPTextField(title: "IANA timezone", text: $timezone, placeholder: profile.timezone)
                            MPButton("Lưu múi giờ", style: .secondary) { Task { await saveTimezone() } }
                        }
                    }
                }
                MPSurfaceCard {
                    VStack(alignment: .leading, spacing: 10) {
                        HStack { MPText("API keys", style: .title); Spacer(); MPIconButton(systemName: "plus", label: "Tạo API key") { showKeyEditor = true } }
                        if keys.isEmpty { MPText("Chưa có API key.", style: .secondary) }
                        ForEach(keys) { key in
                            HStack {
                                VStack(alignment: .leading) { MPText(key.name); MPText(key.scopes.joined(separator: ", "), style: .secondary) }
                                Spacer()
                                MPButton("Thu hồi", style: .destructive) { Task { await revoke(key) } }
                            }
                        }
                        if showKeyEditor {
                            MPTextField(title: "Tên key", text: $keyName, placeholder: "Tích hợp báo cáo")
                            MPWalletSelector(title: "Scope", options: ["finance:read", "advisor:read", "advisor:chat", "feedback:read", "feedback:write"], selection: $keyScope)
                            HStack { MPButton("Hủy", style: .secondary) { showKeyEditor = false }; MPButton("Tạo") { Task { await createKey() } } }
                        }
                        if let createdSecret { MPAssistantCard(title: "Secret chỉ hiển thị một lần") { MPText(createdSecret, style: .secondary) } }
                    }
                }
                MPSurfaceCard {
                    VStack(alignment: .leading, spacing: 10) {
                        MPText("Bảo mật", style: .title)
                        MPText("Face ID chỉ dùng để mở phiên đã lưu trong Keychain.", style: .secondary)
                        MPButton(auth.biometricEnabled ? "Face ID đang bật" : "Bật Face ID", style: .secondary) { Task { await auth.toggleBiometric() } }
                    }
                }
                if let error = auth.errorMessage { MPStatusMessage(text: error, tone: .danger) }
                if let error { MPStatusMessage(text: error, tone: .danger) }
                MPButton("Đăng xuất", style: .destructive, action: auth.logout)
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await loadAccount() }
        .accessibilityIdentifier("account-screen")
    }

    private func loadAccount() async {
        do { async let loadedProfile = auth.financeAPI.profile(); async let loadedKeys = auth.financeAPI.apiKeys(); profile = try await loadedProfile; keys = try await loadedKeys; timezone = profile?.timezone ?? timezone }
        catch { self.error = "Không thể tải thiết lập tài khoản." }
    }

    private func saveTimezone() async {
        do { profile = try await auth.financeAPI.updateTimezone(timezone); error = nil }
        catch { self.error = "Múi giờ không hợp lệ hoặc không thể cập nhật." }
    }

    private func createKey() async {
        guard !keyName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { error = "Nhập tên API key."; return }
        do { let created = try await auth.financeAPI.createAPIKey(name: keyName, scopes: [keyScope]); createdSecret = created.secret; keyName = ""; showKeyEditor = false; keys = try await auth.financeAPI.apiKeys() }
        catch { self.error = "Không thể tạo API key." }
    }

    private func revoke(_ key: APIKeyModel) async {
        do { try await auth.financeAPI.revokeAPIKey(id: key.id); keys.removeAll { $0.id == key.id } }
        catch { self.error = "Không thể thu hồi API key." }
    }
}
