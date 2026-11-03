import Foundation
import XCTest
@testable import MyPocketIOS

final class ReportAnalysisTests: XCTestCase {
    func testAnalysisRequestUsesSnapshotAndClientRequestIdentifiers() throws {
        let request = ReportAnalysisRequestModel(snapshotID: "snapshot-1", clientRequestID: "request-1")
        let encoder = JSONEncoder()
        encoder.keyEncodingStrategy = .convertToSnakeCase
        let data = try encoder.encode(request)
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

        let decoder = JSONDecoder.mypocket
        let analysis = try decoder.decode(ReportAnalysisModel.self, from: payload)

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

    func testFindingEvidenceDecodesAndSurvivesCacheRoundTrip() throws {
        let wire = Data(#"{"id":"budget-over:shopping","title":"Mua sắm vượt 350.000 đ","body":"Đã chi theo kỳ ngân sách.","metrics":[{"label":"Vượt","amount":350000}],"period_label":"29/10/2026 – 04/11/2026","action_hint":"Xem giao dịch","fact_refs":["budget:shopping"],"drilldown":{"target_id":"budget:shopping","from":"2026-10-28T17:00:00Z","end_exclusive":"2026-11-04T17:00:00Z","wallet_id":"wallet-a","type":"expense"}}"#.utf8)
        let decoder = JSONDecoder.mypocket
        let finding = try decoder.decode(ReportInsightModel.self, from: wire)
        XCTAssertEqual(finding.metrics.first?.amount, 350000)
        XCTAssertEqual(finding.periodLabel, "29/10/2026 – 04/11/2026")
        XCTAssertEqual(finding.factRefs, ["budget:shopping"])
        XCTAssertEqual(finding.drilldown?.targetID, "budget:shopping")
        let restored = try JSONDecoder().decode(ReportInsightModel.self, from: JSONEncoder().encode(finding))
        XCTAssertEqual(restored, finding)
        let historical = try decoder.decode(ReportInsightModel.self, from: Data(#"{"id":"old","body":"Bản cũ"}"#.utf8))
        XCTAssertTrue(historical.metrics.isEmpty)
        XCTAssertNil(historical.periodLabel)
    }

    func testFollowUpUsesSavedSnapshotAndHidesIdentifiersFromPrompt() throws {
        let analysis = ReportAnalysisModel(id: "analysis-secret-id", snapshotID: "saved-snapshot-id", status: "completed")
        let finding = ReportInsightModel(id: "finding-secret-id", title: "Mua sắm vượt 350.000 đ", body: "Chi trong kỳ")
        let handoff = try XCTUnwrap(ReportAssistantHandoff.make(analysis: analysis, finding: finding, month: "2026-11", walletLabel: "Ví tiền mặt"))
        XCTAssertEqual(handoff.context?.snapshotID, "saved-snapshot-id")
        XCTAssertEqual(handoff.context?.findingID, "finding-secret-id")
        XCTAssertFalse(handoff.prompt.contains("350.000"), "Generated draft must not reveal masked amounts")
        for identifier in [analysis.id, "saved-snapshot-id", finding.id] {
            XCTAssertFalse(handoff.prompt.contains(identifier))
            XCTAssertFalse(handoff.label.contains(identifier))
        }
        let encoder = JSONEncoder()
        encoder.keyEncodingStrategy = .convertToSnakeCase
        let request = AdvisorReportMessageInputModel(clientRequestID: "retry-id", text: handoff.prompt, reportContext: handoff.context)
        let object = try XCTUnwrap(JSONSerialization.jsonObject(with: encoder.encode(request)) as? [String: Any])
        let context = try XCTUnwrap(object["report_context"] as? [String: String])
        XCTAssertEqual(context["analysis_id"], "analysis-secret-id")
        XCTAssertEqual(context["snapshot_id"], "saved-snapshot-id")
        XCTAssertEqual(object["text"] as? String, handoff.prompt)
    }
}
