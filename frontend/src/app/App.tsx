import { useEffect, useState, type ReactNode } from "react";
import {
  Bell,
  Activity,
  BookOpen,
  BriefcaseBusiness,
  CalendarDays,
  ChevronRight,
  Eye,
  Home,
  ImagePlus,
  Info,
  KeyRound,
  Link,
  List,
  MapPin,
  Plus,
  RefreshCw,
  Search,
  User,
  Users,
  Wallet,
} from "lucide-react";

import { useOnlineStatus } from "./offline";
import { apiBaseURL } from "./apiClient";
import { Card, IconButton, PillButton, SectionTitle, SheetFrame, cx } from "./components";
import { loadCurrentUser, logout, type AuthState } from "./auth";
import { clearOfflineStore, initializeOfflineStore, queueReceiptUpload, saveFinanceMirror } from "../offline/db";
import { listPendingReceiptUploads, markReceiptUploadComplete } from "../offline/receipts";
import {
  discardLocalConflict,
  editAndRetryTransactionConflict,
  fullResync,
  keepServerConflict,
  listOpenConflicts,
} from "../offline/conflicts";
import {
  archiveTransaction,
  archiveWallet,
  createTransaction,
  createWallet,
  loadCategories,
  loadTransactions,
  loadWallets,
  setDefaultAIWallet,
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
  archiveRecurringSchedule,
  archiveObligation,
  createBudget,
  createEvent,
  createRecurringSchedule,
  createObligation,
  linkEventTransaction,
  linkObligationRepayment,
  loadBudgets,
  loadEvents,
  loadRecurringSchedules,
  loadObligations,
  loadTransactionDrafts,
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
  type RecurrenceFrequency,
  type RecurringSchedule,
  type RecurringScheduleInput,
  type TransactionDraft,
} from "./planning";
import { drainOutbox, readOutbox } from "./outbox";
import { loadNotifications, markNotificationRead, subscribeToPush, type NotificationNotice } from "./notifications";
import { loadDashboard, loadInsider, loadReport, loadWalletDetail, searchRecords, type Dashboard, type InsiderReport, type Report, type SearchResult, type WalletDetail } from "./analytics";
import { addAssetPrice, addAssetTrade, archiveAsset, createAsset, loadAssets, loadPortfolioSummary, type AssetPosition, type AssetType, type PortfolioSummary, type TradeSide } from "./portfolio";
import { createAPIKey, loadAPIKeys, revokeAPIKey, type APIKeySummary, type CreatedAPIKey } from "./apiKeys";
import { checkAuditAccess, loadAuditEvents, type AuditEvent } from "./audit";
import { uploadReceipt } from "./receipts";
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
  const [schedules, setSchedules] = useState<RecurringSchedule[] | null>(null);
  const [drafts, setDrafts] = useState<TransactionDraft[] | null>(null);
  const [notifications, setNotifications] = useState<NotificationNotice[] | null>(null);
  const [notificationOpen, setNotificationOpen] = useState(false);
  const [pushState, setPushState] = useState<"idle" | "enabled" | "denied" | "unsupported" | "offline" | "failed">("idle");
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [report, setReport] = useState<Report | null>(null);
  const [insider, setInsider] = useState<InsiderReport | null>(null);
  const [privacyMasked, setPrivacyMasked] = useState(() => localStorage.getItem("mypocket:privacy-masked") === "true");
  const [searchOpen, setSearchOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [walletDetail, setWalletDetail] = useState<WalletDetail | null>(null);
  const [assets, setAssets] = useState<AssetPosition[] | null>(null);
  const [portfolioSummary, setPortfolioSummary] = useState<PortfolioSummary | null>(null);
  const [assetSheetOpen, setAssetSheetOpen] = useState(false);
  const [budgetSheetOpen, setBudgetSheetOpen] = useState(false);
  const [eventSheetOpen, setEventSheetOpen] = useState(false);
  const [obligationSheetOpen, setObligationSheetOpen] = useState(false);
  const [scheduleSheetOpen, setScheduleSheetOpen] = useState(false);
  const [editingBudget, setEditingBudget] = useState<BudgetProgress | null>(null);
  const [editingEvent, setEditingEvent] = useState<EventSummary | null>(null);
  const [editingObligation, setEditingObligation] = useState<ObligationSummary | null>(null);
  const [editingSchedule, setEditingSchedule] = useState<RecurringSchedule | null>(null);
  const [offlineStatus, setOfflineStatus] = useState<{ mode: "ready" | "degraded"; pending: number; reason?: string }>({ mode: "ready", pending: 0 });
  const [conflicts, setConflicts] = useState<OfflineConflict[]>([]);
  const offlineReadOnly = !online && offlineStatus.mode === "degraded";
  const headerWallets = wallets ?? [];
  const totalBalance = dashboard?.net_worth_vnd ?? totalIncludedVND(headerWallets);

  useEffect(() => {
    if (window.location.pathname === "/auth/google") {
      window.location.replace(`${apiBaseURL()}/api/v1/auth/google`);
      return;
    }
    let cancelled = false;
    void loadCurrentUser().then((nextAuthState) => {
      if (!cancelled) {
        setAuthState(nextAuthState);
        // Accept a backend/provider redirect landing on the explicit FE
        // callback path, then return to the app once the cookie is verified.
        if (nextAuthState.status === "authenticated" && window.location.pathname === "/auth/callback") {
          window.history.replaceState({}, "", "/");
        }
      }
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
      setSchedules(null);
      setDrafts(null);
      setNotifications(null);
      setNotificationOpen(false);
      setDashboard(null);
      setReport(null);
      setInsider(null);
      setWalletDetail(null);
      setAssets(null);
      setPortfolioSummary(null);
      setConflicts([]);
      return;
    }
    let cancelled = false;
    const load = online ? refreshFinanceData : hydrateOfflineData;
    void load().catch(() => {
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
        .then(() => drainPendingReceiptUploads())
        .then(() => refreshConflictState())
        .catch(() => undefined);
    });
  }, [authState.status, online]);

  async function drainPendingReceiptUploads() {
    if (!online) return;
    const records = await listPendingReceiptUploads();
    if (records.length === 0) return;
    const currentTransactions = await loadTransactions();
    for (const record of records) {
      const current = currentTransactions.find((item) => item.id === record.transaction_id);
      if (!current) continue;
      const receipt = await uploadReceipt(new File([record.file], record.filename, { type: record.content_type }));
      await updateTransaction(current.id, {
        type: current.type,
        source_wallet_id: current.source_wallet_id,
        destination_wallet_id: current.destination_wallet_id,
        category_id: current.category_id,
        receipt_object_id: receipt.id,
        amount_vnd: current.amount_vnd,
        occurred_at: current.occurred_at,
        note: current.note,
        with_person: current.with_person,
        event_ref: current.event_ref,
        excluded_from_reports: current.excluded_from_reports,
        base_version: current.version,
      });
      await markReceiptUploadComplete(record);
    }
  }

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
    setSchedules(null);
    setDrafts(null);
    setNotifications(null);
    setNotificationOpen(false);
    setDashboard(null);
    setReport(null);
    setInsider(null);
    setWalletDetail(null);
    setAssets(null);
    setPortfolioSummary(null);
    setConflicts([]);
    setOfflineStatus({ mode: "ready", pending: 0 });
    setActiveTab("overview");
  }

  async function refreshFinanceData() {
    const [nextWallets, nextCategories, nextTransactions] = await Promise.all([loadWallets(), loadCategories(), loadTransactions()]);
    setWallets(nextWallets);
    setCategories(nextCategories);
    const [nextBudgets, nextEvents, nextObligations, nextSchedules, nextDrafts, nextNotifications, nextDashboard, nextReport, nextInsider, nextAssets, nextPortfolioSummary] = await Promise.all([loadBudgets(), loadEvents(), loadObligations(), loadRecurringSchedules(), loadTransactionDrafts(), loadNotifications().catch(() => []), loadDashboard().catch(() => null), loadReport("daily").catch(() => null), loadInsider().catch(() => null), loadAssets().catch(() => []), loadPortfolioSummary().catch(() => null)]);
    setWallets(nextWallets);
    setCategories(nextCategories);
    setBudgets(nextBudgets);
    setEvents(nextEvents);
    setObligations(nextObligations);
    setSchedules(nextSchedules);
    setDrafts(nextDrafts);
    setNotifications(nextNotifications);
    setDashboard(nextDashboard);
    setReport(nextReport);
    setInsider(nextInsider);
    setAssets(nextAssets);
    setPortfolioSummary(nextPortfolioSummary);
    await saveFinanceMirror({ wallets: nextWallets, categories: nextCategories, transactions: nextTransactions, assets: nextAssets });
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
    if (snapshot.assets.length > 0) setAssets(snapshot.assets);
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

  if (authState.status !== "authenticated") {
    return <AuthGate authState={authState} />;
  }

  return (
    <div className="app-shell">
      <main className="phone-frame">
        <header className="home-header">
          <div>
            <div className="balance-line">
              <strong>{privacyMasked ? "••••••" : formatVND(totalBalance)}</strong>
              <button className="icon-button" aria-label={privacyMasked ? "Hiện số dư" : "Ẩn số dư"} type="button" onClick={() => setPrivacyMasked((masked) => { const next = !masked; localStorage.setItem("mypocket:privacy-masked", String(next)); return next; })}>
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
            <IconButton aria-label="Tìm kiếm" onClick={() => setSearchOpen((open) => !open)}>
              <Search size={30} />
            </IconButton>
            <IconButton className="notification" aria-label="Thông báo" onClick={() => setNotificationOpen((open) => !open)}>
              <Bell size={30} />
              {notifications?.filter((item) => !item.read_at).length ? <span>{notifications.filter((item) => !item.read_at).length}</span> : null}
            </IconButton>
          </div>
        </header>

        {searchOpen ? <SearchPanel online={online} query={searchQuery} results={searchResults} onQueryChange={(query) => { setSearchQuery(query); if (!online || query.trim().length < 2) { setSearchResults([]); return; } void searchRecords(query).then(setSearchResults).catch(() => setSearchResults([])); }} /> : null}

        {notificationOpen ? <NotificationInbox online={online} notices={notifications ?? []} pushState={pushState} onEnablePush={() => { void subscribeToPush().then(() => setPushState("enabled")).catch((error: Error) => setPushState(error.message === "denied" ? "denied" : error.message === "unsupported" || error.message === "unconfigured" ? "unsupported" : error.message === "offline" ? "offline" : "failed")); }} onRead={(id) => { void markNotificationRead(id).then(() => setNotifications((current) => (current ?? []).map((item) => item.id === id ? { ...item, read_at: new Date().toISOString() } : item))).catch(() => undefined); }} /> : null}

        <AuthBanner authState={authState} />
        {authState.status === "authenticated" ? <ConflictInbox conflicts={conflicts} onResolve={handleConflictAction} /> : null}
        {authState.status === "forbidden" ? <ForbiddenState authState={authState} onLogout={handleLogout} /> : null}
        {authState.status !== "forbidden" && activeTab === "overview" ? <Overview online={online} wallets={wallets} dashboard={dashboard} report={report} insider={insider} privacyMasked={privacyMasked} walletDetail={walletDetail} onCloseWalletDetail={() => setWalletDetail(null)} onManageWallets={() => setWalletSheetOpen(true)} onRefreshInsider={() => void loadInsider().then(setInsider).catch(() => undefined)} /> : null}
        {authState.status !== "forbidden" && activeTab === "transactions" ? <Transactions transactions={transactions ?? []} onEdit={setEditingTransaction} /> : null}
        {authState.status !== "forbidden" && activeTab === "budgets" ? <Budgets budgets={budgets} events={events ?? []} obligations={obligations ?? []} schedules={schedules ?? []} drafts={drafts ?? []} categories={categories ?? []} wallets={wallets ?? []} transactions={transactions ?? []} online={online} onCreate={() => setBudgetSheetOpen(true)} onCreateEvent={() => setEventSheetOpen(true)} onCreateObligation={() => setObligationSheetOpen(true)} onCreateSchedule={() => setScheduleSheetOpen(true)} onEdit={setEditingBudget} onEditEvent={setEditingEvent} onEditObligation={setEditingObligation} onEditSchedule={setEditingSchedule} /> : null}
        {authState.status !== "forbidden" && activeTab === "account" ? <Account authState={authState} assets={assets ?? []} portfolioSummary={portfolioSummary} privacyMasked={privacyMasked} online={online} onCreateAsset={() => setAssetSheetOpen(true)} onAssetChanged={(asset) => { setAssets((current) => [asset, ...(current ?? []).filter((item) => item.id !== asset.id)]); void refreshFinanceData(); }} onAssetArchived={(assetID) => { setAssets((current) => (current ?? []).filter((item) => item.id !== assetID)); void refreshFinanceData(); }} onLogout={handleLogout} /> : null}
      </main>

      <nav className="bottom-nav" aria-label="Điều hướng chính">
        {tabs.slice(0, 2).map((tab) => (
          <TabButton key={tab.id} tab={tab} active={activeTab === tab.id} onClick={() => setActiveTab(tab.id)} />
        ))}
        <button className="add-button" aria-label="Thêm giao dịch" type="button" disabled={(wallets ?? []).length === 0} onClick={() => setSheetOpen(true)}>
          <Plus size={36} />
        </button>
        {tabs.slice(2).map((tab) => (
          <TabButton key={tab.id} tab={tab} active={activeTab === tab.id} onClick={() => setActiveTab(tab.id)} />
        ))}
      </nav>

      {sheetOpen ? <AddTransactionSheet categories={categories ?? []} wallets={wallets ?? []} readOnly={offlineReadOnly} onCreated={upsertTransaction} onDebtCreated={() => void refreshFinanceData()} onClose={() => setSheetOpen(false)} /> : null}
      {walletSheetOpen ? <WalletManagerSheet wallets={wallets ?? []} readOnly={offlineReadOnly} onWalletChanged={(wallet) => setWallets((current) => [wallet, ...(current ?? []).filter((item) => item.id !== wallet.id)])} onWalletArchived={(walletID) => setWallets((current) => (current ?? []).filter((item) => item.id !== walletID))} onChanged={() => void reconcileAfterLocalChange().catch(() => undefined)} onClose={() => setWalletSheetOpen(false)} /> : null}
      {editingTransaction ? <EditTransactionSheet categories={categories ?? []} wallets={wallets ?? []} readOnly={offlineReadOnly} transaction={editingTransaction} onChanged={upsertTransaction} onArchived={() => { setTransactions((current) => (current ?? []).filter((item) => item.id !== editingTransaction.id)); setEditingTransaction(null); void reconcileAfterLocalChange().catch(() => undefined); }} onClose={() => setEditingTransaction(null)} /> : null}
      {budgetSheetOpen ? <BudgetSheet categories={categories ?? []} onSaved={() => { setBudgetSheetOpen(false); void refreshFinanceData(); }} onClose={() => setBudgetSheetOpen(false)} /> : null}
      {editingBudget ? <BudgetSheet budget={editingBudget} categories={categories ?? []} onSaved={() => { setEditingBudget(null); void refreshFinanceData(); }} onArchived={() => { setEditingBudget(null); void refreshFinanceData(); }} onClose={() => setEditingBudget(null)} /> : null}
      {eventSheetOpen ? <EventSheet transactions={transactions ?? []} onSaved={() => { setEventSheetOpen(false); void refreshFinanceData(); }} onClose={() => setEventSheetOpen(false)} /> : null}
      {editingEvent ? <EventSheet event={editingEvent} transactions={transactions ?? []} onSaved={() => { setEditingEvent(null); void refreshFinanceData(); }} onArchived={() => { setEditingEvent(null); void refreshFinanceData(); }} onClose={() => setEditingEvent(null)} /> : null}
      {obligationSheetOpen ? <ObligationSheet transactions={transactions ?? []} onSaved={() => { setObligationSheetOpen(false); void refreshFinanceData(); }} onClose={() => setObligationSheetOpen(false)} /> : null}
      {editingObligation ? <ObligationSheet obligation={editingObligation} transactions={transactions ?? []} onSaved={() => { setEditingObligation(null); void refreshFinanceData(); }} onArchived={() => { setEditingObligation(null); void refreshFinanceData(); }} onClose={() => setEditingObligation(null)} /> : null}
      {scheduleSheetOpen ? <ScheduleSheet wallets={wallets ?? []} categories={categories ?? []} onSaved={() => { setScheduleSheetOpen(false); void refreshFinanceData(); }} onClose={() => setScheduleSheetOpen(false)} /> : null}
      {editingSchedule ? <ScheduleSheet schedule={editingSchedule} wallets={wallets ?? []} categories={categories ?? []} onSaved={() => { setEditingSchedule(null); void refreshFinanceData(); }} onArchived={() => { setEditingSchedule(null); void refreshFinanceData(); }} onClose={() => setEditingSchedule(null)} /> : null}
      {assetSheetOpen ? <AssetSheet onSaved={(asset) => { setAssetSheetOpen(false); setAssets((current) => [asset, ...(current ?? [])]); void refreshFinanceData(); }} onClose={() => setAssetSheetOpen(false)} /> : null}
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

function NotificationInbox({
  online,
  notices,
  pushState,
  onEnablePush,
  onRead,
}: {
  online: boolean;
  notices: NotificationNotice[];
  pushState: "idle" | "enabled" | "denied" | "unsupported" | "offline" | "failed";
  onEnablePush: () => void;
  onRead: (id: string) => void;
}) {
  const pushMessage = pushState === "denied" ? "Bạn đã từ chối quyền thông báo" : pushState === "unsupported" ? "Trình duyệt chưa hỗ trợ Web Push" : pushState === "offline" ? "Kết nối mạng để bật Web Push" : pushState === "failed" ? "Không thể bật Web Push lúc này" : pushState === "enabled" ? "Web Push đã bật" : "";
  return (
    <section className="card notification-inbox" role="dialog" aria-modal="false" aria-label="Hộp thư thông báo">
      <div className="section-title"><h2>Thông báo</h2><button type="button" onClick={onEnablePush} disabled={!online || pushState === "enabled"}>Bật Web Push</button></div>
      {pushMessage ? <p className="notification-status">{pushMessage}</p> : null}
      {!online && notices.length === 0 ? <p className="notification-status">Đang offline. Hộp thư sẽ tải lại khi có mạng.</p> : null}
      {notices.length === 0 && online ? <p className="notification-status">Chưa có thông báo mới.</p> : null}
      {notices.map((notice) => <button className={notice.read_at ? "notice-row read" : "notice-row"} key={notice.id} type="button" onClick={() => !notice.read_at && onRead(notice.id)}><span><strong>{notice.title}</strong><small>{notice.body}</small></span><time>{new Date(notice.created_at).toLocaleDateString("vi-VN")}</time></button>)}
    </section>
  );
}

function SearchPanel({ online, query, results, onQueryChange }: { online: boolean; query: string; results: SearchResult[]; onQueryChange: (query: string) => void }) {
  return <section className="card search-panel" aria-label="Tìm kiếm"><input autoFocus aria-label="Tìm kiếm giao dịch và ví" placeholder="Tìm ví, giao dịch, nhóm..." value={query} onChange={(event) => onQueryChange(event.target.value)} />{!online ? <p className="notification-status">Kết quả offline có thể cũ và chỉ đọc.</p> : null}{query.trim().length >= 2 && results.length === 0 ? <p className="notification-status">Không tìm thấy kết quả.</p> : null}{results.map((item) => <div className="search-result" key={`${item.kind}-${item.id}`}><strong>{item.label || "Không có ghi chú"}</strong><small>{item.detail ?? item.kind}</small></div>)}</section>;
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
    return null;
  }
  return null;
}

function AuthGate({ authState }: { authState: AuthState }) {
  if (authState.status === "loading") {
    return <main className="auth-gate"><section className="auth-gate-panel"><div className="brand-mark">MyPocket</div><p>Đang kiểm tra phiên đăng nhập...</p></section></main>;
  }
  if (authState.status === "forbidden") {
    return <main className="auth-gate"><section className="auth-gate-panel"><h1>Không có quyền truy cập</h1><p>Phiên hiện tại không thể mở dữ liệu này.</p>{authState.correlationID ? <small>{authState.correlationID}</small> : null}</section></main>;
  }
  return <main className="auth-gate"><section className="auth-gate-panel"><div className="brand-mark">MyPocket</div><h1>Quản lý tiền rõ ràng hơn</h1><p>Đăng nhập để xem ví, giao dịch và kế hoạch của bạn.</p><button className="primary-cta login-button" type="button" onClick={() => { window.location.href = `${apiBaseURL()}/api/v1/auth/google`; }}>Đăng nhập bằng Google</button></section></main>;
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
    <button className={cx("tab-button", active && "active")} aria-label={tab.label} type="button" onClick={onClick}>
      <Icon size={27} strokeWidth={2.4} />
      <span>{tab.label}</span>
    </button>
  );
}

function Overview({ online, wallets, dashboard, report, insider, privacyMasked, walletDetail, onCloseWalletDetail, onManageWallets, onRefreshInsider }: { online: boolean; wallets: WalletSummary[] | null; dashboard: Dashboard | null; report: Report | null; insider: InsiderReport | null; privacyMasked: boolean; walletDetail: WalletDetail | null; onCloseWalletDetail: () => void; onManageWallets: () => void; onRefreshInsider: () => void }) {
  const visibleWallets = wallets ?? [];

  return (
    <section className="content-stack">
      <Card className="wallet-card">
        {!online ? <p className="offline-warning neutral">Dữ liệu đang hiển thị từ lần đồng bộ cuối</p> : null}
        <SectionTitle title="Ví của tôi" action={<button type="button" onClick={onManageWallets}>Xem tất cả</button>} />
        {visibleWallets.length === 0 ? <p className="empty-state">Chưa có ví</p> : null}
        {visibleWallets.map((wallet) => (
          <WalletRow key={wallet.id} icon={walletIcon(wallet.type)} name={wallet.name} amount={privacyMasked ? "••••••" : formatVND(wallet.balance_vnd)} onClick={() => { if (online) void loadWalletDetail(wallet.id).then(setWalletDetail).catch(() => undefined); }} />
        ))}
      </Card>
      {walletDetail ? <WalletDetailPanel detail={walletDetail} onClose={onCloseWalletDetail} privacyMasked={privacyMasked} /> : null}
      {dashboard?.investment_market_value_vnd || dashboard?.missing_asset_price_count ? (
        <Card className="wallet-card">
          <SectionTitle title="Tài sản đầu tư" />
          <WalletRow icon="◆" name="Giá trị đầu tư" amount={privacyMasked ? "••••••" : formatVND(dashboard.investment_market_value_vnd ?? 0)} />
          <WalletRow icon="₫" name="Tổng tài sản" amount={privacyMasked ? "••••••" : formatVND(dashboard.combined_net_worth_vnd ?? totalBalance)} />
          {dashboard.missing_asset_price_count ? <p className="notification-status">{dashboard.missing_asset_price_count} tài sản chưa có giá hiện tại</p> : null}
        </Card>
      ) : null}

      <SectionHeading title="Money Insider" action={<RefreshCw size={20} />} actionLabel="Làm mới Money Insider" onAction={onRefreshInsider} />
      <section className="card insider-card">
        {!insider ? <p className="empty-state">Đang tổng hợp chi tiêu...</p> : !insider.selected_category ? <p className="empty-state">Chưa đủ dữ liệu chi tiêu tháng này</p> : <>
          <h2>{insider.selected_category.name} <span className="info" title="Danh mục có nhiều giao dịch nhất">i</span></h2>
          <p className="muted">Tổng đã chi <strong className="expense">{privacyMasked ? "••••••" : formatVND(insider.spent_vnd)}</strong></p>
          <div className="insider-grid">
            <div>
              <h3>Tháng này</h3>
              <p className="muted">Trung bình trong {insider.elapsed_days} ngày</p>
              <strong>{privacyMasked ? "••••••" : formatVND(insider.average_daily_vnd)}<span>/ngày</span></strong>
            </div>
            <div className="ring" aria-label={insider.not_comparable ? "Chưa đủ dữ liệu tháng trước" : `${formatPercent(insider.change_percent ?? 0)} so với tháng trước`}>{insider.not_comparable ? "—" : formatPercent(insider.change_percent ?? 0)}</div>
          </div>
          <p className="insider-comparison">{insider.not_comparable ? "Chưa đủ dữ liệu tháng trước để so sánh" : `${formatPercent(insider.change_percent ?? 0)} so với trung bình mỗi ngày tháng trước`}</p>
        </>}
      </section>

      <SectionHeading title="Báo cáo tháng này" action="Xem báo cáo" />
      <section className="card report-card" aria-label="Báo cáo chi tiêu">
        <div className="segmented"><span>Tuần</span><strong>Tháng</strong></div>
        <div className="chart">
          <span className="chart-line red" />
          <span className="chart-line gray" />
        </div>
        <div className="report-stats">
              <p>Tổng đã chi <strong className="expense">{formatVND(report?.summary.expense_vnd ?? 0)}</strong></p>
              <p>Tổng thu <strong className="income">{formatVND(report?.summary.income_vnd ?? 0)}</strong></p>
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
  schedules,
  drafts,
  categories,
  wallets,
  transactions,
  online,
  onCreate,
  onCreateEvent,
  onCreateObligation,
  onCreateSchedule,
  onEdit,
  onEditEvent,
  onEditObligation,
  onEditSchedule,
}: {
  budgets: BudgetProgress[] | null;
  events: EventSummary[];
  obligations: ObligationSummary[];
  schedules: RecurringSchedule[];
  drafts: TransactionDraft[];
  categories: CategorySummary[];
  wallets: WalletSummary[];
  transactions: Transaction[];
  online: boolean;
  onCreate: () => void;
  onCreateEvent: () => void;
  onCreateObligation: () => void;
  onCreateSchedule: () => void;
  onEdit: (budget: BudgetProgress) => void;
  onEditEvent: (event: EventSummary) => void;
  onEditObligation: (obligation: ObligationSummary) => void;
  onEditSchedule: (schedule: RecurringSchedule) => void;
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
      <section className="card list-card planning-list">
        <div className="section-title">
          <h2>Lặp lại</h2>
          <button type="button" disabled={!online || wallets.length === 0} onClick={onCreateSchedule}>Tạo lịch</button>
        </div>
        {schedules.length === 0 ? <p className="empty-state">Chưa có lịch lặp</p> : schedules.map((schedule) => (
          <button className="planning-row" type="button" key={schedule.id} disabled={!online} onClick={() => onEditSchedule(schedule)}>
            <span className="category-dot schedule-dot" />
            <div>
              <strong>{schedule.name}</strong>
              <p>{recurrenceLabel(schedule.frequency)} · Tiếp theo {new Date(schedule.next_occurs_at).toLocaleDateString("vi-VN")}</p>
              <p>{formatVND(schedule.amount_vnd)} · {transactionTypeLabel(schedule.type)}</p>
            </div>
            <ChevronRight size={22} />
          </button>
        ))}
      </section>
      <section className="card list-card planning-list">
        <div className="section-title">
          <h2>Bản nháp</h2>
          <button type="button">Xem</button>
        </div>
        {drafts.filter((draft) => draft.status === "pending").length === 0 ? <p className="empty-state">Chưa có bản nháp cần duyệt</p> : drafts.filter((draft) => draft.status === "pending").map((draft) => (
          <div className="planning-row draft-row" key={draft.id}>
            <span className="category-dot draft-dot" />
            <div>
              <strong>{draft.note || "Bản nháp lặp lại"}</strong>
              <p>{new Date(draft.occurred_at).toLocaleDateString("vi-VN")} · {transactionTypeLabel(draft.type)}</p>
              <p>{formatVND(draft.amount_vnd)} · Chờ duyệt</p>
            </div>
          </div>
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

function ScheduleSheet({ schedule, wallets, categories, onSaved, onArchived, onClose }: { schedule?: RecurringSchedule; wallets: WalletSummary[]; categories: CategorySummary[]; onSaved: () => void; onArchived?: () => void; onClose: () => void }) {
  const [name, setName] = useState(schedule?.name ?? "");
  const [frequency, setFrequency] = useState<RecurrenceFrequency>(schedule?.frequency ?? "monthly");
  const [amount, setAmount] = useState(String(schedule?.amount_vnd ?? ""));
  const [type, setType] = useState<"income" | "expense" | "transfer">(schedule?.type ?? "expense");
  const [sourceWalletID, setSourceWalletID] = useState(schedule?.source_wallet_id ?? wallets[0]?.id ?? "");
  const [destinationWalletID, setDestinationWalletID] = useState(schedule?.destination_wallet_id ?? wallets.find((wallet) => wallet.id !== sourceWalletID)?.id ?? "");
  const [categoryID, setCategoryID] = useState(schedule?.category_id ?? "");
  const [startsOn, setStartsOn] = useState(schedule ? new Date(schedule.starts_at).toISOString().slice(0, 10) : new Date().toISOString().slice(0, 10));
  const [note, setNote] = useState(schedule?.note ?? "");
  const [saving, setSaving] = useState(false);
  const selectableCategories = categories.filter((category) => category.kind === type);
  const chosenCategoryID = type === "transfer" ? "" : categoryID || selectableCategories[0]?.id || "";
  const canSave = name.trim() !== "" && Number(amount) > 0 && sourceWalletID !== "" && startsOn !== "" && (type === "transfer" ? destinationWalletID !== "" && destinationWalletID !== sourceWalletID : chosenCategoryID !== "");
  useEffect(() => {
    if (!sourceWalletID && wallets[0]) setSourceWalletID(wallets[0].id);
    if (type === "transfer" && (!destinationWalletID || destinationWalletID === sourceWalletID)) {
      setDestinationWalletID(wallets.find((wallet) => wallet.id !== sourceWalletID)?.id ?? "");
    }
  }, [destinationWalletID, sourceWalletID, type, wallets]);
  async function save() {
    if (!canSave || saving || schedule) return;
    setSaving(true);
    try {
      const input: RecurringScheduleInput = {
        name,
        frequency,
        timezone: "Asia/Ho_Chi_Minh",
        starts_at: `${startsOn}T09:00:00+07:00`,
        type,
        source_wallet_id: sourceWalletID,
        destination_wallet_id: type === "transfer" ? destinationWalletID : undefined,
        category_id: type === "transfer" ? undefined : chosenCategoryID,
        amount_vnd: Number(amount),
        note,
      };
      await createRecurringSchedule(input);
      onSaved();
    } finally {
      setSaving(false);
    }
  }
  async function archive() {
    if (!schedule || saving) return;
    setSaving(true);
    try {
      await archiveRecurringSchedule(schedule.id);
      onArchived?.();
    } finally {
      setSaving(false);
    }
  }
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label={schedule ? "Sửa Lịch Lặp" : "Tạo Lịch Lặp"}>
        <header>
          <button className="pill-button" type="button" onClick={onClose}>Hủy</button>
          <h2>{schedule ? "Sửa Lịch Lặp" : "Tạo Lịch Lặp"}</h2>
          <span />
        </header>
        {schedule ? <p className="sheet-meta">Lịch đang chỉ hỗ trợ lưu trữ; tạo lịch mới để đổi mẫu.</p> : null}
        <label className="sheet-row"><CalendarDays /><input aria-label="Tên lịch lặp" value={name} onChange={(change) => setName(change.target.value)} placeholder="Tên lịch lặp" disabled={Boolean(schedule)} /></label>
        <label className="amount-row"><span>VND</span><input aria-label="Số tiền lịch lặp" inputMode="numeric" value={amount} onChange={(change) => setAmount(change.target.value.replace(/\D/g, ""))} placeholder="0" disabled={Boolean(schedule)} /></label>
        <label className="sheet-row"><CalendarDays /><select aria-label="Chu kỳ lặp" value={frequency} onChange={(change) => setFrequency(change.target.value as RecurrenceFrequency)} disabled={Boolean(schedule)}><option value="daily">Hàng ngày</option><option value="weekly">Hàng tuần</option><option value="monthly">Hàng tháng</option></select></label>
        <label className="sheet-row"><CalendarDays /><input aria-label="Ngày bắt đầu lịch lặp" type="date" value={startsOn} onChange={(change) => setStartsOn(change.target.value)} disabled={Boolean(schedule)} /></label>
        <div className="segmented sheet-segmented">
          {(["expense", "income", "transfer"] as const).map((option) => <button className={type === option ? "active" : ""} type="button" key={option} disabled={Boolean(schedule)} onClick={() => setType(option)}>{transactionTypeLabel(option)}</button>)}
        </div>
        <label className="sheet-row"><Wallet /><select aria-label="Ví lịch lặp" value={sourceWalletID} onChange={(change) => setSourceWalletID(change.target.value)} disabled={Boolean(schedule)}>{wallets.map((wallet) => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</select></label>
        {type === "transfer" ? <label className="sheet-row"><Wallet /><select aria-label="Ví nhận lịch lặp" value={destinationWalletID} onChange={(change) => setDestinationWalletID(change.target.value)} disabled={Boolean(schedule)}>{wallets.filter((wallet) => wallet.id !== sourceWalletID).map((wallet) => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</select></label> : null}
        {type !== "transfer" ? <label className="sheet-row"><span className="dot-icon" /><select aria-label="Nhóm lịch lặp" value={chosenCategoryID} onChange={(change) => setCategoryID(change.target.value)} disabled={Boolean(schedule)}>{selectableCategories.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}</select></label> : null}
        <label className="sheet-row"><List /><input aria-label="Ghi chú lịch lặp" value={note} onChange={(change) => setNote(change.target.value)} placeholder="Ghi chú" disabled={Boolean(schedule)} /></label>
        <div className="sheet-actions">
          {schedule ? <button className="wide-pill destructive" type="button" disabled={saving} onClick={() => void archive()}>Lưu trữ</button> : null}
          {!schedule ? <button className="primary-cta" type="button" disabled={!canSave || saving} onClick={() => void save()}>{saving ? "Đang lưu" : "Lưu"}</button> : null}
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

function Account({
  authState,
  assets,
  portfolioSummary,
  privacyMasked,
  online,
  onCreateAsset,
  onAssetChanged,
  onAssetArchived,
  onLogout,
}: {
  authState: AuthState;
  assets: AssetPosition[];
  portfolioSummary: PortfolioSummary | null;
  privacyMasked: boolean;
  online: boolean;
  onCreateAsset: () => void;
  onAssetChanged: (asset: AssetPosition) => void;
  onAssetArchived: (assetID: string) => void;
  onLogout: () => void;
}) {
  const email = authState.status === "authenticated" ? authState.user.email : "Chưa đăng nhập";
  const displayName = authState.status === "authenticated" ? authState.user.display_name || authState.user.email : "Tài khoản MyPocket";
  const [apiKeys, setAPIKeys] = useState<APIKeySummary[]>([]);
  const [apiKeyName, setAPIKeyName] = useState("AI Agent");
  const [createdKey, setCreatedKey] = useState<CreatedAPIKey | null>(null);
  const [apiKeyBusy, setAPIKeyBusy] = useState(false);
  const [auditAllowed, setAuditAllowed] = useState(false);
  const [auditEvents, setAuditEvents] = useState<AuditEvent[]>([]);
  const [auditCorrelationID, setAuditCorrelationID] = useState("");
  const [auditBusy, setAuditBusy] = useState(false);

  useEffect(() => {
    if (!online || authState.status !== "authenticated") return;
    void loadAPIKeys().then(setAPIKeys).catch(() => undefined);
  }, [authState.status, online]);

  useEffect(() => {
    if (!online || authState.status !== "authenticated") {
      setAuditAllowed(false);
      setAuditEvents([]);
      return;
    }
    let active = true;
    setAuditBusy(true);
    void checkAuditAccess()
      .then((result) => {
        if (!active) return;
        setAuditAllowed(result.allowed);
        if (result.allowed) return loadAuditEvents().then((events) => { if (active) setAuditEvents(events); });
        return undefined;
      })
      .catch(() => { if (active) setAuditAllowed(false); })
      .finally(() => { if (active) setAuditBusy(false); });
    return () => { active = false; };
  }, [authState.status, online]);

  async function handleCreateAPIKey() {
    if (!online || apiKeyName.trim() === "") return;
    setAPIKeyBusy(true);
    try {
      const key = await createAPIKey(apiKeyName.trim());
      setCreatedKey(key);
      setAPIKeys((current) => [key, ...current.filter((item) => item.id !== key.id)]);
    } finally {
      setAPIKeyBusy(false);
    }
  }

  async function handleRevokeAPIKey(id: string) {
    if (!online) return;
    setAPIKeyBusy(true);
    try {
      await revokeAPIKey(id);
      setAPIKeys((current) => current.map((item) => item.id === id ? { ...item, revoked_at: new Date().toISOString() } : item));
    } finally {
      setAPIKeyBusy(false);
    }
  }

  async function refreshAuditEvents() {
    if (!auditAllowed || !online) return;
    setAuditBusy(true);
    try {
      setAuditEvents(await loadAuditEvents(auditCorrelationID));
    } finally {
      setAuditBusy(false);
    }
  }

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
      <section className="card list-card">
        <SectionTitle title="Tài sản" action={<button type="button" onClick={onCreateAsset}>Thêm</button>} />
        <div className="asset-total-row">
          <span>Giá trị đầu tư</span>
          <strong>{privacyMasked ? "••••••" : formatVND(portfolioSummary?.investment_market_value_vnd ?? 0)}</strong>
        </div>
        {portfolioSummary?.missing_price_count ? <p className="notification-status">{portfolioSummary.missing_price_count} tài sản chưa có giá hiện tại</p> : null}
        {assets.length === 0 ? <p className="empty-state">Chưa có tài sản</p> : assets.map((asset) => <AssetRow key={asset.id} asset={asset} privacyMasked={privacyMasked} online={online} onChanged={onAssetChanged} onArchived={onAssetArchived} />)}
      </section>
      <section className="card list-card">
        <TransactionRow title="iPhone" subtitle="Thiết bị này" amount="" positive />
      </section>
      <section className="card list-card api-key-card">
        <SectionTitle title="API keys" action={<KeyRound size={18} />} />
        <div className="manager-form api-key-form">
          <input aria-label="Tên API key" value={apiKeyName} onChange={(event) => setAPIKeyName(event.target.value)} placeholder="Tên key" disabled={!online || apiKeyBusy} />
          <button type="button" disabled={!online || apiKeyBusy || apiKeyName.trim() === ""} onClick={() => void handleCreateAPIKey()}>{apiKeyBusy ? "Đang xử lý" : "Tạo key"}</button>
        </div>
        {createdKey ? (
          <div className="api-key-secret">
            <span>Chỉ hiển thị một lần</span>
            <code>{createdKey.plaintext}</code>
            <button type="button" onClick={() => void navigator.clipboard?.writeText(createdKey.plaintext)}>Copy</button>
          </div>
        ) : null}
        {apiKeys.length === 0 ? <p className="empty-state">{online ? "Chưa có API key" : "Cần online để quản lý API key"}</p> : (
          <div className="api-key-list">
            {apiKeys.map((key) => (
              <div className="api-key-row" key={key.id}>
                <div>
                  <strong>{key.name}</strong>
                  <p>{key.key_prefix}... · {key.revoked_at ? "Đã revoke" : key.last_used_at ? `Dùng ${new Date(key.last_used_at).toLocaleDateString("vi-VN")}` : "Chưa dùng"}</p>
                </div>
                <button className="danger-text" type="button" disabled={!online || apiKeyBusy || Boolean(key.revoked_at)} onClick={() => void handleRevokeAPIKey(key.id)}>Revoke</button>
              </div>
            ))}
          </div>
        )}
      </section>
      {auditAllowed ? (
        <section className="card list-card audit-log-card">
          <SectionTitle title="Nhật ký hệ thống" action={<Activity size={18} />} />
          <div className="manager-form audit-log-form">
            <input aria-label="Correlation ID" value={auditCorrelationID} onChange={(event) => setAuditCorrelationID(event.target.value)} placeholder="Correlation ID" />
            <button type="button" disabled={auditBusy || !online} onClick={() => void refreshAuditEvents()}>{auditBusy ? "Đang tải" : "Làm mới"}</button>
          </div>
          {auditEvents.length === 0 ? <p className="empty-state">Không có sự kiện phù hợp</p> : (
            <div className="audit-log-list">
              {auditEvents.map((event) => (
                <article className="audit-log-row" key={event.id}>
                  <div><strong>{event.action}</strong><p>{event.request_method ?? ""} {event.request_path ?? ""}</p></div>
                  <span className={`audit-severity ${event.severity}`}>{event.outcome}</span>
                  <time dateTime={event.occurred_at}>{new Date(event.occurred_at).toLocaleString("vi-VN")}</time>
                  <code>{event.correlation_id}</code>
                </article>
              ))}
            </div>
          )}
        </section>
      ) : null}
      {auditBusy && !auditAllowed ? <p className="notification-status">Đang kiểm tra quyền nhật ký…</p> : null}
      <section className="card list-card">
        <a className="planning-row" href="/docs/" target="_blank" rel="noopener noreferrer">
          <BookOpen size={24} />
          <div>
            <strong>Tài liệu</strong>
            <p>Sơ đồ CSDL và tài liệu API</p>
          </div>
          <ChevronRight size={22} />
        </a>
      </section>
      <button className="wide-pill destructive" type="button" onClick={onLogout}>Đăng xuất</button>
    </section>
  );
}

function AssetRow({ asset, privacyMasked, online, onChanged, onArchived }: { asset: AssetPosition; privacyMasked: boolean; online: boolean; onChanged: (asset: AssetPosition) => void; onArchived: (assetID: string) => void }) {
  const [mode, setMode] = useState<"idle" | "buy" | "sell" | "price">("idle");
  const marketValue = asset.summary.market_value_vnd;
  const pnl = asset.summary.unrealized_pnl_vnd;
  return (
    <article className="asset-row">
      <button type="button" className="asset-main" onClick={() => setMode((current) => current === "idle" ? "price" : "idle")}>
        <span className="category-dot" />
        <div>
          <strong>{asset.name}</strong>
          <p>{assetLabel(asset)} · {asset.summary.quantity} {unitLabel(asset.unit)}</p>
        </div>
        <b>{privacyMasked ? "••••••" : marketValue == null ? "Chưa có giá" : formatVND(marketValue)}</b>
      </button>
      <div className="asset-meta">
        <span>{asset.pricing_mode === "automatic" ? "Tự động" : "Thủ công"}</span>
        <span>{asset.latest_price?.priced_at ? `Giá ${new Date(asset.latest_price.priced_at).toLocaleDateString("vi-VN")}` : "Chưa định giá"}</span>
        {pnl != null ? <strong className={pnl >= 0 ? "income" : "expense"}>{privacyMasked ? "••••••" : formatVND(pnl)}</strong> : null}
      </div>
      <div className="asset-actions">
        <button type="button" onClick={() => setMode("buy")}>Mua</button>
        <button type="button" onClick={() => setMode("sell")}>Bán</button>
        <button type="button" onClick={() => setMode("price")}>Giá</button>
        <button type="button" className="danger-text" onClick={() => void archiveAsset(asset.id, asset.version).then(() => onArchived(asset.id)).catch(() => undefined)}>Ẩn</button>
      </div>
      {mode !== "idle" ? <AssetActionForm asset={asset} mode={mode} onSaved={(next) => { setMode("idle"); onChanged(next); }} /> : null}
    </article>
  );
}

function AssetActionForm({ asset, mode, onSaved }: { asset: AssetPosition; mode: "buy" | "sell" | "price"; onSaved: (asset: AssetPosition) => void }) {
  const [quantity, setQuantity] = useState("");
  const [price, setPrice] = useState("");
  const [fee, setFee] = useState("");
  const [busy, setBusy] = useState(false);
  const today = new Date().toISOString();
  const canSave = Number(price) > 0 && (mode === "price" || Number(quantity) > 0);
  async function save() {
    if (!canSave) return;
    setBusy(true);
    try {
      const next = mode === "price"
        ? await addAssetPrice(asset.id, { unit_price_vnd: Number(price), priced_at: today, base_version: asset.version })
        : await addAssetTrade(asset.id, { side: mode as TradeSide, quantity, unit_price_vnd: Number(price), fee_vnd: Number(fee || 0), occurred_at: today, base_version: asset.version });
      onSaved(next);
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="manager-form asset-form">
      {mode !== "price" ? <input aria-label="Số lượng tài sản" inputMode="decimal" value={quantity} onChange={(event) => setQuantity(event.target.value.replace(/[^0-9.]/g, ""))} placeholder="Số lượng" /> : null}
      <input aria-label="Giá VND" inputMode="numeric" value={price} onChange={(event) => setPrice(event.target.value.replace(/\D/g, ""))} placeholder="Giá VND" />
      {mode !== "price" ? <input aria-label="Phí VND" inputMode="numeric" value={fee} onChange={(event) => setFee(event.target.value.replace(/\D/g, ""))} placeholder="Phí" /> : null}
      <button type="button" disabled={!canSave || busy} onClick={() => void save()}>{busy ? "Đang lưu" : "Lưu"}</button>
    </div>
  );
}

function AssetSheet({ onSaved, onClose }: { onSaved: (asset: AssetPosition) => void; onClose: () => void }) {
  const [type, setType] = useState<AssetType>("gold");
  const [name, setName] = useState("");
  const [symbol, setSymbol] = useState("");
  const [unit, setUnit] = useState("tael");
  const [pricingMode, setPricingMode] = useState<"manual" | "automatic">("manual");
  const [providerKey, setProviderKey] = useState("static");
  const [providerSymbol, setProviderSymbol] = useState("");
  const [busy, setBusy] = useState(false);
  const canSave = name.trim() !== "" && (pricingMode === "manual" || (providerKey.trim() !== "" && providerSymbol.trim() !== ""));
  useEffect(() => {
    setUnit(type === "gold" ? "tael" : type === "stock" ? "share" : type === "crypto" ? "token" : "unit");
  }, [type]);
  async function save() {
    if (!canSave) return;
    setBusy(true);
    try {
      const asset = await createAsset({
        type,
        name,
        symbol,
        unit,
        pricing_mode: pricingMode,
        provider_key: pricingMode === "automatic" ? providerKey : "",
        provider_symbol: pricingMode === "automatic" ? providerSymbol : "",
        include_in_net_worth: true,
      });
      onSaved(asset);
    } finally {
      setBusy(false);
    }
  }
  return (
    <SheetFrame title="Thêm tài sản" label="Thêm tài sản" leading={<button type="button" onClick={onClose}>Đóng</button>} trailing={<Info size={22} />}>
      <section className="manager-section">
        <label className="sheet-row"><BriefcaseBusiness /><select aria-label="Loại tài sản" value={type} onChange={(event) => setType(event.target.value as AssetType)}>{assetTypes.map((item) => <option key={item} value={item}>{assetTypeLabel(item)}</option>)}</select></label>
        <label className="sheet-row"><List /><input aria-label="Tên tài sản" value={name} onChange={(event) => setName(event.target.value)} placeholder="Tên tài sản" /></label>
        <label className="sheet-row"><Search /><input aria-label="Mã tài sản" value={symbol} onChange={(event) => setSymbol(event.target.value)} placeholder="Mã, ví dụ FPT hoặc BTC" /></label>
        <label className="sheet-row"><Wallet /><select aria-label="Đơn vị tài sản" value={unit} onChange={(event) => setUnit(event.target.value)}>{unitsForAsset(type).map((item) => <option key={item} value={item}>{unitLabel(item)}</option>)}</select></label>
        <div className="segmented sheet-segmented">
          <button type="button" className={pricingMode === "manual" ? "active" : ""} onClick={() => setPricingMode("manual")}>Thủ công</button>
          <button type="button" className={pricingMode === "automatic" ? "active" : ""} onClick={() => setPricingMode("automatic")}>Tự động</button>
        </div>
        {pricingMode === "automatic" ? (
          <>
            <label className="sheet-row"><List /><input aria-label="Provider key" value={providerKey} onChange={(event) => setProviderKey(event.target.value)} placeholder="Provider" /></label>
            <label className="sheet-row"><Search /><input aria-label="Provider symbol" value={providerSymbol} onChange={(event) => setProviderSymbol(event.target.value)} placeholder="Symbol provider" /></label>
          </>
        ) : null}
        <button className="primary-cta" type="button" disabled={!canSave || busy} onClick={() => void save()}>{busy ? "Đang lưu" : "Lưu"}</button>
      </section>
    </SheetFrame>
  );
}

function AddTransactionSheet({ categories, wallets, readOnly, onCreated, onDebtCreated, onClose }: { categories: CategorySummary[]; wallets: WalletSummary[]; readOnly: boolean; onCreated: (transaction: Transaction) => void; onDebtCreated: () => void; onClose: () => void }) {
  const [type, setType] = useState<"expense" | "income" | "debt">("expense");
  const [amount, setAmount] = useState("");
  const [note, setNote] = useState("");
  const [sourceWalletID, setSourceWalletID] = useState(wallets[0]?.id ?? "");
  const [categoryID, setCategoryID] = useState("");
  const [excludedFromReports, setExcludedFromReports] = useState(false);
  const [debtDirection, setDebtDirection] = useState<ObligationDirection>("borrowed");
  const [counterparty, setCounterparty] = useState("");
  const [dueOn, setDueOn] = useState(() => new Date().toISOString().slice(0, 10));
  const [saving, setSaving] = useState(false);
  const [receiptFile, setReceiptFile] = useState<File | null>(null);
  const filteredCategories = categories.filter((category) => category.kind === (type === "income" ? "income" : "expense"));
  const chosenCategoryID = type !== "debt" ? categoryID || filteredCategories[0]?.id : "";
  const canSave = !readOnly && wallets.length > 0 && Number(amount) > 0 && (type === "debt" ? counterparty.trim() !== "" && dueOn !== "" : Boolean(sourceWalletID));
  useEffect(() => {
    if (!sourceWalletID && wallets[0]) setSourceWalletID(wallets[0].id);
  }, [sourceWalletID, wallets]);
  async function save() {
    if (!canSave || saving) return;
    setSaving(true);
    try {
      if (type === "debt") {
        await createObligation({ direction: debtDirection, principal_vnd: Number(amount), counterparty: counterparty.trim(), due_on: dueOn, note });
        onDebtCreated();
        onClose();
        return;
      }
      const receipt = receiptFile && navigator.onLine ? await uploadReceipt(receiptFile) : undefined;
      const transaction = await createTransaction(buildTransactionInput({ type, amount, sourceWalletID, categoryID: chosenCategoryID, note, excludedFromReports, receiptObjectID: receipt?.id }));
      if (receiptFile && !navigator.onLine) {
        await queueReceiptUpload({ transaction_id: transaction.id, file: receiptFile, filename: receiptFile.name, content_type: receiptFile.type });
      }
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
          {(["expense", "income", "debt"] as const).map((option) => (
            <button className={type === option ? "active" : ""} type="button" key={option} onClick={() => setType(option)}>{quickAddTypeLabel(option)}</button>
          ))}
        </div>
        <label className="amount-row"><span>VND</span><input aria-label="Số tiền" inputMode="numeric" value={amount} onChange={(event) => setAmount(event.target.value.replace(/\D/g, ""))} placeholder="0" /></label>
        {type !== "debt" ? <label className="sheet-row"><Wallet /><select aria-label="Ví nguồn" value={sourceWalletID} onChange={(event) => setSourceWalletID(event.target.value)}>{wallets.map((wallet) => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</select></label> : null}
        {type !== "debt" ? <label className="sheet-row"><span className="dot-icon" /><select aria-label="Nhóm" value={chosenCategoryID} onChange={(event) => setCategoryID(event.target.value)}><option value="">Chọn nhóm</option>{filteredCategories.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}</select></label> : null}
        {type === "debt" ? <div className="segmented sheet-segmented"><button type="button" className={debtDirection === "borrowed" ? "active" : ""} onClick={() => setDebtDirection("borrowed")}>Tôi vay</button><button type="button" className={debtDirection === "lent" ? "active" : ""} onClick={() => setDebtDirection("lent")}>Tôi cho vay</button></div> : null}
        {type === "debt" ? <label className="sheet-row"><Users /><input aria-label="Đối tác" value={counterparty} onChange={(event) => setCounterparty(event.target.value)} placeholder="Người liên quan" /></label> : null}
        {type === "debt" ? <label className="sheet-row"><CalendarDays /><input aria-label="Ngày đến hạn" type="date" value={dueOn} onChange={(event) => setDueOn(event.target.value)} /></label> : null}
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
        <label className="image-row" htmlFor="receipt-image"><ImagePlus size={28} />{receiptFile ? receiptFile.name : "Thêm Hình Ảnh"}<input id="receipt-image" type="file" accept="image/jpeg,image/png,image/webp" onChange={(event) => setReceiptFile(event.target.files?.[0] ?? null)} hidden /></label>
        {type !== "debt" ? <button className={excludedFromReports ? "toggle-row active" : "toggle-row"} type="button" onClick={() => setExcludedFromReports((current) => !current)}>Không tính vào báo cáo<span /></button> : null}
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
  wallets,
  readOnly,
  onWalletChanged,
  onWalletArchived,
  onChanged,
  onClose,
}: {
  wallets: WalletSummary[];
  readOnly: boolean;
  onWalletChanged: (wallet: WalletSummary) => void;
  onWalletArchived: (walletID: string) => void;
  onChanged: () => void;
  onClose: () => void;
}) {
  const [walletName, setWalletName] = useState("");
  const [walletType, setWalletType] = useState<WalletType>("cash");
  const [creatingWallet, setCreatingWallet] = useState(false);
  const [editingWallets, setEditingWallets] = useState(false);
  const [busy, setBusy] = useState(false);
  const includedWallets = wallets.filter((wallet) => wallet.include_in_total);
  const excludedWallets = wallets.filter((wallet) => !wallet.include_in_total);
  const totalVND = totalIncludedVND(wallets);
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
    <SheetFrame
      title="Ví Của Tôi"
      className="wallet-sheet"
      headerClassName="wallet-sheet-header"
      leading={<PillButton onClick={onClose}>Đóng</PillButton>}
      trailing={(
        <div className="wallet-sheet-tools">
          <PillButton className="icon-pill" aria-label="Thông tin ví"><Info size={24} /></PillButton>
          <PillButton onClick={() => setEditingWallets((current) => !current)}>{editingWallets ? "Xong" : "Sửa"}</PillButton>
        </div>
      )}
    >
        {readOnly ? <p className="offline-warning">Offline storage chưa sẵn sàng. Mở mạng lại để chỉnh ví và nhóm.</p> : null}

        <section className="wallet-total-card">
          <span className="wallet-sheet-icon total">🌐</span>
          <div>
            <strong>Tổng cộng</strong>
            <b>{formatVND(totalVND)}</b>
          </div>
          <span className="wallet-check">✓</span>
        </section>

        <h3 className="wallet-group-title">Tính vào tổng</h3>
        <section className="wallet-list-card" aria-label="Ví tính vào tổng">
          {includedWallets.length === 0 ? <p className="empty-state">Chưa có ví tính vào tổng</p> : null}
          {includedWallets.map((wallet) => <WalletManageRow key={wallet.id} wallet={wallet} wallets={wallets} readOnly={readOnly} busy={busy} editMode={editingWallets} onWalletChanged={onWalletChanged} onWalletArchived={onWalletArchived} onChanged={run} />)}
        </section>

        {excludedWallets.length > 0 ? <h3 className="wallet-group-title">Không tính vào tổng</h3> : null}
        {excludedWallets.length > 0 ? (
          <section className="wallet-list-card" aria-label="Ví không tính vào tổng">
            {excludedWallets.map((wallet) => <WalletManageRow key={wallet.id} wallet={wallet} wallets={wallets} readOnly={readOnly} busy={busy} editMode={editingWallets} onWalletChanged={onWalletChanged} onWalletArchived={onWalletArchived} onChanged={run} />)}
          </section>
        ) : null}

        <section className="wallet-actions-card" aria-label="Thao tác ví">
          <button className="wallet-action-row" type="button" disabled={readOnly} onClick={() => setCreatingWallet((current) => !current)}>
            <span className="wallet-action-icon"><Plus size={28} /></span>
            <strong>Thêm ví</strong>
          </button>
          <button className="wallet-action-row" type="button" disabled>
            <span className="wallet-action-icon"><Link size={24} /></span>
            <strong>Liên kết dịch vụ</strong>
          </button>
        </section>

        {creatingWallet ? (
          <section className="manager-section wallet-create-panel">
            <h3>Thêm ví</h3>
            <div className="manager-form">
              <input aria-label="Tên ví mới" value={walletName} onChange={(event) => setWalletName(event.target.value)} placeholder="Tên ví mới" disabled={readOnly} />
              <select aria-label="Loại ví" value={walletType} onChange={(event) => setWalletType(event.target.value as WalletType)} disabled={readOnly}>{walletTypes.map((type) => <option key={type} value={type}>{walletTypeLabel(type)}</option>)}</select>
              <button type="button" disabled={readOnly || !walletName.trim() || busy} onClick={() => void run(async () => { const wallet = await createWallet({ name: walletName, type: walletType }); if (wallet) onWalletChanged(wallet); setWalletName(""); setCreatingWallet(false); })}>Tạo ví</button>
            </div>
          </section>
        ) : null}

    </SheetFrame>
  );
}

function WalletManageRow({
  wallet,
  wallets,
  readOnly,
  busy,
  editMode,
  onWalletChanged,
  onWalletArchived,
  onChanged,
}: {
  wallet: WalletSummary;
  wallets: WalletSummary[];
  readOnly: boolean;
  busy: boolean;
  editMode: boolean;
  onWalletChanged: (wallet: WalletSummary) => void;
  onWalletArchived: (walletID: string) => void;
  onChanged: (action: () => Promise<unknown>) => Promise<void>;
}) {
  const [name, setName] = useState(wallet.name);
  const [include, setInclude] = useState(wallet.include_in_total);
  if (!editMode) {
    return (
      <article className="wallet-manage-display">
        <span className="wallet-sheet-icon">{walletIcon(wallet.type)}</span>
        <div>
          <strong>{wallet.name}</strong>
          <b>{formatVND(wallet.balance_vnd)}</b>
        </div>
      </article>
    );
  }
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

function WalletRow({ icon, name, amount, onClick }: { icon: string; name: string; amount: string; onClick?: () => void }) {
  return <button className="wallet-row" type="button" onClick={onClick}><span>{icon}</span><strong>{name}</strong><b>{amount}</b></button>;
}

function WalletDetailPanel({ detail, onClose, privacyMasked }: { detail: WalletDetail; onClose: () => void; privacyMasked: boolean }) {
  return <section className="card wallet-detail" aria-label={`Chi tiết ${detail.wallet.name}`}><div className="section-title"><h2>{detail.wallet.name}</h2><button type="button" onClick={onClose}>Đóng</button></div><p>Trạng thái: {detail.wallet.include_in_total ? "Đang tính tổng" : "Không tính tổng"}</p><strong>{privacyMasked ? "••••••" : formatVND(detail.wallet.balance_vnd)}</strong>{detail.transactions.length === 0 ? <p className="empty-state">Chưa có giao dịch</p> : detail.transactions.map((transaction) => <div className="search-result" key={transaction.id}><span>{transaction.note || "Giao dịch"}</span><b>{privacyMasked ? "••••••" : formatVND(transaction.amount_vnd)}</b></div>)}</section>;
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

function SectionHeading({ title, action, actionLabel, onAction }: { title: string; action: ReactNode; actionLabel?: string; onAction?: () => void }) {
  return <div className="section-heading"><h2>{title}</h2><button type="button" aria-label={actionLabel} onClick={onAction}>{action}</button></div>;
}

function formatVND(amount: number) {
  return `${new Intl.NumberFormat("vi-VN").format(amount)} đ`;
}

function formatPercent(value: number) {
  return `${value > 0 ? "+" : ""}${new Intl.NumberFormat("vi-VN", { maximumFractionDigits: 1 }).format(value)}%`;
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
const assetTypes: AssetType[] = ["gold", "stock", "crypto", "foreign_currency", "other"];
const budgetPeriods: BudgetPeriodType[] = ["weekly", "monthly", "quarterly", "yearly", "custom"];

function assetLabel(asset: AssetPosition) {
  const code = [asset.symbol, asset.exchange].filter(Boolean).join("/");
  return code || assetTypeLabel(asset.type);
}

function assetTypeLabel(type: AssetType) {
  switch (type) {
    case "gold":
      return "Vàng";
    case "stock":
      return "Cổ phiếu";
    case "crypto":
      return "Crypto";
    case "foreign_currency":
      return "Ngoại tệ";
    default:
      return "Tài sản khác";
  }
}

function unitsForAsset(type: AssetType) {
  switch (type) {
    case "gold":
      return ["tael", "gram", "ounce"];
    case "stock":
      return ["share"];
    case "crypto":
      return ["token"];
    default:
      return ["unit"];
  }
}

function unitLabel(unit: string) {
  switch (unit) {
    case "tael":
      return "lượng";
    case "gram":
      return "gram";
    case "ounce":
      return "ounce";
    case "share":
      return "cổ phiếu";
    case "token":
      return "token";
    default:
      return "đơn vị";
  }
}

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
  sourceWalletID,
  categoryID,
  note,
  excludedFromReports,
  receiptObjectID,
}: {
  type: "expense" | "income";
  amount: string;
  sourceWalletID: string;
  categoryID: string;
  note: string;
  excludedFromReports: boolean;
  receiptObjectID?: string;
}): TransactionInput {
  return {
    type,
    source_wallet_id: sourceWalletID,
    category_id: categoryID,
    receipt_object_id: receiptObjectID,
    amount_vnd: Number(amount),
    target_balance_vnd: null,
    occurred_at: new Date().toISOString(),
    note,
    excluded_from_reports: excludedFromReports,
  };
}

function quickAddTypeLabel(type: "expense" | "income" | "debt") {
  switch (type) {
    case "income":
      return "Khoản thu";
    case "debt":
      return "Vay/nợ";
    case "expense":
    default:
      return "Khoản chi";
  }
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

function recurrenceLabel(frequency: RecurrenceFrequency) {
  switch (frequency) {
    case "daily":
      return "Hàng ngày";
    case "weekly":
      return "Hàng tuần";
    case "monthly":
    default:
      return "Hàng tháng";
  }
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
