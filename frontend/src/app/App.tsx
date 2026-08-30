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
import { drainOutbox, readOutbox } from "./outbox";

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
  const [offlineStatus, setOfflineStatus] = useState<{ mode: "ready" | "degraded"; pending: number; reason?: string }>({ mode: "ready", pending: 0 });
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
        .catch(() => undefined);
    });
  }, [authState.status, online]);

  async function handleLogout() {
    await logout();
    await clearOfflineStore().catch(() => undefined);
    setAuthState({ status: "unauthenticated" });
    setWallets(null);
    setCategories(null);
    setTransactions(null);
    setOfflineStatus({ mode: "ready", pending: 0 });
    setActiveTab("overview");
  }

  async function refreshFinanceData() {
    const [nextWallets, nextCategories, nextTransactions] = await Promise.all([loadWallets(), loadCategories(), loadTransactions()]);
    setWallets(nextWallets);
    setCategories(nextCategories);
    await saveFinanceMirror({ wallets: nextWallets, categories: nextCategories, transactions: nextTransactions });
    const queued = await readOutbox();
    setOfflineStatus((current) => ({ ...current, pending: queued.length }));
    const pending = queued.map((item) => ({ id: item.id, ...item.input, amount_vnd: Number(item.input.amount_vnd), balance_after_vnd: 0, occurred_at: String(item.input.occurred_at), note: String(item.input.note ?? ""), with_person: "", event_ref: "", excluded_from_reports: Boolean(item.input.excluded_from_reports), version: 0 } as Transaction));
    setTransactions([...pending, ...nextTransactions]);
  }

  async function hydrateOfflineData() {
    const snapshot = await initializeOfflineStore();
    setOfflineStatus({ mode: snapshot.mode, pending: snapshot.outbox.length, reason: snapshot.reason });
    if (snapshot.wallets.length > 0) setWallets(snapshot.wallets);
    if (snapshot.categories.length > 0) setCategories(snapshot.categories);
    if (snapshot.transactions.length > 0 || snapshot.outbox.length > 0) setTransactions(snapshot.transactions);
  }

  async function reconcileAfterLocalChange() {
    if (online) {
      await refreshFinanceData();
      return;
    }
    const queued = await readOutbox();
    setOfflineStatus((current) => ({ ...current, pending: queued.length }));
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
        {authState.status === "forbidden" ? <ForbiddenState authState={authState} onLogout={handleLogout} /> : null}
        {authState.status !== "forbidden" && activeTab === "overview" ? <Overview wallets={wallets} onManageWallets={() => setWalletSheetOpen(true)} /> : null}
        {authState.status !== "forbidden" && activeTab === "transactions" ? <Transactions transactions={transactions ?? []} onEdit={setEditingTransaction} /> : null}
        {authState.status !== "forbidden" && activeTab === "budgets" ? <Budgets /> : null}
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
    </div>
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

function Budgets() {
  return (
    <section className="content-stack">
      <div className="sub-header">
        <h1>Ngân sách Đang áp dụng</h1>
        <button className="pill-button" type="button">🌐</button>
      </div>
      <section className="card budget-hero">
        <p>Tháng này</p>
        <strong>45.075.000 đ</strong>
        <div className="budget-stats"><span>65 M đ<br />Tổng ngân sách</span><span>19,92 M đ<br />Tổng đã chi</span><span>12 ngày<br />Đến cuối tháng</span></div>
        <button className="primary-cta compact" type="button">Tạo Ngân sách</button>
      </section>
      <BudgetRow name="Mua sắm" amount="8.000.000 đ" remaining="Còn lại 5.810.000 đ" progress={28} />
      <BudgetRow name="Ăn uống" amount="5.000.000 đ" remaining="Còn lại 2.005.000 đ" progress={62} />
    </section>
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

function BudgetRow({ name, amount, remaining, progress }: { name: string; amount: string; remaining: string; progress: number }) {
  return <section className="card budget-row"><div><span className="category-dot" /><strong>{name}</strong></div><b>{amount}</b><p>{remaining}</p><span className="progress"><i style={{ width: `${progress}%` }} /></span></section>;
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
