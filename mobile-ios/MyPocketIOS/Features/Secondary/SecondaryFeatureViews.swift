import SwiftUI

struct BudgetsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @State private var summary: BudgetSummaryModel?
    @State private var loading = false
    @State private var error: String?
    @State private var name = ""
    @State private var limit = ""
    @State private var start = MPFormat.currentMonth + "-01"
    @State private var end = MPFormat.currentMonth + "-28"
    @State private var wallets: [WalletModel] = []
    @State private var categories: [CategoryModel] = []
    @State private var walletScopeID = "__all__"
    @State private var categoryScopeID = "__all__"
    @State private var showEditor = false
    @State private var editTarget: BudgetModel?
    @State private var deleteTarget: BudgetModel?
    @State private var accountTimezone = TimeZone.current.identifier

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Ngân sách", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { MPIconButton(systemName: "plus", label: "Tạo ngân sách") { showEditor = true } })
                if loading && summary == nil { MPLoadingView(label: "Đang tải ngân sách...") }
                else if let error, summary == nil { MPErrorView(title: "Không tải được ngân sách", detail: error) { Task { await load() } } }
                else if let summary {
                    LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                        MPMetricCard(title: "Đã chi", value: MPFormat.money(summary.spent))
                        MPMetricCard(title: "Giới hạn", value: MPFormat.money(summary.limitAmount))
                    }
                    if summary.items.isEmpty { MPEmptyView(title: "Chưa có ngân sách", detail: "Tạo ngân sách theo ví hoặc danh mục.") }
                    ForEach(summary.items) { budget in
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 8) {
                                MPText(budget.name)
                                MPProgressBar(progress: Double(budget.spent) / Double(max(1, budget.limitAmount)), label: "\(MPFormat.money(budget.spent)) / \(MPFormat.money(budget.limitAmount))")
                                MPText("\(budget.startDate) → \(budget.endDate)", style: .secondary)
                                HStack {
                                    if !budget.ended {
                                        MPButton("Sửa", style: .secondary) { beginEditing(budget) }
                                    } else {
                                        MPText("Đã kết thúc", style: .secondary)
                                    }
                                    MPButton("Xóa", style: .destructive) { deleteTarget = budget }
                                }
                            }
                        }
                    }
                }
                if showEditor {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 12) {
                            MPText(editTarget == nil ? "Tạo ngân sách" : "Sửa ngân sách", style: .title)
                            MPTextField(title: "Tên", text: $name, placeholder: "Ăn uống tháng này")
                            MPAmountField(title: "Giới hạn", text: $limit)
                            MPTextField(title: "Ngày bắt đầu (YYYY-MM-DD)", text: $start)
                            MPTextField(title: "Ngày kết thúc (YYYY-MM-DD)", text: $end)
                            MPOptionSelector(title: "Phạm vi ví", options: [MPSelectionOption(id: "__all__", label: "Tất cả ví")] + wallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $walletScopeID, emptyLabel: "Tất cả ví")
                            MPOptionSelector(title: "Phạm vi danh mục", options: [MPSelectionOption(id: "__all__", label: "Tất cả danh mục")] + categories.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $categoryScopeID, emptyLabel: "Tất cả danh mục")
                            HStack { MPButton("Hủy", style: .secondary) { showEditor = false; editTarget = nil }; MPButton("Lưu") { Task { await create() } } }
                        }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await load() }
        .mpConfirmDialog(isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }), title: "Xóa ngân sách?", message: "Các giao dịch không bị xóa.", confirmTitle: "Xóa") {
            if let target = deleteTarget { Task { do { try await api.deleteBudget(id: target.id); await load() } catch { self.error = "Không thể xóa ngân sách." } } }
            deleteTarget = nil
        }
        .accessibilityIdentifier("budgets-screen")
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
                        start = range.start
                        end = range.end
                    }
                }
            }
            self.error = nil
        } catch { self.error = "Không thể tải ngân sách." }
    }

    private func create() async {
        guard let parsed = MPFormat.parsePositiveInteger(limit), !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { error = "Nhập tên và giới hạn hợp lệ."; return }
        guard let startDate = MPFormat.parseCalendarDate(start), let endDate = MPFormat.parseCalendarDate(end), startDate <= endDate else { error = "Khoảng thời gian ngân sách không hợp lệ."; return }
        do {
            let input = BudgetInputModel(name: name.trimmingCharacters(in: .whitespacesAndNewlines), limitAmount: parsed, walletID: walletScopeID == "__all__" ? nil : walletScopeID, categoryID: categoryScopeID == "__all__" ? nil : categoryScopeID, startDate: start, endDate: end)
            if let editTarget { _ = try await api.updateBudget(id: editTarget.id, input: input) } else { _ = try await api.createBudget(input) }
            showEditor = false; self.editTarget = nil; await load()
        }
        catch { self.error = "Không thể tạo ngân sách." }
    }

    private func beginEditing(_ budget: BudgetModel) {
        guard !budget.ended else { return }
        editTarget = budget
        name = budget.name
        limit = String(budget.limitAmount)
        start = budget.startDate
        end = budget.endDate
        walletScopeID = budget.walletID ?? "__all__"
        categoryScopeID = budget.categoryID ?? "__all__"
        showEditor = true
    }
}

struct ReportsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @State private var month = MPFormat.currentMonth
    @State private var summary: InsiderSummaryModel?
    @State private var monthDetail: MonthSummaryModel?
    @State private var loading = false
    @State private var error: String?
    @State private var accountTimezone = TimeZone.current.identifier

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Báo cáo", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                MPTextField(title: "Tháng (YYYY-MM)", text: $month)
                MPButton("Xem báo cáo", isLoading: loading) { Task { await load() } }
                if let summary {
                    LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                        MPMetricCard(title: "Thu nhập", value: MPFormat.money(summary.income))
                        MPMetricCard(title: "Chi tiêu", value: MPFormat.money(summary.expense))
                        MPMetricCard(title: "Còn lại", value: MPFormat.money(summary.net))
                        MPMetricCard(title: "Trung bình/ngày", value: MPFormat.money(summary.averageDailyExpense))
                        MPMetricCard(
                            title: "So với tháng trước",
                            value: MPFormat.signedMoney(summary.expenseDelta),
                            detail: summary.expenseChangeBPS.map { MPFormat.percentageFromBPS($0) }
                        )
                        MPMetricCard(
                            title: "Tỷ lệ chi / thu",
                            value: MPFormat.percentageFromBPS(summary.spendingIncomeRatioBPS),
                            detail: summary.spendingIncomeRatioBPS == nil ? "Chưa có thu nhập để tính tỷ lệ" : nil
                        )
                    }
                    MPAssistantCard(title: "Theo danh mục") {
                        MPChart(values: summary.topCategories.map { ($0.name, min(180, max(4, Double($0.amount) / 100_000))) })
                        ForEach(summary.topCategories, id: \.name) { item in MPText("\(item.name): \(MPFormat.money(item.amount))", style: .secondary) }
                    }
                    if let monthDetail {
                        MPAssistantCard(title: "Theo nhóm trong sổ tháng") {
                            if monthDetail.categories.isEmpty {
                                MPText("Chưa có giao dịch được tính vào báo cáo.", style: .secondary)
                            } else {
                                ForEach(monthDetail.categories, id: \.stableID) { item in
                                    HStack {
                                        VStack(alignment: .leading, spacing: 3) {
                                            MPText(item.name)
                                            MPText("\(item.type == "income" ? "Thu" : "Chi") · \(item.count) giao dịch", style: .secondary)
                                        }
                                        Spacer()
                                        MPText(MPFormat.money(item.amount), style: .secondary)
                                    }
                                }
                            }
                        }
                        MPAssistantCard(title: "Hoạt động trong tháng") {
                            MPText("\(monthDetail.transactionCount) giao dịch được tính vào báo cáo. Chuyển ví và điều chỉnh số dư không làm tăng thu/chi.", style: .secondary)
                            MPText("Số liệu được tính lại từ sổ giao dịch theo múi giờ tài khoản.", style: .secondary)
                        }
                    }
                    MPAssistantCard(title: "Khoản chi lớn") {
                        ForEach(summary.topExpenses) { item in MPTransactionRow(title: item.categoryName ?? "Chi tiêu", subtitle: MPFormat.shortDate(item.occurredAt, timeZone: TimeZone(identifier: accountTimezone) ?? .current), amount: MPFormat.money(item.amount), isIncome: false) }
                    }
                } else if let error { MPErrorView(title: "Không tải được báo cáo", detail: error) { Task { await load() } } }
                else { MPEmptyView(title: "Chọn tháng để xem báo cáo", detail: "Số liệu luôn được tính theo timezone tài khoản.") }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await load() }
        .accessibilityIdentifier("reports-screen")
    }

    private func load() async {
        guard !loading else { return }
        loading = true; defer { loading = false }
        do {
            if let profile = try? await api.profile() {
                accountTimezone = profile.timezone
                if summary == nil { month = MPFormat.currentMonth(timeZone: TimeZone(identifier: accountTimezone) ?? .current) }
            }
            async let insiderResult = api.insider(month: month)
            async let monthResult = api.monthSummary(month: month)
            summary = try await insiderResult
            monthDetail = try? await monthResult
            self.error = nil
        } catch { self.error = "Không thể tải báo cáo." }
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
    @State private var adding = false
    @State private var editTarget: CategoryModel?
    @State private var deleteTarget: CategoryModel?
    @State private var wallets: [WalletModel] = []
    @State private var selectedWalletIDs: Set<String> = []
    @State private var parentID = ""

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Danh mục", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { MPIconButton(systemName: "plus", label: "Thêm danh mục") { adding = true } })
                if loading && items.isEmpty { MPLoadingView(label: "Đang tải danh mục...") }
                else if let error, items.isEmpty { MPErrorView(title: "Không tải được danh mục", detail: error) { Task { await load() } } }
                else {
                    MPSurfaceCard {
                        VStack(spacing: 0) {
                            categoryTree(parentID: nil, depth: 0)
                        }
                    }
                }
                if adding {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 10) {
                            MPText(editTarget.map { CategoryEditPolicy.metadataIsReadOnly($0) ? "Ví áp dụng" : "Sửa danh mục" } ?? "Danh mục mới", style: .title)
                            if let editTarget, CategoryEditPolicy.metadataIsReadOnly(editTarget) {
                                MPSurfaceCard {
                                    VStack(alignment: .leading, spacing: 6) {
                                        MPText(editTarget.name, style: .title)
                                        MPText("\(MPSelectionLabels.display(for: editTarget.kind)) · Nhóm hệ thống", style: .secondary)
                                        MPText("Chỉ có thể thay đổi các ví áp dụng.", style: .secondary)
                                    }
                                }
                            } else {
                                MPTextField(title: "Tên", text: $name)
                                MPWalletSelector(title: "Loại", options: ["expense", "income", "debt"], selection: $kind)
                                MPOptionSelector(title: "Danh mục cha", options: availableParentOptions, selectionID: $parentID, emptyLabel: "Không có danh mục cha")
                            }
                            MPMultiOptionSelector(title: "Áp dụng cho ví", options: wallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionIDs: $selectedWalletIDs, allLabel: "Tất cả ví")
                            HStack { MPButton("Hủy", style: .secondary) { adding = false; editTarget = nil }; MPButton("Lưu") { Task { await create() } } }
                        }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .mpConfirmDialog(isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }), title: "Xóa danh mục?", message: "Danh mục con sẽ được xử lý theo chính sách máy chủ.", confirmTitle: "Xóa") {
            if let target = deleteTarget { Task { do { try await api.deleteCategory(id: target.id); await load() } catch { self.error = "Không thể xóa danh mục." } } }
            deleteTarget = nil
        }
        .accessibilityIdentifier("categories-screen")
    }

    private func categoryRow(_ item: CategoryModel, depth: Int) -> some View {
        HStack {
            Image(systemName: depth == 0 ? "folder.fill" : "arrow.turn.down.right")
                .foregroundStyle(MPColor.action)
            MPText(item.name)
            Spacer()
            MPText(item.kind, style: .secondary)
            MPIconButton(systemName: "pencil", label: "Sửa \(item.name)") { beginEditing(item) }
            if CategoryEditPolicy.canDelete(item) {
                MPIconButton(systemName: "trash", label: "Xóa \(item.name)") { deleteTarget = item }
            } else {
                Image(systemName: "lock.fill")
                    .foregroundStyle(MPColor.secondary)
                    .accessibilityLabel("Nhóm hệ thống")
            }
        }
        .padding(.leading, CGFloat(depth * 20))
        .padding(.vertical, 8)
    }
    private func categoryTree(parentID: String?, depth: Int) -> AnyView {
        AnyView(ForEach(items.filter { $0.parentID == parentID }) { item in
            categoryRow(item, depth: depth)
            categoryTree(parentID: item.id, depth: depth + 1)
        })
    }
    private var availableParentOptions: [MPSelectionOption] {
        let excluded = excludedParentIDs()
        return [MPSelectionOption(id: "", label: "Không có danh mục cha")] + items.filter { !excluded.contains($0.id) }.map { MPSelectionOption(id: $0.id, label: $0.name) }
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
    private func create() async {
        if let editTarget, CategoryEditPolicy.metadataIsReadOnly(editTarget) {
            do {
                _ = try await api.updateCategoryWallets(id: editTarget.id, walletIDs: Array(selectedWalletIDs))
                adding = false; self.editTarget = nil; selectedWalletIDs = []; await load()
            } catch { self.error = "Không thể lưu ví áp dụng." }
            return
        }
        guard !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { self.error = "Nhập tên danh mục."; return }
        do {
            let input = CategoryModelInput(name: name.trimmingCharacters(in: .whitespacesAndNewlines), kind: kind, parentID: parentID.isEmpty ? nil : parentID, walletIDs: Array(selectedWalletIDs).sorted(), iconKey: editTarget?.iconKey ?? "tag")
            if let editTarget { _ = try await api.updateCategory(id: editTarget.id, input: input) } else { _ = try await api.createCategory(input) }
            adding = false; name = ""; parentID = ""; selectedWalletIDs = []; self.editTarget = nil; await load()
        } catch { self.error = "Không thể lưu danh mục." }
    }
    private func beginEditing(_ item: CategoryModel) { editTarget = item; name = item.name; kind = item.kind; parentID = item.parentID ?? ""; selectedWalletIDs = Set(item.walletIDs); adding = true }
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
    @State private var summary: JarMonthSummaryModel?
    @State private var month = MPFormat.currentMonth
    @State private var loading = false
    @State private var error: String?
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
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Hũ chi tiêu", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { MPIconButton(systemName: "plus", label: "Thêm hũ") { editTarget = nil; resetEditor(); showEditor = true } })
                MPOptionSelector(title: "Tháng", options: MPFormat.monthOptions(around: month, timeZone: TimeZone(identifier: accountTimezone) ?? .current).map { MPSelectionOption(id: $0, label: MPFormat.monthLabel($0)) }, selectionID: $month)
                if loading { MPLoadingView(label: "Đang tải hũ tháng này...") }
                if let summary {
                    LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                        MPMetricCard(title: "Thu thực tế", value: MPFormat.money(summary.actualIncome))
                        MPMetricCard(title: "Chi thực tế", value: MPFormat.money(summary.totalSpent))
                        MPMetricCard(title: "Đã phân bổ", value: MPFormat.money(summary.totalAllocated))
                        MPMetricCard(title: "Chi chưa gắn hũ", value: MPFormat.money(summary.unassignedSpent))
                    }
                    if summary.overIncome || summary.overOneHundredPercent {
                        MPStatusMessage(text: "Tổng phân bổ vượt thu thực tế. Đây là cảnh báo, không chặn giao dịch.", tone: .danger)
                    }
                    MPText("Hũ tháng", style: .title)
                    if summary.items.isEmpty { MPEmptyView(title: "Chưa có hũ cho tháng này", detail: "Chi tiêu vẫn được ghi nhận bình thường.") }
                    ForEach(summary.items) { item in
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 8) {
                                MPText(item.name)
                                if let allocation = item.calculatedAllocation ?? (item.allocationMode == "fixed" ? item.allocationAmount : nil) {
                                    MPProgressBar(progress: Double(item.spent) / Double(max(1, allocation)), label: "Đã chi \(MPFormat.money(item.spent)) / \(MPFormat.money(allocation))")
                                } else {
                                    MPText("Chưa đặt phân bổ · Đã chi \(MPFormat.money(item.spent))", style: .secondary)
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
                                    MPMetricCard(title: "Chi thực tế", value: MPFormat.money(report.totalSpent))
                                    MPMetricCard(title: "Đã phân bổ", value: MPFormat.money(report.totalAllocated))
                                }
                                MPText("\(MPFormat.monthLabel(report.fromMonth)) → \(MPFormat.monthLabel(report.toMonth)) · Không cộng dồn số dư.", style: .secondary)
                                ForEach(report.months, id: \.month) { item in
                                    HStack { MPText(MPFormat.monthLabel(item.month)); Spacer(); MPText(MPFormat.money(item.spent), style: .secondary) }
                                }
                            }
                        }
                    }
                } else if let error { MPErrorView(title: "Không tải được hũ", detail: error) { Task { await load() } } }
                else if !loading { MPEmptyView(title: "Chưa chọn tháng", detail: "Hũ là lớp ngân sách theo tháng và không thay đổi số dư ví.") }
                if showEditor {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 10) {
                            MPText(editTarget == nil ? "Tạo hũ" : "Sửa hũ", style: .title)
                            MPTextField(title: "Tên", text: $name, placeholder: "Ăn uống")
                            MPWalletSelector(title: "Cách phân bổ", options: ["none", "fixed", "percent"], selection: $mode)
                            if mode == "fixed" { MPAmountField(title: "Số tiền mỗi tháng", text: $amount) }
                            if mode == "percent" { MPTextField(title: "Tỷ lệ (%)", text: $percent, placeholder: "20") }
                            HStack { MPButton("Hủy", style: .secondary) { showEditor = false }; MPButton("Lưu") { Task { await save() } } }
                        }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await load() }
        .onChange(of: month) { _, _ in
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
        .accessibilityIdentifier("jars-screen")
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
    private func beginEditing(_ item: JarItemModel) { editTarget = item; name = item.name; mode = item.allocationMode; amount = item.allocationAmount.map { String($0) } ?? ""; percent = item.allocationPercent.map { String($0) } ?? ""; showEditor = true }
    private func save() async {
        guard !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { error = "Nhập tên hũ."; return }
        let parsedAmount = MPFormat.parseNonNegativeInteger(amount)
        let parsedPercent = Double(percent.replacingOccurrences(of: ",", with: "."))
        guard JarAllocationDraftValidator.isValid(mode: mode, amount: parsedAmount, percent: parsedPercent) else { error = "Cách phân bổ không hợp lệ."; return }
        let input = JarConfigInputModel(month: month, name: name.trimmingCharacters(in: .whitespacesAndNewlines), allocationMode: mode, allocationAmount: mode == "fixed" ? parsedAmount : nil, allocationPercent: mode == "percent" ? parsedPercent : nil)
        do { if let editTarget { _ = try await api.updateJarMonth(jarID: editTarget.jarID, month: month, input: input) } else { _ = try await api.createJar(input) }; showEditor = false; await load() }
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

struct MonthsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    let onOpenJars: () -> Void
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
                            HStack {
                                VStack(alignment: .leading, spacing: 4) {
                                    MPText(MPFormat.monthLabel(summary.month), style: .title)
                                    MPText(summary.isComplete ? "Đã hoàn tất" : "Đang diễn ra", style: .secondary)
                                }
                                Spacer()
                                MPButton("Hũ tháng", style: .secondary, action: onOpenJars)
                            }
                            LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                                MPMetricCard(title: "Thu", value: MPFormat.money(summary.income))
                                MPMetricCard(title: "Chi", value: MPFormat.money(summary.expense))
                                MPMetricCard(title: "Chênh lệch", value: MPFormat.signedMoney(summary.net))
                                MPMetricCard(title: "Giao dịch", value: "\(summary.transactionCount)")
                            }
                            MPText("Kỳ báo cáo: \(MPFormat.shortDate(summary.startAt, timeZone: accountTimezoneValue)) – \(MPFormat.shortDate(summary.nextStartAt, timeZone: accountTimezoneValue))", style: .secondary)
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
                                    MPText(MPFormat.money(item.amount), style: .secondary)
                                }
                            }
                        }
                    }
                    MPAssistantCard(title: "Hũ chi tiêu") {
                        if let jarSummary = summary.jarSummary, !jarSummary.items.isEmpty {
                            ForEach(jarSummary.items) { item in
                                let allocation = item.calculatedAllocation ?? (item.allocationMode == "fixed" ? item.allocationAmount : nil)
                                VStack(alignment: .leading, spacing: 5) {
                                    HStack { MPText(item.name); Spacer(); MPText(MPFormat.money(item.spent), style: .secondary) }
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
        .accessibilityIdentifier("months-screen")
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
        let digits = value.filter(\.isNumber)
        guard !digits.isEmpty, let parsed = Int64(digits), parsed > 0 else { return nil }
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
              positiveInteger(interval) != nil,
              type == "expense" || type == "income",
              ["daily", "weekly", "monthly", "yearly"].contains(frequency) else { return false }
        return !endsAtEnabled || endsAtDate >= nextRunDate
    }
}

struct RecurringView: View {
    let api: FinanceAPI
    let onBack: () -> Void
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
    @State private var interval = "1"
    @State private var active = "Có"
    @State private var note = ""
    @State private var nextRunDate = Date()
    @State private var endsAtDate = Date()
    @State private var endsAtEnabled = "Không"
    @State private var accountTimezone = TimeZone.current.identifier
    @State private var result: String?
    @State private var error: String?
    @State private var deleteTarget: RecurringModel?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Giao dịch định kỳ", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { MPIconButton(systemName: "plus", label: "Thêm lịch") { editTarget = nil; resetEditor(); showEditor = true } })
                MPSurfaceCard {
                    VStack(alignment: .leading, spacing: 8) {
                        MPText("Giao dịch định kỳ", style: .title)
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
                                MPText(item.active ? "Đang bật" : "Tạm dừng", style: .secondary)
                            }
                            MPText("\(item.type == "expense" ? "Khoản chi" : "Khoản thu") · \(frequencyLabel(item.frequency, interval: item.interval)) · \(MPFormat.money(item.amount))", style: .secondary)
                            MPText("Ví: \(walletName(for: item.walletID))", style: .secondary)
                            MPText("Lần chạy kế tiếp: \(MPFormat.shortDateTime(item.nextRunAt, timeZone: TimeZone(identifier: accountTimezone) ?? .current))", style: .secondary)
                            if let note = item.note, !note.isEmpty { MPText(note, style: .secondary) }
                            HStack { MPButton("Sửa", style: .secondary) { beginEditing(item) }; MPButton("Xóa", style: .destructive) { deleteTarget = item } }
                        }
                    }
                }
                if showEditor {
                    MPSurfaceCard { VStack(alignment: .leading, spacing: 10) {
                        MPText(editTarget == nil ? "Tạo lịch" : "Sửa lịch", style: .title)
                        MPTextField(title: "Tên", text: $name)
                        MPAmountField(title: "Số tiền", text: $amount)
                        MPOptionSelector(title: "Ví", options: wallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $walletID)
                        MPWalletSelector(title: "Loại", options: ["expense", "income"], selection: $type)
                        MPOptionSelector(title: "Danh mục", options: [MPSelectionOption(id: "", label: "Không chọn")] + categoriesForType.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $categoryID, emptyLabel: "Không chọn")
                        MPWalletSelector(title: "Tần suất", options: ["daily", "weekly", "monthly", "yearly"], selection: $frequency)
                        MPAmountField(title: "Mỗi bao nhiêu kỳ", text: $interval)
                        MPDateTimeField(title: "Lần chạy kế tiếp", date: $nextRunDate)
                        MPWalletSelector(title: "Có ngày kết thúc", options: ["Có", "Không"], selection: $endsAtEnabled)
                        if endsAtEnabled == "Có" { MPDateTimeField(title: "Ngày kết thúc", date: $endsAtDate) }
                        MPWalletSelector(title: "Trạng thái", options: ["Có", "Không"], selection: $active)
                        MPTextField(title: "Ghi chú", text: $note, placeholder: "Không bắt buộc")
                        HStack { MPButton("Hủy", style: .secondary) { showEditor = false; editTarget = nil }; MPButton("Lưu") { Task { await save() } } }
                    } }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await load() }
        .onChange(of: type) { _, _ in
            if !categoriesForType.contains(where: { $0.id == categoryID }) {
                categoryID = ""
            }
        }
        .mpConfirmDialog(isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }), title: "Xóa lịch định kỳ?", message: "Các giao dịch đã tạo vẫn giữ nguyên trong sổ giao dịch.", confirmTitle: "Xóa") {
            if let target = deleteTarget { Task { await delete(target) } }
            deleteTarget = nil
        }
        .accessibilityIdentifier("recurring-screen")
    }

    private var categoriesForType: [CategoryModel] { categories.filter { $0.kind == type } }
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
    private func load() async { do { async let schedules = api.recurring(); async let walletList = api.wallets(); async let categoryList = api.categories(); async let profile = api.profile(); items = try await schedules; wallets = (try await walletList).filter { $0.type != "credit" }; categories = try await categoryList; if let profile = try? await profile { accountTimezone = profile.timezone }; if walletID.isEmpty { walletID = wallets.first?.id ?? "" }; self.error = nil } catch { self.error = "Không thể tải lịch định kỳ." } }
    private func runDue() async { do { let response = try await api.runDueRecurring(); result = "Đã tạo \(response.created) giao dịch."; await load() } catch { self.error = "Không thể chạy giao dịch đến hạn." } }
    private func resetEditor() { name = ""; amount = ""; walletID = wallets.first?.id ?? ""; frequency = "monthly"; type = "expense"; categoryID = ""; interval = "1"; note = ""; nextRunDate = Date(); endsAtDate = Date(); endsAtEnabled = "Không"; active = "Có" }
    private func beginEditing(_ item: RecurringModel) { editTarget = item; name = item.name; amount = String(item.amount); note = item.note ?? ""; walletID = item.walletID; type = item.type; categoryID = item.categoryID ?? ""; frequency = item.frequency; interval = String(item.interval); nextRunDate = MPFormat.parseInstant(item.nextRunAt) ?? Date(); if let endsAt = item.endsAt, let parsed = MPFormat.parseInstant(endsAt) { endsAtDate = parsed; endsAtEnabled = "Có" } else { endsAtDate = Date(); endsAtEnabled = "Không" }; active = item.active ? "Có" : "Không"; showEditor = true }
    private func save() async {
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
              let parsedInterval = RecurringDraftValidator.positiveInteger(interval).flatMap(Int.init) else {
            self.error = "Nhập đủ tên, ví, chu kỳ và số tiền hợp lệ."
            return
        }
        let trimmedNote = note.trimmingCharacters(in: .whitespacesAndNewlines)
        let input = RecurringInputModel(
            name: name.trimmingCharacters(in: .whitespacesAndNewlines),
            walletID: wallet.id,
            categoryID: categoryID.isEmpty ? nil : categoryID,
            type: type,
            amount: parsed,
            note: trimmedNote.isEmpty ? nil : trimmedNote,
            frequency: frequency,
            interval: parsedInterval,
            nextRunAt: MPFormat.instantString(nextRunDate),
            endsAt: endsAtEnabled == "Có" ? MPFormat.instantString(endsAtDate) : nil,
            active: active == "Có"
        )
        do {
            if let editTarget { _ = try await api.updateRecurring(id: editTarget.id, input: input) }
            else { _ = try await api.createRecurring(input) }
            showEditor = false
            self.editTarget = nil
            await load()
        } catch { self.error = "Không thể lưu lịch định kỳ." }
    }
    private func delete(_ item: RecurringModel) async { do { try await api.deleteRecurring(id: item.id); await load() } catch { self.error = "Không thể xóa lịch định kỳ." } }
}

struct TravelView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @State private var events: [TravelEventModel] = []
    @State private var error: String?
    @State private var name = ""
    @State private var context = ""
    @State private var adding = false
    @State private var editTarget: TravelEventModel?
    @State private var deleteTarget: TravelEventModel?
    @State private var startsOn = ""
    @State private var endsOn = ""

    var body: some View {
        ScrollView { VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: "Travel Mode", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { MPIconButton(systemName: "plus", label: "Thêm chuyến đi") { editTarget = nil; resetEditor(); adding = true } })
            MPSurfaceCard {
                VStack(alignment: .leading, spacing: 8) {
                    MPText("Travel Mode", style: .title)
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
                        }
                        MPText(eventDateLabel(event), style: .secondary)
                        if let context = event.context, !context.isEmpty { MPText(context, style: .secondary) }
                        HStack {
                            MPButton("Sửa", style: .secondary) { beginEditing(event) }
                            MPButton("Xóa", style: .destructive) { deleteTarget = event }
                        }
                    }
                }
            }
            if adding {
                MPSurfaceCard {
                    VStack(alignment: .leading, spacing: 10) {
                        MPText(editTarget == nil ? "Chuyến đi mới" : "Sửa chuyến đi", style: .title)
                        MPText("Chỉ giao dịch thu/chi tạo mới mới tự nhận chuyến đang bật. Giao dịch định kỳ, chuyển ví và điều chỉnh số dư không tự gắn.", style: .secondary)
                        MPTextField(title: "Tên", text: $name)
                        MPTextEditor(title: "Mô tả", text: $context)
                        MPTextField(title: "Ngày bắt đầu (YYYY-MM-DD)", text: $startsOn)
                        MPTextField(title: "Ngày kết thúc (YYYY-MM-DD)", text: $endsOn)
                        HStack { MPButton("Hủy", style: .secondary) { adding = false }; MPButton("Lưu") { Task { await save() } } }
                    }
                }
            }
        }.padding(MPMetrics.pagePadding) }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .mpConfirmDialog(isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }), title: "Xóa chuyến đi?", message: "Giao dịch vẫn được giữ nguyên; chỉ gỡ liên kết chuyến.", confirmTitle: "Xóa") {
            if let target = deleteTarget { Task { await remove(target) } }
            deleteTarget = nil
        }
        .accessibilityIdentifier("travel-screen")
    }
    private func load() async { do { events = try await api.travelEvents(); self.error = nil } catch { self.error = "Không thể tải sự kiện du lịch." } }
    private func toggle(_ event: TravelEventModel) async { do { _ = try await api.setTravelEventActive(id: event.id, active: !event.active); await load() } catch { self.error = "Không thể cập nhật Travel Mode." } }
    private func resetEditor() { name = ""; context = ""; startsOn = ""; endsOn = "" }
    private func beginEditing(_ event: TravelEventModel) { editTarget = event; name = event.name; context = event.context ?? ""; startsOn = event.startsOn ?? ""; endsOn = event.endsOn ?? ""; adding = true }
    private func eventDateLabel(_ event: TravelEventModel) -> String {
        switch (event.startsOn, event.endsOn) {
        case let (start?, end?): return "\(start) → \(end)"
        case let (start?, nil): return "Từ \(start) · Không giới hạn ngày kết thúc"
        case let (nil, end?): return "Đến \(end) · Không giới hạn ngày bắt đầu"
        default: return "Không giới hạn ngày"
        }
    }
    private func save() async { guard TravelEventDraftValidator.isValid(name: name, startsOn: startsOn, endsOn: endsOn) else { self.error = "Tên chuyến và khoảng ngày không hợp lệ."; return }; let trimmedName = name.trimmingCharacters(in: .whitespacesAndNewlines); let trimmedContext = context.trimmingCharacters(in: .whitespacesAndNewlines); let input = TravelEventInputModel(name: trimmedName, context: trimmedContext.isEmpty ? nil : trimmedContext, startsOn: startsOn.isEmpty ? nil : startsOn, endsOn: endsOn.isEmpty ? nil : endsOn, active: editTarget?.active ?? false); do { if let editTarget { _ = try await api.updateTravelEvent(id: editTarget.id, input: input) } else { _ = try await api.createTravelEvent(input) }; resetEditor(); adding = false; await load() } catch { self.error = "Không thể lưu chuyến đi." } }
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
    @State private var summary: PortfolioSummaryModel?
    @State private var error: String?
    @State private var adding = false
    @State private var symbol = ""
    @State private var assetName = ""
    @State private var tradeAsset: PortfolioAssetModel?
    @State private var tradeSide = "buy"
    @State private var tradeQuantity = ""
    @State private var tradePrice = ""
    @State private var tradeFee = ""
    @State private var tradeDate = Date()
    @State private var priceByAsset: [String: String] = [:]
    @State private var accountTimezone = TimeZone.current.identifier

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Danh mục đầu tư", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: {
                    HStack(spacing: 8) {
                        MPIconButton(systemName: "arrow.clockwise", label: "Làm mới") { Task { await load() } }
                        MPIconButton(systemName: "plus", label: "Thêm tài sản") { adding = true }
                    }
                })
                if let summary {
                    LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                        MPMetricCard(title: "Giá vốn", value: MPFormat.money(summary.totalCostBasis))
                        MPMetricCard(title: "Giá trị hiện tại", value: summary.totalMarketValue.map(MPFormat.money) ?? "Chưa có giá")
                        MPMetricCard(title: "Lãi/lỗ chưa thực hiện", value: summary.totalUnrealizedPnL.map(MPFormat.signedMoney) ?? "Chưa có giá")
                    }
                    ForEach(summary.assets) { asset in
                        portfolioAssetCard(asset, position: summary.positions.first(where: { $0.assetID == asset.id }), trades: summary.trades)
                    }
                } else if let error {
                    MPErrorView(title: "Không tải được đầu tư", detail: error) { Task { await load() } }
                } else {
                    MPEmptyView(title: "Chưa có danh mục đầu tư", detail: "Giá chưa biết được hiển thị là chưa định giá, không phải bằng 0.")
                }
                if adding {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 10) {
                            MPText("Tài sản mới", style: .title)
                            MPTextField(title: "Mã", text: $symbol)
                            MPTextField(title: "Tên", text: $assetName)
                            HStack { MPButton("Hủy", style: .secondary) { adding = false }; MPButton("Tạo") { Task { await createAsset() } } }
                        }
                    }
                }
                if let tradeAsset {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 10) {
                            MPText("Ghi nhận \(tradeAsset.symbol)", style: .title)
                            MPWalletSelector(title: "Loại", options: ["buy", "sell"], selection: $tradeSide)
                            MPTextField(title: "Khối lượng", text: $tradeQuantity, placeholder: "0.25")
                            MPAmountField(title: "Giá mỗi đơn vị", text: $tradePrice)
                            MPAmountField(title: "Phí", text: $tradeFee)
                            MPDateField(title: "Ngày giao dịch", date: $tradeDate)
                            HStack { MPButton("Hủy", style: .secondary) { self.tradeAsset = nil }; MPButton("Lưu") { Task { await saveTrade(tradeAsset) } } }
                        }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current)
        .task { await load() }
        .accessibilityIdentifier("portfolio-screen")
    }

    @ViewBuilder
    private func portfolioAssetCard(_ asset: PortfolioAssetModel, position: PortfolioPositionModel?, trades: [PortfolioTradeModel]) -> some View {
        MPAssistantCard(title: "\(asset.symbol) · \(asset.name)") {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    MPText("Số lượng: \(position?.quantity ?? "0")")
                    MPText(asset.latestPrice.map { "Giá hiện tại: \(MPFormat.money($0))" } ?? "Chưa có giá", style: .secondary)
                }
                Spacer()
                MPButton("Giao dịch", style: .secondary) { tradeAsset = asset; tradeSide = "buy"; tradeQuantity = ""; tradePrice = asset.latestPrice.map(String.init) ?? ""; tradeFee = ""; tradeDate = Date() }
            }
            LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 10) {
                MPMetricCard(title: "Giá vốn", value: MPFormat.money(position?.costBasis ?? 0))
                MPMetricCard(title: "Bình quân", value: MPFormat.money(position?.averageCost ?? 0))
                MPMetricCard(title: "Đã thực hiện", value: MPFormat.signedMoney(position?.realizedPnL ?? 0))
                MPMetricCard(title: "Giá trị hiện tại", value: position?.marketValue.map(MPFormat.money) ?? "Chưa có giá")
                MPMetricCard(title: "Lãi/lỗ chưa thực hiện", value: position?.unrealizedPnL.map(MPFormat.signedMoney) ?? "Chưa có giá")
            }
            MPAmountField(title: "Giá hiện tại", text: Binding(get: { priceByAsset[asset.id] ?? asset.latestPrice.map(String.init) ?? "" }, set: { priceByAsset[asset.id] = $0 }))
            MPButton("Cập nhật giá", style: .secondary) { Task { await updatePrice(asset) } }
            MPAssistantCard(title: "Lịch sử mua bán") {
                let assetTrades = Self.tradesForAsset(trades, assetID: asset.id)
                if assetTrades.isEmpty {
                    MPText("Chưa có giao dịch mua bán.", style: .secondary)
                } else {
                    ForEach(assetTrades) { trade in
                        MPTransactionRow(
                            title: "\(Self.tradeSideLabel(trade.side)) · \(trade.quantity)",
                            subtitle: Self.tradeSubtitle(trade, timeZone: TimeZone(identifier: accountTimezone) ?? .current),
                            amount: MPFormat.money(trade.unitPrice),
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

    static func tradeSubtitle(_ trade: PortfolioTradeModel, timeZone: TimeZone) -> String {
        let date = MPFormat.shortDateTime(trade.occurredAt, timeZone: timeZone)
        let fee = MPFormat.money(trade.fee)
        if let note = trade.note?.trimmingCharacters(in: .whitespacesAndNewlines), !note.isEmpty {
            return "Giá \(MPFormat.money(trade.unitPrice)) · Phí \(fee) · \(date) · \(note)"
        }
        return "Giá \(MPFormat.money(trade.unitPrice)) · Phí \(fee) · \(date)"
    }

    private func load() async {
        do {
            async let summaryResult = api.portfolioSummary()
            async let profile = api.profile()
            summary = try await summaryResult
            if let profile = try? await profile { accountTimezone = profile.timezone }
            self.error = nil
        } catch { self.error = "Không thể tải danh mục đầu tư." }
    }
    private func createAsset() async {
        guard let draft = PortfolioAssetDraftValidator.normalized(symbol: symbol, name: assetName) else { error = "Nhập mã và tên tài sản."; return }
        do { _ = try await api.createPortfolioAsset(symbol: draft.symbol, name: draft.name); adding = false; symbol = ""; assetName = ""; await load() }
        catch { self.error = "Không thể tạo tài sản." }
    }
    private func updatePrice(_ asset: PortfolioAssetModel) async {
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
        do { _ = try await api.updatePortfolioPrice(assetID: asset.id, latestPrice: parsed); await load() }
        catch { self.error = "Không thể cập nhật giá." }
    }
    private func saveTrade(_ asset: PortfolioAssetModel) async {
        guard PortfolioTradeDraftValidator.isValid(side: tradeSide, quantity: tradeQuantity, unitPrice: tradePrice, fee: tradeFee),
              let parsedPrice = PortfolioTradeDraftValidator.nonNegativeInteger(tradePrice),
              let parsedFee = PortfolioTradeDraftValidator.nonNegativeInteger(tradeFee, allowingEmpty: true) else {
            error = "Nhập loại, khối lượng, giá và phí hợp lệ. Khối lượng tối đa 8 chữ số thập phân."
            return
        }
        do {
            _ = try await api.createPortfolioTrade(assetID: asset.id, input: PortfolioTradeInputModel(side: tradeSide, quantity: tradeQuantity.trimmingCharacters(in: .whitespacesAndNewlines), unitPrice: parsedPrice, fee: parsedFee, occurredAt: Self.accountLocalNoonTimestamp(tradeDate, timezone: accountTimezone), note: nil))
            tradeAsset = nil
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
        let value = raw.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !value.isEmpty, !value.contains("+") && !value.contains("-") && !value.contains("e") && !value.contains("E") else { return false }
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

struct CreditView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    let preferredWalletID: String?
    @State private var wallets: [WalletModel] = []
    @State private var sourceWallets: [WalletModel] = []
    @State private var categories: [CategoryModel] = []
    @State private var selectedID = ""
    @State private var statement: CreditStatementModel?
    @State private var error: String?
    @State private var action: String?
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
                    LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) {
                        MPMetricCard(title: "Dư nợ", value: MPFormat.money(abs(statement.balance)))
                        MPMetricCard(title: "Khả dụng", value: MPFormat.money(statement.availableCredit))
                        MPMetricCard(title: "Đến hạn", value: MPFormat.money(statement.amountDue))
                        MPMetricCard(title: "Trạng thái", value: creditPaymentStatusLabel(statement.paymentStatus))
                    }
                    if statement.paymentStatus == "not_configured" {
                        MPText("Chưa cấu hình dư nợ sao kê và ngày đến hạn.", style: .secondary)
                    } else {
                        VStack(alignment: .leading, spacing: 4) {
                            MPText(creditPaymentStatusLabel(statement.paymentStatus), style: .secondary)
                            if let paymentDueAt = statement.paymentDueAt {
                                MPText("Hạn thanh toán: \(MPFormat.shortDate(paymentDueAt, timeZone: TimeZone(identifier: accountTimezone) ?? .current))", style: .secondary)
                            }
                        }
                    }
                    MPAssistantCard(title: "Giao dịch tín dụng") {
                        ForEach(statement.items) { item in
                            MPTransactionRow(title: creditKindLabel(item.creditKind), subtitle: creditItemSubtitle(item), amount: MPFormat.money(item.amount), isIncome: item.creditKind == "refund" || item.creditKind == "payment")
                        }
                    }
                    HStack {
                        MPButton("Thêm giao dịch", style: .secondary) { beginAction("entry") }
                        MPButton("Thanh toán") { beginAction("payment") }
                    }
                    if let action {
                        MPSurfaceCard {
                            VStack(alignment: .leading, spacing: 10) {
                                MPText(action == "payment" ? "Thanh toán thẻ" : "Thêm giao dịch tín dụng", style: .title)
                                if action == "entry" {
                                    MPWalletSelector(title: "Loại giao dịch", options: ["purchase", "refund", "fee", "interest"], selection: $entryKind)
                                    MPOptionSelector(title: "Nhóm", options: [MPSelectionOption(id: "", label: "Không chọn nhóm")] + categoriesForEntry.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $categoryID, emptyLabel: "Không chọn nhóm")
                                } else {
                                    MPOptionSelector(title: "Ví thanh toán", options: sourceWallets.map { MPSelectionOption(id: $0.id, label: $0.name) }, selectionID: $sourceWalletID, emptyLabel: "Chọn ví thanh toán")
                                }
                                MPAmountField(title: "Số tiền", text: $entryAmount)
                                MPTextField(title: "Ghi chú", text: $note, placeholder: "Không bắt buộc")
                                HStack {
                                    MPButton("Hủy", style: .secondary) { cancelAction() }
                                    MPButton("Lưu") { Task { await saveAction() } }
                                }
                            }
                        }
                    }
                }
            }
            if let error { MPStatusMessage(text: error, tone: .danger) }
        }.padding(MPMetrics.pagePadding) }.background(MPColor.canvas.ignoresSafeArea()).environment(\.timeZone, TimeZone(identifier: accountTimezone) ?? .current).task { await loadWallets() }.onChange(of: entryKind) { _, _ in
            if !categoriesForEntry.contains(where: { $0.id == categoryID }) { categoryID = categoriesForEntry.first?.id ?? "" }
        }.accessibilityIdentifier("credit-screen")
    }
    private var categoriesForEntry: [CategoryModel] {
        let transactionType = entryKind == "refund" ? "income" : "expense"
        return categories.filter { $0.applies(to: transactionType, walletID: selectedID, walletType: "credit") }
    }

    private func creditPaymentStatusLabel(_ value: String) -> String {
        switch value {
        case "paid": return "Đã thanh toán"
        case "partial": return "Đã trả một phần"
        case "overdue": return "Đã quá hạn"
        case "due": return "Đang đến hạn"
        default: return "Chưa cấu hình"
        }
    }

    private func creditKindLabel(_ value: String?) -> String {
        switch value {
        case "purchase": return "Mua hàng"
        case "refund": return "Hoàn tiền"
        case "fee": return "Phí"
        case "interest": return "Lãi"
        case "payment": return "Thanh toán"
        default: return "Giao dịch tín dụng"
        }
    }

    private func creditItemSubtitle(_ item: TransactionModel) -> String {
        if let note = item.note?.trimmingCharacters(in: .whitespacesAndNewlines), !note.isEmpty { return note }
        return MPFormat.shortDate(item.occurredAt, timeZone: TimeZone(identifier: accountTimezone) ?? .current)
    }

    private func beginAction(_ value: String) {
        action = value
        entryAmount = ""
        note = ""
        if value == "entry" { categoryID = categoriesForEntry.first?.id ?? "" }
    }

    private func cancelAction() {
        action = nil
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
            // Category metadata is optional for the ledger surface. Keep the
            // entry/payment actions usable even if this secondary request fails.
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
        guard action != nil else { return }
        if action == "entry" { await addEntry() } else { await pay() }
    }
    private func addEntry() async { guard !selectedID.isEmpty, let amount = MPFormat.parsePositiveInteger(entryAmount) else { error = "Chọn ví tín dụng và nhập số tiền."; return }; do { _ = try await api.createCreditEntry(walletID: selectedID, input: CreditEntryInputModel(kind: entryKind, categoryID: categoryID.isEmpty ? nil : categoryID, amount: amount, occurredAt: MPFormat.instantString(Date()), note: note.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? nil : note.trimmingCharacters(in: .whitespacesAndNewlines))); cancelAction(); await loadStatement() } catch { self.error = "Không thể ghi nhận bút toán tín dụng." } }
    private func pay() async { guard !selectedID.isEmpty, let amount = MPFormat.parsePositiveInteger(entryAmount), !sourceWalletID.isEmpty else { error = "Chọn nguồn thanh toán và nhập số tiền."; return }; do { _ = try await api.createCreditPayment(walletID: selectedID, input: CreditPaymentInputModel(sourceWalletID: sourceWalletID, amount: amount, occurredAt: MPFormat.instantString(Date()), note: note.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ? nil : note.trimmingCharacters(in: .whitespacesAndNewlines))); cancelAction(); await loadStatement() } catch { self.error = "Không thể tạo thanh toán tín dụng." } }
}
