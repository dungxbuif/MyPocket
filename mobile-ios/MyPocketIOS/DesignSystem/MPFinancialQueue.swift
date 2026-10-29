import SwiftUI

private struct MPFinancialQueueKey: EnvironmentKey {
    static let defaultValue: FinancialMutationQueue? = nil
}

extension EnvironmentValues {
    var mpFinancialQueue: FinancialMutationQueue? {
        get { self[MPFinancialQueueKey.self] }
        set { self[MPFinancialQueueKey.self] = newValue }
    }
}

struct MPFinancialQueueStatus: View {
    @ObservedObject var queue: FinancialMutationQueue
    var onEdit: ((OutboxMutation) -> Void)? = nil
    var canEdit: ((OutboxMutation) -> Bool)? = nil
    var onDiscarded: () -> Void = {}
    @Environment(\.mpFinancialValuesMasked) private var masked
    @State private var discardID: UUID?
    @State private var confirmDiscard = false
    @State private var busyID: UUID?
    @State private var actionError: String?

    var body: some View {
        if !queue.items.isEmpty {
            MPSurfaceCard {
                VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                    MPText("Chờ xác nhận · \(queue.items.count)", style: .title)
                    MPText("Chưa tính vào số dư và báo cáo trên máy chủ.", style: .secondary)
                    ForEach(queue.items, id: \.id) { item in
                        HStack(alignment: .top, spacing: MPMetrics.sectionSpacing) {
                            VStack(alignment: .leading, spacing: 4) {
                                HStack {
                                    MPText(title(item.payload))
                                    Spacer()
                                    MPText(MPFormat.money(amount(item.payload), masked: masked))
                                }
                                if let note = note(item.payload), !note.isEmpty { MPText(note, style: .secondary) }
                                MPStatusBadge(text: label(item.state), tone: item.state == .conflict || item.state == .rejected ? .danger : .muted)
                            }
                            if item.state != .sending {
                                MPOverflowMenu(label: "Thao tác bản nháp", actions: actions(for: item))
                                    .disabled(busyID != nil || queue.isSyncing)
                            }
                        }
                        .accessibilityElement(children: .contain)
                        .accessibilityIdentifier("outbox-item-\(item.id.uuidString)")
                    }
                    if queue.result == .unsupported {
                        MPStatusMessage(text: "Máy chủ chưa xác nhận hỗ trợ đồng bộ. Bản nháp vẫn được giữ trên thiết bị.")
                    }
                    if let message = queue.errorMessage { MPStatusMessage(text: message, tone: .danger) }
                    if let actionError { MPStatusMessage(text: actionError, tone: .danger) }
                    MPButton("Đồng bộ", style: .secondary, isLoading: queue.isSyncing) {
                        Task { try? await queue.drain(reason: .manualRetry) }
                    }
                }
            }
            .accessibilityElement(children: .contain)
            .accessibilityIdentifier("financial-outbox-section")
            .mpConfirmDialog(isPresented: $confirmDiscard, title: "Bỏ bản nháp trên thiết bị?",
                message: "Nếu lần gửi trước mất kết nối, giao dịch có thể đã được lưu trên máy chủ. Thao tác này chỉ dừng bản nháp, không xóa giao dịch đã lưu. Hãy tải lại dữ liệu để kiểm tra.", confirmTitle: "Bỏ bản nháp") {
                    guard let id = discardID else { return }
                    perform(id) { try await queue.discard(id: id); onDiscarded() }
                }
        }
    }

    private func actions(for item: OutboxMutation) -> [MPOverflowAction] {
        var actions: [MPOverflowAction] = []
        if item.state == .retryWait {
            actions.append(MPOverflowAction(title: "Thử lại ngay") {
                perform(item.id) { try await queue.retryNow(id: item.id) }
            })
        }
        if (item.state == .conflict || item.state == .rejected), let onEdit,
           canEdit?(item) ?? true {
            actions.append(MPOverflowAction(title: "Sửa bản nháp") { onEdit(item) })
        }
        if item.state != .sending {
            actions.append(MPOverflowAction(title: "Bỏ bản nháp") {
                discardID = item.id; confirmDiscard = true
            })
        }
        return actions
    }

    private func perform(_ id: UUID, action: @escaping @MainActor () async throws -> Void) {
        guard busyID == nil else { return }
        busyID = id; actionError = nil
        Task { @MainActor in
            defer { busyID = nil }
            do { try await action() }
            catch { actionError = "Trạng thái bản nháp đã thay đổi hoặc chưa thể xử lý. Hãy kiểm tra tài khoản và đồng bộ lại." }
        }
    }

    private func title(_ payload: OutboxPayload) -> String {
        switch payload {
        case .transaction(let input): return input.type == "income" ? "Thu nhập" : "Chi tiêu"
        case .transfer: return "Chuyển ví"
        }
    }
    private func amount(_ payload: OutboxPayload) -> Int64 {
        switch payload { case .transaction(let input): input.amount; case .transfer(let input): input.amount }
    }
    private func note(_ payload: OutboxPayload) -> String? {
        switch payload { case .transaction(let input): input.note; case .transfer(let input): input.note }
    }
    private func label(_ state: OutboxState) -> String {
        switch state {
        case .pending: "Đang chờ gửi"
        case .sending: "Đang xác nhận"
        case .retryWait: "Chờ thử lại"
        case .blockedAuth: "Cần đăng nhập lại đúng tài khoản"
        case .conflict: "Có xung đột — chưa tự xử lý"
        case .rejected: "Máy chủ từ chối — cần kiểm tra"
        case .applied: "Đã xác nhận"
        case .cancelled: "Đã bỏ bản nháp"
        }
    }
}
