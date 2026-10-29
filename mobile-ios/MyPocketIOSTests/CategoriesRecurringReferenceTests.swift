import XCTest
@testable import MyPocketIOS

@MainActor
final class CategoriesRecurringReferenceTests: XCTestCase {
    func testCategoryReferenceContractKeepsHierarchyAndFiltersStable() {
        XCTAssertEqual(CategoriesReferenceCopy.addAction, "Nhóm mới")
        XCTAssertEqual(CategoriesReferenceCopy.kindFilterAccessibilityID, "category-kind-filter")
        XCTAssertEqual(CategoriesReferenceCopy.searchAccessibilityID, "category-search")
        XCTAssertEqual(CategoriesReferenceCopy.scopeSubtitle(walletIDs: []), "Áp dụng tất cả ví")
        XCTAssertEqual(CategoriesReferenceCopy.scopeSubtitle(walletIDs: ["w1", "w2"]), "Hoạt động trong 2 ví")

        let system = CategoryModel(
            id: "transfer-out", ownerID: nil, parentID: nil, kind: "expense",
            name: "Chuyển đi", systemKey: "expense_transfer_out",
            iconKey: "building.columns.fill", isSystem: true, walletIDs: []
        )
        XCTAssertTrue(CategoryEditPolicy.metadataIsReadOnly(system))
        XCTAssertFalse(CategoryEditPolicy.canDelete(system))
        XCTAssertFalse(CategoryEditPolicy.canArchive(system))
    }

    func testCategoryEditorGeometryUsesSharedSpacingRoles() {
        XCTAssertEqual(CategoriesReferencePresentation.editorSectionSpacing, MPMetrics.sectionSpacing)
        XCTAssertEqual(CategoriesReferencePresentation.editorFieldSpacing, MPMetrics.sectionSpacing)
        XCTAssertEqual(CategoriesReferencePresentation.editorActionSpacing, MPMetrics.sectionSpacing)
    }

    func testRecurringReferenceContractUsesCompactSharedActionAndCadenceCopy() {
        XCTAssertEqual(RecurringReferenceCopy.addAction, "Thêm")
        XCTAssertEqual(RecurringReferenceCopy.addActionAccessibilityID, "recurring-add")
        XCTAssertEqual(RecurringDraftValidator.positiveInteger("5.000.000"), 5_000_000)
        XCTAssertNil(RecurringDraftValidator.positiveInteger("5.00.000"))
        XCTAssertEqual(RecurringReferencePresentation.frequencyLabel("monthly", interval: 1), "Hàng tháng")
        XCTAssertEqual(RecurringReferencePresentation.frequencyLabel("monthly", interval: 2), "Mỗi 2 kỳ · hàng tháng")
    }

    func testEditorDraftSessionKeepsSameItemDraftAcrossCancelAndReopen() {
        var session = SecondaryEditorDraftSession()

        XCTAssertTrue(session.begin(itemID: "schedule-1", scopeID: "recurring"))
        session.cancel()
        XCTAssertFalse(session.isPresented)
        XCTAssertFalse(session.begin(itemID: "schedule-1", scopeID: "recurring"))
        XCTAssertTrue(session.isPresented)
        XCTAssertTrue(session.begin(itemID: "schedule-2", scopeID: "recurring"))

        session.complete()
        XCTAssertNil(session.itemID)
        XCTAssertFalse(session.isPresented)
    }
}
