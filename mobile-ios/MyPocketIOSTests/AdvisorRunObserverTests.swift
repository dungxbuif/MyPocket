import Foundation
import XCTest
@testable import MyPocketIOS

@MainActor
final class AdvisorRunObserverTests: XCTestCase {
    func testAdvisorObserverWaitsForNinetySecondRun() async {
        var responses = Array(repeating: run(status: "running"), count: 30) + [run(status: "completed")]
        var sleeps: [Duration] = []
        let observer = AdvisorRunObserver(
            fetchRun: { _ in responses.removeFirst() },
            sleep: { duration in sleeps.append(duration) }
        )

        await observer.observe(ownerID: "owner-1", runID: "run-1")

        XCTAssertEqual(observer.terminalResult?.status, "completed")
        XCTAssertEqual(sleeps.prefix(4), [.seconds(1), .seconds(2), .seconds(3), .seconds(3)])
        XCTAssertEqual(observer.observedRunID, "run-1")
    }

    func testAdvisorObserverForegroundRecoversByGET() async {
        var fetchCount = 0
        let observer = AdvisorRunObserver(
            fetchRun: { _ in
                fetchCount += 1
                return fetchCount == 1 ? self.run(status: "running") : self.run(status: "completed")
            },
            sleep: { _ in }
        )

        await observer.observe(ownerID: "owner-1", runID: "run-1")

        XCTAssertEqual(fetchCount, 2)
        XCTAssertEqual(observer.terminalResult?.status, "completed")
    }

    func testAdvisorObserverRejectsPreviousOwnerResponse() async {
        var responses = [run(status: "completed", ownerID: "previous-owner"), run(status: "completed", ownerID: "owner-1")]
        let observer = AdvisorRunObserver(
            fetchRun: { _ in responses.removeFirst() },
            sleep: { _ in }
        )

        await observer.observe(ownerID: "owner-1", runID: "run-1")

        XCTAssertEqual(observer.terminalResult?.ownerID, "owner-1")
        XCTAssertNil(observer.lastError)
    }

    func testAdvisorObserverNetworkFailureKeepsRunIdentity() async {
        var attempts = 0
        var sleeps: [Duration] = []
        let observer = AdvisorRunObserver(
            fetchRun: { _ in
                attempts += 1
                if attempts == 1 { throw URLError(.notConnectedToInternet) }
                return self.run(status: "completed")
            },
            sleep: { duration in sleeps.append(duration) }
        )

        await observer.observe(ownerID: "owner-1", runID: "run-1")

        XCTAssertEqual(observer.observedRunID, "run-1")
        XCTAssertEqual(observer.terminalResult?.status, "completed")
        XCTAssertEqual(sleeps.first, .seconds(3))
    }

    func testAdvisorObserverRelaunchUsesOnlyTheNewRunIdentity() async {
        var requestedIDs: [String] = []
        let observer = AdvisorRunObserver(
            fetchRun: { id in
                requestedIDs.append(id)
                return self.run(id: id, status: "completed")
            },
            sleep: { _ in }
        )

        await observer.observe(ownerID: "owner-1", runID: "run-before-relaunch")
        XCTAssertEqual(observer.terminalResult?.id, "run-before-relaunch")

        observer.reset()
        XCTAssertNil(observer.terminalResult)
        XCTAssertNil(observer.observedRunID)

        await observer.observe(ownerID: "owner-1", runID: "run-after-relaunch")

        XCTAssertEqual(requestedIDs, ["run-before-relaunch", "run-after-relaunch"])
        XCTAssertEqual(observer.observedRunID, "run-after-relaunch")
        XCTAssertEqual(observer.terminalResult?.id, "run-after-relaunch")
    }

    func testUnifiedAssistantBatchPreviewUsesServerStateAndCanonicalTransactionRow() throws {
        let draft = EntryDraftModel(
            type: "expense",
            amount: 45_000,
            walletID: "wallet",
            categoryID: "coffee",
            jarID: nil,
            occurredAt: "2026-10-07T08:00:00+07:00",
            note: "Cà phê sáng",
            includedInReports: true
        )
        let pending = AdvisorEntryBatchItemModel(
            id: "pending", version: 2, status: "review_ready", draft: draft,
            proposalID: "proposal", error: nil, sourceLabel: "Hóa đơn"
        )
        let applied = AdvisorEntryBatchItemModel(
            id: "applied", version: 2, status: "applied", draft: draft,
            proposalID: "proposal", error: nil, sourceLabel: nil
        )
        let conflict = AdvisorEntryBatchItemModel(
            id: "conflict", version: 2, status: "conflict", draft: draft,
            proposalID: "proposal", error: "Đã có giao dịch tương tự", sourceLabel: nil
        )

        XCTAssertEqual(AssistantEntryBatchPolicy.initialState(for: pending), .pending)
        XCTAssertEqual(AssistantEntryBatchPolicy.initialState(for: applied), .applied)
        XCTAssertEqual(AssistantEntryBatchPolicy.initialState(for: conflict), .conflict)
        XCTAssertEqual(AssistantEntryBatchPolicy.state(after: .apply), .saving)
        XCTAssertEqual(AssistantEntryBatchPolicy.state(after: .reject), .rejected)
        XCTAssertTrue(AssistantEntryBatchPolicy.canApply(pending))
        XCTAssertFalse(AssistantEntryBatchPolicy.canApply(applied))

        let row = try XCTUnwrap(
            AssistantEntryBatchPresentation.transactionRow(
                for: pending,
                timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!,
                masked: false
            )
        )
        XCTAssertEqual(row.title, "Khoản chi")
        XCTAssertEqual(row.subtitle, "Cà phê sáng")
        XCTAssertEqual(row.amount, "−45.000\u{00a0}₫")
        XCTAssertEqual(row.amountTone, .expense)
    }

    func testUnifiedAssistantUnknownPartRemainsReadOnly() throws {
        let fixture = #"{"type":"future_part","data":{"opaque":{"value":7}}}"#
        let part = try JSONDecoder.mypocket.decode(AdvisorPartModel.self, from: Data(fixture.utf8))

        XCTAssertTrue(AssistantPartPolicy.isReadOnly(part))
        XCTAssertFalse(AssistantPartPolicy.allowsMutation(for: part))
    }

    func testUnifiedAssistantLifecycleAllowsSafeRetryAndOnlyCancelsActiveRuns() {
        XCTAssertTrue(AssistantChatStatePolicy.acceptsSubmit(.idle))
        XCTAssertTrue(AssistantChatStatePolicy.acceptsSubmit(.failed(message: "Lỗi tạm thời")))
        XCTAssertTrue(AssistantChatStatePolicy.acceptsSubmit(.cancelled))
        XCTAssertFalse(AssistantChatStatePolicy.acceptsSubmit(.processing(runID: "run-1")))
        XCTAssertFalse(AssistantChatStatePolicy.acceptsSubmit(.uncertain(requestID: "request-1")))

        XCTAssertTrue(AssistantChatStatePolicy.canCancel(.accepted(runID: "run-1")))
        XCTAssertTrue(AssistantChatStatePolicy.canCancel(.processing(runID: "run-1")))
        XCTAssertFalse(AssistantChatStatePolicy.canCancel(.uncertain(requestID: "request-1")))
        XCTAssertEqual(AssistantChatStatePolicy.displayLabel(.processing(runID: "run-1")), "Trợ lý đang phân tích…")
        XCTAssertEqual(AssistantChatStatePolicy.displayLabel(.uncertain(requestID: "request-1")), "Yêu cầu có thể đã được tiếp nhận")
    }

    private func run(id: String = "run-1", status: String, ownerID: String? = "owner-1") -> AdvisorRunModel {
        AdvisorRunModel(id: id, conversationID: "conversation-1", status: status, errorCode: nil, ownerID: ownerID, serverTime: nil, deadlineAt: nil, retryAfterMs: nil, createdAt: nil)
    }
}
