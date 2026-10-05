import SwiftUI

struct AssistantView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @State private var capabilities: AdvisorCapabilitiesModel?
    @State private var messages: [AdvisorMessageModel] = []
    @State private var conversationID = ""
    @State private var input = ""
    @State private var sending = false
    @State private var error: String?

    var body: some View {
        VStack(spacing: 0) {
            MPScreenHeader(title: "Finance Assistant", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                .padding(.horizontal, MPMetrics.pagePadding)
            ScrollView {
                VStack(alignment: .leading, spacing: 12) {
                    if let capabilities, !capabilities.enabled { MPStatusMessage(text: "Finance Assistant chưa được bật.", tone: .info) }
                    if messages.isEmpty { MPEmptyView(title: "Hỏi về tài chính của bạn", detail: "Ví dụ: Tháng này tôi chi ăn uống bao nhiêu?") }
                    ForEach(messages) { message in
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 8) {
                                MPText(message.role == "user" ? "Bạn" : "Trợ lý", style: .secondary)
                                ForEach(Array(message.parts.enumerated()), id: \.offset) { _, part in
                                    if let text = part.text { MPText(text) }
                                    else if let data = part.data {
                                        MPAssistantCard(title: part.type.replacingOccurrences(of: "_", with: " ")) {
                                            MPText(data.prettyDescription, style: .secondary)
                                        }
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
                MPButton("Gửi", isLoading: sending) { Task { await send() } }
            }
            .padding(MPMetrics.pagePadding)
            .background(.bar)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await loadCapabilities() }
        .accessibilityIdentifier("assistant-screen")
    }

    private func loadCapabilities() async {
        do { capabilities = try await api.advisorCapabilities() }
        catch { self.error = "Không thể kết nối Finance Assistant." }
    }

    private func send() async {
        let text = input.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty else { return }
        sending = true; error = nil; defer { sending = false }
        do {
            let result = try await api.submitAdvisorMessage(AdvisorMessageInputModel(clientRequestID: UUID().uuidString, text: text))
            conversationID = result.conversationID
            input = ""
            messages = try await api.advisorMessages(conversationID: conversationID)
        } catch { self.error = "Không thể xử lý yêu cầu trợ lý tài chính." }
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
                ForEach(Array(files.enumerated()), id: \.offset) { index, file in
                    MPAttachmentThumb(name: file.fileName) { files.remove(at: index) }
                }
                MPButton("Phân tích đề xuất", isLoading: processing) { Task { await processEntry() } }
                if let error { MPStatusMessage(text: error, tone: .danger) }
                if let process {
                    if let reply = process.reply { MPAssistantCard(title: "Phản hồi của AI") { MPText(reply) } }
                    if process.proposals.isEmpty { MPEmptyView(title: "Chưa có đề xuất", detail: "Hãy thêm mô tả rõ hơn hoặc đính kèm chứng từ.") }
                    ForEach(process.proposals) { proposal in
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 8) {
                                MPText(proposal.draft.type == "income" ? "Thu nhập" : "Chi tiêu")
                                MPText("\(MPFormat.money(proposal.draft.amount)) · \(MPFormat.shortDate(proposal.draft.occurredAt))", style: .secondary)
                                if !proposal.draft.note.isEmpty { MPText(proposal.draft.note, style: .secondary) }
                                if proposal.status == "pending" {
                                    HStack { MPButton("Sửa", style: .secondary) { editProposal = proposal }; MPButton("Bỏ qua", style: .secondary) { Task { await decide(proposal, as: "reject") } }; MPButton("Lưu giao dịch") { Task { await decide(proposal, as: "approve") } } }
                                } else { MPStatusMessage(text: proposal.status == "approved" ? "Đã lưu" : "Đã bỏ qua", tone: proposal.status == "approved" ? .success : .info) }
                            }
                        }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
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
    private func loadReferenceData() async { do { async let walletList = api.wallets(); async let categoryList = api.categories(); wallets = try await walletList; categories = try await categoryList } catch { self.error = "Không thể tải danh sách ví và danh mục." } }
    private func processEntry() async {
        guard !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { error = "Nhập mô tả trước khi phân tích."; return }
        processing = true; error = nil; defer { processing = false }
        do { process = try await api.processEntry(requestID: UUID().uuidString, text: text, timezone: TimeZone.current.identifier, files: files) }
        catch { self.error = "Không thể xử lý yêu cầu. Không có giao dịch nào được lưu." }
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
            let mime = url.pathExtension.lowercased() == "pdf" ? "application/pdf" : "image/\(url.pathExtension.lowercased() == "jpg" ? "jpeg" : url.pathExtension.lowercased())"
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
    @State private var walletName: String
    @State private var categoryName: String
    @State private var note: String
    @State private var date: Date
    @State private var saving = false
    @State private var error: String?

    init(api: FinanceAPI, proposal: EntryProposalModel, wallets: [WalletModel], categories: [CategoryModel], onSaved: @escaping (EntryProposalModel) -> Void, onCancel: @escaping () -> Void) {
        self.api = api; self.proposal = proposal; self.wallets = wallets; self.categories = categories; self.onSaved = onSaved; self.onCancel = onCancel
        _amount = State(initialValue: String(proposal.draft.amount))
        _walletName = State(initialValue: wallets.first(where: { $0.id == proposal.draft.walletID })?.name ?? wallets.first?.name ?? "")
        _categoryName = State(initialValue: categories.first(where: { $0.id == proposal.draft.categoryID })?.name ?? "Không chọn")
        _note = State(initialValue: proposal.draft.note)
        _date = State(initialValue: ISO8601DateFormatter().date(from: proposal.draft.occurredAt) ?? Date())
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Sửa đề xuất", leading: { MPIconButton(systemName: "xmark", label: "Đóng", action: onCancel) }, trailing: { EmptyView() })
                MPAmountField(title: "Số tiền", text: $amount)
                MPWalletSelector(title: "Ví", options: wallets.map(\.name), selection: $walletName)
                MPWalletSelector(title: "Danh mục", options: ["Không chọn"] + categories.map(\.name), selection: $categoryName)
                MPDateField(title: "Ngày giao dịch", date: $date)
                MPTextEditor(title: "Ghi chú", text: $note)
                if let error { MPStatusMessage(text: error, tone: .danger) }
                MPButton("Lưu đề xuất", isLoading: saving) { Task { await save() } }
            }
            .padding(MPMetrics.pagePadding)
        }
    }

    private func save() async {
        guard let parsed = Int64(amount.filter(\.isNumber)), parsed > 0, let wallet = wallets.first(where: { $0.name == walletName }) else { error = "Chọn ví và nhập số tiền hợp lệ."; return }
        saving = true; defer { saving = false }
        let draft = EntryDraftModel(type: proposal.draft.type, amount: parsed, walletID: wallet.id, categoryID: categories.first(where: { $0.name == categoryName })?.id, jarID: proposal.draft.jarID, occurredAt: ISO8601DateFormatter().string(from: date), note: note, includedInReports: proposal.draft.includedInReports)
        do { onSaved(try await api.editEntryProposal(id: proposal.id, version: proposal.version, draft: draft)) }
        catch { self.error = "Không thể cập nhật đề xuất." }
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
