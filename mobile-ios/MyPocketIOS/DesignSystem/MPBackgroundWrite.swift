import SwiftUI

/// Local-first write coordinator for forms that cannot use the offline outbox (the server's
/// offline contract only covers transaction/transfer creates). The form closes at once; the
/// request runs in the background, the owner screen reloads on success, and a failure is shown
/// inline with Retry/Dismiss instead of blocking the form or losing the user's change.
@MainActor
final class MPBackgroundWriteCenter: ObservableObject {
    struct Failure: Identifiable {
        let id = UUID()
        /// What could not be saved, written by the form that started the write.
        let message: String
        /// The server's explanation when it gave one (capitalised for display).
        let detail: String?
        let retry: () -> Void
    }

    @Published private(set) var pending = 0
    @Published private(set) var failures: [Failure] = []

    func submit(
        failure message: String,
        request: @escaping () async throws -> Void,
        onSuccess: @escaping () -> Void = {}
    ) {
        pending += 1
        Task {
            do {
                try await request()
                pending -= 1
                onSuccess()
            } catch {
                pending -= 1
                let detail = (error as? APIError)?.userMessage.trimmingCharacters(in: .whitespacesAndNewlines)
                let shown = detail.flatMap { $0.isEmpty ? nil : $0.prefix(1).uppercased() + $0.dropFirst() }
                failures.append(Failure(message: message, detail: shown) { [weak self] in
                    self?.submit(failure: message, request: request, onSuccess: onSuccess)
                })
            }
        }
    }

    func retry(_ failure: Failure) {
        failures.removeAll { $0.id == failure.id }
        failure.retry()
    }

    func dismiss(_ failure: Failure) { failures.removeAll { $0.id == failure.id } }
}

private struct MPBackgroundWriteKey: EnvironmentKey {
    // Environment defaults are read on the main thread; a fallback keeps previews/tests writing.
    nonisolated(unsafe) static let defaultValue: MPBackgroundWriteCenter = MainActor.assumeIsolated { MPBackgroundWriteCenter() }
}

extension EnvironmentValues {
    var mpBackgroundWrites: MPBackgroundWriteCenter {
        get { self[MPBackgroundWriteKey.self] }
        set { self[MPBackgroundWriteKey.self] = newValue }
    }
}

struct MPBackgroundWriteStatus: View {
    @ObservedObject var center: MPBackgroundWriteCenter

    /// The progress row is visible exactly while at least one background write is in flight.
    static func showsProgress(pending: Int) -> Bool { pending > 0 }

    /// Shown as an overlay by the shell so it never pushes the content or the bottom navigation.
    var body: some View {
        if center.pending > 0 || !center.failures.isEmpty {
            VStack(spacing: 8) {
                if Self.showsProgress(pending: center.pending) { MPInlineRefreshProgress(label: "Đang lưu thay đổi…") }
                ForEach(center.failures) { failure in
                    VStack(alignment: .leading, spacing: 8) {
                        MPStatusMessage(text: failure.detail.map { "\(failure.message) \($0)" } ?? failure.message, tone: .danger)
                            .fixedSize(horizontal: false, vertical: true)
                        HStack(spacing: 8) {
                            MPButton("Thử lại", style: .secondary) { center.retry(failure) }
                            MPButton("Bỏ qua", style: .secondary) { center.dismiss(failure) }
                        }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
            .frame(maxWidth: .infinity)
            .background(MPColor.canvas)
            .overlay(alignment: .bottom) { Divider() }
            .shadow(color: .black.opacity(0.08), radius: 6, y: 3)
            .zIndex(1)
            .accessibilityElement(children: .contain)
            .accessibilityIdentifier("background-write-status")
        }
    }
}
