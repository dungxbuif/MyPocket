import XCTest
@testable import MyPocketIOS

final class JarAssignmentCardTests: XCTestCase {
    private let viewJSON = #"""
    {"action_id":"act-1","action":"assign_jar","month":"2026-11","revision":2,"status":"pending","conversation_id":"conv-1","generation":3,
     "assignments":[
       {"transaction_id":"t1","jar_id":"j1","jar_name":"Ăn uống","note":"Cà phê","amount":45000,"occurred_at":"2026-11-02T03:00:00Z","category_name":"Đồ uống","reason":"Giống các khoản cà phê trước","status":"proposed","expected_version":4},
       {"transaction_id":"t2","jar_id":"","jar_name":"","note":"Lương","amount":10,"occurred_at":"2026-11-02T03:00:00Z","status":"rejected","reject_reason":"Chỉ khoản chi thường mới được gắn hũ.","reason":""}
     ]}
    """#

    private func view() throws -> [String: JSONValue] {
        let value = try JSONDecoder().decode(JSONValue.self, from: Data(viewJSON.utf8))
        guard case .object(let object) = value else { throw XCTSkip("not an object") }
        return object
    }

    func testDecodesRowsAndRegistryTitle() throws {
        let rows = try XCTUnwrap(JarAssignmentRow.decode(try view()))
        XCTAssertEqual(rows.count, 2)
        XCTAssertEqual(rows[0].jarName, "Ăn uống")
        XCTAssertEqual(rows[0].amount, 45000)
        XCTAssertFalse(rows[0].rejected)
        XCTAssertTrue(rows[1].rejected)
        XCTAssertEqual(rows[1].rejectReason, "Chỉ khoản chi thường mới được gắn hũ.")
        XCTAssertEqual(AssistantCardType.title(for: "jar_assignment_confirmation"), "Đề xuất gán hũ")
        XCTAssertEqual(AssistantJarAssignmentPresentation.revision(try view()), 2)
    }

    func testPerRowResultCopyEvenWhenNothingApplied() throws {
        let json = #"{"status":"applied","applied_count":0,"total_count":3,"results":[{"transaction_id":"a","status":"applied"},{"transaction_id":"b","status":"unchanged","detail":"Giao dịch đã ở hũ này."},{"transaction_id":"c","status":"conflict"},{"transaction_id":"d","status":"skipped"}]}"#
        let model = try JSONDecoder.mypocket.decode(JarAssignmentConfirmModel.self, from: Data(json.utf8))
        XCTAssertEqual(model.status, "applied")
        XCTAssertEqual(model.summary, "Đã gán 0/3 khoản")
        XCTAssertEqual(model.results.map { $0.badge.text }, ["Đã gán", "Không đổi", "Xung đột – đã đổi ở nơi khác", "Bỏ qua"])
        XCTAssertEqual(model.results[1].detail, "Giao dịch đã ở hũ này.")
    }

    func testPersistedAssignmentResultsDecodeFromView() throws {
        var object = try view()
        object["assignment_results"] = .array([.object(["transaction_id": .string("t1"), "status": .string("conflict"), "detail": .string("x")])])
        let results = AssistantJarAssignmentPresentation.decodeResults(object)
        XCTAssertEqual(results.first?.badge.text, "Xung đột – đã đổi ở nơi khác")
    }

    func testEditPayloadKeepsOnlyChosenRowsAndSnakeCase() throws {
        let rows = try XCTUnwrap(JarAssignmentRow.decode(try view())).filter { !$0.rejected }
        let body = AssistantJarAssignmentPresentation.editBody(rows: rows, jarByTransaction: ["t1": ""], conversationID: "conv-1", generation: 3, revision: 2)
        let encoder = JSONEncoder()
        encoder.keyEncodingStrategy = .convertToSnakeCase
        let object = try XCTUnwrap(JSONSerialization.jsonObject(with: encoder.encode(body)) as? [String: Any])
        XCTAssertEqual(object["conversation_id"] as? String, "conv-1")
        XCTAssertEqual(object["revision"] as? Int, 2)
        let patch = try XCTUnwrap(object["patch"] as? [String: Any])
        let assignments = try XCTUnwrap(patch["assignments"] as? [[String: String]])
        XCTAssertEqual(assignments, [["transaction_id": "t1", "jar_id": ""]])
    }

    func testReplacedActionIDAcceptsTopLevelAndNestedView() throws {
        XCTAssertEqual(AssistantJarAssignmentPresentation.replacedActionID(.object(["action_id": .string("n1")])), "n1")
        XCTAssertEqual(AssistantJarAssignmentPresentation.replacedActionID(.object(["view": .object(["action_id": .string("n2")])])), "n2")
        XCTAssertNil(AssistantJarAssignmentPresentation.replacedActionID(.object([:])))
    }

    func testJarContextEncodingIsAdditive() throws {
        let handoff = ReportAssistantHandoff.jar(month: "2026-11", jarID: "jar-1", jarName: "Ăn uống")
        XCTAssertEqual(handoff.prompt, "Gợi ý hũ cho các khoản chi chưa phân bổ trong tháng này.")
        let encoder = JSONEncoder()
        encoder.keyEncodingStrategy = .convertToSnakeCase
        let request = AdvisorReportMessageInputModel(clientRequestID: "r", text: handoff.prompt, reportContext: handoff.context, jarContext: handoff.jarContext)
        let object = try XCTUnwrap(JSONSerialization.jsonObject(with: encoder.encode(request)) as? [String: Any])
        let context = try XCTUnwrap(object["jar_context"] as? [String: String])
        XCTAssertEqual(context, ["month": "2026-11", "jar_id": "jar-1"])
        XCTAssertNil(object["report_context"])
        let plain = AdvisorReportMessageInputModel(clientRequestID: "r", text: "hi", reportContext: nil)
        let plainObject = try XCTUnwrap(JSONSerialization.jsonObject(with: encoder.encode(plain)) as? [String: Any])
        XCTAssertNil(plainObject["jar_context"])
        let old = try JSONDecoder().decode(AdvisorReportMessageInputModel.self, from: Data(#"{"clientRequestID":"a","text":"b"}"#.utf8))
        XCTAssertNil(old.jarContext)
    }

    func testReplayedConfirmWithOnlyReceiptStillDecodesAndDerivesCounts() throws {
        let json = #"{"status":"applied","receipt":{"action_id":"a","assignment_results":[{"transaction_id":"t1","jar_id":"j","status":"applied","detail":""},{"transaction_id":"t2","jar_id":"j","status":"conflict","detail":"đã đổi"}]}}"#
        let decoded = try JSONDecoder.mypocket.decode(JarAssignmentConfirmModel.self, from: Data(json.utf8))
        XCTAssertEqual(decoded.totalCount, 2)
        XCTAssertEqual(decoded.appliedCount, 1)
        XCTAssertEqual(decoded.results.map(\.status), ["applied", "conflict"])
        XCTAssertEqual(decoded.summary, "Đã gán 1/2 khoản")
    }
}
