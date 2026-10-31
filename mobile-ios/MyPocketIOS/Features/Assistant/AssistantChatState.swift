import Foundation

enum AssistantChatState: Equatable, Sendable {
    case idle
    case sending(requestID: String)
    case accepted(runID: String)
    case processing(runID: String)
    case reviewReady(batchID: String)
    case failed(message: String)
    case cancelled
    case uncertain(requestID: String)
    case resetting
}

/// Small, UI-owned state machine for the single full-screen assistant
/// surface. Server polling stays in the existing observer; this type keeps
/// composer labels, retry policy and action affordances deterministic.
enum AssistantChatStatePolicy {
    static func acceptsSubmit(_ state: AssistantChatState) -> Bool {
        switch state {
        case .idle, .failed, .cancelled: return true
        case .sending, .accepted, .processing, .reviewReady, .uncertain, .resetting: return false
        }
    }

    static func canCancel(_ state: AssistantChatState) -> Bool {
        switch state {
        case .accepted, .processing: return true
        default: return false
        }
    }

    static func displayLabel(_ state: AssistantChatState) -> String? {
        switch state {
        case .sending: return "Đang gửi…"
        case .accepted: return "Đã tiếp nhận"
        case .processing: return "Trợ lý đang phân tích…"
        case .reviewReady: return "Có đề xuất cần xem lại"
        case .failed(let message): return message
        case .cancelled: return "Đã hủy yêu cầu"
        case .uncertain: return "Yêu cầu có thể đã được tiếp nhận"
        case .resetting: return "Đang tạo cuộc trò chuyện mới…"
        case .idle: return nil
        }
    }
}
