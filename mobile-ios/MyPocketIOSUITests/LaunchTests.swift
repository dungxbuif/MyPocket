import XCTest

#if LOCAL_API_UI
private struct FixtureEnvelope<Value: Decodable>: Decodable {
    let data: Value
}

private struct FixtureLogin: Decodable {
    let token: String
}

private struct FixtureWallet: Decodable, Sendable {
    let id: String
    let name: String
}

private struct FixtureWalletHandle: Sendable {
    let token: String
    let wallet: FixtureWallet
}

private enum LocalFixtureError: Error {
    case invalidResponse
    case http(Int, String)
}

private enum LocalAPIFixture {
    private static let baseURL = URL(string: "http://127.0.0.1:8080/api/v1")!

    static func wallet(name: String, type: String) async throws -> FixtureWalletHandle {
        let token = try await loginToken()
        let existing = try await request(
            path: "/wallets",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureWallet]>.self
        ).data
        for item in existing where item.name == name {
            try await deleteWallet(id: item.id, token: token)
        }

        var payload: [String: Any] = [
            "name": name,
            "type": type,
            "opening_balance": 100000,
            "is_in_total": true,
        ]
        if type == "credit" {
            payload["credit_limit"] = 5_000_000
            payload["last_statement_balance"] = 0
            payload["statement_day"] = 15
            payload["payment_due_day"] = 25
        }
        let data = try JSONSerialization.data(withJSONObject: payload)
        let created = try await request(
            path: "/wallets",
            method: "POST",
            token: token,
            body: data,
            decode: FixtureEnvelope<FixtureWallet>.self
        ).data
        return FixtureWalletHandle(token: token, wallet: created)
    }

    static func delete(_ fixture: FixtureWalletHandle) async throws {
        try await deleteWallet(id: fixture.wallet.id, token: fixture.token)
    }

    static func deleteWallets(named name: String) async throws {
        let token = try await loginToken()
        let existing = try await request(
            path: "/wallets",
            method: "GET",
            token: token,
            body: nil,
            decode: FixtureEnvelope<[FixtureWallet]>.self
        ).data
        for item in existing where item.name == name {
            try await deleteWallet(id: item.id, token: token)
        }
    }

    private static func loginToken() async throws -> String {
        let body = try JSONSerialization.data(withJSONObject: [
            "code": "fixture",
            "code_verifier": "fixture",
            "redirect_uri": "mypocket://oauth/callback",
            "platform": "ios",
        ])
        return try await request(
            path: "/auth/mobile/google/exchange",
            method: "POST",
            token: nil,
            body: body,
            decode: FixtureEnvelope<FixtureLogin>.self
        ).data.token
    }

    private static func deleteWallet(id: String, token: String) async throws {
        _ = try await request(
            path: "/wallets/\(id)",
            method: "DELETE",
            token: token,
            body: nil,
            decode: EmptyFixtureResponse.self
        )
    }

    private static func request<Value: Decodable>(
        path: String,
        method: String,
        token: String?,
        body: Data?,
        decode: Value.Type
    ) async throws -> Value {
        var request = URLRequest(url: baseURL.appendingPathComponent(String(path.dropFirst())))
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        if let token { request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
        request.httpBody = body
        let (data, response) = try await URLSession.shared.data(for: request)
        guard let response = response as? HTTPURLResponse else { throw LocalFixtureError.invalidResponse }
        guard (200..<300).contains(response.statusCode) else {
            throw LocalFixtureError.http(response.statusCode, String(data: data, encoding: .utf8) ?? "")
        }
        if response.statusCode == 204 { return try JSONDecoder().decode(Value.self, from: Data("{}".utf8)) }
        return try JSONDecoder().decode(Value.self, from: data)
    }
}

private struct EmptyFixtureResponse: Decodable {}

private extension XCUIElement {
    func clearAndType(_ text: String) {
        tap()
        if let value = value as? String, !value.isEmpty {
            typeText(String(repeating: XCUIKeyboardKey.delete.rawValue, count: value.count))
        }
        typeText(text)
    }
}
#endif

@MainActor
final class LaunchTests: XCTestCase {
    func testAppLaunches() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestUnauthenticated"]
        app.launch()
        XCTAssertTrue(
            app.buttons["Đăng nhập bằng Google"].waitForExistence(timeout: 5),
            "Auth screen should expose the Google sign-in action after launch."
        )
    }

    func testAuthenticatedShellExposesPrimaryRoutes() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        XCTAssertTrue(app.buttons["Tổng quan"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Giao dịch"].exists)
        XCTAssertTrue(app.buttons["Ví"].exists)
        app.buttons["Giao dịch"].tap()
        XCTAssertTrue(app.staticTexts["Giao dịch"].firstMatch.waitForExistence(timeout: 5))
        app.buttons["Ví"].tap()
        XCTAssertTrue(app.staticTexts["Ví"].firstMatch.waitForExistence(timeout: 5))
    }

    func testTransactionOverflowMenuKeepsTransferAndAdjustmentActions() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        XCTAssertTrue(app.buttons["Giao dịch"].waitForExistence(timeout: 5))
        app.buttons["Giao dịch"].tap()
        let menu = app.buttons["Tùy chọn giao dịch"]
        XCTAssertTrue(menu.waitForExistence(timeout: 5))
        menu.tap()
        XCTAssertTrue(app.buttons["Chuyển tiền đến ví khác"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Điều chỉnh số dư"].exists)
    }

    func testAuthenticatedMoreRoutesAndFeedbackEntry() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()
        XCTAssertTrue(app.buttons["Thêm"].waitForExistence(timeout: 5))
        app.buttons["Thêm"].tap()
        XCTAssertTrue(app.buttons["Báo cáo"].waitForExistence(timeout: 5))
        app.buttons["Báo cáo"].tap()
        XCTAssertTrue(app.staticTexts["Báo cáo"].firstMatch.waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Xem báo cáo"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Gửi phản hồi"].waitForExistence(timeout: 5))
        app.buttons["Gửi phản hồi"].firstMatch.tap()
        XCTAssertTrue(app.staticTexts["Phản hồi"].firstMatch.waitForExistence(timeout: 5))
    }

    func testAuthenticatedMoreMenuOpensEveryDestination() throws {
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated"]
        app.launch()

        let destinations = [
            "Ngân sách",
            "Báo cáo",
            "Danh mục",
            "Hũ chi tiêu",
            "Tháng",
            "Giao dịch định kỳ",
            "Travel Mode",
            "Danh mục đầu tư",
            "Sổ tín dụng",
            "Finance Assistant",
            "Nhập giao dịch bằng AI",
            "Phản hồi & changelog",
            "Tài khoản"
        ]

        for (index, title) in destinations.enumerated() {
            if index > 0 {
                app.terminate()
                app.launchArguments = ["-UITestAuthenticated"]
                app.launch()
            }
            XCTAssertTrue(app.buttons["Thêm"].waitForExistence(timeout: 5))
            app.buttons["Thêm"].tap()

            let entry = app.buttons[title].firstMatch
            XCTAssertTrue(entry.waitForExistence(timeout: 5), "Missing More destination: \(title)")
            for _ in 0..<5 where !entry.isHittable {
                app.swipeUp()
            }
            XCTAssertTrue(entry.isHittable, "More destination is not reachable: \(title)")
            entry.tap()
            XCTAssertTrue(app.staticTexts[title].firstMatch.waitForExistence(timeout: 5), "Destination did not render: \(title)")

            let back = app.buttons["Quay lại"].firstMatch
            guard back.waitForExistence(timeout: 5) else {
                XCTFail("Missing back action for: \(title)")
                return
            }
            back.tap()
            XCTAssertTrue(app.buttons["Thêm"].waitForExistence(timeout: 5), "Could not return from: \(title)")
        }
    }

    func testAPIBackedAssistantLoadsOverview() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Thêm"].waitForExistence(timeout: 10))
        app.buttons["Thêm"].tap()
        XCTAssertTrue(app.buttons["Finance Assistant"].waitForExistence(timeout: 10))
        app.buttons["Finance Assistant"].tap()
        XCTAssertTrue(app.staticTexts["Finance Assistant"].firstMatch.waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts["Finance Assistant chưa được bật."].waitForExistence(timeout: 10))
        XCTAssertFalse(app.staticTexts["Không thể tải tổng quan trợ lý tài chính."].exists)
#endif
    }

    func testAPIBackedAIEntryDisablesAttachmentsWhenFilesUnavailable() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Thêm"].waitForExistence(timeout: 10))
        app.buttons["Thêm"].tap()
        XCTAssertTrue(app.buttons["Nhập giao dịch bằng AI"].waitForExistence(timeout: 10))
        app.buttons["Nhập giao dịch bằng AI"].tap()
        XCTAssertTrue(app.staticTexts["Nhập giao dịch bằng AI"].firstMatch.waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts["OCR chưa được cấu hình."].waitForExistence(timeout: 10))
        let attachmentPicker = app.buttons["Đính kèm ảnh hoặc PDF"]
        XCTAssertTrue(attachmentPicker.waitForExistence(timeout: 10))
        XCTAssertFalse(attachmentPicker.isEnabled)
#endif
    }

    func testAIEntryProposalReviewRequiresExplicitType() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI to run the deterministic review fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticated", "-UITestAIEntryProposal"]
        app.launch()
        XCTAssertTrue(app.buttons["Thêm"].waitForExistence(timeout: 10))
        app.buttons["Thêm"].tap()
        XCTAssertTrue(app.buttons["Nhập giao dịch bằng AI"].waitForExistence(timeout: 10))
        app.buttons["Nhập giao dịch bằng AI"].tap()
        XCTAssertTrue(app.staticTexts["Chưa xác định — chọn loại"].waitForExistence(timeout: 10))
        app.buttons["Sửa"].tap()
        XCTAssertTrue(app.buttons["Loại giao dịch: Chưa xác định — chọn loại"].waitForExistence(timeout: 10))
        app.buttons["Lưu đề xuất"].tap()
        XCTAssertTrue(app.staticTexts["Chọn loại giao dịch là thu nhập hoặc chi tiêu trước khi lưu."].waitForExistence(timeout: 5))
        app.buttons["Loại giao dịch: Chưa xác định — chọn loại"].tap()
        XCTAssertTrue(app.buttons["Chi tiêu"].waitForExistence(timeout: 5))
        app.buttons["Chi tiêu"].tap()
        XCTAssertTrue(app.buttons["Hũ chi tiêu: Không chọn"].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons["Tính vào báo cáo: Có"].waitForExistence(timeout: 5))
#endif
    }

    func testAPIBackedLedgerUsesWeekModeAndCashflowCard() throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Giao dịch"].waitForExistence(timeout: 10))
        app.buttons["Giao dịch"].tap()
        XCTAssertTrue(app.staticTexts["Giao dịch"].firstMatch.waitForExistence(timeout: 10))
        XCTAssertTrue(app.buttons["Theo tuần"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.buttons["Tùy chọn"].exists)
        XCTAssertTrue(app.buttons["Tuần trước"].exists)
        XCTAssertTrue(app.staticTexts["Tiền vào"].exists)
        XCTAssertTrue(app.staticTexts["Tiền ra"].exists)
        XCTAssertFalse(app.buttons["Tất cả"].exists)
#endif
    }

    func testAPIBackedWalletDetailExposesTransactionEditor() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable fixture wallet.")
#else
        let fixture = try await LocalAPIFixture.wallet(name: "Fixture Wallet", type: "basic")
        addTeardownBlock { try? await LocalAPIFixture.delete(fixture) }
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Ví"].waitForExistence(timeout: 10))
        app.buttons["Ví"].tap()
        XCTAssertTrue(app.staticTexts["Fixture Wallet"].waitForExistence(timeout: 10))
        app.staticTexts["Fixture Wallet"].tap()
        XCTAssertTrue(app.buttons["Thêm giao dịch"].waitForExistence(timeout: 10))
        app.buttons["Thêm giao dịch"].tap()
        XCTAssertTrue(app.staticTexts["Thêm giao dịch"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts["Danh mục"].exists)
        XCTAssertTrue(app.staticTexts["Ghi chú"].exists)
#endif
    }

    func testAPIBackedWalletDetailCreatesAndUpdatesTransaction() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable fixture wallet.")
#else
        let fixture = try await LocalAPIFixture.wallet(name: "Fixture Mutation Wallet", type: "basic")
        addTeardownBlock { try? await LocalAPIFixture.delete(fixture) }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Ví"].waitForExistence(timeout: 10))
        app.buttons["Ví"].tap()
        XCTAssertTrue(app.staticTexts["Fixture Mutation Wallet"].waitForExistence(timeout: 10))
        app.staticTexts["Fixture Mutation Wallet"].tap()

        XCTAssertTrue(app.buttons["Thêm giao dịch"].waitForExistence(timeout: 10))
        app.buttons["Thêm giao dịch"].tap()
        XCTAssertTrue(app.staticTexts["Thêm giao dịch"].waitForExistence(timeout: 10))

        let amount = app.textFields["Số tiền"]
        XCTAssertTrue(amount.waitForExistence(timeout: 10))
        amount.clearAndType("25000")
        app.buttons["Lưu giao dịch"].tap()

        let createdRow = app.staticTexts.matching(NSPredicate(format: "label CONTAINS[c] %@", "Khoản chi")).firstMatch
        XCTAssertTrue(createdRow.waitForExistence(timeout: 15), "Created transaction should be visible in wallet detail.")
        createdRow.tap()
        XCTAssertTrue(app.staticTexts["Sửa giao dịch"].waitForExistence(timeout: 10))

        let editableNote = app.textFields["Ghi chú"]
        XCTAssertTrue(editableNote.waitForExistence(timeout: 10))
        for _ in 0..<4 where !editableNote.isHittable { app.swipeUp() }
        XCTAssertTrue(editableNote.isHittable)
        editableNote.clearAndType("UI Fixture Expense Updated")
        app.buttons["Lưu thay đổi"].tap()

        let updatedRow = app.staticTexts.matching(NSPredicate(format: "label CONTAINS[c] %@", "UI Fixture Expense Updated")).firstMatch
        XCTAssertTrue(updatedRow.waitForExistence(timeout: 15), "Updated transaction should be visible after saving.")
#endif
    }

    func testAPIBackedCreditWalletOpensLedgerActions() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and seed the disposable credit wallet.")
#else
        let fixture = try await LocalAPIFixture.wallet(name: "Fixture Credit", type: "credit")
        addTeardownBlock { try? await LocalAPIFixture.delete(fixture) }
        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Giao dịch"].waitForExistence(timeout: 10))
        app.buttons["Giao dịch"].tap()
        let walletSelector = app.buttons["Ví: Tổng cộng"]
        XCTAssertTrue(walletSelector.waitForExistence(timeout: 10))
        walletSelector.tap()
        XCTAssertTrue(app.buttons["Fixture Credit"].waitForExistence(timeout: 10))
        app.buttons["Fixture Credit"].tap()
        XCTAssertTrue(app.staticTexts["Sổ tín dụng"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.buttons["Thêm giao dịch"].waitForExistence(timeout: 30))
        XCTAssertTrue(app.buttons["Thanh toán"].waitForExistence(timeout: 5))
        app.buttons["Quay lại"].tap()
        XCTAssertTrue(app.buttons["Ví"].waitForExistence(timeout: 10))
        app.buttons["Ví"].tap()
        XCTAssertTrue(app.staticTexts["Fixture Credit"].waitForExistence(timeout: 10))
        app.staticTexts["Fixture Credit"].tap()
        XCTAssertTrue(app.staticTexts["Sổ tín dụng"].waitForExistence(timeout: 10))
        let addEntry = app.buttons["Thêm giao dịch"]
        if !addEntry.waitForExistence(timeout: 30) {
            print(app.debugDescription)
            XCTFail("Credit ledger actions did not render")
            return
        }
        XCTAssertTrue(app.buttons["Thanh toán"].waitForExistence(timeout: 5))
#endif
    }

    func testAPIBackedWalletCreateAndUpdateFromNativeEditor() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        try? await LocalAPIFixture.deleteWallets(named: "Fixture UI Wallet")
        try? await LocalAPIFixture.deleteWallets(named: "Fixture UI Wallet Updated")
        addTeardownBlock {
            try? await LocalAPIFixture.deleteWallets(named: "Fixture UI Wallet")
            try? await LocalAPIFixture.deleteWallets(named: "Fixture UI Wallet Updated")
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Ví"].waitForExistence(timeout: 10))
        app.buttons["Ví"].tap()
        XCTAssertTrue(app.buttons["Thêm ví"].waitForExistence(timeout: 10))
        app.buttons["Thêm ví"].tap()
        XCTAssertTrue(app.staticTexts["Thêm ví"].waitForExistence(timeout: 10))

        let name = app.textFields["Tên ví"]
        XCTAssertTrue(name.waitForExistence(timeout: 10))
        name.clearAndType("Fixture UI Wallet")
        let openingBalance = app.textFields["Số dư ban đầu"]
        XCTAssertTrue(openingBalance.waitForExistence(timeout: 10))
        openingBalance.clearAndType("100000")
        app.buttons["Tạo ví"].tap()
        XCTAssertTrue(app.staticTexts["Fixture UI Wallet"].waitForExistence(timeout: 15))

        let edit = app.buttons["Sửa"].firstMatch
        XCTAssertTrue(edit.waitForExistence(timeout: 10))
        edit.tap()
        XCTAssertTrue(app.staticTexts["Sửa ví"].waitForExistence(timeout: 10))
        let editedName = app.textFields["Tên ví"]
        XCTAssertTrue(editedName.waitForExistence(timeout: 10))
        let currentNameValue = editedName.value as? String ?? "<nil>"
        print("[Wallet UI] edit field value=\(currentNameValue)")
        XCTAssertEqual(currentNameValue, "Fixture UI Wallet")
        editedName.clearAndType("Fixture UI Wallet Updated")
        XCTAssertEqual(editedName.value as? String, "Fixture UI Wallet Updated")
        try await Task.sleep(for: .milliseconds(500))
        let returnKey = app.keyboards.buttons.matching(NSPredicate(format: "label ==[c] 'return' OR label ==[c] 'done'" )).firstMatch
        if returnKey.exists { returnKey.tap() }
        let save = app.buttons["Lưu thay đổi"]
        XCTAssertTrue(save.isHittable)
        save.tap()
        XCTAssertTrue(app.staticTexts["Fixture UI Wallet Updated"].waitForExistence(timeout: 15))
#endif
    }

    func testAPIBackedTransferCreatesAndDeletesAtomicPairFromNativeUI() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let source = try await LocalAPIFixture.wallet(name: "Fixture Transfer Source", type: "basic")
        let destination = try await LocalAPIFixture.wallet(name: "Fixture Transfer Destination", type: "basic")
        addTeardownBlock {
            try? await LocalAPIFixture.delete(source)
            try? await LocalAPIFixture.delete(destination)
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Giao dịch"].waitForExistence(timeout: 10))
        app.buttons["Giao dịch"].tap()
        let menu = app.buttons["Tùy chọn giao dịch"]
        XCTAssertTrue(menu.waitForExistence(timeout: 10))
        menu.tap()
        XCTAssertTrue(app.buttons["Chuyển tiền đến ví khác"].waitForExistence(timeout: 10))
        app.buttons["Chuyển tiền đến ví khác"].tap()
        XCTAssertTrue(app.staticTexts["Chuyển tiền"].waitForExistence(timeout: 10))

        let amount = app.textFields["Số tiền"]
        XCTAssertTrue(amount.waitForExistence(timeout: 10))
        amount.clearAndType("12500")
        let amountDone = app.buttons["Xong"]
        if amountDone.exists { amountDone.tap() }
        let note = app.textFields["Ghi chú"]
        XCTAssertTrue(note.waitForExistence(timeout: 10))
        note.clearAndType("UI Transfer")
        app.buttons["Xác nhận chuyển"].tap()

        let transferRow = app.staticTexts.matching(NSPredicate(format: "label CONTAINS[c] %@", "UI Transfer")).firstMatch
        XCTAssertTrue(transferRow.waitForExistence(timeout: 15), "The atomic transfer pair should be visible after save.")
        transferRow.tap()
        XCTAssertTrue(app.staticTexts["Chi tiết giao dịch"].waitForExistence(timeout: 10))
        XCTAssertTrue(app.buttons["Xóa cả cặp chuyển ví"].waitForExistence(timeout: 10))
        app.buttons["Xóa cả cặp chuyển ví"].tap()
        XCTAssertTrue(app.buttons["Xóa"].waitForExistence(timeout: 5))
        app.buttons["Xóa"].tap()
        XCTAssertFalse(transferRow.waitForExistence(timeout: 5), "Deleting a transfer must remove both rows from the ledger.")
#endif
    }

    func testAPIBackedCreditEntryAndPaymentRefreshStatementFromNativeUI() async throws {
#if !LOCAL_API_UI
        throw XCTSkip("Build with SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI and run against the disposable local API fixture.")
#else
        let source = try await LocalAPIFixture.wallet(name: "Fixture Credit Source", type: "basic")
        let credit = try await LocalAPIFixture.wallet(name: "Fixture Credit Mutation", type: "credit")
        addTeardownBlock {
            try? await LocalAPIFixture.delete(source)
            try? await LocalAPIFixture.delete(credit)
        }

        let app = XCUIApplication()
        app.launchArguments = ["-UITestAuthenticatedAPI"]
        app.launch()
        XCTAssertTrue(app.buttons["Ví"].waitForExistence(timeout: 10))
        app.buttons["Ví"].tap()
        XCTAssertTrue(app.staticTexts["Fixture Credit Mutation"].waitForExistence(timeout: 10))
        app.staticTexts["Fixture Credit Mutation"].tap()
        XCTAssertTrue(app.staticTexts["Sổ tín dụng"].waitForExistence(timeout: 10))

        app.buttons["Thêm giao dịch"].tap()
        XCTAssertTrue(app.staticTexts["Thêm giao dịch tín dụng"].waitForExistence(timeout: 10))
        let entryAmount = app.textFields["Số tiền"]
        XCTAssertTrue(entryAmount.waitForExistence(timeout: 10))
        entryAmount.clearAndType("12000")
        let entryAmountDone = app.buttons["Xong"]
        XCTAssertTrue(entryAmountDone.waitForExistence(timeout: 5))
        entryAmountDone.tap()
        let entryNote = app.textFields["Ghi chú"]
        XCTAssertTrue(entryNote.waitForExistence(timeout: 10))
        entryNote.clearAndType("UI Credit Purchase")
        let entryReturn = app.keyboards.buttons.matching(NSPredicate(format: "label ==[c] 'return' OR label ==[c] 'done'" )).firstMatch
        if entryReturn.exists { entryReturn.tap() }
        app.buttons["Lưu"].tap()
        XCTAssertTrue(app.staticTexts["Mua hàng"].waitForExistence(timeout: 15), "Credit entry should refresh the statement.")

        app.buttons["Thanh toán"].tap()
        XCTAssertTrue(app.staticTexts["Thanh toán thẻ"].waitForExistence(timeout: 10))
        let paymentAmount = app.textFields["Số tiền"]
        XCTAssertTrue(paymentAmount.waitForExistence(timeout: 10))
        paymentAmount.clearAndType("7500")
        let paymentAmountDone = app.buttons["Xong"]
        XCTAssertTrue(paymentAmountDone.waitForExistence(timeout: 5))
        paymentAmountDone.tap()
        let paymentNote = app.textFields["Ghi chú"]
        XCTAssertTrue(paymentNote.waitForExistence(timeout: 10))
        paymentNote.clearAndType("UI Credit Payment")
        let paymentReturn = app.keyboards.buttons.matching(NSPredicate(format: "label ==[c] 'return' OR label ==[c] 'done'" )).firstMatch
        if paymentReturn.exists { paymentReturn.tap() }
        app.buttons["Lưu"].tap()
        XCTAssertTrue(app.staticTexts["Thanh toán"].waitForExistence(timeout: 15), "Credit payment should refresh the statement.")
#endif
    }
}
