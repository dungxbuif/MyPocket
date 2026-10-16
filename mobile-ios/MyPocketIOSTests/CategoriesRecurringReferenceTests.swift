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
    }

    func testRecurringReferenceContractUsesCompactSharedActionAndCadenceCopy() {
        XCTAssertEqual(RecurringReferenceCopy.addAction, "Thêm")
        XCTAssertEqual(RecurringReferenceCopy.addActionAccessibilityID, "recurring-add")
        XCTAssertEqual(RecurringDraftValidator.positiveInteger("5.000.000"), 5_000_000)
        XCTAssertNil(RecurringDraftValidator.positiveInteger("5.00.000"))
    }
}
