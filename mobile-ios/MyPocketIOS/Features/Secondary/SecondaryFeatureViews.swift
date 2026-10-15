import SwiftUI

enum BudgetPresentation {
    static func actionTitle(ended: Bool) -> String { ended ? "Xem / Xóa" : "Sửa" }
    static func isReadOnly(ended: Bool) -> Bool { ended }
}

struct BudgetsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    let masked: Bool
    @State private var summary: BudgetSummaryModel?
    @State private var loading = false
    @State private var error: String?
    @State private var saving = false
    @State private var name = ""
    @State private var limit = ""
    @State private var startDate = Date()
    @State private var endDate = Date()
    @State private var wallets: [WalletModel] = []
    @State private var categories: [CategoryModel] = []
    @State private var walletScopeID = "__all__"
    @State private var categoryScopeID = "__all__"
    @State private var categoryPickerPresented = false
    @State private var showEditor = false
    @State private var editTarget: BudgetModel?
    @State private var deleteTarget: BudgetModel?
    @State private var accountTimezone = TimeZone.current.identifier

    var body: some View {
        Group {
            if showEditor { editorContent } else { listContent }
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await load() }
        .mpConfirmDialog(isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }), title: "Xóa ngân sách?", message: "Các giao dịch không bị xóa.", confirmTitle: "Xóa") {
            if let target = deleteTarget { Task { do { try await api.deleteBudget(id: target.id); await load() } catch { self.error = "Không thể xóa ngân sách." } } }
            deleteTarget = nil
        }
        .mpFormBottomSheet(
            isPresented: $categoryPickerPresented,
            title: "Chọn danh mục",
            accessibilityIdentifier: "budget-category-picker-sheet",
            onClose: { categoryPickerPresented = false },
            fields: {
                MPCategoryPicker(
                    categories: categories,
                    kind: "expense",
                    selectableIDs: nil,
                    selectedID: categoryScopeID,
                    clearLabel: "Tất cả danh mục",
                    clearID: "__all__",
                    onSelect: { categoryScopeID = $0; categoryPickerPresented = false }
                )
            },
            footer: {
                MPButton("Đóng", style: .secondary) { categoryPickerPresented = false }
            }
        )
    }

    private var listContent: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Ngân sách", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: {
                    MPButton(BudgetReferenceCopy.addAction, style: .secondary, fullWidth: false) { beginCreating() }
                        .accessibilityIdentifier(BudgetReferenceCopy.addActionAccessibilityID)
                })
                if loading && summary == nil { MPLoadingView(label: "Đang tải ngân sách...") }
                else if let error, summary == nil { MPErrorView(title: "Không tải được ngân sách", detail: error) { Task { await load() } } }
                else if let summary {
                    if BudgetReferenceCopy.shouldShowSummary(itemCount: summary.items.count) {
                        MPSurfaceCard(style: .reference) {
                            VStack(alignment: .leading, spacing: 10) {
                                MPText("Tổng quan ngân sách", style: .emphasis)
                                MPBudgetGauge(
                                    value: Double(summary.spent) / Double(max(1, summary.limitAmount)) * 100,
                                    label: "Đã dùng \(MPFormat.money(summary.spent, masked: masked)) / \(MPFormat.money(summary.limitAmount, masked: masked))"
                                )
                                LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible()), GridItem(.flexible())], spacing: 8) {
                                    MPMetric(title: "Ngân sách", value: MPFormat.money(summary.limitAmount, masked: masked))
                                    MPMetric(title: "Đã chi", value: MPFormat.money(summary.spent, masked: masked))
                                    MPMetric(title: "Còn lại", value: MPFormat.money(summary.limitAmount - summary.spent, masked: masked))
                                }
                            }
                        }
                    }
                    if let error { MPStatusMessage(text: error, tone: .danger) }
                    if summary.items.isEmpty { MPEmptyView(title: "Chưa có ngân sách", detail: "Tạo ngân sách theo ví hoặc danh mục.") }
                    ForEach(summary.items) { budget in
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 8) {
                                MPText(budget.name)
                                MPProgressBar(progress: Double(budget.spent) / Double(max(1, budget.limitAmount)), label: "\(MPFormat.money(budget.spent, masked: masked)) / \(MPFormat.money(budget.limitAmount, masked: masked))")
                                MPText("\(budget.startDate) → \(budget.endDate) · \(MPBudgetCopy.timing(ended: budget.ended, startDate: budget.startDate, today: MPFormat.todayISO(timeZone: TimeZone(identifier: accountTimezone) ?? .current), daysRemaining: budget.daysRemaining))", style: .secondary)
                                if !budget.ended {
                                    MPText(MPBudgetCopy.dailySuggestion(limit: budget.limitAmount, spent: budget.spent, daysRemaining: budget.daysRemaining, masked: masked), style: .secondary)
                                }
                                if let threshold = MPBudgetCopy.threshold(spent: budget.spent, limit: budget.limitAmount) {
                                    MPStatusBadge(text: threshold.text, tone: threshold.tone == .danger ? .danger : .muted)
                                }
                                HStack {
                                    MPButton(BudgetPresentation.actionTitle(ended: budget.ended), style: .secondary) { beginEditing(budget) }
                                        .accessibilityIdentifier(budget.ended ? "budget-preview-\(budget.id)" : "budget-edit-\(budget.id)")
                                    MPButton("Xóa", style: .destructive) { deleteTarget = budget }
                                        .accessibilityIdentifier("budget-delete-\(budget.id)")
                                }
                            }
                        }
                    }
                }

            }
            .padding(MPMetrics.pagePadding)
        }
    }

    private var editorContent: some View {
        MPEditorScaffold(title: editTarget.map { $0.ended ? "Xem ngân sách" : "Sửa ngân sách" } ?? "Tạo ngân sách", isBusy: saving, onCancel: { showEditor = false }) {
            if let endedBudget = editTarget, endedBudget.ended {
                MPStatusMessage(text: "Ngân sách đã kết thúc; chỉ có thể xem hoặc xóa.", tone: .info)
                MPText(endedBudget.name, style: .emphasis)
                MPText("Giới hạn: \(MPFormat.money(endedBudget.limitAmount, masked: masked))", style: .secondary)
                MPText("Đã chi: \(MPFormat.money(endedBudget.spent, masked: masked))", style: .secondary)
                MPText("Thời gian: \(endedBudget.startDate) → \(endedBudget.endDate)", style: .secondary)
                MPText("Ví: \(walletName(for: endedBudget))", style: .secondary)
                MPText("Danh mục: \(categoryName(for: endedBudget))", style: .secondary)
            } else {
                MPTextField(title: "Tên", text: $name, placeholder: "Ăn uống tháng này")
                MPAmountField(title: "Giới hạn", text: $limit)
                MPDateField(title: "Ngày bắt đầu", date: $startDate, timeZone: accountTimezone)
                MPDateField(title: "Ngày kết thúc", date: $endDate, timeZone: accountTimezone)
                MPOptionSelector(title: "Phạm vi ví", options: [MPSelectionOption(id: "__all__", label: "Tất cả ví")] + wallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $walletScopeID, emptyLabel: "Tất cả ví")
                MPFormRow(
                    title: "Phạm vi danh mục",
                    iconName: "tag",
                    value: categoryScopeLabel,
                    accessibilityIdentifier: "budget-category-scope-row"
                ) { categoryPickerPresented = true }
            }
            if let error { MPStatusMessage(text: error, tone: .danger) }
        } footer: {
            if let endedBudget = editTarget, endedBudget.ended {
                MPButton("Xóa", style: .destructive) { deleteTarget = endedBudget; showEditor = false }
            } else {
                MPButton("Lưu", isLoading: saving) { Task { await create() } }
            }
        }
    }

    private func load() async {
        guard !loading else { return }
        let firstLoad = summary == nil
        loading = true; defer { loading = false }
        do {
            async let summaryResult = api.budgets()
            async let walletResult = api.wallets()
            async let categoryResult = api.categories()
            async let profile = api.profile()
            summary = try await summaryResult
            wallets = (try await walletResult).filter { $0.type != "credit" }
            categories = (try await categoryResult).filter { $0.kind == "expense" }
            if let profile = try? await profile {
                accountTimezone = profile.timezone
                if firstLoad {
                    let currentMonth = MPFormat.currentMonth(timeZone: TimeZone(identifier: accountTimezone) ?? .current)
                    if let range = MPFormat.monthRange(month: currentMonth, timeZone: TimeZone(identifier: accountTimezone) ?? .current) {
                        let zone = TimeZone(identifier: accountTimezone) ?? .current
                        startDate = MPFormat.parseCalendarDate(range.start, timeZone: zone) ?? Date()
                        endDate = MPFormat.parseCalendarDate(range.end, timeZone: zone) ?? Date()
                    }
                }
            }
            self.error = nil
        } catch { self.error = "Không thể tải ngân sách." }
    }

    private func create() async {
        guard !saving, editTarget?.ended != true else { return }
        guard let parsed = MPFormat.parsePositiveInteger(limit), !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { error = "Nhập tên và giới hạn hợp lệ."; return }
        guard startDate <= endDate else { error = "Khoảng thời gian ngân sách không hợp lệ."; return }
        let zone = TimeZone(identifier: accountTimezone) ?? .current
        let start = MPFormat.calendarDateString(startDate, timeZone: zone)
        let end = MPFormat.calendarDateString(endDate, timeZone: zone)
        saving = true; error = nil; defer { saving = false }
        do {
            let input = BudgetInputModel(name: name.trimmingCharacters(in: .whitespacesAndNewlines), limitAmount: parsed, walletID: walletScopeID == "__all__" ? nil : walletScopeID, categoryID: categoryScopeID == "__all__" ? nil : categoryScopeID, startDate: start, endDate: end)
            if let editTarget { _ = try await api.updateBudget(id: editTarget.id, input: input) } else { _ = try await api.createBudget(input) }
            showEditor = false; self.editTarget = nil; name = ""; limit = ""; walletScopeID = "__all__"; categoryScopeID = "__all__"; await load()
        }
        catch {
#if DEBUG
            print("[Budgets] save failed: \(APIError.diagnosticDescription(error))")
#endif
            self.error = (error as? APIError)?.userMessage ?? "Không thể lưu ngân sách."
        }
    }

    private func beginCreating() {
        if editTarget != nil {
            name = ""; limit = ""; walletScopeID = "__all__"; categoryScopeID = "__all__"
            let zone = TimeZone(identifier: accountTimezone) ?? .current
            if let range = MPFormat.monthRange(month: MPFormat.currentMonth(timeZone: zone), timeZone: zone) {
                startDate = MPFormat.parseCalendarDate(range.start, timeZone: zone) ?? Date()
                endDate = MPFormat.parseCalendarDate(range.end, timeZone: zone) ?? Date()
            }
        }
        editTarget = nil; error = nil; showEditor = true
    }

    private func beginEditing(_ budget: BudgetModel) {
        if editTarget?.id == budget.id {
            editTarget = budget
            showEditor = true
            return
        }
        editTarget = budget
        name = budget.name
        limit = String(budget.limitAmount)
        let zone = TimeZone(identifier: accountTimezone) ?? .current
        startDate = MPFormat.parseCalendarDate(budget.startDate, timeZone: zone) ?? Date()
        endDate = MPFormat.parseCalendarDate(budget.endDate, timeZone: zone) ?? Date()
        walletScopeID = budget.walletID ?? "__all__"
        categoryScopeID = budget.categoryID ?? "__all__"
        showEditor = true
    }

    private func walletName(for budget: BudgetModel) -> String {
        guard let walletID = budget.walletID else { return "Tất cả ví" }
        return wallets.first(where: { $0.id == walletID })?.name ?? "Ví đã xóa"
    }

    private func categoryName(for budget: BudgetModel) -> String {
        guard let categoryID = budget.categoryID else { return "Tất cả danh mục" }
        return categories.first(where: { $0.id == categoryID })?.name ?? "Danh mục đã xóa"
    }

    private var categoryScopeLabel: String {
        guard categoryScopeID != "__all__" else { return "Tất cả danh mục" }
        return categories.first(where: { $0.id == categoryScopeID })?.name ?? "Chọn danh mục"
    }
}

/// Stable composition contract for the Stitch reports screen.
enum ReportsReferenceSection: String, CaseIterable {
    case header
    case filter
    case summary
    case trend
    case breakdown
    case activity
    case insider

    static let stitchOrder: [ReportsReferenceSection] = [
        .header,
        .filter,
        .summary,
        .trend,
        .breakdown,
        .activity,
        .insider
    ]
}

enum ReportsReferenceCopy {
    static func summaryAction(month: String) -> String { month }
}

enum ReportsOfflinePolicy {
    static func accepts(_ snapshot: FinancialSnapshot?, month: String) -> Bool {
        guard let snapshot,
              snapshot.monthSummary?.month == month,
              snapshot.insider?.month == month else { return false }
        return true
    }
}

enum ReportsMonthPolicy {
    static func isValid(_ value: String) -> Bool {
        let parts = value.split(separator: "-", omittingEmptySubsequences: false)
        guard parts.count == 2,
              parts[0].count == 4,
              parts[1].count == 2,
              parts[0].allSatisfy(\.isNumber),
              parts[1].allSatisfy(\.isNumber),
              let year = Int(parts[0]),
              let month = Int(parts[1]) else { return false }
        return year > 0 && (1...12).contains(month)
    }
}

struct ReportsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    let masked: Bool
    let ownerID: String?
    let snapshotCache: FinancialSnapshotCache?
    @State private var month = MPFormat.currentMonth
    @State private var summary: InsiderSummaryModel?
    @State private var monthDetail: MonthSummaryModel?
    @State private var loading = false
    @State private var error: String?
    @State private var accountTimezone = TimeZone.current.identifier
    @State private var transactions: [TransactionModel] = []
    @State private var cachedAt: Date?

    init(
        api: FinanceAPI,
        onBack: @escaping () -> Void,
        masked: Bool,
        ownerID: String? = nil,
        snapshotCache: FinancialSnapshotCache? = nil
    ) {
        self.api = api
        self.onBack = onBack
        self.masked = masked
        self.ownerID = ownerID
        self.snapshotCache = snapshotCache
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Báo cáo", leading: {
                    MPButton(style: .plain, fullWidth: false, action: onBack) {
                        Label("Quay lại", systemImage: "chevron.left")
                    }
                    .accessibilityIdentifier("screen-back")
                }, trailing: { EmptyView() })
                MPSurfaceCard(style: .reference) {
                    VStack(alignment: .leading, spacing: 10) {
                        MPTextField(title: "Tháng (YYYY-MM)", text: $month, accessibilityIdentifier: "reports-month")
                        MPButton("Xem báo cáo", isLoading: loading) { Task { await load() } }
                            .accessibilityIdentifier("reports-load-button")
                    }
                }
                if let summary {
                    if let cachedAt {
                        MPStatusMessage(text: "Bản lưu \(cachedAt.formatted(date: .abbreviated, time: .shortened)) · Chưa cập nhật từ máy chủ.", tone: .info)
                    }
                    MPSurfaceCard(style: .reference) {
                        VStack(alignment: .leading, spacing: 12) {
                            MPSectionHeader(title: "Tổng quan báo cáo", actionTitle: ReportsReferenceCopy.summaryAction(month: month))
                            MPReportComparisonChart(income: summary.income, expense: summary.expense)
                            HStack(spacing: 12) {
                                reportMetric(label: "Chi tiêu", value: summary.expense, color: MPColor.danger)
                                reportMetric(label: "Thu nhập", value: summary.income, color: MPColor.action)
                                reportMetric(label: "Chênh lệch", value: summary.net, color: MPColor.ink)
                            }
                        }
                    }
                    .accessibilityElement(children: .contain)
                    .accessibilityIdentifier("reports-section-summary")
                    MPAssistantCard(title: "Xu hướng chi tiêu") {
                        MPTrendChart(
                            points: MPTrendData.points(
                                transactions: transactions,
                                month: month,
                                timeZone: TimeZone(identifier: accountTimezone) ?? .current
                            ),
                            masked: masked
                        )
                    }
                    .accessibilityElement(children: .contain)
                    .accessibilityIdentifier("reports-section-trend")
                    MPAssistantCard(title: "Theo nhóm") {
                        if let categories = monthDetail?.categories, !categories.isEmpty {
                            ForEach(categories, id: \.stableID) { item in
                                HStack {
                                    VStack(alignment: .leading, spacing: 3) {
                                        MPText(item.name)
                                        MPText("\(item.type == "income" ? "Thu" : "Chi") · \(item.count) giao dịch", style: .secondary)
                                    }
                                    Spacer()
                                    MPText(MPFormat.money(item.amount, masked: masked), style: .secondary)
                                }
                            }
                        } else {
                            MPText("Chưa có giao dịch được tính vào báo cáo.", style: .secondary)
                        }
                    }
                    .accessibilityElement(children: .contain)
                    .accessibilityIdentifier("reports-section-breakdown")
                    MPAssistantCard(title: "Hoạt động trong tháng") {
                        MPText("\(monthDetail?.transactionCount ?? 0) giao dịch được tính vào báo cáo. Chuyển ví và điều chỉnh số dư không làm tăng thu/chi.", style: .secondary)
                        MPText("Số liệu được tính lại từ sổ giao dịch theo múi giờ tài khoản.", style: .secondary)
                    }
                    .accessibilityElement(children: .contain)
                    .accessibilityIdentifier("reports-section-activity")
                    MPReportsInsiderCard(value: summary, masked: masked)
                } else if let error { MPErrorView(title: "Không tải được báo cáo", detail: error) { Task { await load() } } }
                else { MPEmptyView(title: "Chọn tháng để xem báo cáo", detail: "Số liệu luôn được tính theo timezone tài khoản.") }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await load() }
    }

    private func reportMetric(label: String, value: Int64, color: Color) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            MPText(label, style: .secondary)
            MPText(MPFormat.money(value, masked: masked), style: .emphasis)
                .foregroundStyle(color)
                .lineLimit(1)
                .minimumScaleFactor(0.75)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private func load() async {
        guard !loading else { return }
        loading = true; defer { loading = false }
        error = nil
        cachedAt = nil
        do {
            if let profile = try? await api.profile() {
                accountTimezone = profile.timezone
                if summary == nil { month = MPFormat.currentMonth(timeZone: TimeZone(identifier: accountTimezone) ?? .current) }
            }
            month = month.trimmingCharacters(in: .whitespacesAndNewlines)
            guard ReportsMonthPolicy.isValid(month) else {
                summary = nil
                monthDetail = nil
                transactions = []
                error = "Nhập tháng theo định dạng YYYY-MM (tháng từ 01 đến 12)."
                return
            }
            // Never leave a previous month's totals visible under a new scope
            // while the three report reads are in flight or when one fails.
            summary = nil
            monthDetail = nil
            transactions = []
            async let insiderResult = api.insider(month: month)
            async let monthResult = api.monthSummary(month: month)
            async let transactionResult = api.transactions()
            summary = try await insiderResult
            monthDetail = try? await monthResult
            // Trend data is part of the same verified read scope. If the
            // ledger read fails, do not persist a report with an empty/fake
            // trend; recover the exact cached month instead.
            transactions = try await transactionResult
            await persistSnapshot()
        } catch {
            await recoverFromSnapshot(error)
        }
    }

    private func recoverFromSnapshot(_ failure: Error) async {
        guard FinancialSnapshotCache.allowsFallback(failure), let ownerID, let snapshotCache else {
            summary = nil; monthDetail = nil; transactions = []; self.error = "Không thể tải báo cáo."
            return
        }
        do {
            guard let snapshot = try await snapshotCache.load(ownerID: ownerID),
                  ReportsOfflinePolicy.accepts(snapshot, month: month),
                  let cachedSummary = snapshot.insider,
                  let cachedMonth = snapshot.monthSummary else {
                throw OutboxStoreError.storageUnavailable
            }
            summary = cachedSummary
            monthDetail = cachedMonth
            transactions = snapshot.transactions
            accountTimezone = snapshot.timezone
            cachedAt = snapshot.fetchedAt
            error = nil
        } catch {
            summary = nil; monthDetail = nil; transactions = []; self.error = "Không thể tải báo cáo."
        }
    }

    private func persistSnapshot() async {
        guard let ownerID, let snapshotCache, let summary, let monthDetail else { return }
        do {
            let existing = try await snapshotCache.load(ownerID: ownerID)
            let base = existing ?? FinancialSnapshot(ownerID: ownerID, fetchedAt: Date(), transactions: transactions, wallets: [], categories: [], timezone: accountTimezone)
            try await snapshotCache.save(FinancialSnapshot(
                ownerID: ownerID,
                fetchedAt: Date(),
                transactions: transactions,
                wallets: base.wallets,
                categories: base.categories,
                timezone: accountTimezone,
                monthSummary: monthDetail,
                insider: summary
            ))
        } catch {
            // Online report data remains authoritative even if local storage is unavailable.
        }
    }
}

struct CategoriesView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @State private var items: [CategoryModel] = []
    @State private var loading = false
    @State private var error: String?
    @State private var name = ""
    @State private var kind = "expense"
    @State private var listKind = "expense"
    @State private var showInactiveCategories = false
    @State private var walletScopeID = "__all__"
    @State private var searchText = ""
    @State private var adding = false
    @State private var saving = false
    @State private var editTarget: CategoryModel?
    @State private var deleteTarget: CategoryModel?
    @State private var wallets: [WalletModel] = []
    @State private var selectedWalletIDs: Set<String> = []
    @State private var parentID = ""
    @State private var iconKey = "tag"

    var body: some View {
        Group {
            if adding { categoryEditor }
            else { categoryList }
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .mpConfirmDialog(
            isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }),
            title: "Xóa danh mục?",
            message: "Danh mục con sẽ được xử lý theo chính sách máy chủ.",
            confirmTitle: "Xóa"
        ) {
            guard let target = deleteTarget, !saving else { return }
            saving = true
            Task {
                defer { saving = false }
                do {
                    try await api.deleteCategory(id: target.id)
                    cancelEditor()
                    await load()
                } catch {
                    self.error = "Không thể xóa danh mục."
                }
            }
            deleteTarget = nil
        }
    }

    private var categoryList: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Danh mục", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                MPSegmentedControl(
                    options: [
                        (id: "expense", title: "Khoản chi"),
                        (id: "income", title: "Khoản thu"),
                        (id: "debt", title: "Vay-Nợ")
                    ],
                    selection: $listKind,
                    accessibilityIdentifier: CategoriesReferenceCopy.kindFilterAccessibilityID
                )
                MPButton(CategoriesReferenceCopy.addAction, systemName: "plus.circle", style: .outline) { beginCreating() }
                    .accessibilityIdentifier(CategoriesReferenceCopy.addActionAccessibilityID)
                MPOptionSelector(
                    title: "Phạm vi ví",
                    options: [MPSelectionOption(id: "__all__", label: "Tất cả ví")] + wallets.map { MPSelectionOption(id: $0.id, label: $0.name) },
                    selectionID: $walletScopeID,
                    emptyLabel: "Tất cả ví",
                    accessibilityIdentifier: CategoriesReferenceCopy.walletFilterAccessibilityID
                )
                MPTextField(
                    title: "Tìm danh mục",
                    text: $searchText,
                    placeholder: "Nhập tên nhóm",
                    accessibilityIdentifier: CategoriesReferenceCopy.searchAccessibilityID
                )
                MPButton(showInactiveCategories ? "Đang hiển thị nhóm không hoạt động" : "Hiển thị nhóm không hoạt động", systemName: showInactiveCategories ? "checkmark.circle.fill" : "circle", style: .secondary) {
                    showInactiveCategories.toggle()
                }
                .accessibilityIdentifier("category-show-inactive")
                if loading && items.isEmpty { MPLoadingView(label: "Đang tải danh mục...") }
                else if let error, items.isEmpty { MPErrorView(title: "Không tải được danh mục", detail: error) { Task { await load() } } }
                else if categoryTreeItems.isEmpty {
                    MPText(
                        searchText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
                            ? "Chưa có danh mục phù hợp."
                            : "Không tìm thấy danh mục phù hợp.",
                        style: .secondary
                    ).accessibilityIdentifier("category-empty-filter-result")
                }
                else {
                    MPCategoryTree(items: categoryTreeItems, onSelect: selectCategory)
                        .accessibilityIdentifier("category-tree")
                }
            }
            .padding(MPMetrics.pagePadding)
        }
    }

    private var categoryEditor: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(
                    title: editTarget == nil ? "Nhóm mới" : "Sửa nhóm",
                    accessibilityIdentifier: "category-editor-title",
                    leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại") { if !saving { cancelEditor() } } },
                    trailing: {
                        if let editTarget, CategoryEditPolicy.canDelete(editTarget) {
                            MPIconButton(systemName: "trash", label: "Xóa nhóm") { requestDelete(editTarget.id) }
                                .accessibilityIdentifier("category-delete")
                                .disabled(saving)
                        } else {
                            EmptyView()
                        }
                    }
                )
                if let editTarget, CategoryEditPolicy.metadataIsReadOnly(editTarget) {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 8) {
                            MPText(editTarget.name, style: .title)
                            MPText("\(MPSelectionLabels.display(for: editTarget.kind)) · Nhóm hệ thống", style: .secondary)
                        }
                    }
                    MPText("Đây là nhóm của hệ thống. Bạn chỉ có thể chọn ví áp dụng.", style: .secondary)
                } else {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 12) {
                            MPTextField(title: "Tên nhóm", text: $name)
                            MPOptionSelector(
                                title: "Loại nhóm",
                                options: [
                                    MPSelectionOption(id: "expense", label: "Khoản chi"),
                                    MPSelectionOption(id: "income", label: "Khoản thu"),
                                    MPSelectionOption(id: "debt", label: "Vay/Nợ")
                                ],
                                selectionID: Binding(
                                    get: { kind },
                                    set: { nextKind in if kind != nextKind { kind = nextKind; parentID = "" } }
                                ),
                                accessibilityIdentifier: "category-kind-editor"
                            )
                            MPOptionSelector(
                                title: "Nhóm cha",
                                options: availableParentOptions,
                                selectionID: $parentID,
                                emptyLabel: "Không có nhóm cha",
                                accessibilityIdentifier: "category-parent-editor"
                            )
                            MPCategoryIconSelector(title: "Biểu tượng", selection: $iconKey)
                        }
                    }
                }
                MPWalletApplicabilityCard(wallets: wallets, selectedWalletIDs: $selectedWalletIDs)
                if let error { MPStatusMessage(text: error, tone: .danger) }
                HStack(spacing: 10) {
                    MPButton("Lưu nhóm", isLoading: saving, isDisabled: saving) { Task { await create() } }
                    MPButton("Hủy", style: .secondary, isDisabled: saving) { cancelEditor() }
                }
                if let editTarget, CategoryEditPolicy.canArchive(editTarget) {
                    MPButton(editTarget.isActive == false ? "Khôi phục nhóm" : "Lưu trữ nhóm", style: editTarget.isActive == false ? .secondary : .destructive, isLoading: saving, isDisabled: saving) {
                        Task { await changeActive() }
                    }
                    .accessibilityIdentifier("category-active-toggle")
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .accessibilityIdentifier("category-editor-screen")
    }

    private var categoryTreeItems: [MPCategoryTreeItem] {
        CategoryListPresentation.nodes(
            items: items,
            kind: listKind,
            walletID: walletScopeID == "__all__" ? nil : walletScopeID,
            search: searchText,
            includeInactive: showInactiveCategories
        ).map { node in
            let category = node.item
            let glyph = MPTransactionPresentation.glyph(systemKey: category.systemKey ?? category.iconKey, type: category.kind)
            return MPCategoryTreeItem(
                id: category.id,
                title: category.name,
                subtitle: category.isActive == false ? "Đã lưu trữ · \(CategoriesReferenceCopy.scopeSubtitle(walletIDs: category.walletIDs))" : CategoriesReferenceCopy.scopeSubtitle(walletIDs: category.walletIDs),
                systemName: glyph.systemName,
                tone: glyph.tone,
                depth: node.depth,
                isSystem: category.isSystem,
                isSelectable: !node.isHierarchyContext,
                isHierarchyContext: node.isHierarchyContext,
                continuesBranch: node.continuesBranch
            )
        }
    }
    private var availableParentOptions: [MPSelectionOption] {
        let excluded = excludedParentIDs()
        let roots = items.filter { $0.kind == kind && $0.parentID == nil && !excluded.contains($0.id) }
        return [MPSelectionOption(id: "", label: "Không có nhóm cha")] + roots.map { MPSelectionOption(id: $0.id, label: $0.name) }
    }
    private func excludedParentIDs() -> Set<String> {
        guard let editTarget else { return [] }
        var excluded: Set<String> = [editTarget.id]
        var pending = [editTarget.id]
        while let parentID = pending.popLast() {
            for child in items where child.parentID == parentID {
                if excluded.insert(child.id).inserted { pending.append(child.id) }
            }
        }
        return excluded
    }
    private func load() async { loading = true; defer { loading = false }; do { async let categoryResult = api.categories(); async let walletResult = api.wallets(); items = try await categoryResult; wallets = try await walletResult; self.error = nil } catch { self.error = "Không thể tải danh mục." } }
    private func selectCategory(_ id: String) {
        guard let item = items.first(where: { $0.id == id }) else { return }
        beginEditing(item)
    }
    private func requestDelete(_ id: String) {
        deleteTarget = items.first(where: { $0.id == id })
    }
    private func beginCreating() {
        editTarget = nil
        name = ""
        kind = listKind
        parentID = ""
        iconKey = "tag"
        selectedWalletIDs = []
        error = nil
        adding = true
    }
    private func cancelEditor() {
        adding = false
        editTarget = nil
        name = ""
        parentID = ""
        iconKey = "tag"
        selectedWalletIDs = []
        error = nil
    }
    private func create() async {
        guard !saving else { return }
        saving = true
        defer { saving = false }
        if let editTarget, CategoryEditPolicy.metadataIsReadOnly(editTarget) {
            do {
                _ = try await api.updateCategoryWallets(id: editTarget.id, walletIDs: Array(selectedWalletIDs))
                adding = false; self.editTarget = nil; selectedWalletIDs = []; await load()
            } catch { self.error = "Không thể lưu ví áp dụng." }
            return
        }
        guard !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { self.error = "Nhập tên danh mục."; return }
        do {
            let input = CategoryModelInput(name: name.trimmingCharacters(in: .whitespacesAndNewlines), kind: kind, parentID: parentID.isEmpty ? nil : parentID, walletIDs: Array(selectedWalletIDs).sorted(), iconKey: iconKey)
            if let editTarget { _ = try await api.updateCategory(id: editTarget.id, input: input) } else { _ = try await api.createCategory(input) }
            cancelEditor(); await load()
        } catch { self.error = "Không thể lưu danh mục." }
    }
    private func changeActive() async {
        guard let editTarget, CategoryEditPolicy.canArchive(editTarget), !saving else { return }
        saving = true
        defer { saving = false }
        do {
            _ = try await api.setCategoryActive(id: editTarget.id, active: editTarget.isActive == false)
            cancelEditor()
            await load()
        } catch { self.error = "Không thể cập nhật trạng thái nhóm." }
    }
    private func beginEditing(_ item: CategoryModel) { editTarget = item; name = item.name; kind = item.kind; parentID = item.parentID ?? ""; iconKey = item.iconKey; selectedWalletIDs = Set(item.walletIDs); error = nil; adding = true }
}

enum JarAllocationDraftValidator {
    static let maximumFixedAmount: Int64 = 9_007_199_254_740_991

    static func isValid(mode: String, amount: Int64?, percent: Double?) -> Bool {
        switch mode {
        case "none":
            return amount == nil && percent == nil
        case "fixed":
            return (amount ?? 0) > 0 && (amount ?? 0) <= maximumFixedAmount && percent == nil
        case "percent":
            guard let percent else { return false }
            return percent.isFinite && percent > 0 && percent <= 100 && (percent * 100).rounded() == percent * 100 && amount == nil
        default:
            return false
        }
    }
}

struct JarsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    let masked: Bool
    @State private var summary: JarMonthSummaryModel?
    @State private var month = MPFormat.currentMonth
    @State private var loading = false
    @State private var error: String?
    @State private var saving = false
    @State private var showEditor = false
    @State private var editTarget: JarItemModel?
    @State private var name = ""
    @State private var mode = "none"
    @State private var amount = ""
    @State private var percent = ""
    @State private var report: JarCumulativeSummaryModel?
    @State private var reportJarID = ""
    @State private var reportLoading = false
    @State private var reportError: String?
    @State private var accountTimezone = TimeZone.current.identifier

    var body: some View {
        Group {
            if showEditor { editorContent } else { listContent }
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await load() }
        .onChange(of: month) { _, _ in
            editTarget = nil
            resetEditor()
            showEditor = false
            report = nil
            reportError = nil
            Task { await load() }
        }
        .onChange(of: reportJarID) { _, value in
            guard !value.isEmpty else {
                report = nil
                reportError = nil
                return
            }
            Task { await loadReport(jarID: value) }
        }
        .onChange(of: mode) { _, _ in
            if mode == "fixed" {
                percent = ""
            } else if mode == "percent" {
                amount = ""
            } else {
                amount = ""
                percent = ""
            }
        }
    }

    private var listContent: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Hũ chi tiêu", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                MPOptionSelector(title: "Tháng", options: MPFormat.monthOptions(around: month, timeZone: TimeZone(identifier: accountTimezone) ?? .current).map { MPSelectionOption(id: $0, label: MPFormat.monthLabel($0)) }, selectionID: $month)
                if loading { MPLoadingView(label: "Đang tải hũ tháng này...") }
                if let summary {
                    LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                        MPMetricCard(title: "Thu thực tế", value: MPFormat.money(summary.actualIncome, masked: masked))
                        MPMetricCard(title: "Chi thực tế", value: MPFormat.money(summary.totalSpent, masked: masked))
                        MPMetricCard(title: "Đã phân bổ", value: MPFormat.money(summary.totalAllocated, masked: masked))
                        MPMetricCard(title: "Chi chưa gắn hũ", value: MPFormat.money(summary.unassignedSpent, masked: masked))
                    }
                    if summary.overIncome || summary.overOneHundredPercent {
                        MPStatusMessage(text: "Tổng phân bổ vượt thu thực tế. Đây là cảnh báo, không chặn giao dịch.", tone: .danger)
                    }
                    HStack {
                        MPText("Hũ tháng", style: .title)
                        Spacer()
                        MPButton(JarsReferenceCopy.addAction, style: .secondary, fullWidth: false) {
                            if editTarget != nil { resetEditor() }
                            editTarget = nil
                            error = nil
                            showEditor = true
                        }
                        .accessibilityIdentifier(JarsReferenceCopy.addActionAccessibilityID)
                    }
                    if summary.items.isEmpty { MPEmptyView(title: "Chưa có hũ cho tháng này", detail: "Chi tiêu vẫn được ghi nhận bình thường.") }
                    ForEach(summary.items) { item in
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 8) {
                                MPText(item.name)
                                if let allocation = item.calculatedAllocation ?? (item.allocationMode == "fixed" ? item.allocationAmount : nil) {
                                    MPProgressBar(progress: Double(item.spent) / Double(max(1, allocation)), label: "Đã chi \(MPFormat.money(item.spent, masked: masked)) / \(MPFormat.money(allocation, masked: masked))")
                                } else {
                                    MPText("Chưa đặt phân bổ · Đã chi \(MPFormat.money(item.spent, masked: masked))", style: .secondary)
                                }
                                if item.active {
                                    HStack {
                                        MPButton("Sửa tháng này", style: .secondary) { beginEditing(item) }
                                        MPButton("Gỡ khỏi tháng", style: .destructive) { Task { await remove(item) } }
                                    }
                                } else {
                                    MPText("Cấu hình lịch sử; giao dịch vẫn được giữ.", style: .secondary)
                                }
                            }
                        }
                    }
                    MPAssistantCard(title: "Cộng dồn") {
                        if summary.jars.isEmpty {
                            MPEmptyView(title: "Chưa có hũ", detail: "Tạo hũ để xem chi tiêu cộng dồn.")
                        } else {
                            MPOptionSelector(title: "Chọn hũ", options: summary.jars.map { MPSelectionOption(id: $0.jarID, label: $0.name) }, selectionID: $reportJarID)
                            if reportLoading { MPLoadingView(label: "Đang tải báo cáo cộng dồn...") }
                            if let reportError { MPStatusMessage(text: reportError, tone: .info) }
                            if let report {
                                LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                                    MPMetricCard(title: "Chi thực tế", value: MPFormat.money(report.totalSpent, masked: masked))
                                    MPMetricCard(title: "Đã phân bổ", value: MPFormat.money(report.totalAllocated, masked: masked))
                                }
                                MPText("\(MPFormat.monthLabel(report.fromMonth)) → \(MPFormat.monthLabel(report.toMonth)) · Không cộng dồn số dư.", style: .secondary)
                                ForEach(report.months, id: \.month) { item in
                                    HStack { MPText(MPFormat.monthLabel(item.month)); Spacer(); MPText(MPFormat.money(item.spent, masked: masked), style: .secondary) }
                                }
                            }
                        }
                    }
                } else if let error { MPErrorView(title: "Không tải được hũ", detail: error) { Task { await load() } } }
                else if !loading { MPEmptyView(title: "Chưa chọn tháng", detail: "Hũ là lớp ngân sách theo tháng và không thay đổi số dư ví.") }

            }
            .padding(MPMetrics.pagePadding)
        }
    }

    private var editorContent: some View {
        MPEditorScaffold(title: editTarget == nil ? "Tạo hũ" : "Sửa hũ", isBusy: saving, onCancel: { showEditor = false }) {
            MPTextField(title: "Tên", text: $name, placeholder: "Ăn uống")
            MPWalletSelector(title: "Cách phân bổ", options: ["none", "fixed", "percent"], selection: $mode)
            if mode == "fixed" { MPAmountField(title: "Số tiền mỗi tháng", text: $amount) }
            if mode == "percent" { MPNumberField(title: "Tỷ lệ (%)", text: $percent, kind: .percentage, placeholder: "20") }
            if let error { MPStatusMessage(text: error, tone: .danger) }
        } footer: {
            MPButton("Lưu", isLoading: saving) { Task { await save() } }
        }
    }

    private func load() async {
        guard !loading else { return }
        loading = true; defer { loading = false }
        do {
            let firstLoad = summary == nil
            if let profile = try? await api.profile() {
                accountTimezone = profile.timezone
                if firstLoad { month = MPFormat.currentMonth(timeZone: TimeZone(identifier: accountTimezone) ?? .current) }
            }
            summary = try await api.jarMonth(month: month)
            if let summary {
                if summary.jars.isEmpty {
                    reportJarID = ""
                    report = nil
                    reportError = nil
                } else {
                    if !summary.jars.contains(where: { $0.jarID == reportJarID }) { reportJarID = summary.jars.first?.jarID ?? "" }
                    if !reportJarID.isEmpty { await loadReport(jarID: reportJarID) }
                }
            }
            self.error = nil
        } catch { self.error = "Không thể tải hũ chi tiêu." }
    }
    private func resetEditor() { name = ""; mode = "none"; amount = ""; percent = "" }
    private func beginEditing(_ item: JarItemModel) {
        if editTarget?.jarID == item.jarID {
            editTarget = item
            showEditor = true
            return
        }
        editTarget = item; name = item.name; mode = item.allocationMode; amount = item.allocationAmount.map { String($0) } ?? ""; percent = item.allocationPercent.map { String($0) } ?? ""; showEditor = true
    }
    private func save() async {
        guard !saving else { return }
        guard !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { error = "Nhập tên hũ."; return }
        let parsedAmount = MPFormat.parseNonNegativeInteger(amount)
        let parsedPercent = MPNumberInputKind.percentage.normalized(percent).flatMap(Double.init)
        guard JarAllocationDraftValidator.isValid(mode: mode, amount: parsedAmount, percent: parsedPercent) else { error = "Cách phân bổ không hợp lệ."; return }
        let input = JarConfigInputModel(month: month, name: name.trimmingCharacters(in: .whitespacesAndNewlines), allocationMode: mode, allocationAmount: mode == "fixed" ? parsedAmount : nil, allocationPercent: mode == "percent" ? parsedPercent : nil)
        saving = true; error = nil; defer { saving = false }
        do { if let editTarget { _ = try await api.updateJarMonth(jarID: editTarget.jarID, month: month, input: input) } else { _ = try await api.createJar(input) }; showEditor = false; editTarget = nil; resetEditor(); await load() }
        catch { self.error = "Không thể lưu hũ." }
    }
    private func remove(_ item: JarItemModel) async { do { try await api.removeJarMonth(jarID: item.jarID, month: month); await load() } catch { self.error = "Không thể gỡ hũ khỏi tháng." } }
    private func loadReport(jarID: String) async {
        reportLoading = true; reportError = nil
        defer { reportLoading = false }
        do { report = try await api.jarReport(jarID: jarID, to: month) }
        catch { report = nil; reportError = "Chưa có dữ liệu cộng dồn cho hũ này." }
    }
}

enum MonthsPresentation {
    static func timezoneLabel(_ timezone: String) -> String {
        "theo múi giờ tài khoản"
    }

    static func reportPeriod(startAt: String, nextStartAt: String, timeZone: TimeZone) -> String {
        guard
            let start = MPFormat.parseInstant(startAt),
            let nextStart = MPFormat.parseInstant(nextStartAt)
        else { return "Không xác định kỳ báo cáo" }

        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = timeZone
        let inclusiveEnd = calendar.date(byAdding: .day, value: -1, to: nextStart) ?? nextStart
        return "\(MPFormat.shortDate(MPFormat.instantString(start), timeZone: timeZone)) – \(MPFormat.shortDate(MPFormat.instantString(inclusiveEnd), timeZone: timeZone))"
    }
}

struct MonthsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    let onOpenJars: () -> Void
    let masked: Bool
    @State private var month = MPFormat.currentMonth
    @State private var summary: MonthSummaryModel?
    @State private var note = ""
    @State private var savedNote = ""
    @State private var pendingMonth: String?
    @State private var loading = false
    @State private var error: String?
    @State private var accountTimezone = TimeZone.current.identifier

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Tháng", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                MPOptionSelector(title: "Tháng", options: MPFormat.monthOptions(around: month, timeZone: TimeZone(identifier: accountTimezone) ?? .current).map { MPSelectionOption(id: $0, label: MPFormat.monthLabel($0)) }, selectionID: monthSelection)
                if loading && summary == nil { MPLoadingView(label: "Đang tải tổng kết...") }
                if let summary {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 10) {
                            let reportTimeZone = TimeZone(identifier: summary.timezone) ?? accountTimezoneValue
                            HStack {
                                VStack(alignment: .leading, spacing: 4) {
                                    MPText(MPFormat.monthLabel(summary.month), style: .title)
                                    MPText("\(summary.isComplete ? "Đã hoàn tất" : "Đang diễn ra") · \(MonthsPresentation.timezoneLabel(summary.timezone))", style: .secondary)
                                }
                                Spacer()
                                MPButton("Hũ tháng", style: .secondary, action: onOpenJars)
                            }
                            LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                                MPMetricCard(title: "Thu", value: MPFormat.money(summary.income, masked: masked))
                                MPMetricCard(title: "Chi", value: MPFormat.money(summary.expense, masked: masked))
                                MPMetricCard(title: "Chênh lệch", value: MPFormat.signedMoney(summary.net, masked: masked))
                                MPMetricCard(title: "Giao dịch", value: "\(summary.transactionCount)")
                            }
                            MPText("Kỳ báo cáo: \(MonthsPresentation.reportPeriod(startAt: summary.startAt, nextStartAt: summary.nextStartAt, timeZone: reportTimeZone))", style: .secondary)
                        }
                    }
                    MPAssistantCard(title: "Theo nhóm") {
                        if summary.categories.isEmpty {
                            MPEmptyView(title: "Chưa có giao dịch", detail: "Tháng này chưa có dữ liệu.")
                        } else {
                            ForEach(summary.categories, id: \.stableID) { item in
                                HStack {
                                    VStack(alignment: .leading, spacing: 3) {
                                        MPText(item.name)
                                        MPText("\(item.type == "income" ? "Thu" : "Chi") · \(item.count) giao dịch", style: .secondary)
                                    }
                                    Spacer()
                                    MPText(MPFormat.money(item.amount, masked: masked), style: .secondary)
                                }
                            }
                        }
                    }
                    MPAssistantCard(title: "Hũ chi tiêu") {
                        if let jarSummary = summary.jarSummary, !jarSummary.items.isEmpty {
                            ForEach(jarSummary.items) { item in
                                let allocation = item.calculatedAllocation ?? (item.allocationMode == "fixed" ? item.allocationAmount : nil)
                                VStack(alignment: .leading, spacing: 5) {
                                    HStack { MPText(item.name); Spacer(); MPText(MPFormat.money(item.spent, masked: masked), style: .secondary) }
                                    if let allocation {
                                        MPProgressBar(progress: Double(item.spent) / Double(max(1, allocation)), label: "Mức sử dụng hũ")
                                    }
                                }
                            }
                        } else {
                            MPEmptyView(title: "Chưa có hũ", detail: "Quản lý hũ cho tháng này nếu cần.")
                        }
                    }
                    MPAssistantCard(title: "Ghi chú của tôi") {
                        MPTextEditor(title: "Ghi chú tháng", text: $note, placeholder: "Điều bạn muốn ghi nhớ về tháng này")
                        HStack {
                            MPButton("Lưu ghi chú", style: .secondary, isDisabled: note == savedNote) { Task { await saveNote() } }
                            MPButton("Bỏ thay đổi", style: .secondary, isDisabled: note == savedNote) { note = savedNote }
                            if !savedNote.isEmpty { MPButton("Xóa ghi chú", style: .destructive) { Task { await deleteNote() } } }
                        }
                    }
                } else if let error {
                    MPErrorView(title: "Không tải được tháng", detail: error) { Task { await load() } }
                } else if !loading {
                    MPEmptyView(title: "Chưa có dữ liệu", detail: "Chọn một tháng để xem tổng kết.")
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await load() }
        .mpConfirmDialog(isPresented: Binding(get: { pendingMonth != nil }, set: { if !$0 { pendingMonth = nil } }), title: "Bỏ ghi chú đang sửa?", message: "Nội dung chưa lưu sẽ bị bỏ khi chuyển tháng.", confirmTitle: "Chuyển tháng") {
                guard let next = pendingMonth else { return }
                pendingMonth = nil
                month = next
                Task { await load() }
        }
    }

    private var accountTimezoneValue: TimeZone { TimeZone(identifier: accountTimezone) ?? .current }
    private var monthSelection: Binding<String> {
        Binding(get: { month }, set: { selectMonth($0) })
    }
    private var noteIsDirty: Bool { note != savedNote }

    private func selectMonth(_ next: String) {
        guard next != month else { return }
        if noteIsDirty {
            pendingMonth = next
        } else {
            month = next
            Task { await load() }
        }
    }

    private func load() async {
        guard !loading else { return }
        loading = true
        defer { loading = false }
        do {
            let firstLoad = summary == nil
            if let profile = try? await api.profile() {
                accountTimezone = profile.timezone
                if firstLoad { month = MPFormat.currentMonth(timeZone: TimeZone(identifier: accountTimezone) ?? .current) }
            }
            summary = try await api.monthSummary(month: month)
            note = summary?.note ?? ""
            savedNote = note
            self.error = nil
        } catch { self.error = "Không thể tải tháng." }
    }

    private func saveNote() async {
        do {
            if note.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                try await api.deleteMonthNote(month: month)
                note = ""
            } else {
                try await api.saveMonthNote(month: month, note: note)
            }
            savedNote = note
            error = nil
        } catch { self.error = "Không thể lưu ghi chú." }
    }

    private func deleteNote() async {
        do {
            try await api.deleteMonthNote(month: month)
            note = ""
            savedNote = ""
            error = nil
        } catch { self.error = "Không thể xóa ghi chú." }
    }
}

enum RecurringDraftValidator {
    static func positiveInteger(_ value: String) -> Int64? {
        let trimmed = value.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }

        let groups = trimmed.split(separator: ".", omittingEmptySubsequences: false)
        let grouped = groups.count > 1
        guard groups.allSatisfy({ group in
            let scalars = group.unicodeScalars
            guard scalars.allSatisfy({ $0.value >= 48 && $0.value <= 57 }) else { return false }
            guard !scalars.isEmpty else { return false }
            if !grouped { return true }
            if group != groups.first { return scalars.count == 3 }
            return scalars.count <= 3
        }) else { return nil }

        let digits = groups.joined()
        guard let parsed = MPFormat.parsePositiveInteger(digits) else { return nil }
        return parsed
    }

    static func isValid(
        name: String,
        amount: String,
        interval: String,
        type: String,
        frequency: String,
        nextRunDate: Date,
        endsAtEnabled: Bool,
        endsAtDate: Date
    ) -> Bool {
        guard !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty,
              positiveInteger(amount) != nil,
              MPFormat.parsePositiveInteger(interval) != nil,
              type == "expense" || type == "income",
              ["daily", "weekly", "monthly", "yearly"].contains(frequency) else { return false }
        return !endsAtEnabled || endsAtDate >= nextRunDate
    }
}

struct RecurringView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @Environment(\.mpFinancialValuesMasked) private var masked
    @State private var items: [RecurringModel] = []
    @State private var wallets: [WalletModel] = []
    @State private var categories: [CategoryModel] = []
    @State private var showEditor = false
    @State private var editTarget: RecurringModel?
    @State private var name = ""
    @State private var amount = ""
    @State private var walletID = ""
    @State private var frequency = "monthly"
    @State private var type = "expense"
    @State private var categoryID = ""
    @State private var categoryPickerPresented = false
    @State private var interval = "1"
    @State private var active = "Có"
    @State private var note = ""
    @State private var nextRunDate = Date()
    @State private var endsAtDate = Date()
    @State private var endsAtEnabled = "Không"
    @State private var accountTimezone = TimeZone.current.identifier
    @State private var result: String?
    @State private var error: String?
    @State private var saving = false
    @State private var deleteTarget: RecurringModel?

    var body: some View {
        Group {
            if showEditor { editorContent } else { listContent }
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await load() }
        .onChange(of: type) { _, _ in
            categoryID = TransactionEditorPolicy.requiredCategoryID(type: type, walletID: walletID, walletType: wallets.first(where: { $0.id == walletID })?.type ?? "basic", categories: categories, currentID: categoryID)
        }
        .mpConfirmDialog(isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }), title: "Xóa lịch định kỳ?", message: "Các giao dịch đã tạo vẫn giữ nguyên trong sổ giao dịch.", confirmTitle: "Xóa") {
            if let target = deleteTarget { Task { await delete(target) } }
            deleteTarget = nil
        }
        .mpFormBottomSheet(
            isPresented: $categoryPickerPresented,
            title: "Chọn danh mục",
            accessibilityIdentifier: "recurring-category-picker-sheet",
            onClose: { categoryPickerPresented = false },
            fields: {
                MPCategoryPicker(
                    categories: categories,
                    kind: type,
                    selectableIDs: Set(categoriesForType.map(\.id)),
                    selectedID: categoryID,
                    clearLabel: nil,
                    clearID: nil,
                    onSelect: { categoryID = $0; categoryPickerPresented = false }
                )
            },
            footer: {
                MPButton("Đóng", style: .secondary) { categoryPickerPresented = false }
            }
        )
    }

    private var categoriesForType: [CategoryModel] {
        TransactionEditorPolicy.applicableCategories(type: type, walletID: walletID, walletType: wallets.first(where: { $0.id == walletID })?.type ?? "basic", categories: categories)
    }
    private func walletName(for id: String) -> String { wallets.first(where: { $0.id == id })?.name ?? "Ví đã xóa" }
    private func frequencyLabel(_ value: String, interval: Int) -> String {
        let label: String
        switch value {
        case "daily": label = "Hàng ngày"
        case "weekly": label = "Hàng tuần"
        case "monthly": label = "Hàng tháng"
        case "yearly": label = "Hàng năm"
        default: label = value
        }
        return interval == 1 ? label : "Mỗi \(interval) kỳ · \(label.lowercased())"
    }
    private var listContent: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Giao dịch định kỳ", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                MPSurfaceCard {
                    VStack(alignment: .leading, spacing: 8) {
                        HStack {
                            MPText("Giao dịch định kỳ", style: .title)
                            Spacer()
                            MPButton(RecurringReferenceCopy.addAction, fullWidth: false) {
                                if editTarget != nil { resetEditor() }
                                editTarget = nil
                                error = nil
                                showEditor = true
                            }
                            .accessibilityIdentifier(RecurringReferenceCopy.addActionAccessibilityID)
                        }
                        MPText("Lịch bật sẽ sinh giao dịch thường theo múi giờ tài khoản. Mỗi kỳ chỉ tạo một lần; giao dịch đã tạo vẫn có thể sửa hoặc xóa như bình thường.", style: .secondary)
                        MPButton("Chạy các giao dịch đến hạn", style: .secondary) { Task { await runDue() } }
                    }
                }
                if let result { MPStatusMessage(text: result, tone: .success) }
                if let error { MPStatusMessage(text: error, tone: .danger) }
                if items.isEmpty { MPEmptyView(title: "Chưa có lịch định kỳ", detail: "Lịch định kỳ chỉ tạo giao dịch khi bạn chủ động chạy.") }
                ForEach(items) { item in
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 6) {
                            HStack {
                                MPText(item.name)
                                Spacer()
                                MPStatusBadge(text: item.active ? "Đang bật" : "Tạm dừng", tone: MPStatusBadge.tone(for: item.active))
                            }
                            MPText("\(item.type == "expense" ? "Khoản chi" : "Khoản thu") · \(frequencyLabel(item.frequency, interval: item.interval)) · \(MPFormat.money(item.amount, masked: masked))", style: .secondary)
                            MPText("Ví: \(walletName(for: item.walletID))", style: .secondary)
                            MPText("Lần chạy kế tiếp: \(MPFormat.shortDateTime(item.nextRunAt, timeZone: TimeZone(identifier: accountTimezone) ?? .current))", style: .secondary)
                            if let note = item.note, !note.isEmpty { MPText(note, style: .secondary) }
                            HStack {
                                MPButton("Sửa", style: .secondary) { beginEditing(item) }
                                    .accessibilityIdentifier("recurring-edit-\(item.id)")
                                MPButton("Xóa", style: .destructive) { deleteTarget = item }
                                    .accessibilityIdentifier("recurring-delete-\(item.id)")
                            }
                        }
                    }
                }

            }
            .padding(MPMetrics.pagePadding)
        }
    }

    private var editorContent: some View {
        MPEditorScaffold(title: editTarget == nil ? "Tạo lịch" : "Sửa lịch", isBusy: saving, onCancel: { showEditor = false }) {
            MPTextField(title: "Tên", text: $name, accessibilityIdentifier: "recurring-name")
            MPAmountField(title: "Số tiền", text: $amount)
            MPOptionSelector(title: "Ví", options: wallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $walletID)
            MPWalletSelector(title: "Loại", options: ["expense", "income"], selection: $type)
            MPFormRow(
                title: "Danh mục",
                iconName: "tag",
                value: categories.first(where: { $0.id == categoryID }).map { $0.name + ($0.isActive == false ? " (đã lưu trữ)" : "") } ?? "Chọn danh mục",
                accessibilityIdentifier: "recurring-category-row"
            ) { categoryPickerPresented = true }
            MPWalletSelector(title: "Tần suất", options: ["daily", "weekly", "monthly", "yearly"], selection: $frequency)
            MPNumberField(title: "Mỗi bao nhiêu kỳ", text: $interval, kind: .integer)
            MPDateTimeField(title: "Lần chạy kế tiếp", date: $nextRunDate, timeZone: accountTimezone)
            MPWalletSelector(title: "Có ngày kết thúc", options: ["Có", "Không"], selection: $endsAtEnabled)
            if endsAtEnabled == "Có" { MPDateTimeField(title: "Ngày kết thúc", date: $endsAtDate, timeZone: accountTimezone) }
            MPWalletSelector(title: "Trạng thái", options: ["Có", "Không"], selection: $active)
            MPTextField(title: "Ghi chú", text: $note, placeholder: "Không bắt buộc")
            if categories.first(where: { $0.id == categoryID })?.isActive == false {
                MPStatusMessage(text: "Lịch hiện tại giữ nhóm đã lưu trữ; chọn nhóm đang hoạt động để thay đổi.", tone: .info)
            }
            if let error { MPStatusMessage(text: error, tone: .danger) }
        } footer: {
            MPButton("Lưu", isLoading: saving) { Task { await save() } }
        }
    }

    private func load() async { do { async let schedules = api.recurring(); async let walletList = api.wallets(); async let categoryList = api.categories(); async let profile = api.profile(); items = try await schedules; wallets = (try await walletList).filter { $0.type != "credit" }; categories = try await categoryList; if let profile = try? await profile { accountTimezone = profile.timezone }; if walletID.isEmpty { walletID = wallets.first?.id ?? "" }; categoryID = TransactionEditorPolicy.requiredCategoryID(type: type, walletID: walletID, walletType: wallets.first(where: { $0.id == walletID })?.type ?? "basic", categories: categories, currentID: categoryID); self.error = nil } catch { self.error = "Không thể tải lịch định kỳ." } }
    private func runDue() async { do { let response = try await api.runDueRecurring(); result = "Đã tạo \(response.created) giao dịch."; await load() } catch { self.error = "Không thể chạy giao dịch đến hạn." } }
    private func resetEditor() { name = ""; amount = ""; walletID = wallets.first?.id ?? ""; frequency = "monthly"; type = "expense"; categoryID = TransactionEditorPolicy.requiredCategoryID(type: "expense", walletID: wallets.first?.id ?? "", walletType: wallets.first?.type ?? "basic", categories: categories, currentID: ""); interval = "1"; note = ""; nextRunDate = Date(); endsAtDate = Date(); endsAtEnabled = "Không"; active = "Có" }
    private func beginEditing(_ item: RecurringModel) {
        if editTarget?.id == item.id {
            editTarget = item
            showEditor = true
            return
        }
        editTarget = item; name = item.name; amount = String(item.amount); note = item.note ?? ""; walletID = item.walletID; type = item.type; categoryID = TransactionEditorPolicy.requiredCategoryID(type: item.type, walletID: item.walletID, walletType: wallets.first(where: { $0.id == item.walletID })?.type ?? "basic", categories: categories, currentID: item.categoryID ?? ""); frequency = item.frequency; interval = String(item.interval); nextRunDate = MPFormat.parseInstant(item.nextRunAt) ?? Date(); if let endsAt = item.endsAt, let parsed = MPFormat.parseInstant(endsAt) { endsAtDate = parsed; endsAtEnabled = "Có" } else { endsAtDate = Date(); endsAtEnabled = "Không" }; active = item.active ? "Có" : "Không"; showEditor = true
    }
    private func save() async {
        guard !saving else { return }
        guard let wallet = wallets.first(where: { $0.id == walletID }),
              RecurringDraftValidator.isValid(
                  name: name,
                  amount: amount,
                  interval: interval,
                  type: type,
                  frequency: frequency,
                  nextRunDate: nextRunDate,
                  endsAtEnabled: endsAtEnabled == "Có",
                  endsAtDate: endsAtDate
              ),
              let parsed = RecurringDraftValidator.positiveInteger(amount),
              let parsedInterval = MPFormat.parsePositiveInteger(interval).flatMap(Int.init) else {
            self.error = "Nhập đủ tên, ví, nhóm, chu kỳ và số tiền hợp lệ."
            return
        }
        guard let category = categories.first(where: { $0.id == categoryID }),
              category.applies(to: type, walletID: wallet.id, walletType: wallet.type)
                || (editTarget?.categoryID == category.id && category.isActive == false && category.appliesIgnoringActivity(to: type, walletID: wallet.id, walletType: wallet.type)) else {
            self.error = "Hãy chọn nhóm cho lịch định kỳ."
            return
        }
        let trimmedNote = note.trimmingCharacters(in: .whitespacesAndNewlines)
        let input = RecurringInputModel(
            name: name.trimmingCharacters(in: .whitespacesAndNewlines),
            walletID: wallet.id,
            categoryID: category.id,
            type: type,
            amount: parsed,
            note: trimmedNote.isEmpty ? nil : trimmedNote,
            frequency: frequency,
            interval: parsedInterval,
            nextRunAt: MPFormat.instantString(nextRunDate),
            endsAt: endsAtEnabled == "Có" ? MPFormat.instantString(endsAtDate) : nil,
            active: active == "Có"
        )
        saving = true; error = nil; defer { saving = false }
        do {
            if let editTarget { _ = try await api.updateRecurring(id: editTarget.id, input: input) }
            else { _ = try await api.createRecurring(input) }
            showEditor = false
            self.editTarget = nil
            resetEditor()
            await load()
        } catch { self.error = "Không thể lưu lịch định kỳ." }
    }
    private func delete(_ item: RecurringModel) async { do { try await api.deleteRecurring(id: item.id); await load() } catch { self.error = "Không thể xóa lịch định kỳ." } }
}

struct TravelView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @Environment(\.mpFinancialValuesMasked) private var masked
    @State private var events: [TravelEventModel] = []
    @State private var error: String?
    @State private var saving = false
    @State private var name = ""
    @State private var context = ""
    @State private var adding = false
    @State private var editTarget: TravelEventModel?
    @State private var deleteTarget: TravelEventModel?
    @State private var startsOn = ""
    @State private var endsOn = ""
    @State private var accountTimezone = TimeZone.current.identifier

    var body: some View {
        Group {
            if adding { editorContent } else { listContent }
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .mpConfirmDialog(isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }), title: "Xóa chuyến đi?", message: "Giao dịch vẫn được giữ nguyên; chỉ gỡ liên kết chuyến.", confirmTitle: "Xóa") {
            if let target = deleteTarget { Task { await remove(target) } }
            deleteTarget = nil
        }
    }
    private var listContent: some View {
        ScrollView { VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: "Travel Mode", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
            MPSurfaceCard {
                VStack(alignment: .leading, spacing: 8) {
                    HStack {
                        MPText("Travel Mode", style: .title)
                        Spacer()
                        MPButton(TravelReferenceCopy.addAction, fullWidth: false) {
                            if editTarget != nil { resetEditor() }
                            editTarget = nil
                            error = nil
                            adding = true
                        }
                        .accessibilityIdentifier(TravelReferenceCopy.addActionAccessibilityID)
                    }
                    MPText("Gắn giao dịch thu/chi vào một chuyến để xem bối cảnh. Chuyến không phải là ví và không đổi số dư.", style: .secondary)
                    if let active = events.first(where: { $0.active }) {
                        MPText("Đang bật: \(active.name). Giao dịch thu/chi thường mới sẽ tự gắn chuyến này.", style: .secondary)
                    } else {
                        MPText("Chưa bật chuyến nào. Giao dịch mới sẽ không tự gắn Travel Mode.", style: .secondary)
                    }
                }
            }
            if let error { MPErrorView(title: "Không tải được Travel Mode", detail: error) { Task { await load() } } }
            if events.isEmpty { MPEmptyView(title: "Chưa có chuyến đi", detail: "Tạo một sự kiện để gắn giao dịch du lịch.") }
            ForEach(events) { event in
                MPSurfaceCard {
                    VStack(alignment: .leading, spacing: 8) {
                        HStack {
                            VStack(alignment: .leading, spacing: 4) {
                                MPText(event.name)
                                MPText(event.active ? "Đang bật" : "Đã tắt", style: .secondary)
                            }
                            Spacer()
                            MPButton(event.active ? "Tắt" : "Bật", style: event.active ? .destructive : .secondary) { Task { await toggle(event) } }
                                .accessibilityIdentifier("travel-toggle-\(event.id)")
                        }
                        MPText(eventDateLabel(event), style: .secondary)
                        if let context = event.context, !context.isEmpty { MPText(context, style: .secondary) }
                        HStack {
                            MPButton("Sửa", style: .secondary) { beginEditing(event) }
                                .accessibilityIdentifier("travel-edit-\(event.id)")
                            MPButton("Xóa", style: .destructive) { deleteTarget = event }
                                .accessibilityIdentifier("travel-delete-\(event.id)")
                        }
                    }
                }
            }

        }.padding(MPMetrics.pagePadding) }
    }

    private var editorContent: some View {
        MPEditorScaffold(title: editTarget == nil ? "Chuyến đi mới" : "Sửa chuyến đi", isBusy: saving, onCancel: { adding = false }) {
            MPText("Chỉ giao dịch thu/chi tạo mới mới tự nhận chuyến đang bật. Giao dịch định kỳ, chuyển ví và điều chỉnh số dư không tự gắn.", style: .secondary)
            MPTextField(title: "Tên", text: $name)
            MPTextEditor(title: "Mô tả", text: $context)
            MPOptionalDateField(title: "Ngày bắt đầu", value: $startsOn, timeZone: accountTimezone, accessibilityIdentifier: "travel-start")
            MPOptionalDateField(title: "Ngày kết thúc", value: $endsOn, timeZone: accountTimezone, accessibilityIdentifier: "travel-end")
            if let error { MPStatusMessage(text: error, tone: .danger) }
        } footer: {
            MPButton("Lưu", isLoading: saving) { Task { await save() } }
        }
    }

    private func load() async {
        do {
            async let eventList = api.travelEvents()
            async let profile = api.profile()
            events = try await eventList
            if let owner = try? await profile { accountTimezone = owner.timezone }
            self.error = nil
        } catch { self.error = "Không thể tải sự kiện du lịch." }
    }
    private func toggle(_ event: TravelEventModel) async { do { _ = try await api.setTravelEventActive(id: event.id, active: !event.active); await load() } catch { self.error = "Không thể cập nhật Travel Mode." } }
    private func resetEditor() { name = ""; context = ""; startsOn = ""; endsOn = "" }
    private func beginEditing(_ event: TravelEventModel) {
        if editTarget?.id == event.id {
            editTarget = event
            adding = true
            return
        }
        editTarget = event; name = event.name; context = event.context ?? ""; startsOn = event.startsOn ?? ""; endsOn = event.endsOn ?? ""; adding = true
    }
    private func eventDateLabel(_ event: TravelEventModel) -> String {
        switch (event.startsOn, event.endsOn) {
        case let (start?, end?): return "\(start) → \(end)"
        case let (start?, nil): return "Từ \(start) · Không giới hạn ngày kết thúc"
        case let (nil, end?): return "Đến \(end) · Không giới hạn ngày bắt đầu"
        default: return "Không giới hạn ngày"
        }
    }
    private func save() async { guard !saving else { return }; guard TravelEventDraftValidator.isValid(name: name, startsOn: startsOn, endsOn: endsOn) else { self.error = "Tên chuyến và khoảng ngày không hợp lệ."; return }; let trimmedName = name.trimmingCharacters(in: .whitespacesAndNewlines); let trimmedContext = context.trimmingCharacters(in: .whitespacesAndNewlines); let input = TravelEventInputModel(name: trimmedName, context: trimmedContext.isEmpty ? nil : trimmedContext, startsOn: startsOn.isEmpty ? nil : startsOn, endsOn: endsOn.isEmpty ? nil : endsOn, active: editTarget?.active ?? false); saving = true; error = nil; defer { saving = false }; do { if let editTarget { _ = try await api.updateTravelEvent(id: editTarget.id, input: input) } else { _ = try await api.createTravelEvent(input) }; resetEditor(); editTarget = nil; adding = false; await load() } catch { self.error = "Không thể lưu chuyến đi." } }
    private func remove(_ event: TravelEventModel) async { do { try await api.deleteTravelEvent(id: event.id); await load() } catch { self.error = "Không thể xóa chuyến đi." } }
}

enum TravelEventDraftValidator {
    static func isValid(name: String, startsOn: String, endsOn: String) -> Bool {
        guard !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { return false }
        func validDate(_ value: String) -> Bool {
            guard value.isEmpty else {
                guard let parsed = MPFormat.parseCalendarDate(value) else { return false }
                return MPFormat.calendarDateString(parsed) == value
            }
            return true
        }
        guard validDate(startsOn), validDate(endsOn) else { return false }
        return startsOn.isEmpty || endsOn.isEmpty || startsOn <= endsOn
    }
}

struct PortfolioView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @Environment(\.mpFinancialValuesMasked) private var masked
    @State private var summary: PortfolioSummaryModel?
    @State private var tradeHistory: [PortfolioTradeModel] = []
    @State private var error: String?
    @State private var symbol = ""
    @State private var assetName = ""
    @State private var tradeAssetID = ""
    @State private var tradeSide = "buy"
    @State private var tradeQuantity = ""
    @State private var tradePrice = ""
    @State private var tradeFee = ""
    @State private var tradeDate = Date()
    @State private var priceByAsset: [String: String] = [:]
    @State private var assetEditorPresented = false
    @State private var tradeEditorPresented = false
    @State private var priceTarget: PortfolioAssetModel?
    @State private var deleteTarget: PortfolioAssetModel?
    @State private var saving = false
    @State private var accountTimezone = TimeZone.current.identifier

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                MPScreenHeader(title: "Danh mục đầu tư", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: {
                    MPButton(PortfolioReferenceCopy.refreshAction, style: .secondary, fullWidth: false) { Task { await load() } }
                        .accessibilityIdentifier(PortfolioReferenceCopy.refreshActionAccessibilityID)
                })
                MPText(PortfolioReferenceCopy.subtitle, style: .secondary)
                    .accessibilityIdentifier("portfolio-subtitle")
                HStack(spacing: 8) {
                    MPButton("Thêm tài sản", style: .secondary) { error = nil; assetEditorPresented = true }
                        .accessibilityIdentifier("portfolio-open-asset-editor")
                    MPButton("Ghi giao dịch", style: .secondary, isDisabled: summary?.assets.isEmpty != false) { error = nil; tradeEditorPresented = true }
                        .accessibilityIdentifier("portfolio-open-trade-editor")
                }
                if let error, summary != nil { MPStatusMessage(text: error, tone: .danger) }
                if let summary {
                    MPSurfaceCard(style: .reference) {
                        VStack(alignment: .leading, spacing: MPMetrics.sectionSpacing) {
                            MPSectionHeader(title: PortfolioReferenceCopy.summaryTitle)
                            LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: MPMetrics.sectionSpacing) {
                                MPMetric(title: "Giá vốn", value: MPFormat.money(summary.totalCostBasis, masked: masked), compact: true)
                                MPMetric(title: "Giá trị hiện tại", value: summary.totalMarketValue.map { MPFormat.money($0, masked: masked) } ?? "Chưa có giá", compact: true)
                                MPMetric(title: "Lãi/lỗ chưa thực hiện", value: summary.totalUnrealizedPnL.map { MPFormat.signedMoney($0, masked: masked) } ?? "Chưa có giá", compact: true)
                            }
                        }
                    }
                    .accessibilityIdentifier("portfolio-summary")
                    ForEach(summary.assets) { asset in
                        portfolioAssetCard(asset, position: summary.positions.first(where: { $0.assetID == asset.id }), trades: tradeHistory)
                    }
                } else if let error {
                    MPErrorView(title: "Không tải được đầu tư", detail: error) { Task { await load() } }
                } else {
                    MPEmptyView(title: "Chưa có danh mục đầu tư", detail: "Giá chưa biết được hiển thị là chưa định giá, không phải bằng 0.")
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await load() }
        .mpFullScreenCover(isPresented: $assetEditorPresented) {
            MPEditorScaffold(title: "Thêm tài sản", isBusy: saving, onCancel: { assetEditorPresented = false }) {
                MPTextField(title: "Mã", text: $symbol, placeholder: "VN30F")
                MPTextField(title: "Tên", text: $assetName, placeholder: "Quỹ chỉ số")
                if let error { MPStatusMessage(text: error, tone: .danger) }
            } footer: {
                MPButton(PortfolioReferenceCopy.addAssetAction, isLoading: saving, isDisabled: PortfolioAssetDraftValidator.normalized(symbol: symbol, name: assetName) == nil) { Task { await createAsset() } }
                    .accessibilityIdentifier(PortfolioReferenceCopy.addAssetActionAccessibilityID)
            }
        }
        .mpFullScreenCover(isPresented: $tradeEditorPresented) {
            MPEditorScaffold(title: "Ghi giao dịch đầu tư", isBusy: saving, onCancel: { tradeEditorPresented = false }) {
                MPOptionSelector(title: "Tài sản", options: (summary?.assets ?? []).map { MPSelectionOption(id: $0.id, label: "\($0.symbol) · \($0.name)") }, selectionID: $tradeAssetID, emptyLabel: "Chọn tài sản", accessibilityIdentifier: "portfolio-trade-asset")
                MPWalletSelector(title: "Loại", options: ["buy", "sell"], selection: $tradeSide)
                MPNumberField(title: "Khối lượng", text: $tradeQuantity, kind: .decimal, placeholder: "1,5")
                MPAmountField(title: "Giá đơn vị", text: $tradePrice)
                MPAmountField(title: "Phí", text: $tradeFee)
                MPDateField(title: "Ngày", date: $tradeDate, timeZone: accountTimezone)
                if let error { MPStatusMessage(text: error, tone: .danger) }
            } footer: {
                MPButton(PortfolioReferenceCopy.saveTradeAction, isLoading: saving, isDisabled: selectedTradeAsset == nil || !PortfolioTradeDraftValidator.isValid(side: tradeSide, quantity: tradeQuantity, unitPrice: tradePrice, fee: tradeFee)) { Task { await saveTrade() } }
                    .accessibilityIdentifier(PortfolioReferenceCopy.saveTradeActionAccessibilityID)
            }
        }
        .mpFormBottomSheet(isPresented: Binding(get: { priceTarget != nil }, set: { if !$0 { priceTarget = nil } }), title: "Cập nhật giá", accessibilityIdentifier: "portfolio-price-editor", isDismissDisabled: saving, onClose: { if !saving { priceTarget = nil } }) {
            if let asset = priceTarget {
                MPText("\(asset.symbol) · \(asset.name)", style: .emphasis)
                MPAmountField(title: "Giá hiện tại", text: Binding(get: { priceByAsset[asset.id] ?? asset.latestPrice.map(String.init) ?? "" }, set: { priceByAsset[asset.id] = $0 }), accessibilityIdentifier: "portfolio-price-\(asset.id)")
                    .disabled(saving)
            }
            if let error { MPStatusMessage(text: error, tone: .danger) }
        } footer: {
            if let asset = priceTarget {
                MPButton("Lưu giá", isLoading: saving) { Task { await updatePrice(asset) } }
                    .accessibilityIdentifier("portfolio-price-save")
            }
        }
        .mpConfirmDialog(
            isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }),
            title: "Xóa tài sản?",
            message: "Chỉ tài sản chưa có giao dịch mới có thể xóa. Lịch sử đầu tư không bị thay đổi.",
            confirmTitle: "Xóa"
        ) {
            if let target = deleteTarget { Task { await deleteAsset(target) } }
            deleteTarget = nil
        }
    }

    @ViewBuilder
    private func portfolioAssetCard(_ asset: PortfolioAssetModel, position: PortfolioPositionModel?, trades: [PortfolioTradeModel]) -> some View {
        let assetTrades = Self.tradesForAsset(trades, assetID: asset.id)
        MPAssistantCard(title: "\(asset.symbol) · \(asset.name)") {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    MPText("Số lượng: \(position?.quantity ?? "0")")
                    MPText(asset.latestPrice.map { "Giá hiện tại: \(MPFormat.money($0, masked: masked))" } ?? "Chưa có giá", style: .secondary)
                }
                Spacer()
                MPButton("Giao dịch", style: .secondary) {
                    tradeAssetID = asset.id
                    tradeSide = "buy"
                    tradeQuantity = ""
                    tradePrice = asset.latestPrice.map(String.init) ?? ""
                    tradeFee = ""
                    tradeDate = Date()
                    error = nil
                    tradeEditorPresented = true
                }
            }
            LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: MPMetrics.sectionSpacing) {
                MPMetric(title: "Giá vốn", value: MPFormat.money(position?.costBasis ?? 0, masked: masked), compact: true)
                MPMetric(title: "Bình quân", value: MPFormat.money(position?.averageCost ?? 0, masked: masked), compact: true)
                MPMetric(title: "Đã thực hiện", value: MPFormat.signedMoney(position?.realizedPnL ?? 0, masked: masked), compact: true)
                MPMetric(title: "Giá trị hiện tại", value: position?.marketValue.map { MPFormat.money($0, masked: masked) } ?? "Chưa có giá", compact: true)
                MPMetric(title: "Lãi/lỗ chưa thực hiện", value: position?.unrealizedPnL.map { MPFormat.signedMoney($0, masked: masked) } ?? "Chưa có giá", compact: true)
            }
            MPButton("Cập nhật giá", style: .secondary) { error = nil; priceTarget = asset }
                .accessibilityIdentifier("portfolio-price-update-\(asset.id)")
            if assetTrades.isEmpty {
                MPButton("Xóa tài sản", style: .destructive) { error = nil; deleteTarget = asset }
                    .accessibilityIdentifier("portfolio-delete-\(asset.id)")
            }
            VStack(alignment: .leading, spacing: 8) {
                MPText("Lịch sử mua bán", style: .emphasis)
                if assetTrades.isEmpty {
                    MPText("Chưa có giao dịch mua bán.", style: .secondary)
                } else {
                    ForEach(assetTrades) { trade in
                        MPTransactionRow(
                            title: "\(Self.tradeSideLabel(trade.side)) · \(trade.quantity)",
                            subtitle: Self.tradeSubtitle(trade, timeZone: TimeZone(identifier: accountTimezone) ?? .current, masked: masked),
                            amount: MPFormat.money(trade.unitPrice, masked: masked),
                            isIncome: trade.side == "sell"
                        )
                    }
                }
            }
        }
    }

    static func tradesForAsset(_ trades: [PortfolioTradeModel], assetID: String) -> [PortfolioTradeModel] {
        trades
            .filter { $0.assetID == assetID }
            .sorted { $0.occurredAt > $1.occurredAt }
    }

    static func tradeSideLabel(_ side: String) -> String {
        switch side {
        case "buy": return "Mua"
        case "sell": return "Bán"
        default: return "Giao dịch"
        }
    }

    static func tradeSubtitle(_ trade: PortfolioTradeModel, timeZone: TimeZone, masked: Bool = false) -> String {
        let date = MPFormat.shortDateTime(trade.occurredAt, timeZone: timeZone)
        let fee = MPFormat.money(trade.fee, masked: masked)
        if let note = trade.note?.trimmingCharacters(in: .whitespacesAndNewlines), !note.isEmpty {
            return "Giá \(MPFormat.money(trade.unitPrice, masked: masked)) · Phí \(fee) · \(date) · \(note)"
        }
        return "Giá \(MPFormat.money(trade.unitPrice, masked: masked)) · Phí \(fee) · \(date)"
    }

    private func load() async {
        do {
            async let summaryResult = api.portfolioSummary()
            async let profile = api.profile()
            let loadedSummary = try await summaryResult
            var loadedTrades: [PortfolioTradeModel] = []
            for asset in loadedSummary.assets {
                loadedTrades.append(contentsOf: try await api.portfolioTrades(assetID: asset.id))
            }
            summary = loadedSummary
            tradeHistory = loadedTrades
            if let profile = try? await profile { accountTimezone = profile.timezone }
            self.error = nil
        } catch { self.error = "Không thể tải danh mục đầu tư." }
    }
    private var selectedTradeAsset: PortfolioAssetModel? {
        summary?.assets.first(where: { $0.id == tradeAssetID })
    }
    private func createAsset() async {
        guard !saving else { return }
        guard let draft = PortfolioAssetDraftValidator.normalized(symbol: symbol, name: assetName) else { error = "Nhập mã và tên tài sản."; return }
        saving = true; error = nil
        defer { saving = false }
        do {
            let created = try await api.createPortfolioAsset(symbol: draft.symbol, name: draft.name)
            if tradeAssetID.isEmpty { tradeAssetID = created.id }
            symbol = ""
            assetName = ""
            assetEditorPresented = false
            await load()
        }
        catch { self.error = "Không thể tạo tài sản." }
    }
    private func updatePrice(_ asset: PortfolioAssetModel) async {
        guard !saving else { return }
        let draft = (priceByAsset[asset.id] ?? asset.latestPrice.map(String.init) ?? "").trimmingCharacters(in: .whitespacesAndNewlines)
        let parsed: Int64?
        if draft.isEmpty {
            parsed = nil
        } else if let value = MPFormat.parseNonNegativeInteger(draft) {
            parsed = value
        } else {
            error = "Giá hiện tại phải là số nguyên không âm hoặc để trống."
            return
        }
        saving = true; error = nil
        defer { saving = false }
        do { _ = try await api.updatePortfolioPrice(assetID: asset.id, latestPrice: parsed); priceTarget = nil; await load() }
        catch { self.error = "Không thể cập nhật giá." }
    }
    private func deleteAsset(_ asset: PortfolioAssetModel) async {
        guard !saving else { return }
        saving = true
        error = nil
        defer { saving = false }
        do {
            try await api.deletePortfolioAsset(assetID: asset.id)
            await load()
        } catch {
            self.error = "Không thể xóa tài sản. Tài sản đã có giao dịch sẽ được giữ nguyên."
        }
    }
    private func saveTrade() async {
        guard !saving else { return }
        guard let asset = selectedTradeAsset else {
            error = "Hãy chọn tài sản."
            return
        }
        guard let normalizedQuantity = MPNumberInputKind.decimal.normalized(tradeQuantity),
              PortfolioTradeDraftValidator.isValid(side: tradeSide, quantity: normalizedQuantity, unitPrice: tradePrice, fee: tradeFee),
              let parsedPrice = PortfolioTradeDraftValidator.nonNegativeInteger(tradePrice),
              let parsedFee = PortfolioTradeDraftValidator.nonNegativeInteger(tradeFee, allowingEmpty: true) else {
            error = "Nhập loại, khối lượng, giá và phí hợp lệ. Khối lượng tối đa 8 chữ số thập phân."
            return
        }
        saving = true; error = nil
        defer { saving = false }
        do {
            _ = try await api.createPortfolioTrade(assetID: asset.id, input: PortfolioTradeInputModel(side: tradeSide, quantity: normalizedQuantity, unitPrice: parsedPrice, fee: parsedFee, occurredAt: Self.accountLocalNoonTimestamp(tradeDate, timezone: accountTimezone), note: nil))
            tradeQuantity = ""
            tradePrice = ""
            tradeFee = ""
            tradeEditorPresented = false
            await load()
        } catch { self.error = "Không thể ghi nhận giao dịch danh mục." }
    }

    static func accountLocalNoonTimestamp(_ date: Date, timezone: String) -> String {
        let zone = TimeZone(identifier: timezone) ?? .current
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = zone
        let components = calendar.dateComponents([.year, .month, .day], from: date)
        let localNoon = calendar.date(from: DateComponents(year: components.year, month: components.month, day: components.day, hour: 12)) ?? date
        return MPFormat.instantString(localNoon)
    }
}

enum PortfolioAssetDraftValidator {
    static func normalized(symbol: String, name: String) -> (symbol: String, name: String)? {
        let normalizedSymbol = symbol.trimmingCharacters(in: .whitespacesAndNewlines)
        let normalizedName = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !normalizedSymbol.isEmpty, !normalizedName.isEmpty else { return nil }
        return (normalizedSymbol, normalizedName)
    }
}

enum PortfolioTradeDraftValidator {
    static func nonNegativeInteger(_ raw: String, allowingEmpty: Bool = false) -> Int64? {
        MPFormat.parseNonNegativeInteger(raw, allowingEmpty: allowingEmpty)
    }

    static func isValid(side: String, quantity: String, unitPrice: String, fee: String) -> Bool {
        guard side == "buy" || side == "sell" else { return false }
        guard isPositiveQuantity(quantity), nonNegativeInteger(unitPrice) != nil,
              nonNegativeInteger(fee, allowingEmpty: true) != nil else { return false }
        return true
    }

    private static func isPositiveQuantity(_ raw: String) -> Bool {
        guard let value = MPNumberInputKind.decimal.normalized(raw), !value.isEmpty else { return false }
        let parts = value.split(separator: ".", omittingEmptySubsequences: false)
        guard parts.count <= 2, let whole = parts.first, !whole.isEmpty,
              whole.allSatisfy(\.isNumber) else { return false }
        if parts.count == 2 {
            let fraction = parts[1]
            guard !fraction.isEmpty, fraction.count <= 8, fraction.allSatisfy(\.isNumber) else { return false }
        }
        return value.filter(\.isNumber).contains(where: { $0 != "0" })
    }
}

enum CreditLedgerCopy {
    static let emptyStatement = "Chưa có giao dịch tín dụng."
    static let addEntryAction = "Thêm giao dịch"
    static let addEntryActionAccessibilityID = "credit-add-entry"
    static let paymentAction = "Thanh toán thẻ"
    static let paymentActionAccessibilityID = "credit-add-payment"
}

enum CreditLedgerReferenceSection: String, CaseIterable {
    case header
    case walletSelector
    case summary
    case actions
    case statement

    static let stitchOrder: [CreditLedgerReferenceSection] = [
        .header,
        .walletSelector,
        .summary,
        .actions,
        .statement
    ]
}

struct CreditView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    let preferredWalletID: String?
    @Environment(\.mpFinancialValuesMasked) private var masked
    @State private var wallets: [WalletModel] = []
    @State private var sourceWallets: [WalletModel] = []
    @State private var categories: [CategoryModel] = []
    @State private var selectedID = ""
    @State private var statement: CreditStatementModel?
    @State private var error: String?
    @State private var action: String?
    @State private var isChoosingEntryCategory = false
    @State private var saving = false
    @State private var entryKind = "purchase"
    @State private var categoryID = ""
    @State private var entryAmount = ""
    @State private var sourceWalletID = ""
    @State private var note = ""
    @State private var accountTimezone = TimeZone.current.identifier

    init(api: FinanceAPI, onBack: @escaping () -> Void, preferredWalletID: String? = nil) {
        self.api = api
        self.onBack = onBack
        self.preferredWalletID = preferredWalletID
        _selectedID = State(initialValue: preferredWalletID ?? "")
    }

    var body: some View {
        ScrollView { VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: "Sổ tín dụng", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
            if wallets.isEmpty { MPEmptyView(title: "Chưa có ví tín dụng", detail: "Tạo ví loại tín dụng trong mục Ví.") }
            else {
                MPOptionSelector(title: "Ví tín dụng", options: wallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: Binding(get: { selectedID }, set: { selectedID = $0; categoryID = ""; Task { await loadStatement() } }))
                if let statement {
                    MPSurfaceCard(style: .reference) {
                        VStack(alignment: .leading, spacing: 12) {
                            HStack(spacing: 12) {
                                creditMetric(label: "Dư nợ hiện tại", value: MPFormat.money(abs(statement.balance), masked: masked))
                                creditMetric(label: "Còn khả dụng", value: MPFormat.money(statement.availableCredit, masked: masked), tone: statement.availableCredit < 0 ? .danger : .action)
                            }
                            Divider().overlay(MPColor.line)
                            HStack {
                                MPText("Hạn mức", style: .secondary)
                                Spacer()
                                MPText(MPFormat.money(statement.creditLimit, masked: masked), style: .emphasis)
                            }
                            if statement.paymentStatus == "not_configured" {
                                MPText("Chưa cấu hình dư nợ sao kê và ngày đến hạn.", style: .secondary)
                            } else {
                                HStack {
                                    MPText("Kỳ này cần trả", style: .secondary)
                                    Spacer()
                                    MPText(MPFormat.money(statement.amountDue, masked: masked), style: .emphasis)
                                }
                                MPText(
                                    paymentStatusText(statement),
                                    style: statement.paymentStatus == "overdue" ? .danger : .secondary
                                )
                            }
                        }
                    }
                    HStack {
                        MPButton(CreditLedgerCopy.addEntryAction, style: .secondary) { beginAction("entry") }
                            .accessibilityIdentifier(CreditLedgerCopy.addEntryActionAccessibilityID)
                        MPButton(CreditLedgerCopy.paymentAction) { beginAction("payment") }
                            .accessibilityIdentifier(CreditLedgerCopy.paymentActionAccessibilityID)
                    }
                    if statement.items.isEmpty {
                        MPText("Giao dịch tín dụng", style: .title)
                        MPText(CreditLedgerCopy.emptyStatement, style: .secondary)
                            .accessibilityIdentifier("credit-empty-statement")
                    } else {
                        MPAssistantCard(title: "Giao dịch tín dụng") {
                            ForEach(statement.items) { item in
                                MPTransactionRow(transaction: item, category: categories.first(where: { $0.id == item.categoryID }), timeZone: TimeZone(identifier: accountTimezone) ?? .current, masked: masked)
                            }
                        }
                    }
                }
            }
            if let error { MPStatusMessage(text: error, tone: .danger) }
        }.padding(MPMetrics.pagePadding) }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await loadWallets() }
        .onChange(of: entryKind) { _, _ in
            if !categoriesForEntry.contains(where: { $0.id == categoryID }) { categoryID = categoriesForEntry.first?.id ?? "" }
        }
        .mpFormBottomSheet(
            isPresented: Binding(get: { action != nil }, set: { if !$0 { cancelAction() } }),
            title: isChoosingEntryCategory ? "Chọn nhóm" : (action == "payment" ? "Thanh toán thẻ" : "Thêm giao dịch tín dụng"),
            accessibilityIdentifier: "credit-action-sheet",
            isDismissDisabled: saving,
            onClose: cancelAction,
            fields: {
                if isChoosingEntryCategory {
                    MPCategoryPicker(
                        categories: categories,
                        kind: entryCategoryKind,
                        selectableIDs: Set(categoriesForEntry.map(\.id)),
                        selectedID: categoryID,
                        clearLabel: nil,
                        clearID: nil,
                        onSelect: { categoryID = $0; isChoosingEntryCategory = false }
                    )
                } else {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 12) {
                            if let error { MPStatusMessage(text: error, tone: .danger) }
                            if action == "entry" {
                                MPWalletSelector(title: "Loại giao dịch", options: ["purchase", "refund", "fee", "interest"], selection: $entryKind)
                                MPFormRow(
                                    title: "Nhóm",
                                    iconName: "tag",
                                    value: categoriesForEntry.first(where: { $0.id == categoryID })?.name ?? "Chọn nhóm",
                                    accessibilityIdentifier: "credit-entry-category"
                                ) { isChoosingEntryCategory = true }
                            } else {
                                MPOptionSelector(title: "Ví thanh toán", options: sourceWallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $sourceWalletID, emptyLabel: "Chọn ví thanh toán")
                            }
                            MPAmountField(title: "Số tiền", text: $entryAmount)
                            MPTextField(title: "Ghi chú", text: $note, placeholder: "Không bắt buộc")
                        }
                    }
                }
            },
            footer: {
                if isChoosingEntryCategory {
                    MPButton("Quay lại", style: .secondary, isDisabled: saving) { isChoosingEntryCategory = false }
                } else {
                    MPButton("Lưu", isLoading: saving, isDisabled: saveIsDisabled) { Task { await saveAction() } }
                }
            }
        )
    }
    @ViewBuilder
    private func creditMetric(label: String, value: String, tone: MPTextStyle = .emphasis) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            MPText(label, style: .secondary)
            MPText(value, style: tone)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private func paymentStatusText(_ statement: CreditStatementModel) -> String {
        let status = MPSelectionLabels.display(for: statement.paymentStatus)
        guard let paymentDueAt = statement.paymentDueAt else { return status }
        return "\(status) · hạn \(MPFormat.shortDate(paymentDueAt, timeZone: TimeZone(identifier: accountTimezone) ?? .current))"
    }

    private var categoriesForEntry: [CategoryModel] {
        let transactionType = entryKind == "refund" ? "income" : "expense"
        return categories.filter { $0.applies(to: transactionType, walletID: selectedID, walletType: "credit") }
    }

    private var entryCategoryKind: String { entryKind == "refund" ? "income" : "expense" }

    private func creditItemSubtitle(_ item: TransactionModel) -> String {
        if let note = item.note?.trimmingCharacters(in: .whitespacesAndNewlines), !note.isEmpty { return note }
        return MPFormat.shortDate(item.occurredAt, timeZone: TimeZone(identifier: accountTimezone) ?? .current)
    }

    private func beginAction(_ value: String) {
        action = value
        isChoosingEntryCategory = false
        error = nil
        entryAmount = ""
        note = ""
        if value == "entry" { categoryID = categoriesForEntry.first?.id ?? "" }
    }

    private func cancelAction() {
        action = nil
        isChoosingEntryCategory = false
        error = nil
        entryAmount = ""
        note = ""
        categoryID = ""
    }

    private func loadWallets() async {
        do {
            let all = try await api.wallets()
            wallets = all.filter { $0.type == "credit" }
            sourceWallets = all.filter { $0.type != "credit" }
            sourceWalletID = sourceWallets.first?.id ?? ""
            selectedID = preferredWalletID.flatMap { id in
                wallets.contains(where: { $0.id == id }) ? id : nil
            } ?? wallets.first?.id ?? ""
        } catch {
            self.error = "Không thể tải ví tín dụng."
            return
        }

        do {
            categories = try await api.categories()
        } catch {
            // Keep the statement readable if catalog metadata is unavailable;
            // entry creation remains blocked until an applicable category is loaded.
            categories = []
        }

        if let profile = try? await api.profile() {
            accountTimezone = profile.timezone
        }
        await loadStatement()
    }
    private func loadStatement() async {
        guard !selectedID.isEmpty else { return }
        do {
            statement = try await api.creditStatement(walletID: selectedID)
            self.error = nil
        } catch {
            self.error = "Không thể tải sổ tín dụng."
        }
    }
    private func saveAction() async {
        guard action != nil, !saving else { return }
        saving = true
        defer { saving = false }
        if action == "entry" { await addEntry() } else { await pay() }
    }
    private var saveIsDisabled: Bool {
        guard MPFormat.parsePositiveInteger(entryAmount) != nil else { return true }
        if action == "entry" { return !categoriesForEntry.contains(where: { $0.id == categoryID }) }
        if action == "payment" { return sourceWallets.contains(where: { $0.id == sourceWalletID }) == false }
        return true
    }
    private func addEntry() async { guard !selectedID.isEmpty, let amount = MPFormat.parsePositiveInteger(entryAmount) else { error = "Chọn ví tín dụng và nhập số tiền."; return }; guard let category = categoriesForEntry.first(where: { $0.id == categoryID }) else { error = "Hãy chọn nhóm cho bút toán tín dụng."; return }; do { _ = try await api.createCreditEntry(walletID: selectedID, input: CreditEntryInputModel(kind: entryKind, categoryID: category.id, amount: amount, occurredAt: MPFormat.instantString(Date()), note: note.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? nil : note.trimmingCharacters(in: .whitespacesAndNewlines))); cancelAction(); await loadStatement() } catch { self.error = "Không thể ghi nhận bút toán tín dụng." } }
    private func pay() async { guard !selectedID.isEmpty, let amount = MPFormat.parsePositiveInteger(entryAmount), !sourceWalletID.isEmpty else { error = "Chọn nguồn thanh toán và nhập số tiền."; return }; do { _ = try await api.createCreditPayment(walletID: selectedID, input: CreditPaymentInputModel(sourceWalletID: sourceWalletID, amount: amount, occurredAt: MPFormat.instantString(Date()), note: note.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? nil : note.trimmingCharacters(in: .whitespacesAndNewlines))); cancelAction(); await loadStatement() } catch { self.error = "Không thể tạo thanh toán tín dụng." } }
}

struct NotificationsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    let timeZone: TimeZone
    @State private var items: [NotificationModel] = []
    @State private var loading = true
    @State private var error: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Thông báo", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                if loading && items.isEmpty {
                    MPLoadingView(label: "Đang tải thông báo...")
                } else if let error, items.isEmpty {
                    MPErrorView(title: "Không tải được thông báo", detail: error) { Task { await load() } }
                } else if items.isEmpty {
                    MPEmptyView(title: "Chưa có thông báo", detail: "Thông báo trong ứng dụng sẽ xuất hiện tại đây.")
                } else {
                    ForEach(items) { item in
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 8) {
                                HStack {
                                    MPText(item.title, style: .emphasis)
                                    Spacer()
                                    if item.readAt == nil { Circle().fill(MPColor.action).frame(width: 8, height: 8).accessibilityHidden(true) }
                                }
                                MPText(item.body, style: .secondary)
                                MPText(MPFormat.shortDate(item.createdAt, timeZone: timeZone), style: .secondary)
                                if item.readAt == nil {
                                    MPButton("Đánh dấu đã đọc", style: .secondary) { Task { await markRead(item) } }
                                        .accessibilityIdentifier("notification-mark-read-\(item.id)")
                                }
                            }
                        }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .refreshable { await load() }
    }

    private func load() async {
        do { items = try await api.notifications(); error = nil }
        catch { self.error = "Không thể tải thông báo." }
        loading = false
    }

    private func markRead(_ item: NotificationModel) async {
        do {
            try await api.markNotificationRead(id: item.id)
            if let index = items.firstIndex(where: { $0.id == item.id }) {
                let current = items[index]
                items[index] = NotificationModel(id: current.id, kind: current.kind, title: current.title, body: current.body, readAt: current.readAt ?? MPFormat.instantString(Date()), createdAt: current.createdAt)
            }
        } catch { self.error = "Không thể cập nhật thông báo." }
    }
}
