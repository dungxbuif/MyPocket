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
                                    else { MPText("Đã nhận thẻ dữ liệu tài chính.", style: .secondary) }
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

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Nhập giao dịch bằng AI", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                MPText("AI chỉ đề xuất. Bạn phải duyệt từng giao dịch trước khi lưu.", style: .secondary)
                if let capabilities {
                    MPStatusMessage(text: capabilities.ocrConfigured ? "OCR đã sẵn sàng." : "OCR chưa được cấu hình.", tone: capabilities.ocrConfigured ? .success : .info)
                }
                MPTextEditor(title: "Mô tả", text: $text, placeholder: "Ví dụ: Chi 50.000 cho cà phê hôm nay")
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
                                    HStack { MPButton("Bỏ qua", style: .secondary) { Task { await decide(proposal, as: "reject") } }; MPButton("Lưu giao dịch") { Task { await decide(proposal, as: "approve") } } }
                                } else { MPStatusMessage(text: proposal.status == "approved" ? "Đã lưu" : "Đã bỏ qua", tone: proposal.status == "approved" ? .success : .info) }
                            }
                        }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await loadCapabilities() }
        .accessibilityIdentifier("ai-entry-screen")
    }

    private func loadCapabilities() async { do { capabilities = try await api.entryCapabilities() } catch { self.error = "Không thể tải trạng thái AI entry." } }
    private func processEntry() async {
        guard !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { error = "Nhập mô tả trước khi phân tích."; return }
        processing = true; error = nil; defer { processing = false }
        do { process = try await api.processEntry(requestID: UUID().uuidString, text: text, timezone: TimeZone.current.identifier) }
        catch { self.error = "Không thể xử lý yêu cầu. Không có giao dịch nào được lưu." }
    }
    private func decide(_ proposal: EntryProposalModel, as decision: String) async {
        do { let updated = try await api.decideEntryProposal(id: proposal.id, version: proposal.version, decision: decision); process?.proposals = process?.proposals.map { $0.id == updated.id ? updated : $0 } ?? [] }
        catch { self.error = "Không thể cập nhật đề xuất. Tải lại để kiểm tra trạng thái." }
    }
}

struct AccountView: View {
    @ObservedObject var auth: AuthViewModel
    let onBack: () -> Void

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
                MPSurfaceCard {
                    VStack(alignment: .leading, spacing: 10) {
                        MPText("Bảo mật", style: .title)
                        MPText("Face ID chỉ dùng để mở phiên đã lưu trong Keychain.", style: .secondary)
                        MPButton(auth.biometricEnabled ? "Face ID đang bật" : "Bật Face ID", style: .secondary) { Task { await auth.toggleBiometric() } }
                    }
                }
                if let error = auth.errorMessage { MPStatusMessage(text: error, tone: .danger) }
                MPButton("Đăng xuất", style: .destructive, action: auth.logout)
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .accessibilityIdentifier("account-screen")
    }
}
