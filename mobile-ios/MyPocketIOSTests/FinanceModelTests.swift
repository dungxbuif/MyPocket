import Foundation
import XCTest
@testable import MyPocketIOS

@MainActor
final class FinanceModelTests: XCTestCase {
    func testReportsMonthPolicyAcceptsOnlyRealCalendarMonths() {
        XCTAssertTrue(ReportsMonthPolicy.isValid("2026-01"))
        XCTAssertTrue(ReportsMonthPolicy.isValid("2026-12"))
        XCTAssertFalse(ReportsMonthPolicy.isValid("2026-00"))
        XCTAssertFalse(ReportsMonthPolicy.isValid("2026-13"))
        XCTAssertFalse(ReportsMonthPolicy.isValid("2026-99"))
        XCTAssertFalse(ReportsMonthPolicy.isValid("26-01"))
        XCTAssertFalse(ReportsMonthPolicy.isValid("2026/01"))
        XCTAssertFalse(ReportsMonthPolicy.isValid(""))
    }

    func testPortfolioQuantityAcceptsLocalizedDecimalWithoutAcceptingUnicodeNumerals() {
        XCTAssertTrue(PortfolioTradeDraftValidator.isValid(side: "buy", quantity: "1,5", unitPrice: "10000", fee: ""))
        XCTAssertFalse(PortfolioTradeDraftValidator.isValid(side: "buy", quantity: "١.٥", unitPrice: "10000", fee: ""), "Backend quantity contract accepts ASCII decimal digits only.")
        XCTAssertFalse(PortfolioTradeDraftValidator.isValid(side: "buy", quantity: "1,000.5", unitPrice: "10000", fee: ""), "Do not reinterpret an ambiguous pasted quantity.")
    }

    func testRecurringIntervalDoesNotUseCurrencyThousandsGrouping() {
        XCTAssertFalse(RecurringDraftValidator.isValid(
            name: "Monthly coffee", amount: "15000", interval: "1.000",
            type: "expense", frequency: "monthly", nextRunDate: Date(),
            endsAtEnabled: false, endsAtDate: Date()
        ), "A count must not silently reinterpret decimal-looking input as one thousand periods.")
    }
    func testSharedTransactionRowKeepsCategoryIconAndNoteAcrossScreens() {
        var item = transaction(id: "coffee", occurredAt: "2026-10-13T01:00:00Z")
        item.categoryID = "coffee"
        item.note = "  Highlands sáng  "
        let category = CategoryModel(id: "coffee", ownerID: "owner", parentID: nil, kind: "expense", name: "Cà phê", systemKey: nil, iconKey: "cup.and.saucer.fill", isSystem: false, walletIDs: [])
        let row = MPTransactionRow(transaction: item, category: category, timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!, masked: false)
        XCTAssertEqual(row.title, "Cà phê")
        XCTAssertEqual(row.subtitle, "Highlands sáng")
        XCTAssertEqual(row.iconName, "cup.and.saucer.fill")
        XCTAssertEqual(row.amount, MPFormat.signedMoney(-10_000))
        XCTAssertEqual(row.amountTone, .expense)
    }

    func testSharedTransactionRowNeutralTransferAdjustmentAndMaskedAmounts() {
        var item = transaction(id: "transfer", occurredAt: "2026-10-13T01:00:00Z")
        item.transferID = "transfer-pair"
        item.type = "income"
        let transfer = MPTransactionRow(transaction: item, timeZone: .gmt, masked: false)
        XCTAssertEqual(transfer.title, "Nhận chuyển ví")
        XCTAssertEqual(transfer.iconName, "arrow.left.arrow.right")
        XCTAssertEqual(transfer.amountTone, .neutral)
        XCTAssertEqual(transfer.amount, MPFormat.signedMoney(10_000))
        item.transferID = nil
        item.type = "adjustment"
        item.adjustmentDirection = "increase"
        let adjustment = MPTransactionRow(transaction: item, timeZone: .gmt, masked: false)
        XCTAssertEqual(adjustment.title, "Điều chỉnh số dư")
        XCTAssertEqual(adjustment.amountTone, .neutral)
        XCTAssertEqual(adjustment.amount, MPFormat.signedMoney(10_000))
        let masked = MPTransactionRow(transaction: item, timeZone: .gmt, masked: true)
        XCTAssertEqual(masked.amount, MPFormat.maskedValue)
    }

    func testSharedTransactionRowUsesAccountDateAndCreditDirection() {
        var item = transaction(id: "refund", occurredAt: "2026-10-12T18:00:00Z")
        item.creditKind = "refund"
        item.note = "  "
        let row = MPTransactionRow(transaction: item, timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!, masked: false)
        XCTAssertEqual(row.title, MPSelectionLabels.creditKind("refund"))
        XCTAssertEqual(row.subtitle, "13/10/2026")
        XCTAssertEqual(row.amountTone, .income)
        XCTAssertEqual(row.amount, MPFormat.signedMoney(10_000))
    }

    func testSystemCategoryAllowsOnlyWalletScopeMutation() {
        let system = CategoryModel(id: "system", ownerID: nil, parentID: nil, kind: "expense", name: "Ăn uống", systemKey: "food", iconKey: "fork.knife", isSystem: true, walletIDs: [])
        let personal = CategoryModel(id: "personal", ownerID: "owner", parentID: nil, kind: "expense", name: "Cà phê", systemKey: nil, iconKey: "cup.and.saucer", isSystem: false, walletIDs: [])

        XCTAssertTrue(CategoryEditPolicy.metadataIsReadOnly(system))
        XCTAssertFalse(CategoryEditPolicy.canDelete(system))
        XCTAssertFalse(CategoryEditPolicy.metadataIsReadOnly(personal))
        XCTAssertTrue(CategoryEditPolicy.canDelete(personal))
    }

    func testSharedEnumLabelsNeverExposeFinanceWireValues() {
        XCTAssertEqual(MPSelectionLabels.display(for: "basic"), "Thông thường")
        XCTAssertEqual(MPSelectionLabels.display(for: "expense"), "Chi")
        XCTAssertEqual(MPSelectionLabels.feedbackStatus("in_progress"), "Đang xử lý")
        XCTAssertEqual(MPSelectionLabels.display(for: "overdue"), "Đã quá hạn")
        XCTAssertEqual(MPSelectionLabels.creditKind("payment"), "Thanh toán")
    }

    func testMonthOptionsStayBoundedAndIncludeAnchor() {
        let options = MPFormat.monthOptions(around: "2026-10", past: 2, future: 1, timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!)
        XCTAssertEqual(options, ["2026-08", "2026-09", "2026-10", "2026-11"])
    }

    func testRecurringNextRunFormattingUsesAccountTimezone() {
        XCTAssertEqual(
            MPFormat.shortDateTime("2026-10-06T16:30:00Z", timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!),
            "06/10/2026 23:30"
        )
    }

    func testSavingsProgressClampsTargetAndBalanceSafely() {
        XCTAssertEqual(MPFormat.savingsProgress(balance: 300, target: 1_000), MPSavingsProgress(remaining: 700, percentage: 0.3, reached: false))
        XCTAssertEqual(MPFormat.savingsProgress(balance: 1_100, target: 1_000), MPSavingsProgress(remaining: 0, percentage: 1, reached: true))
        XCTAssertEqual(MPFormat.savingsProgress(balance: -10, target: 0), MPSavingsProgress(remaining: 0, percentage: 0, reached: false))
    }

    func testSavingsQueueEditAllowsTransferWithoutCategoriesButGuardsTransactions() {
        let transfer = OutboxPayload.transfer(TransferInputModel(
            sourceWalletID: "source", destinationWalletID: "goal", amount: 12_500,
            occurredAt: "2026-10-15T00:00:00Z", note: "pending"
        ))
        let transaction = OutboxPayload.transaction(TransactionInputModel(
            walletID: "goal", categoryID: "food", jarID: nil, type: "expense", amount: 12_500,
            occurredAt: "2026-10-15T00:00:00Z", note: "pending", includedInReports: true,
            travelEventID: nil
        ))

        XCTAssertTrue(WalletQueueEditPolicy.canEdit(transfer, walletType: "goal", detailsReady: false))
        XCTAssertFalse(WalletQueueEditPolicy.canEdit(transaction, walletType: "goal", detailsReady: false))
        XCTAssertTrue(WalletQueueEditPolicy.canEdit(transaction, walletType: "goal", detailsReady: true))
        XCTAssertFalse(WalletQueueEditPolicy.canEdit(transfer, walletType: "credit", detailsReady: true))
    }

    func testSavingsTargetDayUsesAccountTimezoneAndCalendarDate() {
        let timezone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        let now = ISO8601DateFormatter().date(from: "2026-10-05T16:30:00Z")!
        XCTAssertEqual(MPFormat.calendarDayDifference(to: "2026-10-07", from: now, timeZone: timezone), 2)
        XCTAssertEqual(MPFormat.calendarDayDifference(to: "2026-10-04", from: now, timeZone: timezone), -1)
        XCTAssertNil(MPFormat.calendarDayDifference(to: nil, from: now, timeZone: timezone))
    }

    func testPortfolioAssetDraftTrimsAndRejectsWhitespaceOnlyValues() {
        XCTAssertEqual(PortfolioAssetDraftValidator.normalized(symbol: " VN30F ", name: " Quỹ chỉ số ")?.symbol, "VN30F")
        XCTAssertEqual(PortfolioAssetDraftValidator.normalized(symbol: " VN30F ", name: " Quỹ chỉ số ")?.name, "Quỹ chỉ số")
        XCTAssertNil(PortfolioAssetDraftValidator.normalized(symbol: "  ", name: "Quỹ chỉ số"))
        XCTAssertNil(PortfolioAssetDraftValidator.normalized(symbol: "VN30F", name: " \n"))
    }

    func testFeedbackDraftTrimsAndRestrictsType() {
        let draft = FeedbackDraftValidator.normalized(type: "bug", title: "  Không lưu  ", description: "  Mô tả lỗi  ")
        XCTAssertEqual(draft?.title, "Không lưu")
        XCTAssertEqual(draft?.description, "Mô tả lỗi")
        XCTAssertNil(FeedbackDraftValidator.normalized(type: "other", title: "Lỗi", description: "Mô tả"))
        XCTAssertNil(FeedbackDraftValidator.normalized(type: "bug", title: " \n", description: "Mô tả"))
    }

    func testFeedbackOwnerResolutionSurvivesWireRoundTripIndependentlyOfDevStatus() throws {
        let input = Data(#"{"id":"fb","type":"bug","title":"Lỗi","description":"Mô tả","status":"open","resolved":true,"resolved_at":"2026-10-14T00:00:00Z","created_at":"2026-10-14T00:00:00Z","updated_at":"2026-10-14T00:00:00Z","screenshot_available":false}"#.utf8)
        let item = try JSONDecoder.mypocket.decode(FeedbackModel.self, from: input)
        let encoded = try XCTUnwrap(JSONSerialization.jsonObject(with: JSONEncoder.mypocket.encode(item)) as? [String: Any])
        XCTAssertEqual(encoded["resolved"] as? Bool, true)
        XCTAssertEqual(encoded["resolved_at"] as? String, "2026-10-14T00:00:00Z")
        XCTAssertEqual(item.status, "open")
    }

    func testFeedbackPresentationResolvesFixedChangelogAndStatusCopy() {
        let fixed = FeedbackModel(
            id: "feedback-1",
            type: "bug",
            title: "Không lưu",
            description: "Mô tả",
            status: "fixed",
            fixedAt: "2026-10-07T10:00:00Z",
            changelogID: "release-1",
            createdAt: "2026-10-06T10:00:00Z",
            updatedAt: "2026-10-07T10:00:00Z",
            screenshotAvailable: false
        )
        let release = ChangelogModel(id: "release-1", version: "1.0.0", title: "Sửa lỗi", description: "Đã sửa.", publishedAt: "2026-10-07T11:00:00Z")

        XCTAssertEqual(FeedbackPresentation.statusLabel(for: fixed.status), "Đã xử lý")
        XCTAssertEqual(FeedbackPresentation.changelogVersion(for: fixed, in: [release]), "1.0.0")
        XCTAssertTrue(FeedbackPresentation.shouldShowChangelog(for: fixed, in: [release]))
        XCTAssertEqual(FeedbackPresentation.statusLabel(for: "unknown"), "unknown")
    }

    func testMonthPresentationUsesReadableInclusiveDatesWithoutTimezoneIdentifiers() {
        let timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        XCTAssertEqual(
            MonthsPresentation.reportPeriod(startAt: "2026-09-30T17:00:00Z", nextStartAt: "2026-10-31T17:00:00Z", timeZone: timeZone),
            "01/10/2026 – 31/10/2026"
        )
        XCTAssertEqual(MonthsPresentation.timezoneLabel("Asia/Ho_Chi_Minh"), "theo múi giờ tài khoản")
    }

    func testTransactionDayGroupPresentationMatchesWebHierarchyAndTimezone() {
        let timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        XCTAssertEqual(
            TransactionsPresentation.dayGroup("2026-10-07", timeZone: timeZone),
            TransactionsPresentation.DayGroup(dayNumber: "7", weekday: "Thứ Tư", monthYear: "tháng 10 2026")
        )
        XCTAssertNil(TransactionsPresentation.dayGroup("not-a-date", timeZone: timeZone))
    }

    func testLedgerRowSubtitleUsesNoteWithoutRedundantWalletName() {
        let timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        XCTAssertEqual(
            MPTransactionPresentation.rowSubtitle(
                note: "Cà phê",
                occurredAt: "2026-10-13T01:00:00Z",
                timeZone: timeZone
            ),
            "Cà phê"
        )
        XCTAssertEqual(
            MPTransactionPresentation.rowSubtitle(
                note: nil,
                occurredAt: "2026-10-13T01:00:00Z",
                timeZone: timeZone
            ),
            MPFormat.shortDate("2026-10-13T01:00:00Z", timeZone: timeZone)
        )
    }

    func testTransactionGlyphPresentationFollowsWebCategoryRegistry() {
        XCTAssertEqual(
            MPTransactionPresentation.glyph(systemKey: "fork.knife", type: "expense"),
            MPTransactionPresentation.Glyph(systemName: "fork.knife", tone: "categoryOrange")
        )
        XCTAssertEqual(
            MPTransactionPresentation.glyph(systemKey: "expense_food_coffee", type: "expense"),
            MPTransactionPresentation.Glyph(systemName: "cup.and.saucer.fill", tone: "categoryAmber")
        )
        XCTAssertEqual(
            MPTransactionPresentation.glyph(systemKey: "income_salary", type: "income"),
            MPTransactionPresentation.Glyph(systemName: "banknote.fill", tone: "categoryIncome")
        )
        XCTAssertEqual(
            MPTransactionPresentation.glyph(systemKey: nil, type: "adjustment"),
            MPTransactionPresentation.Glyph(systemName: "arrow.up.and.down", tone: "categorySlate")
        )
        XCTAssertEqual(
            MPTransactionPresentation.glyph(systemKey: "unknown", type: "expense"),
            MPTransactionPresentation.Glyph(systemName: "tag.fill", tone: "categorySlate")
        )
    }

    func testCategoryIconCatalogMatchesWebPickerKeysAndFallback() {
        XCTAssertEqual(MPCategoryIconCatalog.options.count, 10)
        XCTAssertEqual(MPCategoryIconCatalog.option(for: "expense_food").label, "Ăn uống")
        XCTAssertEqual(MPCategoryIconCatalog.option(for: "expense_travel").systemName, "airplane")
        XCTAssertEqual(MPCategoryIconCatalog.option(for: "unknown").id, "tag")
    }

    func testCategoryListFiltersKindWalletAndSearchWhilePreservingTreeContext() {
        let root = category(id: "root-food", name: "Ăn uống", kind: "expense")
        let coffee = category(id: "coffee", name: "Cà phê", kind: "expense", parentID: root.id, walletIDs: ["wallet-a"])
        let bakery = category(id: "bakery", name: "Tiệm bánh", kind: "expense", parentID: root.id, walletIDs: ["wallet-b"])
        let income = category(id: "salary", name: "Lương", kind: "income")
        let items = [income, bakery, root, coffee]

        let filtered = CategoryListPresentation.nodes(
            items: items,
            kind: "expense",
            walletID: "wallet-a",
            search: "CA PHE"
        )

        XCTAssertEqual(filtered.map(\.item.id), ["root-food", "coffee"])
        XCTAssertEqual(filtered.map(\.depth), [0, 1])
        XCTAssertEqual(filtered.map(\.isSelectable), [true, true])
        XCTAssertEqual(filtered.map(\.isHierarchyContext), [true, false])
    }

    func testCategoryListRootSearchShowsEligibleChildrenInStableTreeOrder() {
        let root = category(id: "root-food", name: "Ăn uống", kind: "expense")
        let coffee = category(id: "coffee", name: "Cà phê", kind: "expense", parentID: root.id)
        let bakery = category(id: "bakery", name: "Tiệm bánh", kind: "expense", parentID: root.id)

        let filtered = CategoryListPresentation.nodes(
            items: [bakery, root, coffee],
            kind: "expense",
            walletID: nil,
            search: "ăn uống"
        )

        XCTAssertEqual(filtered.map(\.item.id), ["root-food", "bakery", "coffee"])
        XCTAssertEqual(filtered.map(\.isHierarchyContext), [false, false, false])
    }

    func testCategoryListRetainsInapplicableParentOnlyAsContextForEligibleChild() {
        let root = category(id: "root-food", name: "Ăn uống", kind: "expense", walletIDs: ["wallet-b"])
        let coffee = category(id: "coffee", name: "Cà phê", kind: "expense", parentID: root.id, walletIDs: ["wallet-a"])

        let filtered = CategoryListPresentation.nodes(
            items: [root, coffee],
            kind: "expense",
            walletID: "wallet-a",
            search: ""
        )

        XCTAssertEqual(filtered.map(\.item.id), ["root-food", "coffee"])
        XCTAssertEqual(filtered.map(\.isSelectable), [false, true])
        XCTAssertEqual(filtered.map(\.isHierarchyContext), [true, false])
    }

    func testArchivedCategoryIsHiddenByDefaultButCanBeShownAndRestored() {
        var archived = category(id: "old", name: "Nhóm cũ", kind: "expense")
        archived.isActive = false
        let hidden = CategoryListPresentation.nodes(items: [archived], kind: "expense", walletID: nil, search: "")
        XCTAssertTrue(hidden.isEmpty)

        let shown = CategoryListPresentation.nodes(items: [archived], kind: "expense", walletID: nil, search: "", includeInactive: true)
        XCTAssertEqual(shown.map(\.item.id), ["old"])
        XCTAssertFalse(shown[0].isHierarchyContext)
        XCTAssertFalse(CategoryEditPolicy.canSelect(archived))
        XCTAssertTrue(CategoryEditPolicy.canArchive(archived))
    }

    func testArchivedParentRemainsContextWhenActiveChildIsSelectable() {
        var parent = category(id: "parent", name: "Nhóm cha", kind: "expense")
        parent.isActive = false
        let child = category(id: "child", name: "Nhóm con", kind: "expense", parentID: parent.id)
        let nodes = CategorySelectionPresentation.nodes(items: [parent, child], kind: "expense", search: "")
        XCTAssertEqual(nodes.map(\.item.id), ["parent", "child"])
        XCTAssertEqual(nodes.map(\.isHierarchyContext), [true, false])
    }

    func testTransactionEditRetainsOnlyMatchingArchivedCategory() {
        var archived = category(id: "old", name: "Nhóm cũ", kind: "expense")
        archived.isActive = false
        let active = category(id: "active", name: "Nhóm mới", kind: "expense")
        XCTAssertEqual(TransactionEditorPolicy.requiredCategoryID(type: "expense", walletID: "wallet", walletType: "basic", categories: [archived, active], currentID: "old"), "old")
        XCTAssertEqual(TransactionEditorPolicy.requiredCategoryID(type: "income", walletID: "wallet", walletType: "basic", categories: [archived, active], currentID: "old"), "")
    }

    func testCategoryPickerFiltersByKindAndSelectableIDsWhileRetainingParentContext() {
        let root = category(id: "root-food", name: "Ăn uống", kind: "expense")
        let coffee = category(id: "coffee", name: "Cà phê", kind: "expense", parentID: root.id)
        let bakery = category(id: "bakery", name: "Tiệm bánh", kind: "expense", parentID: root.id)
        let income = category(id: "salary", name: "Lương", kind: "income")

        let selected = CategorySelectionPresentation.nodes(
            items: [income, bakery, root, coffee],
            kind: "expense",
            selectableIDs: [coffee.id],
            search: ""
        )

        XCTAssertEqual(selected.map(\.item.id), [root.id, coffee.id])
        XCTAssertEqual(selected.map(\.isHierarchyContext), [true, false])
        XCTAssertEqual(selected.map(\.depth), [0, 1])
    }

    func testCategoryPickerPartitionsSeparateRootCardsWithoutLosingParentContext() {
        let food = category(id: "food", name: "Ăn uống", kind: "expense")
        let coffee = category(id: "coffee", name: "Cà phê", kind: "expense", parentID: food.id)
        let transport = category(id: "transport", name: "Đi lại", kind: "expense")
        let bus = category(id: "bus", name: "Xe buýt", kind: "expense", parentID: transport.id)

        let groups = CategorySelectionPresentation.groups(
            items: [food, coffee, transport, bus],
            kind: "expense",
            selectableIDs: [coffee.id, bus.id],
            search: ""
        )

        XCTAssertEqual(groups.map { $0.root.item.id }, [food.id, transport.id])
        XCTAssertEqual(groups.map { $0.children.map(\.item.id) }, [[coffee.id], [bus.id]])
        XCTAssertTrue(groups.allSatisfy { $0.root.isHierarchyContext })
        XCTAssertTrue(groups.allSatisfy { $0.children.allSatisfy { !$0.isHierarchyContext } })
    }

    func testCategoryPickerSearchKeepsMatchingChildAndOnlyItsAncestry() {
        let food = category(id: "food", name: "Ăn uống", kind: "expense")
        let coffee = category(id: "coffee", name: "Cà phê", kind: "expense", parentID: food.id)
        let lunch = category(id: "lunch", name: "Cơm trưa", kind: "expense", parentID: food.id)
        let transport = category(id: "transport", name: "Đi lại", kind: "expense")

        let results = CategorySelectionPresentation.nodes(
            items: [transport, lunch, food, coffee],
            kind: "expense",
            selectableIDs: [coffee.id],
            search: "CA PHE"
        )

        XCTAssertEqual(results.map(\.item.id), [food.id, coffee.id])
        XCTAssertEqual(results.map(\.isHierarchyContext), [true, false])
    }

    func testCategoryPickerSubtitlesMatchSelectableSystemAndPersonalCategories() {
        XCTAssertEqual(CategorySelectionPresentation.subtitle(isSystem: true, isSelectable: true), "Nhóm mặc định")
        XCTAssertEqual(CategorySelectionPresentation.subtitle(isSystem: false, isSelectable: true), "Nhóm cá nhân")
        XCTAssertEqual(CategorySelectionPresentation.subtitle(isSystem: false, isSelectable: false), "Không áp dụng cho lựa chọn này")
    }

    private func category(
        id: String,
        name: String,
        kind: String,
        parentID: String? = nil,
        walletIDs: [String] = []
    ) -> CategoryModel {
        CategoryModel(
            id: id,
            ownerID: "owner",
            parentID: parentID,
            kind: kind,
            name: name,
            systemKey: nil,
            iconKey: "tag",
            isSystem: false,
            walletIDs: walletIDs
        )
    }

    func testAmountCalculatorUsesIntegerVNDArithmeticWithoutEval() {
        XCTAssertEqual(MPAmountCalculator.calculate("1000000+500000"), 1_500_000)
        XCTAssertEqual(MPAmountCalculator.calculate("1000000+500000×2"), 2_000_000)
        XCTAssertEqual(MPAmountCalculator.calculate("1000000÷4"), 250_000)
        XCTAssertEqual(MPAmountCalculator.calculate("1000000-500000"), 500_000)
        XCTAssertNil(MPAmountCalculator.calculate("1000000÷0"))
        XCTAssertNil(MPAmountCalculator.calculate("1000000÷3"))
        XCTAssertNil(MPAmountCalculator.calculate("1000000,"))
        XCTAssertNil(MPAmountCalculator.calculate("1+"))
        XCTAssertNil(MPAmountCalculator.calculate("999999999999999999999999"))
        XCTAssertNil(MPAmountCalculator.calculate("-1"))
    }

    func testEndedBudgetUsesReadOnlyPreviewAction() {
        XCTAssertEqual(BudgetPresentation.actionTitle(ended: false), "Sửa")
        XCTAssertEqual(BudgetPresentation.actionTitle(ended: true), "Xem / Xóa")
        XCTAssertTrue(BudgetPresentation.isReadOnly(ended: true))
        XCTAssertFalse(BudgetPresentation.isReadOnly(ended: false))
    }

    func testWalletRowUsesDescriptionFallbackAndDangerToneForNegativeBalance() {
        XCTAssertEqual(MPWalletRowPolicy.subtitle(description: "  Tiền lương  ", type: "basic"), "Tiền lương")
        XCTAssertEqual(MPWalletRowPolicy.subtitle(description: "  ", type: "goal"), "Tiết kiệm")
        XCTAssertEqual(MPWalletRowPolicy.balanceTone(isNegative: true), .danger)
        XCTAssertEqual(MPWalletRowPolicy.balanceTone(isNegative: false), .ink)
    }

    func testWalletManagementGroupsIncludedAndExcludedAndTotalsOnlyIncludedBalances() {
        let included = WalletModel(
            id: "included", name: "Ví chính", type: "basic", currency: "VND",
            openingBalance: 120_000, currentBalance: 150_000, availableCredit: nil,
            isInTotal: true, description: nil, targetAmount: nil, targetDate: nil,
            creditLimit: nil, lastStatementBalance: nil, statementDay: nil, paymentDueDay: nil
        )
        let excluded = WalletModel(
            id: "excluded", name: "Ví riêng", type: "basic", currency: "VND",
            openingBalance: 80_000, currentBalance: 200_000, availableCredit: nil,
            isInTotal: false, description: nil, targetAmount: nil, targetDate: nil,
            creditLimit: nil, lastStatementBalance: nil, statementDay: nil, paymentDueDay: nil
        )

        let sections = MPWalletManagementPolicy.sections(from: [included, excluded])
        XCTAssertEqual(sections.map(\.id), ["included", "excluded"])
        XCTAssertEqual(sections.map(\.wallets), [[included], [excluded]])
        XCTAssertEqual(MPWalletManagementPolicy.totalBalance(from: [included, excluded]), 150_000)
        XCTAssertEqual(MPWalletManagementPolicy.totalBalance(from: [excluded]), 0)
    }

    func testSnakeCaseFinancialPayloadDecodesAcronyms() throws {
        let data = Data(#"{"id":"tx-1","owner_id":"u-1","wallet_id":"w-1","category_id":"c-1","type":"expense","amount":50000,"occurred_at":"2026-10-05T10:00:00Z","included_in_reports":true,"created_at":"2026-10-05T10:00:00Z","updated_at":"2026-10-05T10:00:00Z"}"#.utf8)
        let transaction = try JSONDecoder.mypocket.decode(TransactionModel.self, from: data)
        XCTAssertEqual(transaction.ownerID, "u-1")
        XCTAssertEqual(transaction.walletID, "w-1")
        XCTAssertEqual(transaction.categoryID, "c-1")
        XCTAssertTrue(transaction.includedInReports)
    }

    func testCategoryPayloadDecodesPluralIDKey() throws {
        let data = Data(#"{"id":"category-1","kind":"expense","name":"Ăn uống","icon_key":"fork.knife","is_system":false,"wallet_ids":["wallet-1","wallet-2"]}"#.utf8)
        let category = try JSONDecoder.mypocket.decode(CategoryModel.self, from: data)
        XCTAssertEqual(category.walletIDs, ["wallet-1", "wallet-2"])
    }

    func testTransactionInputEncodesBackendSnakeCase() throws {
        let input = TransactionInputModel(walletID: "w-1", categoryID: "c-1", jarID: nil, type: "expense", amount: 1000, occurredAt: "2026-10-05", note: nil, includedInReports: true, travelEventID: nil)
        let data = try JSONEncoder.mypocket.encode(input)
        let json = String(decoding: data, as: UTF8.self)
        XCTAssertTrue(json.contains("wallet_id"))
        XCTAssertTrue(json.contains("included_in_reports"))
    }

    func testBulkDeleteInputEncodesUniqueIDContractShape() throws {
        let input = BulkDeleteInputModel(transactionIDs: ["tx-1", "tx-2"])
        let payload = try JSONSerialization.jsonObject(with: JSONEncoder.mypocket.encode(input)) as? [String: [String]]
        XCTAssertEqual(payload?["transaction_ids"], ["tx-1", "tx-2"])
    }

    func testBulkDeleteSelectionExcludesImmutableRows() {
        let ordinary = transaction(id: "ordinary", occurredAt: "2026-10-06T00:00:00Z")
        var transfer = ordinary
        transfer.transferID = "transfer-1"
        var adjustment = ordinary
        adjustment.type = "adjustment"
        var credit = ordinary
        credit.creditKind = "purchase"

        XCTAssertTrue(TransactionsViewModel.bulkDeleteSelectable(ordinary))
        XCTAssertFalse(TransactionsViewModel.bulkDeleteSelectable(transfer))
        XCTAssertFalse(TransactionsViewModel.bulkDeleteSelectable(adjustment))
        XCTAssertFalse(TransactionsViewModel.bulkDeleteSelectable(credit))
    }

    func testBulkDeleteCallsBackendExactlyOnceWithUniqueIDs() async {
        let recorder = BulkDeleteRecorder()
        let viewModel = TransactionsViewModel(api: FinanceAPI()) { ids in
            await recorder.record(ids)
        }

        let deleted = await viewModel.deleteMany(["tx-2", "tx-1", "tx-2"])

        XCTAssertTrue(deleted)
        let calls = await recorder.snapshot()
        XCTAssertEqual(calls, [["tx-1", "tx-2"]])
    }

    func testOrdinaryTransactionWalletsExcludeCreditWallets() {
        let wallets = [
            WalletModel(id: "basic", name: "Tiền mặt", type: "basic", currency: "VND", openingBalance: 0, currentBalance: 0, availableCredit: nil, isInTotal: true, description: nil, targetAmount: nil, targetDate: nil, creditLimit: nil, lastStatementBalance: nil, statementDay: nil, paymentDueDay: nil),
            WalletModel(id: "credit", name: "Thẻ", type: "credit", currency: "VND", openingBalance: 0, currentBalance: 0, availableCredit: 10_000_000, isInTotal: true, description: nil, targetAmount: nil, targetDate: nil, creditLimit: 10_000_000, lastStatementBalance: nil, statementDay: 15, paymentDueDay: 5)
        ]
        XCTAssertEqual(TransactionsViewModel.ordinaryTransactionWallets(wallets).map(\.id), ["basic"])
    }

    func testRequiredCategoryDefaultsToFirstApplicableAndPreservesValidSelection() {
        let food = CategoryModel(id: "food", ownerID: nil, parentID: nil, kind: "expense", name: "Ăn uống", systemKey: nil, iconKey: "tag", isSystem: false, walletIDs: [])
        let shopping = CategoryModel(id: "shopping", ownerID: nil, parentID: nil, kind: "expense", name: "Mua sắm", systemKey: nil, iconKey: "tag", isSystem: false, walletIDs: [])
        let income = CategoryModel(id: "income", ownerID: nil, parentID: nil, kind: "income", name: "Thu nhập", systemKey: nil, iconKey: "tag", isSystem: false, walletIDs: [])
        let categories = [food, shopping, income]

        XCTAssertEqual(TransactionEditorPolicy.requiredCategoryID(type: "expense", walletID: "wallet-1", walletType: "basic", categories: categories, currentID: ""), "food")
        XCTAssertEqual(TransactionEditorPolicy.requiredCategoryID(type: "expense", walletID: "wallet-1", walletType: "basic", categories: categories, currentID: "shopping"), "shopping")
        XCTAssertEqual(TransactionEditorPolicy.requiredCategoryID(type: "income", walletID: "wallet-1", walletType: "basic", categories: categories, currentID: "missing"), "income")
    }

    func testPortfolioTradeDateUsesAccountLocalDay() {
        let date = MPFormat.parseCalendarDate("2026-10-06")!
        let timestamp = PortfolioView.accountLocalNoonTimestamp(date, timezone: "Asia/Ho_Chi_Minh")
        XCTAssertEqual(MPFormat.shortDate(timestamp, timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!), "06/10/2026")
    }

    func testPortfolioDetailsFilterAndSortTradesWithLocalizedLabels() {
        let older = PortfolioTradeModel(id: "older", assetID: "asset-1", side: "buy", quantity: "1.5", unitPrice: 100_000, fee: 500, occurredAt: "2026-10-01T03:00:00Z", note: nil)
        let newer = PortfolioTradeModel(id: "newer", assetID: "asset-1", side: "sell", quantity: "0.5", unitPrice: 120_000, fee: 700, occurredAt: "2026-10-05T03:00:00Z", note: "Chốt một phần")
        let unrelated = PortfolioTradeModel(id: "other", assetID: "asset-2", side: "buy", quantity: "2", unitPrice: 90_000, fee: 0, occurredAt: "2026-10-06T03:00:00Z", note: nil)

        XCTAssertEqual(PortfolioView.tradesForAsset([older, unrelated, newer], assetID: "asset-1").map(\.id), ["newer", "older"])
        XCTAssertEqual(PortfolioView.tradeSideLabel("buy"), "Mua")
        XCTAssertEqual(PortfolioView.tradeSideLabel("sell"), "Bán")
        XCTAssertTrue(PortfolioView.tradeSubtitle(newer, timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!).contains("Chốt một phần"))
    }

    func testTransferWalletsExcludeCreditWalletsAndPreserveAccountLocalDay() {
        let wallets = [
            WalletModel(id: "basic", name: "Tiền mặt", type: "basic", currency: "VND", openingBalance: 0, currentBalance: 0, availableCredit: nil, isInTotal: true, description: nil, targetAmount: nil, targetDate: nil, creditLimit: nil, lastStatementBalance: nil, statementDay: nil, paymentDueDay: nil),
            WalletModel(id: "credit", name: "Thẻ", type: "credit", currency: "VND", openingBalance: 0, currentBalance: 0, availableCredit: 0, isInTotal: true, description: nil, targetAmount: nil, targetDate: nil, creditLimit: 1_000_000, lastStatementBalance: 0, statementDay: 15, paymentDueDay: 5),
            WalletModel(id: "goal", name: "Du lịch", type: "goal", currency: "VND", openingBalance: 0, currentBalance: 0, availableCredit: nil, isInTotal: true, description: nil, targetAmount: 1_000_000, targetDate: "2027-01-01", creditLimit: nil, lastStatementBalance: nil, statementDay: nil, paymentDueDay: nil)
        ]
        XCTAssertEqual(TransferEditorView.transferWallets(wallets).map(\.id), ["basic", "goal"])
        let timestamp = TransferEditorView.timestamp(MPFormat.parseCalendarDate("2026-10-06")!, timezone: "Asia/Ho_Chi_Minh")
        XCTAssertEqual(MPFormat.shortDate(timestamp, timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!), "06/10/2026")
    }

    func testWalletDetailMonthLabelUsesCalendarComponents() {
        XCTAssertEqual(MPFormat.monthLabel("2026-10"), "Tháng 10 năm 2026")
        XCTAssertEqual(MPFormat.monthLabel("invalid"), "Không rõ thời gian")
        XCTAssertEqual(MPFormat.monthLabel("2026-13"), "Không rõ thời gian")
    }

    func testTravelEventDraftValidatorRejectsWhitespaceMalformedAndReversedDates() {
        XCTAssertTrue(TravelEventDraftValidator.isValid(name: "Nhật Bản", startsOn: "2026-10-01", endsOn: "2026-10-10"))
        XCTAssertTrue(TravelEventDraftValidator.isValid(name: "Không giới hạn", startsOn: "", endsOn: ""))
        XCTAssertFalse(TravelEventDraftValidator.isValid(name: "   ", startsOn: "2026-10-01", endsOn: "2026-10-10"))
        XCTAssertFalse(TravelEventDraftValidator.isValid(name: "Sai ngày", startsOn: "2026-10-10", endsOn: "2026-10-01"))
        XCTAssertFalse(TravelEventDraftValidator.isValid(name: "Sai định dạng", startsOn: "2026/10/01", endsOn: "2026-10-10"))
    }

    func testAIEntryAttachmentPolicyRequiresConfiguredFilesAndIdleRequest() {
        XCTAssertFalse(AIEntryAttachmentPolicy.canSelect(filesConfigured: nil, isProcessing: false))
        XCTAssertFalse(AIEntryAttachmentPolicy.canSelect(filesConfigured: false, isProcessing: false))
        XCTAssertTrue(AIEntryAttachmentPolicy.canSelect(filesConfigured: true, isProcessing: false))
        XCTAssertFalse(AIEntryAttachmentPolicy.canSelect(filesConfigured: true, isProcessing: true))
    }

    func testAdvisorOverviewDecodesStructuredViewAndSource() throws {
        let data = Data(#"{"view_kind":"metric_group","view":{"income":1200000,"expense":500000},"source":{"tool":"get_finance_summary"},"status":"ok"}"#.utf8)
        let overview = try JSONDecoder.mypocket.decode(AdvisorOverviewModel.self, from: data)
        XCTAssertEqual(overview.viewKind, "metric_group")
        XCTAssertEqual(overview.status, "ok")
        if case .object(let view)? = overview.view {
            XCTAssertEqual(view["income"]?.prettyDescription, "1200000")
        } else {
            XCTFail("overview view should remain structured JSON")
        }
    }

    func testCreditEntryCategoriesRespectKindAndWalletScope() {
        let creditWalletID = "credit-wallet"
        let purchaseCategory = CategoryModel(id: "purchase", ownerID: nil, parentID: nil, kind: "expense", name: "Mua sắm", systemKey: nil, iconKey: "tag", isSystem: false, walletIDs: [creditWalletID])
        let refundCategory = CategoryModel(id: "refund", ownerID: nil, parentID: nil, kind: "income", name: "Hoàn tiền", systemKey: nil, iconKey: "tag", isSystem: false, walletIDs: [creditWalletID])
        XCTAssertTrue(purchaseCategory.applies(to: "expense", walletID: creditWalletID, walletType: "credit"))
        XCTAssertFalse(purchaseCategory.applies(to: "income", walletID: creditWalletID, walletType: "credit"))
        XCTAssertTrue(refundCategory.applies(to: "income", walletID: creditWalletID, walletType: "credit"))
        XCTAssertFalse(refundCategory.applies(to: "income", walletID: "other-wallet", walletType: "credit"))
    }

    func testLocalAuthFixtureExchangeEncodesIOSRequestContract() throws {
        let request = LocalAuthFixtureExchangeRequest(
            code: "fixture",
            codeVerifier: "fixture",
            redirectURI: "mypocket://oauth/callback",
            platform: "ios"
        )
        let payload = try JSONSerialization.jsonObject(with: JSONEncoder.mypocket.encode(request)) as? [String: String]
        XCTAssertEqual(payload?["code"], "fixture")
        XCTAssertEqual(payload?["code_verifier"], "fixture")
        XCTAssertEqual(payload?["redirect_uri"], "mypocket://oauth/callback")
        XCTAssertEqual(payload?["platform"], "ios")
    }

    func testPortfolioSummaryDecodesTradesAndNullableValuation() throws {
        let data = Data(#"{"assets":[{"id":"asset-1","symbol":"VNM","name":"Vinamilk","latest_price":72000,"latest_price_at":"2026-10-05T10:00:00Z"}],"positions":[{"asset_id":"asset-1","symbol":"VNM","name":"Vinamilk","quantity":"1.25","cost_basis":90000,"average_cost":72000,"realized_pnl":1000,"market_value":95000,"unrealized_pnl":5000,"latest_price":76000}],"trades":[{"id":"trade-1","asset_id":"asset-1","side":"buy","quantity":"1.25","unit_price":72000,"fee":1000,"occurred_at":"2026-10-05T10:00:00Z"}],"total_cost_basis":91000,"total_market_value":95000,"total_unrealized_pnl":5000}"#.utf8)
        let summary = try JSONDecoder.mypocket.decode(PortfolioSummaryModel.self, from: data)
        XCTAssertEqual(summary.trades.first?.quantity, "1.25")
        XCTAssertEqual(summary.positions.first?.realizedPnL, 1000)
        XCTAssertEqual(summary.totalUnrealizedPnL, 5000)
    }

    func testGoalNotificationPreferenceAndInboxDecode() throws {
        let walletData = Data(#"{"id":"goal-1","name":"Du lịch","type":"goal","currency":"VND","opening_balance":0,"current_balance":0,"is_in_total":true,"goal_notifications_enabled":true}"#.utf8)
        let wallet = try JSONDecoder.mypocket.decode(WalletModel.self, from: walletData)
        XCTAssertTrue(wallet.goalNotificationsEnabled)

        let notificationData = Data(#"{"id":"notice-1","kind":"budget","title":"Ngân sách","body":"Sắp chạm hạn mức","created_at":"2026-10-06T10:00:00Z"}"#.utf8)
        let notification = try JSONDecoder.mypocket.decode(NotificationModel.self, from: notificationData)
        XCTAssertEqual(notification.id, "notice-1")
        XCTAssertNil(notification.readAt)
    }

    func testPortfolioTradeDraftValidatorMirrorsBackendNumericContract() {
        XCTAssertTrue(PortfolioTradeDraftValidator.isValid(side: "buy", quantity: "1.25000000", unitPrice: "72000", fee: "1000"))
        XCTAssertTrue(PortfolioTradeDraftValidator.isValid(side: "sell", quantity: "0.25", unitPrice: "0", fee: ""))
        XCTAssertEqual(PortfolioTradeDraftValidator.nonNegativeInteger(" 72000 "), 72000)
        XCTAssertFalse(PortfolioTradeDraftValidator.isValid(side: "transfer", quantity: "1", unitPrice: "1", fee: "0"))
        XCTAssertFalse(PortfolioTradeDraftValidator.isValid(side: "buy", quantity: "0", unitPrice: "1", fee: "0"))
        XCTAssertFalse(PortfolioTradeDraftValidator.isValid(side: "buy", quantity: "1.000000001", unitPrice: "1", fee: "0"))
        XCTAssertFalse(PortfolioTradeDraftValidator.isValid(side: "buy", quantity: "1e2", unitPrice: "1", fee: "0"))
        XCTAssertFalse(PortfolioTradeDraftValidator.isValid(side: "buy", quantity: "1", unitPrice: "12.5", fee: "0"))
        XCTAssertFalse(PortfolioTradeDraftValidator.isValid(side: "buy", quantity: "1", unitPrice: "1", fee: "-1"))
    }

    func testAIEntryDraftEncodesAllEditableFields() throws {
        let draft = EntryDraftModel(type: "expense", amount: 50000, walletID: "wallet-1", categoryID: "category-1", jarID: nil, occurredAt: "2026-10-05T10:00:00Z", note: "Cà phê", includedInReports: true)
        let json = String(decoding: try JSONEncoder.mypocket.encode(draft), as: UTF8.self)
        XCTAssertTrue(json.contains("wallet_id"))
        XCTAssertTrue(json.contains("included_in_reports"))
        XCTAssertTrue(json.contains("Cà phê"))
    }

    func testAIEntryReviewKeepsUnknownTypeExplicitAndScopesJarToExpenses() {
        XCTAssertEqual(AIEntryProposalReview.typeLabel("unknown"), "Chưa xác định — chọn loại")
        XCTAssertEqual(AIEntryProposalReview.typeLabel("transfer"), "Chưa xác định — chọn loại")
        XCTAssertNil(AIEntryProposalReview.normalizedType("unknown"))
        XCTAssertEqual(AIEntryProposalReview.normalizedType("income"), "income")
        XCTAssertTrue(AIEntryProposalReview.canUseJar(type: "expense", walletType: "basic"))
        XCTAssertTrue(AIEntryProposalReview.canUseJar(type: "expense", walletType: "goal"))
        XCTAssertFalse(AIEntryProposalReview.canUseJar(type: "income", walletType: "basic"))
        XCTAssertFalse(AIEntryProposalReview.canUseJar(type: "expense", walletType: "credit"))
        let transferOut = CategoryModel(id: "transfer-out", ownerID: nil, parentID: nil, kind: "expense", name: "Chuyển đi", systemKey: "expense_transfer_out", iconKey: "arrow.up", isSystem: true, walletIDs: [])
        XCTAssertFalse(AIEntryProposalReview.canUseJar(type: "expense", walletType: "basic", category: transferOut))
        XCTAssertNil(AIEntryProposalReview.jarIDForSave(type: "income", walletType: "basic", jarID: "jar-1"))
        XCTAssertEqual(AIEntryProposalReview.jarIDForSave(type: "expense", walletType: "basic", jarID: "jar-1"), "jar-1")

        let summary = JarMonthSummaryModel(
            month: "2026-10", timezone: "Asia/Ho_Chi_Minh", actualIncome: 0,
            totalSpent: 0, unassignedSpent: 0, totalAllocated: 0,
            totalAllocationPercent: 0, overIncome: false, overOneHundredPercent: false,
            jars: [JarIdentityModel(jarID: "active", name: "Ăn uống"), JarIdentityModel(jarID: "old", name: "Đã gỡ")],
            items: [
                JarItemModel(jarID: "active", name: "Ăn uống", allocationMode: "none", allocationAmount: nil, allocationPercent: nil, calculatedAllocation: nil, spent: 0, active: true),
                JarItemModel(jarID: "old", name: "Đã gỡ", allocationMode: "none", allocationAmount: nil, allocationPercent: nil, calculatedAllocation: nil, spent: 0, active: false)
            ]
        )
        XCTAssertEqual(AIEntryProposalReview.activeJars(from: summary).map(\.jarID), ["active"])
    }

    func testAIEntryRecoveryTreatsOnlyPossiblyAcceptedRequestsAsUncertain() {
        XCTAssertTrue(AIEntryRequestRecoveryPolicy.needsResultRefresh(after: .transport("offline")))
        XCTAssertTrue(AIEntryRequestRecoveryPolicy.needsResultRefresh(after: .decoding("invalid response")))
        XCTAssertTrue(AIEntryRequestRecoveryPolicy.needsResultRefresh(after: .http(status: 503)))
        XCTAssertTrue(AIEntryRequestRecoveryPolicy.needsResultRefresh(after: .problem(ProblemDetails(
            title: "Unavailable", status: 503, detail: nil, instance: nil, code: nil, requestID: nil
        ))))
        XCTAssertTrue(AIEntryRequestRecoveryPolicy.needsResultRefresh(after: .problem(ProblemDetails(
            title: "Unknown", status: nil, detail: nil, instance: nil, code: nil, requestID: nil
        ))))

        XCTAssertFalse(AIEntryRequestRecoveryPolicy.needsResultRefresh(after: .http(status: 400)))
        XCTAssertFalse(AIEntryRequestRecoveryPolicy.needsResultRefresh(after: .unauthorized))
        XCTAssertFalse(AIEntryRequestRecoveryPolicy.needsResultRefresh(after: .invalidURL))
        XCTAssertFalse(AIEntryRequestRecoveryPolicy.needsResultRefresh(after: .missingSession))
        XCTAssertFalse(AIEntryRequestRecoveryPolicy.needsResultRefresh(after: .accountMismatch), "Never resolve an old owner's request with the new account")
    }

    func testAIEntryRecoveryBlocksResubmissionUntilRequestIsResolved() {
        XCTAssertFalse(AIEntryRequestRecoveryPolicy.canSubmit(isUncertain: true, isProcessing: false))
        XCTAssertFalse(AIEntryRequestRecoveryPolicy.canSubmit(isUncertain: false, isProcessing: true))
        XCTAssertFalse(AIEntryRequestRecoveryPolicy.canSubmit(isUncertain: true, isProcessing: true))
        XCTAssertTrue(AIEntryRequestRecoveryPolicy.canSubmit(isUncertain: false, isProcessing: false))
    }

    func testWalletInputEncodesGoalAndCreditMetadata() throws {
        let input = WalletModelInput(name: "Credit", type: "credit", openingBalance: 0, isInTotal: true, description: nil, targetAmount: nil, targetDate: nil, creditLimit: 3_000_000, lastStatementBalance: 250_000, statementDay: 15, paymentDueDay: 5)
        let json = String(decoding: try JSONEncoder.mypocket.encode(input), as: UTF8.self)
        XCTAssertTrue(json.contains("credit_limit"))
        XCTAssertTrue(json.contains("last_statement_balance"))
        XCTAssertTrue(json.contains("statement_day"))
        XCTAssertTrue(json.contains("payment_due_day"))
    }

    func testCalendarDateRoundTripDoesNotUseDeviceTimezone() {
        let date = MPFormat.parseCalendarDate("2026-12-31")
        XCTAssertEqual(date.map(MPFormat.calendarDateString), "2026-12-31")
    }

    func testGoalTargetDateRoundTripUsesAccountTimezone() {
        let timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        let date = MPFormat.parseCalendarDate("2026-10-06", timeZone: timeZone)

        XCTAssertNotNil(date)
        XCTAssertEqual(date.map { MPFormat.calendarDateString($0, timeZone: timeZone) }, "2026-10-06")
    }

    func testCategoryInputPreservesParentID() throws {
        let input = CategoryModelInput(name: "Highlands", kind: "expense", parentID: "food", walletIDs: [], iconKey: "tag")
        let json = String(decoding: try JSONEncoder.mypocket.encode(input), as: UTF8.self)
        XCTAssertTrue(json.contains("\"parent_id\":\"food\""))
    }

    func testCategoryInputPreservesMultipleWalletScopes() throws {
        let input = CategoryModelInput(name: "Highlands", kind: "expense", parentID: "food", walletIDs: ["wallet-1", "wallet-2"], iconKey: "tag")
        let encoded = try JSONEncoder.mypocket.encode(input)
        let json = String(decoding: encoded, as: UTF8.self)
        XCTAssertTrue(json.contains("\"wallet_ids\":[\"wallet-1\",\"wallet-2\"]"), "Category API must receive wallet_ids, not a synthesized acronym variant: \(json)")
    }

    func testRecurringInputUsesRFC3339InstantAndNote() throws {
        let input = RecurringInputModel(name: "Tiền nhà", walletID: "wallet-1", categoryID: nil, type: "expense", amount: 5_000_000, note: "Hợp đồng", frequency: "monthly", interval: 1, nextRunAt: "2026-10-05T02:30:00Z", endsAt: nil, active: true)
        let json = String(decoding: try JSONEncoder.mypocket.encode(input), as: UTF8.self)
        XCTAssertTrue(json.contains("2026-10-05T02:30:00Z"))
        XCTAssertTrue(json.contains("Hợp đồng"))
    }

    func testRecurringDraftValidatorRequiresPositiveAmountAndInterval() {
        let nextRun = MPFormat.parseInstant("2026-10-06T02:30:00Z")!
        let end = MPFormat.parseInstant("2026-11-06T02:30:00Z")!

        XCTAssertTrue(RecurringDraftValidator.isValid(name: "Tiền nhà", amount: "5.000.000", interval: "1", type: "expense", frequency: "monthly", nextRunDate: nextRun, endsAtEnabled: true, endsAtDate: end))
        XCTAssertEqual(RecurringDraftValidator.positiveInteger("5.000.000"), 5_000_000)
        XCTAssertNil(RecurringDraftValidator.positiveInteger("1a2"))
        XCTAssertNil(RecurringDraftValidator.positiveInteger("1 2"))
        XCTAssertNil(RecurringDraftValidator.positiveInteger("5.00.000"))
        XCTAssertFalse(RecurringDraftValidator.isValid(name: "Tiền nhà", amount: "0", interval: "1", type: "expense", frequency: "monthly", nextRunDate: nextRun, endsAtEnabled: false, endsAtDate: nextRun))
        XCTAssertFalse(RecurringDraftValidator.isValid(name: "Tiền nhà", amount: "5.000", interval: "0", type: "expense", frequency: "monthly", nextRunDate: nextRun, endsAtEnabled: false, endsAtDate: nextRun))
    }

    func testRecurringDraftValidatorRejectsUnsupportedMetadataAndReversedEndDate() {
        let nextRun = MPFormat.parseInstant("2026-10-06T02:30:00Z")!
        let end = MPFormat.parseInstant("2026-10-05T02:30:00Z")!

        XCTAssertFalse(RecurringDraftValidator.isValid(name: " ", amount: "1", interval: "1", type: "expense", frequency: "monthly", nextRunDate: nextRun, endsAtEnabled: false, endsAtDate: nextRun))
        XCTAssertFalse(RecurringDraftValidator.isValid(name: "Lương", amount: "1", interval: "1", type: "transfer", frequency: "monthly", nextRunDate: nextRun, endsAtEnabled: false, endsAtDate: nextRun))
        XCTAssertFalse(RecurringDraftValidator.isValid(name: "Lương", amount: "1", interval: "1", type: "income", frequency: "quarterly", nextRunDate: nextRun, endsAtEnabled: false, endsAtDate: nextRun))
        XCTAssertFalse(RecurringDraftValidator.isValid(name: "Lương", amount: "1", interval: "1", type: "income", frequency: "monthly", nextRunDate: nextRun, endsAtEnabled: true, endsAtDate: end))
    }

    func testCategoryApplicabilityRespectsKindGoalWalletAndScope() {
        let food = CategoryModel(id: "food", ownerID: "u", parentID: nil, kind: "expense", name: "Ăn uống", systemKey: nil, iconKey: "fork.knife", isSystem: false, walletIDs: [])
        XCTAssertTrue(food.applies(to: "expense", walletID: "basic", walletType: "basic"))
        XCTAssertFalse(food.applies(to: "income", walletID: "basic", walletType: "basic"))
        XCTAssertFalse(food.applies(to: "expense", walletID: "goal", walletType: "goal"))

        let scoped = CategoryModel(id: "scoped", ownerID: "u", parentID: nil, kind: "expense", name: "Ví riêng", systemKey: nil, iconKey: "tag", isSystem: false, walletIDs: ["wallet-1"])
        XCTAssertTrue(scoped.applies(to: "expense", walletID: "wallet-1", walletType: "basic"))
        XCTAssertFalse(scoped.applies(to: "expense", walletID: "wallet-2", walletType: "basic"))
    }

    func testInstantFormatSupportsFractionalSecondsAndAccountTimezone() {
        let instant = "2026-10-01T00:30:00.123Z"
        XCTAssertNotNil(MPFormat.parseInstant(instant))
        XCTAssertEqual(MPFormat.accountMonth(of: instant, timeZone: TimeZone(identifier: "Asia/Ho_Chi_Minh")!), "2026-10")
        XCTAssertEqual(MPFormat.accountMonth(of: instant, timeZone: TimeZone(identifier: "America/Los_Angeles")!), "2026-09")
    }

    func testTransactionPeriodFilterExcludesMalformedTimestamps() {
        let valid = transaction(id: "valid", occurredAt: "2026-10-06T03:00:00Z")
        let malformed = transaction(id: "malformed", occurredAt: "not-a-timestamp")
        let day = MPFormat.parseCalendarDate("2026-10-06")!

        let result = TransactionsViewModel.applyFilters(
            [valid, malformed],
            walletSelection: "__all__",
            periodSelection: "custom",
            customStart: day,
            customEnd: day,
            accountTimezone: "Asia/Ho_Chi_Minh"
        )

        XCTAssertEqual(result.map(\.id), ["valid"])
    }

    func testTransactionWeekFilterUsesAccountLocalMondayThroughSundayAndOffset() {
        let now = MPFormat.parseInstant("2026-10-07T12:00:00Z")!
        let inCurrentWeek = transaction(id: "current", occurredAt: "2026-10-05T01:00:00Z")
        let inPreviousWeek = transaction(id: "previous", occurredAt: "2026-09-28T01:00:00Z")
        let nextWeek = transaction(id: "next", occurredAt: "2026-10-12T01:00:00Z")

        let current = TransactionsViewModel.applyFilters(
            [inCurrentWeek, inPreviousWeek, nextWeek],
            walletSelection: "__all__",
            periodSelection: "week",
            customStart: now,
            customEnd: now,
            accountTimezone: "Asia/Ho_Chi_Minh",
            weekOffset: 0,
            now: now
        )
        XCTAssertEqual(current.map(\.id), ["current"])

        let previous = TransactionsViewModel.applyFilters(
            [inCurrentWeek, inPreviousWeek, nextWeek],
            walletSelection: "__all__",
            periodSelection: "week",
            customStart: now,
            customEnd: now,
            accountTimezone: "Asia/Ho_Chi_Minh",
            weekOffset: -1,
            now: now
        )
        XCTAssertEqual(previous.map(\.id), ["previous"])
    }

    func testInitialLedgerWeekFallsBackToNewestWeekContainingSelectedWalletTransactions() {
        let now = MPFormat.parseInstant("2026-10-13T09:00:00+07:00")!
        let transactions = [
            transaction(id: "older", occurredAt: "2026-09-28T01:00:00Z", walletID: "wallet-b"),
            transaction(id: "newest", occurredAt: "2026-10-06T01:00:00Z", walletID: "wallet-a")
        ]
        XCTAssertEqual(
            TransactionsPresentation.newestWeekOffset(
                transactions: transactions,
                walletSelection: "wallet-a",
                accountTimezone: "Asia/Ho_Chi_Minh",
                now: now
            ),
            -1
        )
        XCTAssertEqual(
            TransactionsPresentation.newestWeekOffset(
                transactions: transactions,
                walletSelection: "wallet-b",
                accountTimezone: "Asia/Ho_Chi_Minh",
                now: now
            ),
            -2
        )
        XCTAssertEqual(
            TransactionsPresentation.newestWeekOffset(
                transactions: [],
                walletSelection: "__all__",
                accountTimezone: "Asia/Ho_Chi_Minh",
                now: now
            ),
            0
        )
    }

    func testWalletScopePreferencePersistsPerOwnerAndClearsAggregateSelection() {
        let suite = "MyPocketIOSTests.wallet-scope.\(UUID().uuidString)"
        let defaults = UserDefaults(suiteName: suite)!
        defer { defaults.removePersistentDomain(forName: suite) }

        XCTAssertNil(MPWalletScopePreference.load(ownerID: "owner-a", defaults: defaults))
        MPWalletScopePreference.store("wallet-a", ownerID: "owner-a", defaults: defaults)
        XCTAssertEqual(MPWalletScopePreference.load(ownerID: "owner-a", defaults: defaults), "wallet-a")
        XCTAssertNil(MPWalletScopePreference.load(ownerID: "owner-b", defaults: defaults))
        MPWalletScopePreference.store("__all__", ownerID: "owner-a", defaults: defaults)
        XCTAssertNil(MPWalletScopePreference.load(ownerID: "owner-a", defaults: defaults))
        XCTAssertNil(MPWalletScopePreference.load(ownerID: nil, defaults: defaults))
    }

    func testAdjustmentTimestampUsesAccountLocalCalendarDay() {
        let source = ISO8601DateFormatter().date(from: "2026-10-01T00:30:00Z")!
        let raw = AdjustmentEditorView.timestamp(source, timezone: "America/Los_Angeles")
        let parsed = MPFormat.parseInstant(raw)!
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "America/Los_Angeles")!
        XCTAssertEqual(calendar.component(.hour, from: parsed), 12)
        XCTAssertEqual(calendar.component(.day, from: parsed), 30)
    }

    func testWeekRangeLabelUsesAccountLocalMondayThroughSundayDates() {
        let now = MPFormat.parseInstant("2026-10-07T12:00:00Z")!
        let timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        let range = MPFormat.weekRange(now: now, timeZone: timeZone)

        XCTAssertEqual(MPFormat.calendarDateString(range.start, timeZone: timeZone), "2026-10-05")
        XCTAssertEqual(MPFormat.calendarDateString(range.endExclusive, timeZone: timeZone), "2026-10-12")
        XCTAssertEqual(MPFormat.weekRangeLabel(start: range.start, endExclusive: range.endExclusive, timeZone: timeZone), "05/10/2026 – 11/10/2026")
    }

    func testOverviewWeekReportTotalsUseAccountWeekAndExcludeNonReportRows() {
        let timeZone = TimeZone(identifier: "Asia/Ho_Chi_Minh")!
        let now = MPFormat.parseInstant("2026-10-11T18:00:00Z")!
        let rows = [
            reportTransaction(id: "expense-start", type: "expense", amount: 120_000, occurredAt: "2026-10-11T17:00:00Z"),
            reportTransaction(id: "income-start", type: "income", amount: 500_000, occurredAt: "2026-10-11T17:00:00Z"),
            reportTransaction(id: "before-week", type: "expense", amount: 1_000_000, occurredAt: "2026-10-11T16:59:59Z"),
            reportTransaction(id: "income-end", type: "income", amount: 200_000, occurredAt: "2026-10-18T16:59:59Z"),
            reportTransaction(id: "after-week", type: "expense", amount: 2_000_000, occurredAt: "2026-10-18T17:00:00Z"),
            reportTransaction(id: "excluded", type: "expense", amount: 900_000, occurredAt: "2026-10-13T03:00:00+07:00", includedInReports: false),
            reportTransaction(id: "transfer", type: "transfer", amount: 700_000, occurredAt: "2026-10-14T10:00:00+07:00", transferID: "transfer-1"),
            reportTransaction(id: "adjustment", type: "adjustment", amount: 600_000, occurredAt: "2026-10-15T10:00:00+07:00"),
        ]

        let totals = MPWeekReportData.totals(transactions: rows, now: now, timeZone: timeZone)

        XCTAssertEqual(totals.income, 700_000)
        XCTAssertEqual(totals.expense, 120_000)
    }

    func testTransactionCustomPeriodRejectsReversedDateRange() {
        let start = MPFormat.parseCalendarDate("2026-10-10")!
        let end = MPFormat.parseCalendarDate("2026-10-01")!
        let row = transaction(id: "row", occurredAt: "2026-10-05T01:00:00Z")

        let result = TransactionsViewModel.applyFilters(
            [row],
            walletSelection: "__all__",
            periodSelection: "custom",
            customStart: start,
            customEnd: end,
            accountTimezone: "Asia/Ho_Chi_Minh"
        )
        XCTAssertTrue(result.isEmpty)
    }

    func testTransactionEditorJarPolicyMatchesBusinessRules() {
        let transferOut = CategoryModel(id: "transfer-out", ownerID: nil, parentID: nil, kind: "expense", name: "Chuyển đi", systemKey: "expense_transfer_out", iconKey: "arrow.up", isSystem: true, walletIDs: [])

        XCTAssertTrue(TransactionEditorPolicy.canUseJar(type: "expense", category: nil, walletType: "basic"))
        XCTAssertTrue(TransactionEditorPolicy.canUseJar(type: "expense", category: nil, walletType: "goal"))
        XCTAssertFalse(TransactionEditorPolicy.canUseJar(type: "income", category: nil, walletType: "basic"))
        XCTAssertFalse(TransactionEditorPolicy.canUseJar(type: "expense", category: transferOut, walletType: "basic"))
        XCTAssertFalse(TransactionEditorPolicy.canUseJar(type: "expense", category: nil, walletType: "credit"))
    }

    func testTransactionEditorJarOptionsKeepInactiveSelectionVisible() {
        let summary = JarMonthSummaryModel(
            month: "2026-10",
            timezone: "Asia/Ho_Chi_Minh",
            actualIncome: 10_000_000,
            totalSpent: 2_000_000,
            unassignedSpent: 0,
            totalAllocated: 10_000_000,
            totalAllocationPercent: 1,
            overIncome: false,
            overOneHundredPercent: false,
            jars: [
                JarIdentityModel(jarID: "active", name: "Ăn uống"),
                JarIdentityModel(jarID: "inactive", name: "Du lịch")
            ],
            items: [
                JarItemModel(jarID: "active", name: "Ăn uống", allocationMode: "amount", allocationAmount: 3_000_000, allocationPercent: nil, calculatedAllocation: 3_000_000, spent: 1_000_000, active: true),
                JarItemModel(jarID: "inactive", name: "Du lịch", allocationMode: "amount", allocationAmount: 7_000_000, allocationPercent: nil, calculatedAllocation: 7_000_000, spent: 0, active: false)
            ]
        )

        XCTAssertEqual(TransactionEditorPolicy.jarOptions(summary: summary, selectedJarID: "", selectedJarName: nil).map(\.id), ["", "active"])
        let inactiveOptions = TransactionEditorPolicy.jarOptions(summary: summary, selectedJarID: "inactive", selectedJarName: nil)
        XCTAssertEqual(inactiveOptions.map(\.id), ["", "active", "inactive"])
        XCTAssertTrue(inactiveOptions.last?.label.contains("đã gỡ") == true)
        XCTAssertTrue(TransactionEditorPolicy.hasActiveJar("active", summary: summary))
        XCTAssertFalse(TransactionEditorPolicy.hasActiveJar("inactive", summary: summary))
        XCTAssertFalse(TransactionEditorPolicy.hasActiveJar("missing", summary: summary))
    }

    func testTransactionDetailPolicyKeepsTransfersAdjustmentsAndCreditRowsSafe() {
        let ordinary = transaction(id: "ordinary", occurredAt: "2026-10-06T00:00:00Z")
        var transfer = ordinary
        transfer.transferID = "transfer-1"
        var adjustment = ordinary
        adjustment.type = "adjustment"
        adjustment.adjustmentDirection = "decrease"
        var credit = ordinary
        credit.creditKind = "purchase"

        XCTAssertTrue(TransactionDetailPolicy.canEditOrdinary(ordinary, isCreditWallet: false))
        XCTAssertTrue(TransactionDetailPolicy.canLinkTravel(ordinary, isCreditWallet: false))
        XCTAssertFalse(TransactionDetailPolicy.canEditOrdinary(transfer, isCreditWallet: false))
        XCTAssertFalse(TransactionDetailPolicy.canLinkTravel(transfer, isCreditWallet: false))
        XCTAssertEqual(TransactionDetailPolicy.actionKind(transfer, isCreditWallet: false), "transfer")
        XCTAssertFalse(TransactionDetailPolicy.canEditOrdinary(adjustment, isCreditWallet: false))
        XCTAssertFalse(TransactionDetailPolicy.canLinkTravel(adjustment, isCreditWallet: false))
        XCTAssertEqual(TransactionDetailPolicy.actionKind(adjustment, isCreditWallet: false), "adjustment")
        XCTAssertFalse(TransactionDetailPolicy.canEditOrdinary(credit, isCreditWallet: true))
        XCTAssertEqual(TransactionDetailPolicy.actionKind(credit, isCreditWallet: true), "credit")

        let viewModel = TransactionsViewModel(api: FinanceAPI())
        XCTAssertEqual(viewModel.signedAmount(for: adjustment), -10_000)
        adjustment.adjustmentDirection = "increase"
        XCTAssertEqual(viewModel.signedAmount(for: adjustment), 10_000)
    }

    func testAssistantHistoryExpiryRecognizesTypedGoneProblemOnly() {
        let expired = APIError.problem(ProblemDetails(
            title: "Gone",
            status: 410,
            detail: "history expired",
            instance: "/api/v1/ai/advisor/messages",
            code: "advisor_history_expired",
            requestID: "request-1"
        ))
        let otherGone = APIError.problem(ProblemDetails(
            title: "Gone",
            status: 410,
            detail: "other resource",
            instance: "/api/v1/other",
            code: "other_gone",
            requestID: "request-2"
        ))
        XCTAssertTrue(AssistantHistoryExpiryPolicy.isExpired(expired))
        XCTAssertTrue(AssistantHistoryExpiryPolicy.isExpired(otherGone), "HTTP 410 remains a safe signal to clear stale local chat")
        XCTAssertTrue(AssistantHistoryExpiryPolicy.isExpired(APIError.http(status: 410)))
        XCTAssertFalse(AssistantHistoryExpiryPolicy.isExpired(APIError.problem(ProblemDetails(
            title: "Bad Request", status: 400, detail: "invalid", instance: "/api/v1/ai/advisor/messages", code: "advisor_invalid", requestID: nil
        ))))
    }

    func testAssistantMarkdownParserBuildsReadableBlocks() {
        let content = """
        ## Tổng quan
        Bạn đã chi **3.250.000 ₫**.

        - Ăn uống
        - Mua sắm

        | Hũ | Tỷ lệ |
        | --- | ---: |
        | NEC | 55% |
        """

        XCTAssertEqual(
            MPMarkdownParser.parse(content),
            [
                .heading(level: 2, text: "Tổng quan"),
                .paragraph("Bạn đã chi **3.250.000 ₫**."),
                .list(ordered: false, items: ["Ăn uống", "Mua sắm"]),
                .table(headers: ["Hũ", "Tỷ lệ"], rows: [["NEC", "55%"]])
            ]
        )
    }

    private func transaction(id: String, occurredAt: String, walletID: String = "wallet") -> TransactionModel {
        TransactionModel(
            id: id,
            ownerID: "owner",
            walletID: walletID,
            categoryID: nil,
            jarID: nil,
            jarName: nil,
            type: "expense",
            amount: 10_000,
            adjustmentDirection: nil,
            creditKind: nil,
            creditPaymentID: nil,
            travelEventID: nil,
            occurredAt: occurredAt,
            note: nil,
            includedInReports: true,
            transferID: nil,
            createdAt: occurredAt,
            updatedAt: occurredAt
        )
    }

    private func reportTransaction(
        id: String,
        type: String,
        amount: Int64,
        occurredAt: String,
        includedInReports: Bool = true,
        transferID: String? = nil
    ) -> TransactionModel {
        TransactionModel(
            id: id,
            ownerID: "owner",
            walletID: "wallet",
            categoryID: nil,
            jarID: nil,
            jarName: nil,
            type: type,
            amount: amount,
            adjustmentDirection: nil,
            creditKind: nil,
            creditPaymentID: nil,
            travelEventID: nil,
            occurredAt: occurredAt,
            note: nil,
            includedInReports: includedInReports,
            transferID: transferID,
            createdAt: occurredAt,
            updatedAt: occurredAt
        )
    }
}

private actor BulkDeleteRecorder {
    private(set) var calls: [[String]] = []

    func record(_ ids: [String]) {
        calls.append(ids)
    }

    func snapshot() -> [[String]] {
        calls
    }
}
