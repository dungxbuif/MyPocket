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
    @State private var showEditor = false
    @State private var editTarget: BudgetModel?
    @State private var deleteTarget: BudgetModel?

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
                                HStack { MPButton("Sửa", style: .secondary) { beginEditing(budget) }; MPButton("Xóa", style: .destructive) { deleteTarget = budget } }
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
                            HStack { MPButton("Hủy", style: .secondary) { showEditor = false; editTarget = nil }; MPButton("Lưu") { Task { await create() } } }
                        }
                    }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .mpConfirmDialog(isPresented: Binding(get: { deleteTarget != nil }, set: { if !$0 { deleteTarget = nil } }), title: "Xóa ngân sách?", message: "Các giao dịch không bị xóa.", confirmTitle: "Xóa") {
            if let target = deleteTarget { Task { do { try await api.deleteBudget(id: target.id); await load() } catch { self.error = "Không thể xóa ngân sách." } } }
            deleteTarget = nil
        }
        .accessibilityIdentifier("budgets-screen")
    }

    private func load() async {
        guard !loading else { return }
        loading = true; defer { loading = false }
        do { summary = try await api.budgets(); self.error = nil } catch { self.error = "Không thể tải ngân sách." }
    }

    private func create() async {
        guard let parsed = Int64(limit.filter(\.isNumber)), !name.isEmpty else { error = "Nhập tên và giới hạn hợp lệ."; return }
        do {
            let input = BudgetInputModel(name: name, limitAmount: parsed, walletID: nil, categoryID: nil, startDate: start, endDate: end)
            if let editTarget { _ = try await api.updateBudget(id: editTarget.id, input: input) } else { _ = try await api.createBudget(input) }
            showEditor = false; self.editTarget = nil; await load()
        }
        catch { self.error = "Không thể tạo ngân sách." }
    }

    private func beginEditing(_ budget: BudgetModel) {
        editTarget = budget
        name = budget.name
        limit = String(budget.limitAmount)
        start = budget.startDate
        end = budget.endDate
        showEditor = true
    }
}

struct ReportsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @State private var month = MPFormat.currentMonth
    @State private var summary: InsiderSummaryModel?
    @State private var loading = false
    @State private var error: String?

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
                    }
                    MPAssistantCard(title: "Theo danh mục") {
                        MPChart(values: summary.topCategories.map { ($0.name, min(180, max(4, Double($0.amount) / 100_000))) })
                        ForEach(summary.topCategories, id: \.name) { item in MPText("\(item.name): \(MPFormat.money(item.amount))", style: .secondary) }
                    }
                    MPAssistantCard(title: "Khoản chi lớn") {
                        ForEach(summary.topExpenses) { item in MPTransactionRow(title: item.categoryName ?? "Chi tiêu", subtitle: MPFormat.shortDate(item.occurredAt), amount: MPFormat.money(item.amount), isIncome: false) }
                    }
                } else if let error { MPErrorView(title: "Không tải được báo cáo", detail: error) { Task { await load() } } }
                else { MPEmptyView(title: "Chọn tháng để xem báo cáo", detail: "Số liệu luôn được tính theo timezone tài khoản.") }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .accessibilityIdentifier("reports-screen")
    }

    private func load() async {
        guard !loading else { return }
        loading = true; defer { loading = false }
        do { summary = try await api.insider(month: month); self.error = nil } catch { self.error = "Không thể tải báo cáo." }
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

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Danh mục", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { MPIconButton(systemName: "plus", label: "Thêm danh mục") { adding = true } })
                if loading && items.isEmpty { MPLoadingView(label: "Đang tải danh mục...") }
                else if let error, items.isEmpty { MPErrorView(title: "Không tải được danh mục", detail: error) { Task { await load() } } }
                else {
                    MPSurfaceCard {
                        VStack(spacing: 0) {
                            ForEach(items) { item in
                                HStack { Image(systemName: "tag.fill").foregroundStyle(MPColor.action); MPText(item.name); Spacer(); MPText(item.kind, style: .secondary); MPIconButton(systemName: "pencil", label: "Sửa \(item.name)") { beginEditing(item) }; MPIconButton(systemName: "trash", label: "Xóa \(item.name)") { deleteTarget = item } }.padding(.vertical, 8)
                            }
                        }
                    }
                }
                if adding {
                    MPSurfaceCard {
                        VStack(alignment: .leading, spacing: 10) {
                            MPText(editTarget == nil ? "Danh mục mới" : "Sửa danh mục", style: .title)
                            MPTextField(title: "Tên", text: $name)
                            MPWalletSelector(title: "Loại", options: ["expense", "income"], selection: $kind)
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

    private func load() async { loading = true; defer { loading = false }; do { items = try await api.categories(); self.error = nil } catch { self.error = "Không thể tải danh mục." } }
    private func create() async { guard !name.isEmpty else { self.error = "Nhập tên danh mục."; return }; do { let input = CategoryModelInput(name: name, kind: kind, parentID: editTarget?.parentID, walletIDs: editTarget?.walletIDs ?? [], iconKey: editTarget?.iconKey ?? "tag"); if let editTarget { _ = try await api.updateCategory(id: editTarget.id, input: input) } else { _ = try await api.createCategory(input) }; adding = false; name = ""; self.editTarget = nil; await load() } catch { self.error = "Không thể lưu danh mục." } }
    private func beginEditing(_ item: CategoryModel) { editTarget = item; name = item.name; kind = item.kind; adding = true }
}

struct JarsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @State private var summary: JarMonthSummaryModel?
    @State private var month = MPFormat.currentMonth
    @State private var loading = false
    @State private var error: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Hũ chi tiêu", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                MPTextField(title: "Tháng (YYYY-MM)", text: $month)
                MPButton("Tải hũ", isLoading: loading) { Task { await load() } }
                if let summary {
                    MPMetricCard(title: "Tổng đã phân bổ", value: MPFormat.money(summary.totalAllocated), detail: "Thu nhập: \(MPFormat.money(summary.actualIncome))")
                    ForEach(summary.items) { item in
                        MPSurfaceCard { VStack(alignment: .leading, spacing: 6) { MPText(item.name); MPProgressBar(progress: Double(item.spent) / Double(max(1, item.calculatedAllocation ?? 1)), label: "Đã chi \(MPFormat.money(item.spent))") } }
                    }
                } else if let error { MPErrorView(title: "Không tải được hũ", detail: error) { Task { await load() } } }
                else { MPEmptyView(title: "Chưa chọn tháng", detail: "Hũ là lớp ngân sách theo tháng và không thay đổi số dư ví.") }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .accessibilityIdentifier("jars-screen")
    }

    private func load() async { loading = true; defer { loading = false }; do { summary = try await api.jarMonth(month: month); self.error = nil } catch { self.error = "Không thể tải hũ chi tiêu." } }
}

struct MonthsView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @State private var month = MPFormat.currentMonth
    @State private var summary: MonthSummaryModel?
    @State private var note = ""
    @State private var error: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Tháng", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
                MPTextField(title: "Tháng (YYYY-MM)", text: $month)
                MPButton("Tải tháng") { Task { await load() } }
                if let summary {
                    LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) { MPMetricCard(title: "Thu", value: MPFormat.money(summary.income)); MPMetricCard(title: "Chi", value: MPFormat.money(summary.expense)); MPMetricCard(title: "Ròng", value: MPFormat.money(summary.net)); MPMetricCard(title: "Số giao dịch", value: "\(summary.transactionCount)") }
                    MPTextEditor(title: "Ghi chú tháng", text: $note, placeholder: "Ghi lại điều quan trọng của tháng")
                    HStack {
                        MPButton("Lưu ghi chú", style: .secondary) { Task { do { try await api.saveMonthNote(month: month, note: note); self.error = nil } catch { self.error = "Không thể lưu ghi chú." } } }
                        MPButton("Xóa ghi chú", style: .destructive) { Task { do { try await api.deleteMonthNote(month: month); note = ""; self.error = nil } catch { self.error = "Không thể xóa ghi chú." } } }
                    }
                    MPAssistantCard(title: "Theo danh mục") { ForEach(summary.categories, id: \.name) { item in MPText("\(item.name): \(MPFormat.money(item.amount))", style: .secondary) } }
                } else if let error { MPErrorView(title: "Không tải được tháng", detail: error) { Task { await load() } } }
                else { MPEmptyView(title: "Chưa có dữ liệu", detail: "Chọn một tháng để xem tổng kết.") }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .accessibilityIdentifier("months-screen")
    }

    private func load() async { do { summary = try await api.monthSummary(month: month); note = summary?.note ?? ""; self.error = nil } catch { self.error = "Không thể tải tháng." } }
}

struct RecurringView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @State private var items: [RecurringModel] = []
    @State private var wallets: [WalletModel] = []
    @State private var showEditor = false
    @State private var editTarget: RecurringModel?
    @State private var name = ""
    @State private var amount = ""
    @State private var walletName = ""
    @State private var frequency = "monthly"
    @State private var nextRun = MPFormat.todayISO
    @State private var result: String?
    @State private var error: String?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Giao dịch định kỳ", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { MPIconButton(systemName: "plus", label: "Thêm lịch") { editTarget = nil; showEditor = true } })
                MPButton("Chạy các giao dịch đến hạn", style: .secondary) { Task { await runDue() } }
                if let result { MPStatusMessage(text: result, tone: .success) }
                if let error { MPStatusMessage(text: error, tone: .danger) }
                if items.isEmpty { MPEmptyView(title: "Chưa có lịch định kỳ", detail: "Lịch định kỳ chỉ tạo giao dịch khi bạn chủ động chạy.") }
                ForEach(items) { item in MPSurfaceCard { VStack(alignment: .leading, spacing: 6) { MPText(item.name); MPText("\(item.frequency) · \(MPFormat.money(item.amount))", style: .secondary); MPText(item.active ? "Đang hoạt động" : "Đã tắt", style: .secondary); HStack { MPButton("Sửa", style: .secondary) { beginEditing(item) }; MPButton("Xóa", style: .destructive) { Task { await delete(item) } } } } } }
                if showEditor {
                    MPSurfaceCard { VStack(alignment: .leading, spacing: 10) {
                        MPText(editTarget == nil ? "Tạo lịch" : "Sửa lịch", style: .title)
                        MPTextField(title: "Tên", text: $name)
                        MPAmountField(title: "Số tiền", text: $amount)
                        MPWalletSelector(title: "Ví", options: wallets.map(\.name), selection: $walletName)
                        MPWalletSelector(title: "Tần suất", options: ["daily", "weekly", "monthly", "yearly"], selection: $frequency)
                        MPTextField(title: "Ngày chạy đầu tiên", text: $nextRun)
                        HStack { MPButton("Hủy", style: .secondary) { showEditor = false; editTarget = nil }; MPButton("Lưu") { Task { await save() } } }
                    } }
                }
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .accessibilityIdentifier("recurring-screen")
    }

    private func load() async { do { async let schedules = api.recurring(); async let walletList = api.wallets(); items = try await schedules; wallets = try await walletList; if walletName.isEmpty { walletName = wallets.first?.name ?? "" }; self.error = nil } catch { self.error = "Không thể tải lịch định kỳ." } }
    private func runDue() async { do { let response = try await api.runDueRecurring(); result = "Đã tạo \(response.created) giao dịch."; await load() } catch { self.error = "Không thể chạy giao dịch đến hạn." } }
    private func beginEditing(_ item: RecurringModel) { editTarget = item; name = item.name; amount = String(item.amount); walletName = wallets.first(where: { $0.id == item.walletID })?.name ?? wallets.first?.name ?? ""; frequency = item.frequency; nextRun = item.nextRunAt; showEditor = true }
    private func save() async { guard let wallet = wallets.first(where: { $0.name == walletName }), let parsed = Int64(amount.filter(\.isNumber)), !name.isEmpty else { self.error = "Nhập đủ tên, ví và số tiền."; return }; let input = RecurringInputModel(name: name, walletID: wallet.id, categoryID: nil, type: "expense", amount: parsed, note: nil, frequency: frequency, interval: 1, nextRunAt: nextRun, endsAt: nil, active: true); do { if let editTarget { _ = try await api.updateRecurring(id: editTarget.id, input: input) } else { _ = try await api.createRecurring(input) }; showEditor = false; self.editTarget = nil; await load() } catch { self.error = "Không thể lưu lịch định kỳ." } }
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

    var body: some View {
        ScrollView { VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: "Travel Mode", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { MPIconButton(systemName: "plus", label: "Thêm chuyến đi") { adding = true } })
            if let error { MPErrorView(title: "Không tải được Travel Mode", detail: error) { Task { await load() } } }
            if events.isEmpty { MPEmptyView(title: "Chưa có chuyến đi", detail: "Tạo một sự kiện để gắn giao dịch du lịch.") }
            ForEach(events) { event in
                MPSurfaceCard { HStack { VStack(alignment: .leading) { MPText(event.name); if let context = event.context { MPText(context, style: .secondary) } }; Spacer(); MPButton(event.active ? "Tắt" : "Bật", style: event.active ? .destructive : .secondary) { Task { await toggle(event) } }; MPButton("Xóa", style: .destructive) { Task { await remove(event) } } } }
            }
            if adding { MPSurfaceCard { VStack(alignment: .leading, spacing: 10) { MPText("Chuyến đi mới", style: .title); MPTextField(title: "Tên", text: $name); MPTextEditor(title: "Mô tả", text: $context); HStack { MPButton("Hủy", style: .secondary) { adding = false }; MPButton("Tạo") { Task { await create() } } } } } }
        }.padding(MPMetrics.pagePadding) }.background(MPColor.canvas.ignoresSafeArea()).task { await load() }.accessibilityIdentifier("travel-screen")
    }
    private func load() async { do { events = try await api.travelEvents(); self.error = nil } catch { self.error = "Không thể tải sự kiện du lịch." } }
    private func toggle(_ event: TravelEventModel) async { do { _ = try await api.setTravelEventActive(id: event.id, active: !event.active); await load() } catch { self.error = "Không thể cập nhật Travel Mode." } }
    private func create() async { guard !name.isEmpty else { self.error = "Nhập tên chuyến đi."; return }; do { _ = try await api.createTravelEvent(TravelEventInputModel(name: name, context: context.isEmpty ? nil : context, startsOn: nil, endsOn: nil, active: false)); name = ""; context = ""; adding = false; await load() } catch { self.error = "Không thể tạo chuyến đi." } }
    private func remove(_ event: TravelEventModel) async { do { try await api.deleteTravelEvent(id: event.id); await load() } catch { self.error = "Không thể xóa chuyến đi." } }
}

struct PortfolioView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @State private var summary: PortfolioSummaryModel?
    @State private var error: String?
    @State private var adding = false
    @State private var symbol = ""
    @State private var assetName = ""
    @State private var price = ""

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                MPScreenHeader(title: "Danh mục đầu tư", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { MPIconButton(systemName: "plus", label: "Thêm tài sản") { adding = true } })
                if let summary {
                    MPMetricCard(title: "Giá vốn", value: MPFormat.money(summary.totalCostBasis))
                    if let market = summary.totalMarketValue { MPMetricCard(title: "Giá trị thị trường", value: MPFormat.money(market)) }
                    ForEach(summary.positions) { position in
                        MPSurfaceCard {
                            VStack(alignment: .leading) {
                                MPText(position.symbol ?? position.name ?? "Tài sản")
                                MPText("SL \(position.quantity) · Giá vốn \(MPFormat.money(position.costBasis))", style: .secondary)
                            }
                        }
                    }
                    ForEach(summary.assets) { asset in
                        MPSurfaceCard {
                            HStack {
                                VStack(alignment: .leading) {
                                    MPText("\(asset.symbol) · \(asset.name)")
                                    MPText(asset.latestPrice.map { MPFormat.money($0) } ?? "Chưa có giá", style: .secondary)
                                }
                                Spacer()
                                MPButton("Cập nhật giá", style: .secondary) {
                                    price = asset.latestPrice.map(String.init) ?? ""
                                    Task { await updatePrice(asset) }
                                }
                            }
                        }
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
            }
            .padding(MPMetrics.pagePadding)
        }
        .background(MPColor.canvas.ignoresSafeArea())
        .task { await load() }
        .accessibilityIdentifier("portfolio-screen")
    }
    private func load() async { do { summary = try await api.portfolioSummary(); self.error = nil } catch { self.error = "Không thể tải danh mục đầu tư." } }
    private func createAsset() async { guard !symbol.isEmpty, !assetName.isEmpty else { error = "Nhập mã và tên tài sản."; return }; do { _ = try await api.createPortfolioAsset(symbol: symbol, name: assetName); adding = false; symbol = ""; assetName = ""; await load() } catch { self.error = "Không thể tạo tài sản." } }
    private func updatePrice(_ asset: PortfolioAssetModel) async { guard let parsed = Int64(price.filter(\.isNumber)) else { error = "Nhập giá hợp lệ."; return }; do { _ = try await api.updatePortfolioPrice(assetID: asset.id, latestPrice: parsed); await load() } catch { self.error = "Không thể cập nhật giá." } }
}

struct CreditView: View {
    let api: FinanceAPI
    let onBack: () -> Void
    @State private var wallets: [WalletModel] = []
    @State private var sourceWallets: [WalletModel] = []
    @State private var selected: WalletModel?
    @State private var statement: CreditStatementModel?
    @State private var error: String?
    @State private var entryKind = "purchase"
    @State private var entryAmount = ""
    @State private var sourceWalletName = ""

    var body: some View {
        ScrollView { VStack(alignment: .leading, spacing: 16) {
            MPScreenHeader(title: "Sổ tín dụng", leading: { MPIconButton(systemName: "chevron.left", label: "Quay lại", action: onBack) }, trailing: { EmptyView() })
            if wallets.isEmpty { MPEmptyView(title: "Chưa có ví tín dụng", detail: "Tạo ví loại tín dụng trong mục Ví.") }
            else { MPWalletSelector(title: "Ví tín dụng", options: wallets.map(\.name), selection: Binding(get: { selected?.name ?? wallets.first?.name ?? "" }, set: { name in selected = wallets.first(where: { $0.name == name }); Task { await loadStatement() } })); if let statement { LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 12) { MPMetricCard(title: "Dư nợ", value: MPFormat.money(statement.balance)); MPMetricCard(title: "Khả dụng", value: MPFormat.money(statement.availableCredit)); MPMetricCard(title: "Đến hạn", value: MPFormat.money(statement.amountDue)); MPMetricCard(title: "Trạng thái", value: statement.paymentStatus) }; MPAssistantCard(title: "Giao dịch tín dụng") { ForEach(statement.items) { item in MPTransactionRow(title: item.creditKind ?? "Giao dịch", subtitle: MPFormat.shortDate(item.occurredAt), amount: MPFormat.money(item.amount), isIncome: item.creditKind == "refund" || item.creditKind == "payment") } }; MPWalletSelector(title: "Nguồn thanh toán", options: sourceWallets.map(\.name), selection: $sourceWalletName); MPWalletSelector(title: "Loại bút toán", options: ["purchase", "refund", "fee", "interest"], selection: $entryKind); MPAmountField(title: "Số tiền", text: $entryAmount); HStack { MPButton("Ghi nhận", style: .secondary) { Task { await addEntry() } }; MPButton("Thanh toán") { Task { await pay() } } } } }
            if let error { MPStatusMessage(text: error, tone: .danger) }
        }.padding(MPMetrics.pagePadding) }.background(MPColor.canvas.ignoresSafeArea()).task { await loadWallets() }.accessibilityIdentifier("credit-screen")
    }
    private func loadWallets() async { do { let all = try await api.wallets(); wallets = all.filter { $0.type == "credit" }; sourceWallets = all.filter { $0.type != "credit" }; sourceWalletName = sourceWallets.first?.name ?? ""; selected = wallets.first; await loadStatement() } catch { self.error = "Không thể tải ví tín dụng." } }
    private func loadStatement() async { guard let selected else { return }; do { statement = try await api.creditStatement(walletID: selected.id); self.error = nil } catch { self.error = "Không thể tải sổ tín dụng." } }
    private func addEntry() async { guard let selected, let amount = Int64(entryAmount.filter(\.isNumber)), amount > 0 else { error = "Chọn ví tín dụng và nhập số tiền."; return }; do { _ = try await api.createCreditEntry(walletID: selected.id, input: CreditEntryInputModel(kind: entryKind, categoryID: nil, amount: amount, occurredAt: ISO8601DateFormatter().string(from: Date()), note: nil)); entryAmount = ""; await loadStatement() } catch { self.error = "Không thể ghi nhận bút toán tín dụng." } }
    private func pay() async { guard let selected, let source = sourceWallets.first(where: { $0.name == sourceWalletName }), let amount = Int64(entryAmount.filter(\.isNumber)), amount > 0 else { error = "Chọn nguồn thanh toán và nhập số tiền."; return }; do { _ = try await api.createCreditPayment(walletID: selected.id, input: CreditPaymentInputModel(sourceWalletID: source.id, amount: amount, occurredAt: ISO8601DateFormatter().string(from: Date()), note: nil)); entryAmount = ""; await loadStatement() } catch { self.error = "Không thể tạo thanh toán tín dụng." } }
}
