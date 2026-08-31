import { useEffect, useState, type ReactNode } from "react";
import {
  Bell,
  BriefcaseBusiness,
  CalendarDays,
  ChevronRight,
  Eye,
  Home,
  ImagePlus,
  List,
  MapPin,
  Plus,
  Search,
  User,
  Users,
  Wallet,
} from "lucide-react";

import { useOnlineStatus } from "./offline";
import { apiBaseURL } from "./apiClient";
import { loadCurrentUser, logout, type AuthState } from "./auth";
import { clearOfflineStore, initializeOfflineStore, saveFinanceMirror } from "../offline/db";
import {
  discardLocalConflict,
  editAndRetryTransactionConflict,
  fullResync,
  keepServerConflict,
  listOpenConflicts,
} from "../offline/conflicts";
import {
  archiveCategory,
  archiveTransaction,
  archiveWallet,
  createCategory,
  createTransaction,
  createWallet,
  loadCategories,
  loadTransactions,
  loadWallets,
  setDefaultAIWallet,
  setWalletCategoryActive,
  updateCategory,
  updateTransaction,
  updateWallet,
  type CategorySummary,
  type Transaction,
  type TransactionInput,
  type TransactionType,
  type WalletSummary,
  type WalletType,
} from "./finance";
import {
  archiveBudget,
  archiveEvent,
  archiveObligation,
  createBudget,
  createEvent,
  createObligation,
  linkEventTransaction,
  linkObligationRepayment,
  loadBudgets,
  loadEvents,
  loadObligations,
  updateBudget,
  updateEvent,
  updateObligation,
  type BudgetInput,
  type BudgetPeriodType,
  type BudgetProgress,
  type EventInput,
  type EventSummary,
  type ObligationDirection,
  type ObligationInput,
  type ObligationSummary,
} from "./planning";
import { drainOutbox, readOutbox } from "./outbox";
import type { OfflineConflict } from "../offline/types";

type Tab = "overview" | "transactions" | "budgets" | "account";

const tabs: Array<{ id: Tab; label: string; icon: typeof Home }> = [
  { id: "overview", label: "Tổng quan", icon: Home },
  { id: "transactions", label: "Sổ giao dịch", icon: Wallet },
  { id: "budgets", label: "Ngân sách", icon: BriefcaseBusiness },
  { id: "account", label: "Tài khoản", icon: User },
];

export function App() {
  const online = useOnlineStatus();
  const [activeTab, setActiveTab] = useState<Tab>("overview");
  const [sheetOpen, setSheetOpen] = useState(false);
  const [walletSheetOpen, setWalletSheetOpen] = useState(false);
  const [editingTransaction, setEditingTransaction] = useState<Transaction | null>(null);
  const [authState, setAuthState] = useState<AuthState>({ status: "loading" });
  const [wallets, setWallets] = useState<WalletSummary[] | null>(null);
  const [categories, setCategories] = useState<CategorySummary[] | null>(null);
  const [transactions, setTransactions] = useState<Transaction[] | null>(null);
  const [budgets, setBudgets] = useState<BudgetProgress[] | null>(null);
  const [events, setEvents] = useState<EventSummary[] | null>(null);
  const [obligations, setObligations] = useState<ObligationSummary[] | null>(null);
  const [budgetSheetOpen, setBudgetSheetOpen] = useState(false);
  const [eventSheetOpen, setEventSheetOpen] = useState(false);
  const [obligationSheetOpen, setObligationSheetOpen] = useState(false);
  const [editingBudget, setEditingBudget] = useState<BudgetProgress | null>(null);
  const [editingEvent, setEditingEvent] = useState<EventSummary | null>(null);
  const [editingObligation, setEditingObligation] = useState<ObligationSummary | null>(null);
  const [offlineStatus, setOfflineStatus] = useState<{ mode: "ready" | "degraded"; pending: number; reason?: string }>({ mode: "ready", pending: 0 });
  const [conflicts, setConflicts] = useState<OfflineConflict[]>([]);
  const offlineReadOnly = !online && offlineStatus.mode === "degraded";
  const headerWallets = wallets && wallets.length > 0 ? wallets : sampleWallets;
  const totalBalance = totalIncludedVND(headerWallets);

  useEffect(() => {
    let cancelled = false;
    void loadCurrentUser().then((nextAuthState) => {
      if (!cancelled) setAuthState(nextAuthState);
    });
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (authState.status !== "authenticated") {
      setWallets(null);
      setCategories(null);
      setTransactions(null);
      setBudgets(null);
      setEvents(null);
      setObligations(null);
      setConflicts([]);
      return;
    }
    let cancelled = false;
    void hydrateOfflineData()
      .then(() => online ? refreshFinanceData() : undefined)
      .catch(() => {
        if (!cancelled) {
          setWallets([]);
          setCategories([]);
          setTransactions([]);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [authState.status, online]);

  useEffect(() => {
    if (authState.status !== "authenticated" || !online) return;
    void readOutbox().then((pending) => {
      if (pending.length === 0) return;
      void drainOutbox(async (input) => { await createTransaction(input as Parameters<typeof createTransaction>[0]); })
        .then(() => refreshFinanceData())
        .then(() => refreshConflictState())
        .catch(() => undefined);
    });
  }, [authState.status, online]);

  async function handleLogout() {
    await logout().catch(() => undefined);
    await clearOfflineStore().catch(() => undefined);
    setAuthState({ status: "unauthenticated" });
    setWallets(null);
    setCategories(null);
    setTransactions(null);
    setBudgets(null);
    setEvents(null);
    setObligations(null);
    setConflicts([]);
    setOfflineStatus({ mode: "ready", pending: 0 });
    setActiveTab("overview");
  }

  async function refreshFinanceData() {
    const [nextWallets, nextCategories, nextTransactions, nextBudgets, nextEvents, nextObligations] = await Promise.all([loadWallets(), loadCategories(), loadTransactions(), loadBudgets(), loadEvents(), loadObligations()]);
    setWallets(nextWallets);
    setCategories(nextCategories);
    setBudgets(nextBudgets);
    setEvents(nextEvents);
    setObligations(nextObligations);
    await saveFinanceMirror({ wallets: nextWallets, categories: nextCategories, transactions: nextTransactions });
    const queued = await readOutbox();
    setOfflineStatus((current) => ({ ...current, pending: queued.length }));
    setConflicts(await listOpenConflicts());
    const pending = queued.map((item) => ({ id: item.id, ...item.input, amount_vnd: Number(item.input.amount_vnd), balance_after_vnd: 0, occurred_at: String(item.input.occurred_at), note: String(item.input.note ?? ""), with_person: "", event_ref: "", excluded_from_reports: Boolean(item.input.excluded_from_reports), version: 0 } as Transaction));
    setTransactions([...pending, ...nextTransactions]);
  }

  async function hydrateOfflineData() {
    const snapshot = await initializeOfflineStore();
    setOfflineStatus({ mode: snapshot.mode, pending: snapshot.outbox.length, reason: snapshot.reason });
    setConflicts(snapshot.conflicts.filter((conflict) => conflict.status === "open"));
    if (snapshot.wallets.length > 0) setWallets(snapshot.wallets);
    if (snapshot.categories.length > 0) setCategories(snapshot.categories);
    if (snapshot.transactions.length > 0 || snapshot.outbox.length > 0) setTransactions(snapshot.transactions);
  }

  async function refreshConflictState() {
    setConflicts(await listOpenConflicts());
    const queued = await readOutbox();
    setOfflineStatus((current) => ({ ...current, pending: queued.length }));
  }

  async function reconcileAfterLocalChange() {
    if (online) {
      await refreshFinanceData();
      return;
    }
    const queued = await readOutbox();
    setOfflineStatus((current) => ({ ...current, pending: queued.length }));
    setConflicts(await listOpenConflicts());
  }

  async function handleConflictAction(action: () => Promise<void>) {
    await action();
    if (online) {
      await drainOutbox(async (input) => { await createTransaction(input as Parameters<typeof createTransaction>[0]); }).catch(() => undefined);
      await refreshFinanceData();
      return;
    }
    await hydrateOfflineData();
  }

  function upsertTransaction(transaction: Transaction) {
    setTransactions((current) => [transaction, ...(current ?? []).filter((item) => item.id !== transaction.id)]);
    void reconcileAfterLocalChange().catch(() => undefined);
  }

  return (
    <div className="app-shell">
      <main className="phone-frame">
        <header className="home-header">
          <div>
            <div className="balance-line">
              <strong>{formatVND(totalBalance)}</strong>
              <button className="icon-button" aria-label="Ẩn số dư" type="button">
                <Eye size={24} />
              </button>
            </div>
            <p>Tổng số dư <span className="help-dot">?</span></p>
          </div>
          <div className="header-actions">
            {!online ? <span className="offline-pill">Offline</span> : null}
            {offlineStatus.pending > 0 ? <span className="offline-pill pending">{offlineStatus.pending} chờ đồng bộ</span> : null}
            {conflicts.length > 0 ? <span className="offline-pill conflict">{conflicts.length} cần xử lý</span> : null}
            {offlineStatus.mode === "degraded" ? <span className="offline-pill degraded">Chỉ đọc offline</span> : null}
            <button className="plain-icon" aria-label="Tìm kiếm" type="button">
              <Search size={30} />
            </button>
            <button className="plain-icon notification" aria-label="Thông báo" type="button">
              <Bell size={30} />
              <span>4</span>
            </button>
          </div>
        </header>

        <AuthBanner authState={authState} />
        {authState.status === "authenticated" ? <ConflictInbox conflicts={conflicts} onResolve={handleConflictAction} /> : null}
        {authState.status === "forbidden" ? <ForbiddenState authState={authState} onLogout={handleLogout} /> : null}
        {authState.status !== "forbidden" && activeTab === "overview" ? <Overview wallets={wallets} onManageWallets={() => setWalletSheetOpen(true)} /> : null}
        {authState.status !== "forbidden" && activeTab === "transactions" ? <Transactions transactions={transactions ?? []} onEdit={setEditingTransaction} /> : null}
        {authState.status !== "forbidden" && activeTab === "budgets" ? <Budgets budgets={budgets} events={events ?? []} obligations={obligations ?? []} categories={categories ?? []} transactions={transactions ?? []} online={online} onCreate={() => setBudgetSheetOpen(true)} onCreateEvent={() => setEventSheetOpen(true)} onCreateObligation={() => setObligationSheetOpen(true)} onEdit={setEditingBudget} onEditEvent={setEditingEvent} onEditObligation={setEditingObligation} /> : null}
        {authState.status !== "forbidden" && activeTab === "account" ? <Account authState={authState} onLogout={handleLogout} /> : null}
      </main>

      <nav className="bottom-nav" aria-label="Điều hướng chính">
        {tabs.slice(0, 2).map((tab) => (
          <TabButton key={tab.id} tab={tab} active={activeTab === tab.id} onClick={() => setActiveTab(tab.id)} />
        ))}
        <button className="add-button" aria-label="Thêm giao dịch" type="button" onClick={() => setSheetOpen(true)}>
          <Plus size={36} />
        </button>
        {tabs.slice(2).map((tab) => (
          <TabButton key={tab.id} tab={tab} active={activeTab === tab.id} onClick={() => setActiveTab(tab.id)} />
        ))}
      </nav>

      {sheetOpen ? <AddTransactionSheet categories={categories ?? []} wallets={wallets ?? []} readOnly={offlineReadOnly} onCreated={upsertTransaction} onClose={() => setSheetOpen(false)} /> : null}
      {walletSheetOpen ? <WalletManagerSheet categories={categories ?? []} wallets={wallets ?? []} readOnly={offlineReadOnly} onWalletChanged={(wallet) => setWallets((current) => [wallet, ...(current ?? []).filter((item) => item.id !== wallet.id)])} onWalletArchived={(walletID) => setWallets((current) => (current ?? []).filter((item) => item.id !== walletID))} onCategoryChanged={(category) => setCategories((current) => [category, ...(current ?? []).filter((item) => item.id !== category.id)])} onCategoryArchived={(categoryID) => setCategories((current) => (current ?? []).filter((item) => item.id !== categoryID))} onChanged={() => void reconcileAfterLocalChange().catch(() => undefined)} onClose={() => setWalletSheetOpen(false)} /> : null}
      {editingTransaction ? <EditTransactionSheet categories={categories ?? []} wallets={wallets ?? []} readOnly={offlineReadOnly} transaction={editingTransaction} onChanged={upsertTransaction} onArchived={() => { setTransactions((current) => (current ?? []).filter((item) => item.id !== editingTransaction.id)); setEditingTransaction(null); void reconcileAfterLocalChange().catch(() => undefined); }} onClose={() => setEditingTransaction(null)} /> : null}
      {budgetSheetOpen ? <BudgetSheet categories={categories ?? []} onSaved={() => { setBudgetSheetOpen(false); void refreshFinanceData(); }} onClose={() => setBudgetSheetOpen(false)} /> : null}
      {editingBudget ? <BudgetSheet budget={editingBudget} categories={categories ?? []} onSaved={() => { setEditingBudget(null); void refreshFinanceData(); }} onArchived={() => { setEditingBudget(null); void refreshFinanceData(); }} onClose={() => setEditingBudget(null)} /> : null}
      {eventSheetOpen ? <EventSheet transactions={transactions ?? []} onSaved={() => { setEventSheetOpen(false); void refreshFinanceData(); }} onClose={() => setEventSheetOpen(false)} /> : null}
      {editingEvent ? <EventSheet event={editingEvent} transactions={transactions ?? []} onSaved={() => { setEditingEvent(null); void refreshFinanceData(); }} onArchived={() => { setEditingEvent(null); void refreshFinanceData(); }} onClose={() => setEditingEvent(null)} /> : null}
      {obligationSheetOpen ? <ObligationSheet transactions={transactions ?? []} onSaved={() => { setObligationSheetOpen(false); void refreshFinanceData(); }} onClose={() => setObligationSheetOpen(false)} /> : null}
      {editingObligation ? <ObligationSheet obligation={editingObligation} transactions={transactions ?? []} onSaved={() => { setEditingObligation(null); void refreshFinanceData(); }} onArchived={() => { setEditingObligation(null); void refreshFinanceData(); }} onClose={() => setEditingObligation(null)} /> : null}
    </div>
  );
}

function ConflictInbox({ conflicts, onResolve }: { conflicts: OfflineConflict[]; onResolve: (action: () => Promise<void>) => Promise<void> }) {
  if (conflicts.length === 0) return null;
  return (
    <section className="card conflict-inbox" aria-label="Xung đột đồng bộ">
      <div className="section-title">
        <h2>Cần xử lý</h2>
        <button type="button" onClick={() => void onResolve(fullResync)}>Đồng bộ lại</button>
      </div>
      {conflicts.map((conflict) => (
        <ConflictRow key={conflict.conflict_id} conflict={conflict} onResolve={onResolve} />
      ))}
    </section>
  );
}

function ConflictRow({ conflict, onResolve }: { conflict: OfflineConflict; onResolve: (action: () => Promise<void>) => Promise<void> }) {
  const serverNote = String(conflict.server_payload.note ?? conflict.server_payload.name ?? "Bản trên server");
  const localNote = String(conflict.local_payload.note ?? conflict.local_payload.name ?? "Thay đổi offline");
  const [amount, setAmount] = useState(String(Number(conflict.local_payload.amount_vnd ?? conflict.server_payload.amount_vnd ?? 0)));
  const [note, setNote] = useState(localNote);
  const canRetryTransaction = conflict.entity_type === "transaction" && conflict.operation === "update" && Number(amount) > 0;
  return (
    <article className="conflict-row">
      <div>
        <strong>{conflictLabel(conflict)}</strong>
        <p>Server: {serverNote}</p>
        <p>Offline: {localNote}</p>
      </div>
      {canRetryTransaction ? (
        <div className="conflict-edit">
          <input aria-label={`Số tiền xử lý ${conflict.entity_id}`} inputMode="numeric" value={amount} onChange={(event) => setAmount(event.target.value.replace(/\D/g, ""))} />
          <input aria-label={`Ghi chú xử lý ${conflict.entity_id}`} value={note} onChange={(event) => setNote(event.target.value)} />
        </div>
      ) : null}
      <div className="conflict-actions">
        <button type="button" onClick={() => void onResolve(() => keepServerConflict(conflict))}>Giữ server</button>
        {canRetryTransaction ? <button type="button" onClick={() => void onResolve(() => editAndRetryTransactionConflict(conflict, { amount_vnd: Number(amount), note }))}>Sửa gửi lại</button> : null}
        <button type="button" onClick={() => void onResolve(() => discardLocalConflict(conflict))}>Bỏ offline</button>
      </div>
    </article>
  );
}

function AuthBanner({ authState }: { authState: AuthState }) {
  if (authState.status === "loading") {
    return <section className="auth-panel"><p>Đang kiểm tra phiên đăng nhập...</p></section>;
  }
  if (authState.status === "unauthenticated") {
    return (
      <section className="auth-panel">
        <button className="primary-cta login-button" type="button" onClick={() => { window.location.href = `${apiBaseURL()}/api/v1/auth/google`; }}>
          Đăng nhập bằng Google
        </button>
      </section>
    );
  }
  if (authState.status === "authenticated") {
    return <section className="auth-panel compact-auth"><p>{authState.user.email}</p></section>;
  }
  return null;
}

function ForbiddenState({ authState, onLogout }: { authState: Extract<AuthState, { status: "forbidden" }>; onLogout: () => void }) {
  return (
    <section className="content-stack">
      <section className="card auth-state-card">
        <h1>Không có quyền truy cập</h1>
        <p>Phiên hiện tại không thể mở dữ liệu này.</p>
        {authState.correlationID ? <small>{authState.correlationID}</small> : null}
        <button className="wide-pill destructive" type="button" onClick={onLogout}>Đăng xuất</button>
      </section>
    </section>
  );
}

function TabButton({
  tab,
  active,
  onClick,
}: {
  tab: { id: Tab; label: string; icon: typeof Home };
  active: boolean;
  onClick: () => void;
}) {
  const Icon = tab.icon;
  return (
    <button className={active ? "tab-button active" : "tab-button"} aria-label={tab.label} type="button" onClick={onClick}>
      <Icon size={27} strokeWidth={2.4} />
      <span>{tab.label}</span>
    </button>
  );
}

function Overview({ wallets, onManageWallets }: { wallets: WalletSummary[] | null; onManageWallets: () => void }) {
  const visibleWallets = wallets && wallets.length > 0 ? wallets : sampleWallets;
  const totalVND = totalIncludedVND(visibleWallets);

  return (
    <section className="content-stack">
      <section className="card wallet-card">
        <div className="section-title">
          <h2>Ví của tôi</h2>
          <button type="button" onClick={onManageWallets}>Xem tất cả</button>
        </div>
        <div className="wallet-total-row"><span>Tổng hiển thị</span><strong>{formatVND(totalVND)}</strong></div>
        {visibleWallets.map((wallet) => (
          <WalletRow key={wallet.id} icon={walletIcon(wallet.type)} name={wallet.name} amount={formatVND(wallet.balance_vnd)} />
        ))}
      </section>

      <SectionHeading title="Money Insider" action="↻" />
      <section className="card insider-card">
        <h2>Ăn uống <span className="info">i</span></h2>
        <p className="muted">Tổng đã chi <strong className="expense">250.000 đ</strong></p>
        <div className="insider-grid">
          <div>
            <h3>Tháng này</h3>
            <p className="muted">Trung bình</p>
            <strong>13.157,89 đ<span>/ngày</span></strong>
          </div>
          <div className="ring">214%</div>
        </div>
        <button className="soft-cta" type="button">Dùng thử miễn phí</button>
        <button className="primary-cta" type="button">Đăng ký ngay</button>
      </section>

      <SectionHeading title="Báo cáo tháng này" action="Xem báo cáo" />
      <section className="card report-card">
        <div className="segmented"><span>Tuần</span><strong>Tháng</strong></div>
        <div className="chart">
          <span className="chart-line red" />
          <span className="chart-line gray" />
        </div>
        <div className="report-stats">
          <p>Tổng đã chi <strong className="expense">250.000 đ</strong></p>
          <p>Tổng thu <strong className="income">0 đ</strong></p>
        </div>
      </section>
    </section>
  );
}

function Transactions({ transactions, onEdit }: { transactions: Transaction[]; onEdit: (transaction: Transaction) => void }) {
  const [query, setQuery] = useState("");
  const visible = transactions.filter((transaction) => `${transaction.note} ${transaction.type}`.toLowerCase().includes(query.toLowerCase()));
  return (
    <section className="content-stack">
      <div className="sub-header">
        <h1>Sổ giao dịch</h1>
        <button className="pill-button" type="button">Tháng 08/2026</button>
      </div>
      <section className="card list-card">
        <label className="transaction-search"><Search size={18} /><input aria-label="Tìm giao dịch" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Tìm giao dịch" /></label>
        {visible.length === 0 ? <p className="empty-state">Chưa có giao dịch</p> : visible.map((transaction) => <TransactionRow key={transaction.id} title={transaction.note || transaction.type} subtitle={new Date(transaction.occurred_at).toLocaleDateString("vi-VN")} amount={signedAmount(transaction)} positive={transaction.type === "income"} onClick={() => onEdit(transaction)} />)}
      </section>
    </section>
  );
}

function Budgets({
  budgets,
  events,
  obligations,
  categories,
  transactions,
  online,
  onCreate,
  onCreateEvent,
  onCreateObligation,
  onEdit,
  onEditEvent,
  onEditObligation,
}: {
  budgets: BudgetProgress[] | null;
  events: EventSummary[];
  obligations: ObligationSummary[];
  categories: CategorySummary[];
  transactions: Transaction[];
  online: boolean;
  onCreate: () => void;
  onCreateEvent: () => void;
  onCreateObligation: () => void;
  onEdit: (budget: BudgetProgress) => void;
  onEditEvent: (event: EventSummary) => void;
  onEditObligation: (obligation: ObligationSummary) => void;
}) {
  const rows = budgets ?? [];
  const totalBudget = rows.reduce((total, item) => total + item.budget.amount_vnd, 0);
  const totalSpent = rows.reduce((total, item) => total + item.spent_vnd, 0);
  const daysLeft = rows[0] ? Math.max(0, Math.ceil((new Date(rows[0].period_end).getTime() - Date.now()) / 86400000)) : 0;
  return (
    <section className="content-stack">
      <div className="sub-header">
        <h1>Ngân sách</h1>
        <button className="pill-button" type="button" disabled={!online} onClick={onCreate}>Tạo</button>
      </div>
      <section className="card budget-hero">
        <p>{rows[0] ? `${formatDate(rows[0].period_start)} - ${formatDate(rows[0].period_end)}` : "Kỳ hiện tại"}</p>
        <strong>{formatVND(totalBudget)}</strong>
        <div className="budget-stats"><span>{formatVND(totalBudget)}<br />Tổng ngân sách</span><span>{formatVND(totalSpent)}<br />Tổng đã chi</span><span>{daysLeft} ngày<br />Còn lại</span></div>
        <button className="primary-cta compact" type="button" disabled={!online} onClick={onCreate}>Tạo Ngân sách</button>
      </section>
      {!online ? <p className="offline-warning">Cần online để tạo hoặc sửa ngân sách. Dữ liệu đã tải vẫn có thể xem.</p> : null}
      {rows.length === 0 ? <section className="card list-card"><p className="empty-state">Chưa có ngân sách</p></section> : rows.map((budget) => (
        <BudgetRow key={budget.budget.id} item={budget} categories={categories} disabled={!online} onEdit={() => onEdit(budget)} />
      ))}
      <section className="card list-card planning-list">
        <div className="section-title">
          <h2>Sự kiện</h2>
          <button type="button" disabled={!online} onClick={onCreateEvent}>Tạo sự kiện</button>
        </div>
        {events.length === 0 ? <p className="empty-state">Chưa có sự kiện</p> : events.map((event) => (
          <button className="planning-row" type="button" key={event.id} disabled={!online} onClick={() => onEditEvent(event)}>
            <span className="category-dot" />
            <div>
              <strong>{event.name}</strong>
              <p>{formatDate(event.starts_on)} - {formatDate(event.ends_on)}</p>
              <p>Đã dùng {formatVND(event.total_vnd)} · {event.transaction_count} giao dịch</p>
            </div>
            <ChevronRight size={22} />
          </button>
        ))}
      </section>
      <section className="card list-card planning-list">
        <div className="section-title">
          <h2>Khoản vay nợ</h2>
          <button type="button" disabled={!online} onClick={onCreateObligation}>Tạo khoản nợ</button>
        </div>
        {obligations.length === 0 ? <p className="empty-state">Chưa có khoản vay nợ</p> : obligations.map((obligation) => (
          <button className="planning-row" type="button" key={obligation.id} disabled={!online} onClick={() => onEditObligation(obligation)}>
            <span className="category-dot debt-dot" />
            <div>
              <strong>{obligation.counterparty}</strong>
              <p>{obligationDirectionLabel(obligation.direction)} · Hạn {formatDate(obligation.due_on)}</p>
              <p>Còn {formatVND(obligation.remaining_vnd)}</p>
              <p>Đã trả {formatVND(obligation.repaid_vnd)}</p>
            </div>
            <ChevronRight size={22} />
          </button>
        ))}
      </section>
      {transactions.length === 0 ? <p className="offline-warning neutral">Tạo giao dịch trước khi gắn chi phí sự kiện hoặc trả nợ.</p> : null}
    </section>
  );
}

function EventSheet({ event, transactions, onSaved, onArchived, onClose }: { event?: EventSummary; transactions: Transaction[]; onSaved: () => void; onArchived?: () => void; onClose: () => void }) {
  const [name, setName] = useState(event?.name ?? "");
  const [startsOn, setStartsOn] = useState(event?.starts_on ?? new Date().toISOString().slice(0, 10));
  const [endsOn, setEndsOn] = useState(event?.ends_on ?? new Date().toISOString().slice(0, 10));
  const [note, setNote] = useState(event?.note ?? "");
  const [transactionID, setTransactionID] = useState("");
  const [saving, setSaving] = useState(false);
  const canSave = name.trim() !== "" && startsOn !== "" && endsOn !== "";
  async function save() {
    if (!canSave || saving) return;
    setSaving(true);
    try {
      const input: EventInput = { name, starts_on: startsOn, ends_on: endsOn, note };
      const saved = event ? await updateEvent(event.id, input) : await createEvent(input);
      if (transactionID) await linkEventTransaction(saved.id, transactionID);
      onSaved();
    } finally {
      setSaving(false);
    }
  }
  async function archive() {
    if (!event || saving) return;
    setSaving(true);
    try {
      await archiveEvent(event.id);
      onArchived?.();
    } finally {
      setSaving(false);
    }
  }
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label={event ? "Sửa Sự Kiện" : "Tạo Sự Kiện"}>
        <header>
          <button className="pill-button" type="button" onClick={onClose}>Hủy</button>
          <h2>{event ? "Sửa Sự Kiện" : "Tạo Sự Kiện"}</h2>
          <span />
        </header>
        <label className="sheet-row"><MapPin /><input aria-label="Tên sự kiện" value={name} onChange={(change) => setName(change.target.value)} placeholder="Tên sự kiện" /></label>
        <div className="manager-form two">
          <input aria-label="Ngày bắt đầu sự kiện" type="date" value={startsOn} onChange={(change) => setStartsOn(change.target.value)} />
          <input aria-label="Ngày kết thúc sự kiện" type="date" value={endsOn} onChange={(change) => setEndsOn(change.target.value)} />
        </div>
        <label className="sheet-row"><List /><input aria-label="Ghi chú sự kiện" value={note} onChange={(change) => setNote(change.target.value)} placeholder="Ghi chú" /></label>
        <label className="sheet-row"><Wallet /><select aria-label="Giao dịch sự kiện" value={transactionID} onChange={(change) => setTransactionID(change.target.value)}><option value="">Không gắn giao dịch</option>{transactions.map((transaction) => <option key={transaction.id} value={transaction.id}>{transaction.note || transaction.type} · {formatVND(transaction.amount_vnd)}</option>)}</select></label>
        <div className="sheet-actions">
          {event ? <button className="wide-pill destructive" type="button" disabled={saving} onClick={() => void archive()}>Lưu trữ</button> : null}
          <button className="primary-cta" type="button" disabled={!canSave || saving} onClick={() => void save()}>{saving ? "Đang lưu" : "Lưu"}</button>
        </div>
      </section>
    </div>
  );
}

function ObligationSheet({ obligation, transactions, onSaved, onArchived, onClose }: { obligation?: ObligationSummary; transactions: Transaction[]; onSaved: () => void; onArchived?: () => void; onClose: () => void }) {
  const [counterparty, setCounterparty] = useState(obligation?.counterparty ?? "");
  const [direction, setDirection] = useState<ObligationDirection>(obligation?.direction ?? "borrowed");
  const [principal, setPrincipal] = useState(String(obligation?.principal_vnd ?? ""));
  const [dueOn, setDueOn] = useState(obligation?.due_on ?? new Date().toISOString().slice(0, 10));
  const [note, setNote] = useState(obligation?.note ?? "");
  const [transactionID, setTransactionID] = useState("");
  const [saving, setSaving] = useState(false);
  const canSave = counterparty.trim() !== "" && Number(principal) > 0 && dueOn !== "";
  async function save() {
    if (!canSave || saving) return;
    setSaving(true);
    try {
      const input: ObligationInput = { direction, principal_vnd: Number(principal), counterparty, due_on: dueOn, note };
      const saved = obligation ? await updateObligation(obligation.id, input) : await createObligation(input);
      if (transactionID) await linkObligationRepayment(saved.id, transactionID);
      onSaved();
    } finally {
      setSaving(false);
    }
  }
  async function archive() {
    if (!obligation || saving) return;
    setSaving(true);
    try {
      await archiveObligation(obligation.id);
      onArchived?.();
    } finally {
      setSaving(false);
    }
  }
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label={obligation ? "Sửa Khoản Nợ" : "Tạo Khoản Nợ"}>
        <header>
          <button className="pill-button" type="button" onClick={onClose}>Hủy</button>
          <h2>{obligation ? "Sửa Khoản Nợ" : "Tạo Khoản Nợ"}</h2>
          <span />
        </header>
        <div className="segmented sheet-segmented"><button type="button" className={direction === "borrowed" ? "active" : ""} onClick={() => setDirection("borrowed")}>Tôi vay</button><button type="button" className={direction === "lent" ? "active" : ""} onClick={() => setDirection("lent")}>Tôi cho vay</button></div>
        <label className="sheet-row"><Users /><input aria-label="Đối tác" value={counterparty} onChange={(change) => setCounterparty(change.target.value)} placeholder="Người liên quan" /></label>
        <label className="amount-row"><span>VND</span><input aria-label="Số tiền gốc" inputMode="numeric" value={principal} onChange={(change) => setPrincipal(change.target.value.replace(/\D/g, ""))} placeholder="0" /></label>
        <label className="sheet-row"><CalendarDays /><input aria-label="Ngày đến hạn" type="date" value={dueOn} onChange={(change) => setDueOn(change.target.value)} /></label>
        <label className="sheet-row"><List /><input aria-label="Ghi chú khoản nợ" value={note} onChange={(change) => setNote(change.target.value)} placeholder="Ghi chú" /></label>
        <label className="sheet-row"><Wallet /><select aria-label="Giao dịch trả nợ" value={transactionID} onChange={(change) => setTransactionID(change.target.value)}><option value="">Không gắn trả nợ</option>{transactions.map((transaction) => <option key={transaction.id} value={transaction.id}>{transaction.note || transaction.type} · {formatVND(transaction.amount_vnd)}</option>)}</select></label>
        {obligation ? <p className="sheet-meta">Còn {formatVND(obligation.remaining_vnd)} · Đã trả {formatVND(obligation.repaid_vnd)}</p> : null}
        <div className="sheet-actions">
          {obligation ? <button className="wide-pill destructive" type="button" disabled={saving} onClick={() => void archive()}>Lưu trữ</button> : null}
          <button className="primary-cta" type="button" disabled={!canSave || saving} onClick={() => void save()}>{saving ? "Đang lưu" : "Lưu"}</button>
        </div>
      </section>
    </div>
  );
}

function BudgetSheet({ budget, categories, onSaved, onArchived, onClose }: { budget?: BudgetProgress; categories: CategorySummary[]; onSaved: () => void; onArchived?: () => void; onClose: () => void }) {
  const editing = budget?.budget;
  const [name, setName] = useState(editing?.name ?? "");
  const [amount, setAmount] = useState(String(editing?.amount_vnd ?? ""));
  const [periodType, setPeriodType] = useState<BudgetPeriodType>(editing?.period_type ?? "monthly");
  const [allCategories, setAllCategories] = useState(editing?.all_categories ?? true);
  const [categoryIDs, setCategoryIDs] = useState<string[]>(editing?.category_ids ?? []);
  const [customStart, setCustomStart] = useState(editing?.custom_start ?? "");
  const [customEnd, setCustomEnd] = useState(editing?.custom_end ?? "");
  const [saving, setSaving] = useState(false);
  const expenseCategories = categories.filter((category) => category.kind === "expense");
  const canSave = name.trim() !== "" && Number(amount) > 0 && (periodType !== "custom" || Boolean(customStart && customEnd));
  function toggleCategory(categoryID: string) {
    setCategoryIDs((current) => current.includes(categoryID) ? current.filter((item) => item !== categoryID) : [...current, categoryID]);
  }
  function input(): BudgetInput {
    return {
      name,
      period_type: periodType,
      amount_vnd: Number(amount),
      category_ids: allCategories ? [] : categoryIDs,
      custom_start: periodType === "custom" ? customStart : undefined,
      custom_end: periodType === "custom" ? customEnd : undefined,
    };
  }
  async function save() {
    if (!canSave || saving) return;
    setSaving(true);
    try {
      if (editing) await updateBudget(editing.id, input());
      else await createBudget(input());
      onSaved();
    } finally {
      setSaving(false);
    }
  }
  async function archive() {
    if (!editing || saving) return;
    setSaving(true);
    try {
      await archiveBudget(editing.id);
      onArchived?.();
    } finally {
      setSaving(false);
    }
  }
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label={editing ? "Sửa Ngân Sách" : "Tạo Ngân Sách"}>
        <header>
          <button className="pill-button" type="button" onClick={onClose}>Hủy</button>
          <h2>{editing ? "Sửa Ngân Sách" : "Tạo Ngân Sách"}</h2>
          <span />
        </header>
        <label className="sheet-row"><List /><input aria-label="Tên ngân sách" value={name} onChange={(event) => setName(event.target.value)} placeholder="Tên ngân sách" /></label>
        <label className="amount-row"><span>VND</span><input aria-label="Số tiền ngân sách" inputMode="numeric" value={amount} onChange={(event) => setAmount(event.target.value.replace(/\D/g, ""))} placeholder="0" /></label>
        <label className="sheet-row"><CalendarDays /><select aria-label="Kỳ ngân sách" value={periodType} onChange={(event) => setPeriodType(event.target.value as BudgetPeriodType)}>{budgetPeriods.map((period) => <option key={period} value={period}>{budgetPeriodLabel(period)}</option>)}</select></label>
        {periodType === "custom" ? (
          <div className="manager-form two">
            <input aria-label="Ngày bắt đầu" type="date" value={customStart} onChange={(event) => setCustomStart(event.target.value)} />
            <input aria-label="Ngày kết thúc" type="date" value={customEnd} onChange={(event) => setCustomEnd(event.target.value)} />
          </div>
        ) : null}
        <button className={allCategories ? "toggle-row active" : "toggle-row"} type="button" onClick={() => setAllCategories((current) => !current)}>Tất cả nhóm chi<span /></button>
        {!allCategories ? (
          <div className="category-picker">
            {expenseCategories.map((category) => <button className={categoryIDs.includes(category.id) ? "mini-toggle active" : "mini-toggle"} type="button" key={category.id} onClick={() => toggleCategory(category.id)}>{category.name}</button>)}
          </div>
        ) : null}
        <div className="sheet-actions">
          {editing ? <button className="wide-pill destructive" type="button" disabled={saving} onClick={() => void archive()}>Lưu trữ</button> : null}
          <button className="primary-cta" type="button" disabled={!canSave || saving} onClick={() => void save()}>{saving ? "Đang lưu" : "Lưu"}</button>
        </div>
      </section>
    </div>
  );
}

function Account({ authState, onLogout }: { authState: AuthState; onLogout: () => void }) {
  const email = authState.status === "authenticated" ? authState.user.email : "Chưa đăng nhập";
  const displayName = authState.status === "authenticated" ? authState.user.display_name || authState.user.email : "Tài khoản MyPocket";

  return (
    <section className="content-stack">
      <div className="sub-header">
        <h1>Quản Lý Tài Khoản</h1>
      </div>
      <section className="card profile-card">
        <div className="avatar">D</div>
        <div className="ribbon">TÀI KHOẢN PREMIUM</div>
        <h2>{displayName}</h2>
        <p>{email}</p>
        <strong className="google-mark">G</strong>
      </section>
      <button className="wide-pill" type="button">Thay đổi mật khẩu</button>
      <section className="card list-card">
        <TransactionRow title="iPhone" subtitle="Thiết bị này" amount="" positive />
      </section>
      <button className="wide-pill destructive" type="button" onClick={onLogout}>Đăng xuất</button>
    </section>
  );
}

function AddTransactionSheet({ categories, wallets, readOnly, onCreated, onClose }: { categories: CategorySummary[]; wallets: WalletSummary[]; readOnly: boolean; onCreated: (transaction: Transaction) => void; onClose: () => void }) {
  const [type, setType] = useState<TransactionType>("expense");
  const [amount, setAmount] = useState("");
  const [targetBalance, setTargetBalance] = useState("");
  const [note, setNote] = useState("");
  const [sourceWalletID, setSourceWalletID] = useState(wallets[0]?.id ?? "");
  const [destinationWalletID, setDestinationWalletID] = useState(wallets.find((wallet) => wallet.id !== (wallets[0]?.id ?? ""))?.id ?? "");
  const [categoryID, setCategoryID] = useState("");
  const [excludedFromReports, setExcludedFromReports] = useState(false);
  const [saving, setSaving] = useState(false);
  const filteredCategories = categories.filter((category) => category.kind === (type === "income" ? "income" : "expense"));
  const chosenCategoryID = type === "income" || type === "expense" ? categoryID || filteredCategories[0]?.id : "";
  const canSave = !readOnly && Number(amount) > 0 && Boolean(sourceWalletID) && (type !== "transfer" || Boolean(destinationWalletID && destinationWalletID !== sourceWalletID));
  useEffect(() => {
    if (!sourceWalletID && wallets[0]) setSourceWalletID(wallets[0].id);
    if (!destinationWalletID) {
      const currentSource = sourceWalletID || wallets[0]?.id || "";
      const nextDestination = wallets.find((wallet) => wallet.id !== currentSource);
      if (nextDestination) setDestinationWalletID(nextDestination.id);
    }
  }, [destinationWalletID, sourceWalletID, wallets]);
  async function save() {
    if (!canSave || saving) return;
    setSaving(true);
    try {
      const transaction = await createTransaction(buildTransactionInput({ type, amount, targetBalance, sourceWalletID, destinationWalletID, categoryID: chosenCategoryID, note, excludedFromReports }));
      onCreated(transaction);
      onClose();
    } finally {
      setSaving(false);
    }
  }
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label="Thêm Giao Dịch">
        <header>
          <button className="pill-button" type="button" onClick={onClose}>Hủy</button>
          <h2>Thêm Giao Dịch</h2>
          <span />
        </header>
        {readOnly ? <p className="offline-warning">Offline storage chưa sẵn sàng. Mở mạng lại để lưu giao dịch.</p> : null}
        <div className="segmented sheet-segmented">
          {(["expense", "income", "transfer", "adjustment"] as TransactionType[]).map((option) => (
            <button className={type === option ? "active" : ""} type="button" key={option} onClick={() => setType(option)}>{transactionTypeLabel(option)}</button>
          ))}
        </div>
        <label className="amount-row"><span>VND</span><input aria-label="Số tiền" inputMode="numeric" value={amount} onChange={(event) => setAmount(event.target.value.replace(/\D/g, ""))} placeholder="0" /></label>
        {type === "adjustment" ? <label className="sheet-row"><Wallet /><input aria-label="Số dư mục tiêu" inputMode="numeric" value={targetBalance} onChange={(event) => setTargetBalance(event.target.value.replace(/\D/g, ""))} placeholder="Số dư sau điều chỉnh" /></label> : null}
        <label className="sheet-row"><Wallet /><select aria-label="Ví nguồn" value={sourceWalletID} onChange={(event) => setSourceWalletID(event.target.value)}>{wallets.map((wallet) => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</select></label>
        {type === "transfer" ? <label className="sheet-row"><Wallet /><select aria-label="Ví nhận" value={destinationWalletID} onChange={(event) => setDestinationWalletID(event.target.value)}>{wallets.filter((wallet) => wallet.id !== sourceWalletID).map((wallet) => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</select></label> : null}
        {type === "income" || type === "expense" ? <label className="sheet-row"><span className="dot-icon" /><select aria-label="Nhóm" value={chosenCategoryID} onChange={(event) => setCategoryID(event.target.value)}><option value="">Chọn nhóm</option>{filteredCategories.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}</select></label> : null}
        <label className="sheet-row"><List /><input aria-label="Ghi chú" value={note} onChange={(event) => setNote(event.target.value)} placeholder="Ghi chú" /></label>
        <SheetRow icon={<CalendarDays />} label="Chủ Nhật, 23/08/2026" green />
        <div className="sheet-group">
          <SheetRow icon={<Users />} label="Với" muted />
        </div>
        <div className="sheet-group">
          <SheetRow icon={<MapPin />} label="Đặt vị trí" muted />
          <SheetRow icon={<BriefcaseBusiness />} label="Chọn sự kiện" muted />
          <SheetRow icon={<Bell />} label="Đặt nhắc nhở" muted />
        </div>
        <button className="image-row" type="button"><ImagePlus size={28} />Thêm Hình Ảnh</button>
        <button className={excludedFromReports ? "toggle-row active" : "toggle-row"} type="button" onClick={() => setExcludedFromReports((current) => !current)}>Không tính vào báo cáo<span /></button>
        <div className="save-bar"><button type="button" disabled={!canSave || saving} onClick={() => void save()}>{saving ? "Đang lưu" : "Lưu"}</button><button type="button" className="receipt"><ImagePlus size={24} /></button></div>
      </section>
    </div>
  );
}

function EditTransactionSheet({ categories, wallets, readOnly, transaction, onChanged, onArchived, onClose }: { categories: CategorySummary[]; wallets: WalletSummary[]; readOnly: boolean; transaction: Transaction; onChanged: (transaction: Transaction) => void; onArchived: () => void; onClose: () => void }) {
  const [amount, setAmount] = useState(String(transaction.amount_vnd));
  const [note, setNote] = useState(transaction.note);
  const [excludedFromReports, setExcludedFromReports] = useState(transaction.excluded_from_reports);
  const [saving, setSaving] = useState(false);
  const category = categories.find((item) => item.id === transaction.category_id);
  const sourceWallet = wallets.find((item) => item.id === transaction.source_wallet_id);
  async function save() {
    if (readOnly || Number(amount) <= 0 || saving) return;
    setSaving(true);
    try {
      const next = await updateTransaction(transaction.id, {
        type: transaction.type,
        source_wallet_id: transaction.source_wallet_id,
        destination_wallet_id: transaction.destination_wallet_id,
        category_id: transaction.category_id,
        amount_vnd: Number(amount),
        occurred_at: transaction.occurred_at,
        note,
        with_person: transaction.with_person,
        event_ref: transaction.event_ref,
        excluded_from_reports: excludedFromReports,
        base_version: transaction.version,
      });
      onChanged(next);
      onClose();
    } finally {
      setSaving(false);
    }
  }
  async function archive() {
    if (readOnly || saving) return;
    setSaving(true);
    try {
      await archiveTransaction(transaction.id, transaction.version);
      onArchived();
    } finally {
      setSaving(false);
    }
  }
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label="Sửa Giao Dịch">
        <header>
          <button className="pill-button" type="button" onClick={onClose}>Hủy</button>
          <h2>Sửa Giao Dịch</h2>
          <span />
        </header>
        {readOnly ? <p className="offline-warning">Offline storage chưa sẵn sàng. Mở mạng lại để sửa giao dịch.</p> : null}
        <p className="sheet-meta">{transactionTypeLabel(transaction.type)} · {sourceWallet?.name ?? "Ví"} · {category?.name ?? "Không nhóm"}</p>
        <label className="amount-row"><span>VND</span><input aria-label="Số tiền" inputMode="numeric" value={amount} onChange={(event) => setAmount(event.target.value.replace(/\D/g, ""))} placeholder="0" disabled={readOnly} /></label>
        <label className="sheet-row"><List /><input aria-label="Ghi chú" value={note} onChange={(event) => setNote(event.target.value)} placeholder="Ghi chú" disabled={readOnly} /></label>
        <button className={excludedFromReports ? "toggle-row active" : "toggle-row"} type="button" disabled={readOnly} onClick={() => setExcludedFromReports((current) => !current)}>Không tính vào báo cáo<span /></button>
        <div className="sheet-actions">
          <button className="wide-pill destructive" type="button" disabled={readOnly || saving} onClick={() => void archive()}>Lưu trữ</button>
          <button className="primary-cta" type="button" disabled={readOnly || saving || Number(amount) <= 0} onClick={() => void save()}>{saving ? "Đang lưu" : "Lưu thay đổi"}</button>
        </div>
      </section>
    </div>
  );
}

function WalletManagerSheet({
  categories,
  wallets,
  readOnly,
  onWalletChanged,
  onWalletArchived,
  onCategoryChanged,
  onCategoryArchived,
  onChanged,
  onClose,
}: {
  categories: CategorySummary[];
  wallets: WalletSummary[];
  readOnly: boolean;
  onWalletChanged: (wallet: WalletSummary) => void;
  onWalletArchived: (walletID: string) => void;
  onCategoryChanged: (category: CategorySummary) => void;
  onCategoryArchived: (categoryID: string) => void;
  onChanged: () => void;
  onClose: () => void;
}) {
  const [walletName, setWalletName] = useState("");
  const [walletType, setWalletType] = useState<WalletType>("cash");
  const [categoryName, setCategoryName] = useState("");
  const [categoryKind, setCategoryKind] = useState<CategorySummary["kind"]>("expense");
  const [busy, setBusy] = useState(false);
  async function run(action: () => Promise<unknown>) {
    if (busy) return;
    setBusy(true);
    try {
      await action();
      onChanged();
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label="Quản lý ví và nhóm">
        <header>
          <button className="pill-button" type="button" onClick={onClose}>Đóng</button>
          <h2>Ví và nhóm</h2>
          <span />
        </header>
        {readOnly ? <p className="offline-warning">Offline storage chưa sẵn sàng. Mở mạng lại để chỉnh ví và nhóm.</p> : null}
        <section className="manager-section">
          <h3>Ví</h3>
          <div className="manager-form">
            <input aria-label="Tên ví mới" value={walletName} onChange={(event) => setWalletName(event.target.value)} placeholder="Tên ví mới" disabled={readOnly} />
            <select aria-label="Loại ví" value={walletType} onChange={(event) => setWalletType(event.target.value as WalletType)} disabled={readOnly}>{walletTypes.map((type) => <option key={type} value={type}>{walletTypeLabel(type)}</option>)}</select>
            <button type="button" disabled={readOnly || !walletName.trim() || busy} onClick={() => void run(async () => { const wallet = await createWallet({ name: walletName, type: walletType }); if (wallet) onWalletChanged(wallet); setWalletName(""); })}>Tạo ví</button>
          </div>
          {wallets.map((wallet) => <WalletManageRow key={wallet.id} wallet={wallet} wallets={wallets} readOnly={readOnly} busy={busy} onWalletChanged={onWalletChanged} onWalletArchived={onWalletArchived} onChanged={run} />)}
        </section>
        <section className="manager-section">
          <h3>Nhóm</h3>
          <div className="manager-form">
            <input aria-label="Tên nhóm mới" value={categoryName} onChange={(event) => setCategoryName(event.target.value)} placeholder="Tên nhóm mới" disabled={readOnly} />
            <select aria-label="Loại nhóm" value={categoryKind} onChange={(event) => setCategoryKind(event.target.value as CategorySummary["kind"])} disabled={readOnly}><option value="expense">Chi</option><option value="income">Thu</option><option value="debt">Nợ</option></select>
            <button type="button" disabled={readOnly || !categoryName.trim() || busy} onClick={() => void run(async () => { const category = await createCategory({ name: categoryName, kind: categoryKind }); if (category) onCategoryChanged(category); setCategoryName(""); })}>Tạo nhóm</button>
          </div>
          {categories.map((category) => <CategoryManageRow key={category.id} category={category} walletID={wallets[0]?.id ?? ""} readOnly={readOnly} busy={busy} onCategoryChanged={onCategoryChanged} onCategoryArchived={onCategoryArchived} onChanged={run} />)}
        </section>
      </section>
    </div>
  );
}

function WalletManageRow({
  wallet,
  wallets,
  readOnly,
  busy,
  onWalletChanged,
  onWalletArchived,
  onChanged,
}: {
  wallet: WalletSummary;
  wallets: WalletSummary[];
  readOnly: boolean;
  busy: boolean;
  onWalletChanged: (wallet: WalletSummary) => void;
  onWalletArchived: (walletID: string) => void;
  onChanged: (action: () => Promise<unknown>) => Promise<void>;
}) {
  const [name, setName] = useState(wallet.name);
  const [include, setInclude] = useState(wallet.include_in_total);
  return (
    <div className="manager-row">
      <input aria-label={`Tên ví ${wallet.name}`} value={name} onChange={(event) => setName(event.target.value)} disabled={readOnly} />
      <button type="button" disabled={readOnly} className={include ? "mini-toggle active" : "mini-toggle"} onClick={() => setInclude((current) => !current)}>{include ? "Tổng" : "Ẩn"}</button>
      <button type="button" disabled={readOnly || busy || !name.trim()} onClick={() => void onChanged(async () => { const next = await updateWallet(wallet.id, { name, include_in_total: include, base_version: wallet.version, current_wallet: wallet }); if (next) onWalletChanged(next); })}>Lưu</button>
      <button type="button" disabled={readOnly || busy} onClick={() => void onChanged(async () => { await setDefaultAIWallet(wallet.id, wallets, wallet.version); wallets.forEach((item) => onWalletChanged({ ...item, is_default_ai: item.id === wallet.id })); })}>{wallet.is_default_ai ? "AI" : "Đặt AI"}</button>
      <button type="button" className="danger-text" disabled={readOnly || busy} onClick={() => void onChanged(async () => { await archiveWallet(wallet.id, wallet.version); onWalletArchived(wallet.id); })}>Ẩn</button>
    </div>
  );
}

function CategoryManageRow({
  category,
  walletID,
  readOnly,
  busy,
  onCategoryChanged,
  onCategoryArchived,
  onChanged,
}: {
  category: CategorySummary;
  walletID: string;
  readOnly: boolean;
  busy: boolean;
  onCategoryChanged: (category: CategorySummary) => void;
  onCategoryArchived: (categoryID: string) => void;
  onChanged: (action: () => Promise<unknown>) => Promise<void>;
}) {
  const [name, setName] = useState(category.name);
  const [active, setActive] = useState(true);
  return (
    <div className="manager-row">
      <input aria-label={`Tên nhóm ${category.name}`} value={name} onChange={(event) => setName(event.target.value)} disabled={readOnly || category.is_system} />
      <button type="button" disabled={readOnly || !walletID || busy} className={active ? "mini-toggle active" : "mini-toggle"} onClick={() => { const next = !active; setActive(next); void onChanged(() => setWalletCategoryActive(walletID, category.id, next)); }}>{active ? "Bật" : "Tắt"}</button>
      <button type="button" disabled={readOnly || busy || category.is_system || !name.trim()} onClick={() => void onChanged(async () => { const next = await updateCategory(category.id, { name, current_category: category }); if (next) onCategoryChanged(next); })}>Lưu</button>
      <button type="button" className="danger-text" disabled={readOnly || busy || category.is_system} onClick={() => void onChanged(async () => { await archiveCategory(category.id); onCategoryArchived(category.id); })}>Ẩn</button>
    </div>
  );
}

function WalletRow({ icon, name, amount }: { icon: string; name: string; amount: string }) {
  return <div className="wallet-row"><span>{icon}</span><strong>{name}</strong><b>{amount}</b></div>;
}

function TransactionRow({ title, subtitle, amount, positive = false, onClick }: { title: string; subtitle: string; amount: string; positive?: boolean; onClick?: () => void }) {
  return <button className="transaction-row" type="button" onClick={onClick}><span className="category-dot" /><div><strong>{title}</strong><p>{subtitle}</p></div><b className={positive ? "income" : "expense"}>{amount}</b><ChevronRight size={22} /></button>;
}

function BudgetRow({ item, categories, disabled, onEdit }: { item: BudgetProgress; categories: CategorySummary[]; disabled: boolean; onEdit: () => void }) {
  const categoryNames = item.budget.all_categories ? "Tất cả nhóm chi" : (item.budget.category_ids ?? []).map((id) => categories.find((category) => category.id === id)?.name ?? "Nhóm").join(", ");
  const progress = Math.min(100, item.percent);
  return (
    <button className="card budget-row" type="button" disabled={disabled} onClick={onEdit}>
      <div><span className="category-dot" /><strong>{item.budget.name}</strong></div>
      <b>{formatVND(item.budget.amount_vnd)}</b>
      <p>{categoryNames}</p>
      <p>Đã chi {formatVND(item.spent_vnd)} · Còn {formatVND(item.remaining_vnd)}</p>
      <span className="progress"><i style={{ width: `${progress}%` }} /></span>
      {item.alert_100 ? <small className="budget-alert">Đã vượt 100%</small> : item.alert_80 ? <small className="budget-alert">Đã chạm 80%</small> : null}
    </button>
  );
}

function SheetRow({ icon, label, muted = false, green = false }: { icon: ReactNode; label: string; muted?: boolean; green?: boolean }) {
  return <button className={muted ? "sheet-row muted" : green ? "sheet-row green" : "sheet-row"} type="button">{icon}<span>{label}</span><ChevronRight size={22} /></button>;
}

function SectionHeading({ title, action }: { title: string; action: string }) {
  return <div className="section-heading"><h2>{title}</h2><button type="button">{action}</button></div>;
}

const sampleWallets: WalletSummary[] = [
  { id: "sample-credit", name: "Ví tín dụng", type: "credit", balance_vnd: -3711104, include_in_total: true, is_default_ai: false, version: 1 },
  { id: "sample-bank", name: "Techcombank", type: "bank", balance_vnd: 4710200, include_in_total: true, is_default_ai: false, version: 1 },
  { id: "sample-cash", name: "Tiền Mặt", type: "cash", balance_vnd: 84000, include_in_total: true, is_default_ai: true, version: 1 },
];

function formatVND(amount: number) {
  return `${new Intl.NumberFormat("vi-VN").format(amount)} đ`;
}

function formatDate(value: string) {
  return new Date(`${value}T00:00:00+07:00`).toLocaleDateString("vi-VN");
}

function totalIncludedVND(wallets: WalletSummary[]) {
  return wallets.filter((wallet) => wallet.include_in_total).reduce((total, wallet) => total + wallet.balance_vnd, 0);
}

function walletIcon(type: WalletSummary["type"]) {
  switch (type) {
    case "credit":
      return "💳";
    case "bank":
      return "◆";
    case "e_wallet":
      return "◎";
    case "savings":
      return "◇";
    case "debt":
      return "!";
    case "cash":
    default:
      return "₫";
  }
}

const walletTypes: WalletType[] = ["cash", "bank", "credit", "e_wallet", "savings", "debt"];
const budgetPeriods: BudgetPeriodType[] = ["weekly", "monthly", "quarterly", "yearly", "custom"];

function budgetPeriodLabel(period: BudgetPeriodType) {
  switch (period) {
    case "weekly":
      return "Hàng tuần";
    case "monthly":
      return "Hàng tháng";
    case "quarterly":
      return "Hàng quý";
    case "yearly":
      return "Hàng năm";
    case "custom":
      return "Tùy chọn";
  }
}

function buildTransactionInput({
  type,
  amount,
  targetBalance,
  sourceWalletID,
  destinationWalletID,
  categoryID,
  note,
  excludedFromReports,
}: {
  type: TransactionType;
  amount: string;
  targetBalance: string;
  sourceWalletID: string;
  destinationWalletID: string;
  categoryID: string;
  note: string;
  excludedFromReports: boolean;
}): TransactionInput {
  return {
    type,
    source_wallet_id: sourceWalletID,
    destination_wallet_id: type === "transfer" ? destinationWalletID : undefined,
    category_id: type === "income" || type === "expense" ? categoryID : undefined,
    amount_vnd: Number(amount),
    target_balance_vnd: type === "adjustment" && targetBalance ? Number(targetBalance) : null,
    occurred_at: new Date().toISOString(),
    note,
    excluded_from_reports: excludedFromReports,
  };
}

function signedAmount(transaction: Transaction) {
  if (transaction.type === "income") return `+${formatVND(transaction.amount_vnd)}`;
  if (transaction.type === "adjustment") return formatVND(transaction.amount_vnd);
  return `-${formatVND(transaction.amount_vnd)}`;
}

function transactionTypeLabel(type: TransactionType) {
  switch (type) {
    case "income":
      return "Thu";
    case "transfer":
      return "Chuyển";
    case "adjustment":
      return "Điều chỉnh";
    case "expense":
    default:
      return "Chi";
  }
}

function obligationDirectionLabel(direction: ObligationDirection) {
  return direction === "borrowed" ? "Tôi vay" : "Tôi cho vay";
}

function conflictLabel(conflict: OfflineConflict) {
  const entity = conflict.entity_type === "transaction" ? "giao dịch" : conflict.entity_type === "wallet" ? "ví" : "nhóm";
  const action = conflict.operation === "archive" ? "ẩn" : conflict.operation === "create" ? "tạo" : "sửa";
  return `Xung đột ${action} ${entity}`;
}

function walletTypeLabel(type: WalletType) {
  switch (type) {
    case "bank":
      return "Ngân hàng";
    case "credit":
      return "Tín dụng";
    case "e_wallet":
      return "Ví điện tử";
    case "savings":
      return "Tiết kiệm";
    case "debt":
      return "Nợ";
    case "cash":
    default:
      return "Tiền mặt";
  }
}

export default App;
