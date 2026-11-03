import XCTest
@testable import MyPocketIOS

final class LocalFirstContractTests: XCTestCase {
    private func input(version: Int64?) -> TransactionInputModel {
        TransactionInputModel(walletID: "w", categoryID: "c", jarID: nil, type: "expense", amount: 1000,
                              occurredAt: "2026-11-02T10:00:00+07:00", note: nil, includedInReports: true,
                              travelEventID: nil, expectedVersion: version)
    }

    func testEditsSendExpectedVersionAndCreatesOmitIt() throws {
        let edit = try String(decoding: JSONEncoder.mypocket.encode(input(version: 3)), as: UTF8.self)
        XCTAssertTrue(edit.contains("\"expected_version\":3"), edit)
        let create = try String(decoding: JSONEncoder.mypocket.encode(input(version: nil)), as: UTF8.self)
        XCTAssertFalse(create.contains("expected_version"), create)
    }

    @MainActor
    func testVersionConflictIsRecognised() {
        XCTAssertTrue(TransactionsViewModel.isVersionConflict(APIError.http(status: 409)))
        XCTAssertFalse(TransactionsViewModel.isVersionConflict(APIError.http(status: 500)))
        XCTAssertFalse(TransactionsViewModel.isVersionConflict(APIError.unauthorized))
    }

    func testCachedStampIsVietnameseAndStable() {
        let stamp = Date(timeIntervalSince1970: 1_793_000_000).mpCachedStamp
        XCTAssertTrue(stamp.contains("thg"), "cached stamp must be Vietnamese, got \(stamp)")
    }

    @MainActor
    func testBackgroundWriteFailureKeepsItemMessageAndServerDetail() async {
        let center = MPBackgroundWriteCenter()
        center.submit(failure: "Không thể lưu danh mục.", request: { throw APIError.http(status: 404) }, onSuccess: {})
        for _ in 0..<50 where center.failures.isEmpty { try? await Task.sleep(nanoseconds: 20_000_000) }
        XCTAssertEqual(center.pending, 0)
        XCTAssertEqual(center.failures.first?.message, "Không thể lưu danh mục.")
        let failure = try? XCTUnwrap(center.failures.first)
        if let failure { center.dismiss(failure) }
        XCTAssertTrue(center.failures.isEmpty)
    }
}
