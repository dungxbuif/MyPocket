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

    private func run(status: String, ownerID: String? = "owner-1") -> AdvisorRunModel {
        AdvisorRunModel(id: "run-1", conversationID: "conversation-1", status: status, errorCode: nil, ownerID: ownerID, serverTime: nil, deadlineAt: nil, retryAfterMs: nil, createdAt: nil)
    }
}
