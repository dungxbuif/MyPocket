import Foundation
import XCTest
@testable import MyPocketIOS

final class AIModelsTests: XCTestCase {
    func testEntryBatchAndUnknownPartDecodeWithoutLosingPayload() throws {
        let fixture = #"{"id":"message-1","conversation_id":"conversation-1","seq":3,"role":"assistant","created_at":"2026-10-31T00:00:00Z","parts":[{"type":"entry_batch","data":{"task_id":"task-1","batch_id":"batch-1","version":2,"item_ids":["item-1"],"item_count":1,"summary":"One candidate"}},{"type":"future_part","data":{"opaque":{"value":7}}}]}"#
        let message = try JSONDecoder.mypocket.decode(AdvisorMessageModel.self, from: Data(fixture.utf8))
        let batch = try XCTUnwrap(message.parts.first?.entryBatch)
        XCTAssertEqual(batch.taskID, "task-1")
        XCTAssertEqual(batch.batchID, "batch-1")
        XCTAssertEqual(batch.version, 2)
        XCTAssertEqual(batch.itemIDs, ["item-1"])
        XCTAssertEqual(batch.itemCount, 1)
        XCTAssertEqual(batch.summary, "One candidate")
        XCTAssertEqual(message.parts[1].type, "future_part")
        guard case .object(let unknown)? = message.parts[1].data,
              case .object(let opaque)? = unknown["opaque"],
              case .number(let value)? = opaque["value"] else { return XCTFail("unknown payload lost") }
        XCTAssertEqual(value, 7)
    }

    func testCapabilitiesStayAbsentForLegacyResponse() throws {
        let fixture = #"{"enabled":true,"write":true,"parts_version":1,"tools":[]}"#
        let capabilities = try JSONDecoder.mypocket.decode(AdvisorCapabilitiesModel.self, from: Data(fixture.utf8))
        XCTAssertNil(capabilities.entryIntegration)
        XCTAssertNil(capabilities.supportedPartKinds)
        XCTAssertNil(capabilities.actionKinds)
        XCTAssertNil(capabilities.attachmentFormats)
        XCTAssertNil(capabilities.attachmentMaxCount)
        XCTAssertNil(capabilities.attachmentMaxBytes)
    }

    func testAdvertisedCapabilityMatrixAndTaskDecode() throws {
        let capabilityFixture = #"{"enabled":true,"write":true,"parts_version":2,"tools":[],"entry_integration":true,"supported_part_kinds":["text","entry_batch"],"action_kinds":["create_transaction"],"attachment_formats":["application/pdf"],"attachment_max_count":3,"attachment_max_bytes":5242880}"#
        let capabilities = try JSONDecoder.mypocket.decode(AdvisorCapabilitiesModel.self, from: Data(capabilityFixture.utf8))
        XCTAssertEqual(capabilities.entryIntegration, true)
        XCTAssertEqual(capabilities.supportedPartKinds, ["text", "entry_batch"])
        XCTAssertEqual(capabilities.actionKinds, ["create_transaction"])
        XCTAssertEqual(capabilities.attachmentFormats, ["application/pdf"])
        XCTAssertEqual(capabilities.attachmentMaxCount, 3)
        XCTAssertEqual(capabilities.attachmentMaxBytes, 5_242_880)
        XCTAssertEqual(capabilities.partsVersion, 2)

        let taskFixture = #"{"id":"task-1","owner_id":"owner-1","conversation_id":"conversation-1","generation":2,"client_request_id":"request-1","session_id":"session-1","status":"review_ready","stage":"review","batch_id":"batch-1","batch_version":2,"created_at":"2026-10-31T00:00:00Z","updated_at":"2026-10-31T00:00:01Z"}"#
        let task = try JSONDecoder.mypocket.decode(AdvisorEntryTaskModel.self, from: Data(taskFixture.utf8))
        XCTAssertEqual(task.generation, 2)
        XCTAssertEqual(task.status, "review_ready")
        XCTAssertEqual(task.batchVersion, 2)
    }
}
