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
import { loadCategories, loadWallets, type CategorySummary, type WalletSummary } from "./finance";

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
  const [authState, setAuthState] = useState<AuthState>({ status: "loading" });
  const [wallets, setWallets] = useState<WalletSummary[] | null>(null);
  const [categories, setCategories] = useState<CategorySummary[] | null>(null);
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
      return;
    }
    let cancelled = false;
    void Promise.all([loadWallets(), loadCategories()])
      .then(([nextWallets, nextCategories]) => {
        if (!cancelled) {
          setWallets(nextWallets);
          setCategories(nextCategories);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setWallets([]);
          setCategories([]);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [authState.status]);

  async function handleLogout() {
    await logout();
    setAuthState({ status: "unauthenticated" });
    setActiveTab("overview");
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
        {authState.status !== "forbidden" && activeTab === "overview" ? <Overview wallets={wallets} /> : null}
        {authState.status !== "forbidden" && activeTab === "transactions" ? <Transactions /> : null}
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

      {sheetOpen ? <AddTransactionSheet categories={categories ?? []} onClose={() => setSheetOpen(false)} /> : null}
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

function Overview({ wallets }: { wallets: WalletSummary[] | null }) {
  const visibleWallets = wallets && wallets.length > 0 ? wallets : sampleWallets;
  const totalVND = totalIncludedVND(visibleWallets);

  return (
    <section className="content-stack">
      <section className="card wallet-card">
        <div className="section-title">
          <h2>Ví của tôi</h2>
          <button type="button">Xem tất cả</button>
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

function Transactions() {
  return (
    <section className="content-stack">
      <div className="sub-header">
        <h1>Sổ giao dịch</h1>
        <button className="pill-button" type="button">Tháng 08/2026</button>
      </div>
      <section className="card list-card">
        <TransactionRow title="Ăn uống" subtitle="Techcombank" amount="-250.000 đ" />
        <TransactionRow title="Tiền mặt" subtitle="Điều chỉnh số dư" amount="+84.000 đ" positive />
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

function AddTransactionSheet({ categories, onClose }: { categories: CategorySummary[]; onClose: () => void }) {
  const firstExpenseCategory = categories.find((category) => category.kind === "expense");
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label="Thêm Giao Dịch">
        <header>
          <button className="pill-button" type="button" onClick={onClose}>Hủy</button>
          <h2>Thêm Giao Dịch</h2>
          <span />
        </header>
        <div className="amount-row"><span>VND</span><strong>0</strong></div>
        <SheetRow icon={<span className="dot-icon" />} label={firstExpenseCategory?.name ?? "Chọn nhóm"} muted={!firstExpenseCategory} />
        <SheetRow icon={<List />} label="Ghi chú" />
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
        <label className="toggle-row">Không tính vào báo cáo<span /></label>
        <div className="save-bar"><button type="button" disabled>Lưu</button><button type="button" className="receipt"><ImagePlus size={24} /></button></div>
      </section>
    </div>
  );
}

function WalletRow({ icon, name, amount }: { icon: string; name: string; amount: string }) {
  return <div className="wallet-row"><span>{icon}</span><strong>{name}</strong><b>{amount}</b></div>;
}

function TransactionRow({ title, subtitle, amount, positive = false }: { title: string; subtitle: string; amount: string; positive?: boolean }) {
  return <div className="transaction-row"><span className="category-dot" /><div><strong>{title}</strong><p>{subtitle}</p></div><b className={positive ? "income" : "expense"}>{amount}</b><ChevronRight size={22} /></div>;
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

export default App;
