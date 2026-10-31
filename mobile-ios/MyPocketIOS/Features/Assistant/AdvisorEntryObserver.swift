import Foundation
import Combine

/// GET-only observer for the optional typed-entry task contract. It is safe to
/// instantiate before the matching backend is deployed: callers can keep the
/// transcript read-only when the endpoint is unavailable.
@MainActor
final class AdvisorEntryObserver: ObservableObject {
    typealias FetchTask = (String) async throws -> AdvisorEntryTaskModel

    @Published private(set) var task: AdvisorEntryTaskModel?
    @Published private(set) var state: AssistantChatState = .idle
    @Published private(set) var lastError: String?

    private let fetchTask: FetchTask
    private var token = UUID()

    init(fetchTask: @escaping FetchTask) {
        self.fetchTask = fetchTask
    }

    func observe(taskID: String) async {
        let id = taskID.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !id.isEmpty else { return }
        let currentToken = UUID()
        token = currentToken
        state = .processing(runID: id)
        lastError = nil
        for delay in [Duration.seconds(0), .seconds(1), .seconds(2), .seconds(3), .seconds(5)] {
            guard token == currentToken else { return }
            do {
                if delay > .zero { try await Task.sleep(for: delay) }
                let result = try await fetchTask(id)
                guard token == currentToken else { return }
                task = result
                switch result.status.lowercased() {
                case "completed", "review_ready": state = .reviewReady(batchID: result.batchID ?? id); return
                case "cancelled": state = .cancelled; return
                case "failed", "rejected": state = .failed(message: result.errorCode ?? "Không thể xử lý đề xuất."); return
                default: state = .processing(runID: id)
                }
            } catch is CancellationError {
                return
            } catch {
                guard token == currentToken else { return }
                lastError = error.localizedDescription
                state = .failed(message: "Chưa thể tải trạng thái đề xuất.")
                return
            }
        }
    }

    func reset() {
        token = UUID()
        task = nil
        lastError = nil
        state = .idle
    }
}
