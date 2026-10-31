import Foundation
import XCTest
@testable import MyPocketIOS

final class ReportAnalysisTests: XCTestCase {
    func testAnalysisRequestUsesSnapshotAndClientRequestIdentifiers() throws {
        let request = ReportAnalysisRequestModel(snapshotID: "snapshot-1", clientRequestID: "request-1")
        let data = try JSONEncoder().encode(request)
        let object = try XCTUnwrap(JSONSerialization.jsonObject(with: data) as? [String: String])

        XCTAssertEqual(object["snapshot_id"], "snapshot-1")
        XCTAssertEqual(object["client_request_id"], "request-1")
        XCTAssertNil(object["ledger"])
    }

    func testAnalysisPayloadRendersGroundedSectionsAndPreservesStatus() throws {
        let payload = """
        {
          "id": "analysis-1",
          "snapshot_id": "snapshot-1",
          "status": "completed",
          "revision": 7,
          "summary": "Tóm tắt từ báo cáo.",
          "observations": [{"id":"observation-1","body":"Có thay đổi trong dữ liệu."}],
          "suggestions": [{"id":"suggestion-1","body":"Theo dõi kỳ tiếp theo."}]
        }
        """.data(using: .utf8)!

        let analysis = try JSONDecoder().decode(ReportAnalysisModel.self, from: payload)

        XCTAssertEqual(analysis.id, "analysis-1")
        XCTAssertEqual(analysis.status, "completed")
        XCTAssertEqual(analysis.snapshotID, "snapshot-1")
        XCTAssertEqual(analysis.revision, 7)
        XCTAssertTrue(analysis.renderedArticle.contains("Tóm tắt từ báo cáo."))
        XCTAssertTrue(analysis.renderedArticle.contains("Có thay đổi trong dữ liệu."))
        XCTAssertTrue(analysis.renderedArticle.contains("Theo dõi kỳ tiếp theo."))
    }

    func testAnalysisStateKeepsRevisionOnlyOnCompletedSnapshots() {
        let completed = ReportAnalysisState.completed(requestID: "request-1", revision: 7, article: "Báo cáo")
        let queued = ReportAnalysisState.queued(requestID: "request-1")

        XCTAssertEqual(completed, .completed(requestID: "request-1", revision: 7, article: "Báo cáo"))
        XCTAssertNotEqual(completed, queued)
    }
}
