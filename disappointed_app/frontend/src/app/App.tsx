import { useEffect, useRef, useState, type ReactNode } from "react";
import { mergePendingTransactions } from "./pendingTransactions";
import {
  Bell,
  Bot,
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
  Search,
  User,
  Users,
  Wallet,
} from "lucide-react";

import { useOnlineStatus } from "./offline";
import { apiBaseURL } from "./apiClient";
import { ActionButton, Card, IconButton, InputControl, PillButton, SectionTitle, SheetFrame, cx } from "./components";
import { loadCurrentUser, logout, type AuthState } from "./auth";
import { clearUserDataCaches } from "./userDataCache";
import { buildTransactionInput, calendarDateInHoChiMinh } from "./transactionInput";
import { SearchPanel } from "./SearchPanel";
import { initializeOfflineStoreForUser, queueReceiptUpload, saveFinanceMirror } from "../offline/db";
import { OverviewScreen } from "../screens/OverviewScreen";
import { ReportsPanel } from "../screens/ReportsPanel";
import { TransactionsScreen } from "../screens/TransactionsScreen";
import { BudgetsScreen } from "../screens/BudgetsScreen";
import { AccountScreen } from "../screens/AccountScreen";
import { AgentScreen } from "../screens/AgentScreen";
import { PWAInstallPrompt, type InstallPromptEvent } from "../components/feedback/PWAInstallPrompt";
import { OperationError, operationFailure, type OperationFailure } from "../components/feedback/OperationError";
import { UnavailableAction } from "../components/feedback/UnavailableAction";
import { FilePickerInput } from "../components/inputs/FilePickerInput";
import { Select } from "../components/ui/select";
import { listPendingReceiptUploads, markReceiptUploadComplete } from "../offline/receipts";
import {
  discardLocalConflict,
  editAndRetryTransactionConflict,
  fullResync,
  keepServerConflict,
  listOpenConflicts,
  reconcileServerEpoch,
} from "../offline/conflicts";
import {
  archiveTransaction,
  archiveWallet,
  createTransaction,
  createCategory,
  createWallet,
  loadCategories,
  loadWalletCategorySettings,
  loadTransactions,
  loadWallets,
  setDefaultAIWallet,
  setWalletCategoryActive,
  updateCategory,
  updateTransaction,
  updateWallet,
  type CategorySummary,
  type WalletCategorySetting,
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
  confirmTransactionDraft,
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
  pauseRecurringSchedule,
  rejectTransactionDraft,
  resumeRecurringSchedule,
  updateBudget,
  updateEvent,
  updateObligation,
  updateRecurringSchedule,
  type BudgetInput,
  type BudgetPeriodType,
  type BudgetProgress,
  type EventInput,
  type EventSummary,
  type ObligationDirection,
  type ObligationInput,
  type ObligationSummary,
  type RecurrenceFrequency,
  type RecurringPostingMode,
  type RecurringSchedule,
  type RecurringScheduleInput,
  type TransactionDraft,
  type TransactionDraftDecision,
} from "./planning";
import { drainOutbox, readOutbox } from "./outbox";
import { loadNotifications, markNotificationRead, subscribeToPush, type NotificationNotice } from "./notifications";
import { loadDashboard, loadInsider, loadReport, loadWalletDetail, type Dashboard, type InsiderReport, type Report, type WalletDetail } from "./analytics";
import { addAssetPrice, addAssetTrade, archiveAsset, createAsset, loadAssets, loadPortfolioSummary, type AssetPosition, type AssetType, type PortfolioSummary, type TradeSide } from "./portfolio";
import { createAPIKey, loadAPIKeys, revokeAPIKey, type APIKeySummary, type CreatedAPIKey } from "./apiKeys";
import { checkAuditAccess, loadAuditEvents, type AuditEvent } from "./audit";
import { uploadFile } from "./receipts";
import type { OfflineConflict } from "../offline/types";

type Tab = "overview" | "transactions" | "agent" | "budgets" | "account";

const tabs: Array<{ id: Tab; label: string; icon: typeof Home }> = [
  { id: "overview", label: "Tổng quan", icon: Home },
  { id: "transactions", label: "Sổ giao dịch", icon: Wallet },
  { id: "agent", label: "Trợ lý", icon: Bot },
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
  const [installPrompt, setInstallPrompt] = useState<InstallPromptEvent | null>(null);
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
  const [reportsOpen, setReportsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [walletDetail, setWalletDetail] = useState<WalletDetail | null>(null);
  const [walletDetailFailure, setWalletDetailFailure] = useState<OperationFailure | null>(null);
  const [walletDetailBusy, setWalletDetailBusy] = useState(false);
  const [readFailure, setReadFailure] = useState<OperationFailure | null>(null);
  const [readBusy, setReadBusy] = useState(false);
  const [notificationFailure, setNotificationFailure] = useState<OperationFailure | null>(null);
  const [notificationBusyID, setNotificationBusyID] = useState<string | null>(null);
  const [failedNotificationID, setFailedNotificationID] = useState<string | null>(null);
  const [insiderFailure, setInsiderFailure] = useState<OperationFailure | null>(null);
  const [insiderBusy, setInsiderBusy] = useState(false);
  const [logoutFailure, setLogoutFailure] = useState<OperationFailure | null>(null);
  const [logoutBusy, setLogoutBusy] = useState(false);
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
  const authenticatedUserID = authState.status === "authenticated" ? authState.user.id : "";
  const ownerRef = useRef(authenticatedUserID);
  const refreshGeneration = useRef(0);
  const walletDetailGeneration = useRef(0);
  const walletDetailSelection = useRef<string | null>(null);
  const walletDetailBusyID = useRef<string | null>(null);
  const notificationGeneration = useRef(0);
  const notificationBusyIDRef = useRef<string | null>(null);
  const insiderGeneration = useRef(0);
  const insiderBusyRef = useRef(false);
  const logoutGeneration = useRef(0);
  const logoutBusyRef = useRef(false);
  const draftRefreshGeneration = useRef(0);
  ownerRef.current = authenticatedUserID;
  const headerWallets = wallets ?? [];
  const totalBalance = dashboard?.net_worth_vnd ?? totalIncludedVND(headerWallets);

  useEffect(() => () => {
    ownerRef.current = "";
    ++refreshGeneration.current;
    ++walletDetailGeneration.current;
    ++notificationGeneration.current;
    ++insiderGeneration.current;
    ++logoutGeneration.current;
    ++draftRefreshGeneration.current;
  }, []);

  useEffect(() => {
    const onInstallPrompt = (event: Event) => { event.preventDefault(); setInstallPrompt(event as InstallPromptEvent); };
    const onInstalled = () => setInstallPrompt(null);
    window.addEventListener("beforeinstallprompt", onInstallPrompt);
    window.addEventListener("appinstalled", onInstalled);
    return () => { window.removeEventListener("beforeinstallprompt", onInstallPrompt); window.removeEventListener("appinstalled", onInstalled); };
  }, []);

  useEffect(() => {
    if (window.location.pathname === "/auth/google") {
      window.location.replace(`${apiBaseURL()}/api/v1/auth/google`);
      return;
    }
    let cancelled = false;
    void loadCurrentUser().then(async (nextAuthState) => {
      if (nextAuthState.status === "authenticated") {
        await initializeOfflineStoreForUser(nextAuthState.user.id);
      }
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
      ++refreshGeneration.current;
      ++walletDetailGeneration.current;
      ++notificationGeneration.current;
      ++insiderGeneration.current;
      ++draftRefreshGeneration.current;
      walletDetailSelection.current = null;
      walletDetailBusyID.current = null;
      notificationBusyIDRef.current = null;
      insiderBusyRef.current = false;
      setSearchQuery("");
      setSearchOpen(false);
      setReportsOpen(false);
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
      setWalletDetailFailure(null);
      setWalletDetailBusy(false);
      setReadFailure(null);
      setReadBusy(false);
      setNotificationFailure(null);
      setNotificationBusyID(null);
      setFailedNotificationID(null);
      setInsiderFailure(null);
      setInsiderBusy(false);
      setLogoutFailure(null);
      setLogoutBusy(false);
      setAssets(null);
      setPortfolioSummary(null);
      setConflicts([]);
      return;
    }
    let cancelled = false;
    const load = online
      ? async () => {
        await initializeOfflineStoreForUser(authState.user.id);
        // Epoch reconciliation is a retryable cache upgrade. A temporarily old
        // or unavailable sync endpoint must never block the live finance reads.
        await reconcileServerEpoch().catch(() => false);
        if (!cancelled) await refreshFinanceData();
      }
      : () => hydrateOfflineData(authState.user.id);
    void load().catch((error) => {
      if (!cancelled) setReadFailure(operationFailure(error, "Không tải được dữ liệu. Dữ liệu đã xác nhận vẫn được giữ lại."));
    });
    return () => {
      cancelled = true;
      ++refreshGeneration.current;
    };
  }, [authState.status, authenticatedUserID, online]);

  useEffect(() => {
    if (authState.status !== "authenticated" || !online) return;
    void readOutbox().then((pending) => {
      if (pending.length === 0) return;
      void drainOutbox(async (input) => { await createTransaction(input as Parameters<typeof createTransaction>[0]); })
        .then(() => refreshFinanceData())
        .then(() => drainPendingReceiptUploads())
        .then(() => refreshConflictState())
        .catch(() => undefined);
    }).catch(() => undefined);
  }, [authState.status, online]);

  async function drainPendingReceiptUploads() {
    if (!online) return;
    const records = await listPendingReceiptUploads();
    if (records.length === 0) return;
    const currentTransactions = await loadTransactions();
    for (const record of records) {
      const current = currentTransactions.find((item) => item.id === record.transaction_id);
      if (!current) continue;
      const receipt = await uploadFile(new File([record.file], record.filename, { type: record.content_type }));
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
    if (authState.status !== "authenticated" || logoutBusyRef.current) return;
    const userID = authState.user.id;
    const generation = ++logoutGeneration.current;
    logoutBusyRef.current = true;
    setLogoutBusy(true);
    setLogoutFailure(null);
    closeWalletDetail();
    ++refreshGeneration.current;
    ++notificationGeneration.current;
    ++insiderGeneration.current;
    notificationBusyIDRef.current = null;
    insiderBusyRef.current = false;
    setNotificationBusyID(null);
    setInsiderBusy(false);
    try {
      await logout();
    } catch (error) {
      if (generation === logoutGeneration.current && ownerRef.current === userID) {
        setLogoutFailure(operationFailure(error, "Không đăng xuất được. Phiên trên máy chủ có thể vẫn còn hoạt động; bạn vẫn đang đăng nhập. Hãy thử lại."));
      }
      return;
    } finally {
      if (generation === logoutGeneration.current) {
        logoutBusyRef.current = false;
        setLogoutBusy(false);
      }
    }
    if (generation !== logoutGeneration.current || ownerRef.current !== userID) return;
    clearUserDataCaches(userID);
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
    if (authState.status !== "authenticated") return;
    const userID = authState.user.id;
    const generation = ++refreshGeneration.current;
    const ownsRefresh = () => generation === refreshGeneration.current && ownerRef.current === userID;
    let financeData: { wallets: WalletSummary[]; categories: CategorySummary[]; transactions: Transaction[] } | null = null;
    let refreshedAssets: AssetPosition[] | null = null;
    setReadFailure(null);
    setReadBusy(true);
    const recordReadFailure = (error: unknown) => {
      if (ownsRefresh()) setReadFailure(operationFailure(error, "Không tải được một phần dữ liệu. Dữ liệu đã xác nhận vẫn được giữ lại."));
    };

    const financeRefresh = Promise.all([loadWallets(), loadCategories(), loadTransactions()])
      .then(([nextWallets, nextCategories, nextTransactions]) => {
        if (!ownsRefresh()) return;
        financeData = { wallets: nextWallets, categories: nextCategories, transactions: nextTransactions };
        setWallets(nextWallets);
        setCategories(nextCategories);
        setTransactions(nextTransactions);
      })
      .catch(recordReadFailure);
    const planningRefresh = Promise.all([loadBudgets(), loadEvents(), loadObligations(), loadRecurringSchedules(), loadTransactionDrafts()])
      .then(([nextBudgets, nextEvents, nextObligations, nextSchedules, nextDrafts]) => {
        if (!ownsRefresh()) return;
        setBudgets(nextBudgets);
        setEvents(nextEvents);
        setObligations(nextObligations);
        setSchedules(nextSchedules);
        setDrafts(nextDrafts);
      })
      .catch(recordReadFailure);
    const notificationRefresh = loadNotifications()
      .then((nextNotifications) => { if (ownsRefresh()) setNotifications(nextNotifications); })
      .catch(recordReadFailure);
    const dashboardRefresh = loadDashboard(userID)
      .then((nextDashboard) => { if (ownsRefresh()) setDashboard(nextDashboard); })
      .catch(recordReadFailure);
    const reportRefresh = loadReport(userID, "daily")
      .then((nextReport) => { if (ownsRefresh()) setReport(nextReport); })
      .catch(recordReadFailure);
    const insiderRefresh = refreshInsider(userID);
    const assetRefresh = loadAssets(userID)
      .then((nextAssets) => { if (ownsRefresh()) { refreshedAssets = nextAssets; setAssets(nextAssets); } })
      .catch(recordReadFailure);
    const portfolioRefresh = loadPortfolioSummary(userID)
      .then((nextPortfolioSummary) => { if (ownsRefresh()) setPortfolioSummary(nextPortfolioSummary); })
      .catch(recordReadFailure);

    await Promise.all([financeRefresh, planningRefresh, notificationRefresh, dashboardRefresh, reportRefresh, insiderRefresh, assetRefresh, portfolioRefresh]);
    if (!ownsRefresh()) return;
    if (financeData) {
      const confirmed = financeData as { wallets: WalletSummary[]; categories: CategorySummary[]; transactions: Transaction[] };
      try {
        await saveFinanceMirror({ userID, wallets: confirmed.wallets, categories: confirmed.categories, transactions: confirmed.transactions, assets: refreshedAssets ?? assets ?? [] });
        if (!ownsRefresh()) return;
        const queued = await readOutbox();
        if (!ownsRefresh()) return;
        setOfflineStatus((current) => ({ ...current, pending: queued.length }));
        const nextConflicts = await listOpenConflicts();
        if (!ownsRefresh()) return;
        setConflicts(nextConflicts);
        setTransactions(mergePendingTransactions(confirmed.transactions, queued));
      } catch (error) {
        recordReadFailure(error);
      }
    }
    if (ownsRefresh()) setReadBusy(false);
  }

  async function refreshInsider(userID = authenticatedUserID) {
    if (!userID || insiderBusyRef.current) return;
    const generation = ++insiderGeneration.current;
    insiderBusyRef.current = true;
    setInsiderBusy(true);
    setInsiderFailure(null);
    try {
      const nextInsider = await loadInsider(userID);
      if (generation === insiderGeneration.current && ownerRef.current === userID) setInsider(nextInsider);
    } catch (error) {
      if (generation === insiderGeneration.current && ownerRef.current === userID && insider) {
        setInsiderFailure(operationFailure(error, "Không tải được Money Insider. Dữ liệu đã xác nhận vẫn được giữ lại."));
      }
    } finally {
      if (generation === insiderGeneration.current && ownerRef.current === userID) {
        insiderBusyRef.current = false;
        setInsiderBusy(false);
      }
    }
  }

  function handleDraftDecision(decision: TransactionDraftDecision) {
    if (authState.status !== "authenticated") return;
    ++refreshGeneration.current;
    setReadBusy(false);
    setDrafts((current) => [decision.draft, ...(current ?? []).filter((draft) => draft.id !== decision.draft.id)]);
    if (decision.transaction) {
      setTransactions((current) => [decision.transaction!, ...(current ?? []).filter((transaction) => transaction.id !== decision.transaction!.id)]);
    }
    void refreshAfterDraftDecision(authState.user.id);
  }

  async function refreshAfterDraftDecision(userID: string) {
    const generation = ++draftRefreshGeneration.current;
    const ownsRefresh = () => generation === draftRefreshGeneration.current && ownerRef.current === userID;
    try {
      const [nextDrafts, nextWallets, nextTransactions, nextBudgets, nextDashboard] = await Promise.all([
        loadTransactionDrafts(),
        loadWallets(),
        loadTransactions(),
        loadBudgets(),
        loadDashboard(userID),
      ]);
      if (!ownsRefresh()) return;
      setDrafts(nextDrafts);
      setWallets(nextWallets);
      setTransactions(nextTransactions);
      setBudgets(nextBudgets);
      setDashboard(nextDashboard);
      await saveFinanceMirror({ userID, wallets: nextWallets, categories: categories ?? [], transactions: nextTransactions, assets: assets ?? [] });
    } catch (error) {
      if (ownsRefresh()) {
        setReadFailure(operationFailure(error, "Quyết định bản nháp đã được ghi nhận nhưng chưa tải lại được toàn bộ dữ liệu đã xác nhận."));
      }
    }
  }

  async function openWalletDetail(walletID: string) {
    if (!online || authState.status !== "authenticated" || walletDetailBusyID.current === walletID) return;
    const userID = authState.user.id;
    const generation = ++walletDetailGeneration.current;
    walletDetailSelection.current = walletID;
    walletDetailBusyID.current = walletID;
    setWalletDetailBusy(true);
    setWalletDetailFailure(null);
    try {
      const detail = await loadWalletDetail(walletID);
      if (generation === walletDetailGeneration.current && ownerRef.current === userID && walletDetailSelection.current === walletID) setWalletDetail(detail);
    } catch (error) {
      if (generation === walletDetailGeneration.current && ownerRef.current === userID && walletDetailSelection.current === walletID) {
        setWalletDetailFailure(operationFailure(error, "Không tải được chi tiết ví. Hãy thử lại."));
      }
    } finally {
      if (generation === walletDetailGeneration.current && ownerRef.current === userID && walletDetailSelection.current === walletID) {
        walletDetailBusyID.current = null;
        setWalletDetailBusy(false);
      }
    }
  }

  function closeWalletDetail() {
    ++walletDetailGeneration.current;
    walletDetailSelection.current = null;
    walletDetailBusyID.current = null;
    setWalletDetail(null);
    setWalletDetailFailure(null);
    setWalletDetailBusy(false);
  }

  async function markNoticeRead(notificationID: string) {
    if (authState.status !== "authenticated" || notificationBusyIDRef.current) return;
    const userID = authState.user.id;
    const generation = ++notificationGeneration.current;
    notificationBusyIDRef.current = notificationID;
    setNotificationBusyID(notificationID);
    setNotificationFailure(null);
    try {
      await markNotificationRead(notificationID);
      if (generation === notificationGeneration.current && ownerRef.current === userID) {
        setNotifications((current) => (current ?? []).map((item) => item.id === notificationID ? { ...item, read_at: new Date().toISOString() } : item));
        setFailedNotificationID(null);
      }
    } catch (error) {
      if (generation === notificationGeneration.current && ownerRef.current === userID) {
        setFailedNotificationID(notificationID);
        setNotificationFailure(operationFailure(error, "Chưa đánh dấu đã đọc. Thông báo vẫn được giữ là chưa đọc."));
      }
    } finally {
      if (generation === notificationGeneration.current && ownerRef.current === userID) {
        notificationBusyIDRef.current = null;
        setNotificationBusyID(null);
      }
    }
  }

  async function hydrateOfflineData(userID: string) {
    const snapshot = await initializeOfflineStoreForUser(userID);
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
    if (authState.status === "authenticated") await hydrateOfflineData(authState.user.id);
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
        <header className="home-header command-header">
          <div>
            <span className="command-label">Money Command</span>
            <h1>Today Desk</h1>
            <div className="balance-line">
              <strong>{privacyMasked ? "••••••" : formatVND(totalBalance)}</strong>
              <ActionButton className="icon-button" aria-label={privacyMasked ? "Hiện số dư" : "Ẩn số dư"} type="button" onClick={() => setPrivacyMasked((masked) => { const next = !masked; localStorage.setItem("mypocket:privacy-masked", String(next)); return next; })}>
                <Eye size={24} />
              </ActionButton>
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

        {searchOpen ? <SearchPanel key={authenticatedUserID} userID={authenticatedUserID} online={online} query={searchQuery} onQueryChange={setSearchQuery} /> : null}

        {notificationOpen ? <NotificationInbox online={online} notices={notifications ?? []} pushState={pushState} failure={notificationFailure} busyID={notificationBusyID} onEnablePush={() => { void subscribeToPush().then(() => setPushState("enabled")).catch((error: Error) => setPushState(error.message === "denied" ? "denied" : error.message === "unsupported" || error.message === "unconfigured" ? "unsupported" : error.message === "offline" ? "offline" : "failed")); }} onRead={(id) => { void markNoticeRead(id); }} onRetry={() => { if (failedNotificationID) void markNoticeRead(failedNotificationID); }} /> : null}

        <AuthBanner authState={authState} />
        {activeTab === "budgets" ? <OperationError failure={readFailure} onRetry={() => { void refreshFinanceData(); }} retryLabel="Thử lại" busy={readBusy} /> : null}
        {authState.status === "authenticated" ? <ConflictInbox conflicts={conflicts} onResolve={handleConflictAction} /> : null}
        {activeTab === "overview" ? reportsOpen && authState.status === "authenticated" ? <ReportsPanel key={authState.user.id} userID={authState.user.id} online={online} privacyMasked={privacyMasked} wallets={wallets ?? []} onClose={() => setReportsOpen(false)} /> : <>
          <OperationError failure={walletDetailFailure} onRetry={() => { if (walletDetailSelection.current) void openWalletDetail(walletDetailSelection.current); }} retryLabel="Thử lại" busy={walletDetailBusy} />
          <OperationError failure={insiderFailure} onRetry={() => { void refreshInsider(); }} retryLabel="Thử lại" busy={insiderBusy} />
          <OverviewScreen online={online} wallets={wallets} dashboard={dashboard} report={report} insider={insider} privacyMasked={privacyMasked} walletDetail={walletDetail} onCloseWalletDetail={closeWalletDetail} onManageWallets={() => setWalletSheetOpen(true)} onViewReports={() => setReportsOpen(true)} onWalletClick={(walletID) => { void openWalletDetail(walletID); }} onRefreshInsider={() => { void refreshInsider(); }} formatVND={formatVND} formatPercent={formatPercent} />
        </> : null}
        {activeTab === "transactions" ? <TransactionsScreen transactions={transactions ?? []} onEdit={setEditingTransaction} /> : null}
        {activeTab === "agent" ? <AgentScreen online={online} onDraftReady={() => { setActiveTab("budgets"); void refreshFinanceData(); }} /> : null}
        {activeTab === "budgets" ? <>
          <BudgetsScreen budgets={budgets} events={events ?? []} obligations={obligations ?? []} schedules={schedules ?? []} drafts={[]} categories={categories ?? []} wallets={wallets ?? []} transactions={transactions ?? []} online={online} onCreate={() => setBudgetSheetOpen(true)} onCreateEvent={() => setEventSheetOpen(true)} onCreateObligation={() => setObligationSheetOpen(true)} onCreateSchedule={() => setScheduleSheetOpen(true)} onEdit={setEditingBudget} onEditEvent={setEditingEvent} onEditObligation={setEditingObligation} onEditSchedule={setEditingSchedule} formatVND={formatVND} formatDate={formatDate} obligationDirectionLabel={obligationDirectionLabel} recurrenceLabel={recurrenceLabel} transactionTypeLabel={transactionTypeLabel} />
          <DraftDecisionPanel key={authenticatedUserID} drafts={drafts ?? []} wallets={wallets ?? []} online={online} onDecided={handleDraftDecision} />
        </> : null}
        {activeTab === "account" ? <>
          {logoutBusy ? <p role="status">Đang đăng xuất…</p> : null}
          <OperationError failure={logoutFailure} onRetry={() => { void handleLogout(); }} retryLabel="Thử lại" busy={logoutBusy} />
          <AccountScreen authState={authState} assets={assets ?? []} portfolioSummary={portfolioSummary} privacyMasked={privacyMasked} online={online} onCreateAsset={() => setAssetSheetOpen(true)} onAssetChanged={(asset) => { setAssets((current) => [asset, ...(current ?? []).filter((item) => item.id !== asset.id)]); void refreshFinanceData(); }} onAssetArchived={(assetID) => { setAssets((current) => (current ?? []).filter((item) => item.id !== assetID)); void refreshFinanceData(); }} onLogout={handleLogout} formatVND={formatVND} />
        </> : null}
      </main>

      <nav className="bottom-nav dock-nav" aria-label="Điều hướng chính">
        {tabs.map((tab) => (
          <TabButton key={tab.id} tab={tab} active={activeTab === tab.id} onClick={() => setActiveTab(tab.id)} />
        ))}
      </nav>
      <ActionButton className="add-button floating-add" aria-label="Thêm giao dịch" type="button" disabled={(wallets ?? []).length === 0} onClick={() => setSheetOpen(true)}>
        <Plus size={30} />
      </ActionButton>
      <PWAInstallPrompt prompt={installPrompt} onConsumed={() => setInstallPrompt(null)} />

      {sheetOpen ? <AddTransactionSheet categories={categories ?? []} wallets={wallets ?? []} budgets={budgets ?? []} readOnly={offlineReadOnly} onCreated={upsertTransaction} onDebtCreated={() => void refreshFinanceData()} onClose={() => setSheetOpen(false)} /> : null}
      {walletSheetOpen ? <WalletManagerSheet wallets={wallets ?? []} categories={categories ?? []} online={online} readOnly={offlineReadOnly} onWalletChanged={(wallet) => setWallets((current) => [wallet, ...(current ?? []).filter((item) => item.id !== wallet.id)])} onWalletArchived={(walletID) => setWallets((current) => (current ?? []).filter((item) => item.id !== walletID))} onCategoryChanged={(category) => setCategories((current) => [category, ...(current ?? []).filter((item) => item.id !== category.id)])} onChanged={() => reconcileAfterLocalChange()} onClose={() => setWalletSheetOpen(false)} /> : null}
      {editingTransaction ? <EditTransactionSheet categories={categories ?? []} wallets={wallets ?? []} budgets={budgets ?? []} readOnly={offlineReadOnly} transaction={editingTransaction} onChanged={upsertTransaction} onArchived={() => { setTransactions((current) => (current ?? []).filter((item) => item.id !== editingTransaction.id)); setEditingTransaction(null); void reconcileAfterLocalChange().catch(() => undefined); }} onClose={() => setEditingTransaction(null)} /> : null}
      {budgetSheetOpen ? <BudgetSheet categories={categories ?? []} onSaved={() => { setBudgetSheetOpen(false); void refreshFinanceData(); }} onClose={() => setBudgetSheetOpen(false)} /> : null}
      {editingBudget ? <BudgetSheet budget={editingBudget} categories={categories ?? []} onSaved={() => { setEditingBudget(null); void refreshFinanceData(); }} onArchived={() => { setEditingBudget(null); void refreshFinanceData(); }} onClose={() => setEditingBudget(null)} /> : null}
      {eventSheetOpen ? <EventSheet transactions={transactions ?? []} onSaved={() => { setEventSheetOpen(false); void refreshFinanceData(); }} onClose={() => setEventSheetOpen(false)} /> : null}
      {editingEvent ? <EventSheet event={editingEvent} transactions={transactions ?? []} onSaved={() => { setEditingEvent(null); void refreshFinanceData(); }} onArchived={() => { setEditingEvent(null); void refreshFinanceData(); }} onClose={() => setEditingEvent(null)} /> : null}
      {obligationSheetOpen ? <ObligationSheet transactions={transactions ?? []} onSaved={() => { setObligationSheetOpen(false); void refreshFinanceData(); }} onClose={() => setObligationSheetOpen(false)} /> : null}
      {editingObligation ? <ObligationSheet obligation={editingObligation} transactions={transactions ?? []} onSaved={() => { setEditingObligation(null); void refreshFinanceData(); }} onArchived={() => { setEditingObligation(null); void refreshFinanceData(); }} onClose={() => setEditingObligation(null)} /> : null}
      {scheduleSheetOpen ? <ScheduleSheet wallets={wallets ?? []} categories={categories ?? []} budgets={budgets ?? []} onSaved={() => { setScheduleSheetOpen(false); void refreshFinanceData(); }} onClose={() => setScheduleSheetOpen(false)} /> : null}
      {editingSchedule ? <ScheduleSheet schedule={editingSchedule} wallets={wallets ?? []} categories={categories ?? []} budgets={budgets ?? []} onSaved={() => { setEditingSchedule(null); void refreshFinanceData(); }} onArchived={() => { setEditingSchedule(null); void refreshFinanceData(); }} onClose={() => setEditingSchedule(null)} /> : null}
      {assetSheetOpen ? <AssetSheet onSaved={(asset) => { setAssetSheetOpen(false); setAssets((current) => [asset, ...(current ?? [])]); void refreshFinanceData(); }} onClose={() => setAssetSheetOpen(false)} /> : null}
    </div>
  );
}

function DraftDecisionPanel({ drafts, wallets, online, onDecided }: {
  drafts: TransactionDraft[];
  wallets: WalletSummary[];
  online: boolean;
  onDecided: (decision: TransactionDraftDecision) => void;
}) {
  return (
    <Card className="list-card planning-list" aria-label="Duyệt bản nháp giao dịch">
      <SectionTitle title="Bản nháp giao dịch" />
      {!online && drafts.some((draft) => draft.status === "pending") ? (
        <p className="offline-warning">Cần online để xác nhận hoặc từ chối bản nháp. Thay đổi chưa được xếp hàng chờ.</p>
      ) : null}
      {drafts.length === 0 ? <p className="empty-state">Chưa có bản nháp giao dịch</p> : drafts.map((draft) => (
        <DraftDecisionRow key={draft.id} draft={draft} wallets={wallets} online={online} onDecided={onDecided} />
      ))}
    </Card>
  );
}

function DraftDecisionRow({ draft, wallets, online, onDecided }: {
  draft: TransactionDraft;
  wallets: WalletSummary[];
  online: boolean;
  onDecided: (decision: TransactionDraftDecision) => void;
}) {
  const label = draft.note.trim() || `Bản nháp ${draft.id}`;
  const [amount, setAmount] = useState(String(draft.amount_vnd));
  const [note, setNote] = useState(draft.note);
  const [busy, setBusy] = useState(false);
  const [failure, setFailure] = useState<OperationFailure | null>(null);
  const [failedAction, setFailedAction] = useState<"confirm" | "reject" | null>(null);
  const busyRef = useRef(false);
  const confirmAttemptRef = useRef<{ key: string; input: { version: number; amount_vnd: number; note: string } } | null>(null);
  const mountedRef = useRef(true);

  useEffect(() => () => { mountedRef.current = false; }, []);

  const sourceName = wallets.find((wallet) => wallet.id === draft.source_wallet_id)?.name ?? draft.source_wallet_id;
  const destinationName = draft.destination_wallet_id
    ? wallets.find((wallet) => wallet.id === draft.destination_wallet_id)?.name ?? draft.destination_wallet_id
    : "";
  const route = draft.type === "transfer" ? `${sourceName} → ${destinationName}` : sourceName;
  const numericAmount = Number(amount);
  const canConfirm = Number.isSafeInteger(numericAmount) && numericAmount > 0;

  function clearConfirmRetry() {
    if (!confirmAttemptRef.current) return;
    confirmAttemptRef.current = null;
    if (failedAction === "confirm") {
      setFailedAction(null);
      setFailure(null);
    }
  }

  async function decide(action: "confirm" | "reject") {
    if (!online || draft.status !== "pending" || busyRef.current || (action === "confirm" && !canConfirm)) return;
    busyRef.current = true;
    setBusy(true);
    setFailure(null);
    try {
      let decision: TransactionDraftDecision;
      if (action === "confirm") {
        const attempt = confirmAttemptRef.current ?? {
          key: createDraftDecisionKey(),
          input: { version: draft.version, amount_vnd: numericAmount, note },
        };
        confirmAttemptRef.current = attempt;
        decision = await confirmTransactionDraft(draft.id, attempt.input, attempt.key);
      } else {
        decision = await rejectTransactionDraft(draft.id, { version: draft.version });
      }
      if (!mountedRef.current) return;
      setFailedAction(null);
      onDecided(decision);
    } catch (error) {
      if (!mountedRef.current) return;
      setFailedAction(action);
      setFailure(operationFailure(error, action === "confirm"
        ? "Chưa xác nhận được bản nháp. Số tiền và ghi chú vẫn được giữ; hãy thử lại."
        : "Chưa từ chối được bản nháp. Nội dung vẫn được giữ; hãy thử lại."));
    } finally {
      if (mountedRef.current) {
        busyRef.current = false;
        setBusy(false);
      }
    }
  }

  return (
    <article className="planning-row draft-row" aria-label={`Bản nháp ${label}`}>
      <span className="category-dot draft-dot" />
      <div>
        <strong>{label}</strong>
        <p>{transactionTypeLabel(draft.type)} · {route}</p>
        <p>{new Date(draft.occurred_at).toLocaleDateString("vi-VN")} · {formatVND(draft.amount_vnd)}</p>
        {draft.status === "pending" ? <p>{formatVND(draft.amount_vnd)} · Chờ duyệt</p> : null}
        {draft.status === "pending" ? <>
          <label className="form-row">
            Số tiền
            <InputControl
              aria-label={`Số tiền bản nháp ${label}`}
              inputMode="numeric"
              value={amount}
              disabled={busy}
              onChange={(event) => {
                setAmount(event.target.value.replace(/\D/g, ""));
                clearConfirmRetry();
              }}
            />
          </label>
          <label className="form-row">
            Ghi chú
            <InputControl aria-label={`Ghi chú bản nháp ${label}`} value={note} disabled={busy} onChange={(event) => {
              setNote(event.target.value);
              clearConfirmRetry();
            }} />
          </label>
          <OperationError
            failure={failure}
            onRetry={failedAction ? () => { void decide(failedAction); } : undefined}
            retryLabel={failedAction === "reject" ? "Thử từ chối" : "Thử xác nhận"}
            busy={busy}
          />
          <div className="conflict-actions">
            <PillButton disabled={!online || busy || !canConfirm} onClick={() => { void decide("confirm"); }}>Xác nhận {label}</PillButton>
            <PillButton disabled={!online || busy} onClick={() => { void decide("reject"); }}>Từ chối {label}</PillButton>
          </div>
        </> : <>
          <p>{draft.status === "confirmed" ? "Đã xác nhận" : "Đã từ chối"}</p>
          {draft.confirmed_transaction_id ? <p>Giao dịch: {draft.confirmed_transaction_id}</p> : null}
        </>}
      </div>
    </article>
  );
}

function createDraftDecisionKey() {
  return `draft-confirm-${globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`}`;
}

function ConflictInbox({ conflicts, onResolve }: { conflicts: OfflineConflict[]; onResolve: (action: () => Promise<void>) => Promise<void> }) {
  if (conflicts.length === 0) return null;
  return (
    <section className="card conflict-inbox" aria-label="Xung đột đồng bộ">
      <div className="section-title">
        <h2>Cần xử lý</h2>
        <ActionButton type="button" onClick={() => void onResolve(fullResync)}>Đồng bộ lại</ActionButton>
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
  failure,
  busyID,
  onEnablePush,
  onRead,
  onRetry,
}: {
  online: boolean;
  notices: NotificationNotice[];
  pushState: "idle" | "enabled" | "denied" | "unsupported" | "offline" | "failed";
  failure: OperationFailure | null;
  busyID: string | null;
  onEnablePush: () => void;
  onRead: (id: string) => void;
  onRetry: () => void;
}) {
  const pushMessage = pushState === "denied" ? "Bạn đã từ chối quyền thông báo" : pushState === "unsupported" ? "Trình duyệt chưa hỗ trợ Web Push" : pushState === "offline" ? "Kết nối mạng để bật Web Push" : pushState === "failed" ? "Không thể bật Web Push lúc này" : pushState === "enabled" ? "Web Push đã bật" : "";
  return (
    <section className="card notification-inbox" role="dialog" aria-modal="false" aria-label="Hộp thư thông báo">
      <div className="section-title"><h2>Thông báo</h2><ActionButton type="button" onClick={onEnablePush} disabled={!online || pushState === "enabled"}>Bật Web Push</ActionButton></div>
      {pushMessage ? <p className="notification-status">{pushMessage}</p> : null}
      <OperationError failure={failure} onRetry={onRetry} retryLabel="Thử lại" busy={Boolean(busyID)} />
      {!online && notices.length === 0 ? <p className="notification-status">Đang offline. Hộp thư sẽ tải lại khi có mạng.</p> : null}
      {notices.length === 0 && online ? <p className="notification-status">Chưa có thông báo mới.</p> : null}
      {notices.map((notice) => <ActionButton className={notice.read_at ? "notice-row read" : "notice-row"} disabled={busyID === notice.id} key={notice.id} type="button" onClick={() => !notice.read_at && onRead(notice.id)}><span><strong>{notice.title}</strong><small>{notice.body}</small></span><time>{new Date(notice.created_at).toLocaleDateString("vi-VN")}</time></ActionButton>)}
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
          <InputControl aria-label={`Số tiền xử lý ${conflict.entity_id}`} inputMode="numeric" value={amount} onChange={(event) => setAmount(event.target.value.replace(/\D/g, ""))} />
          <InputControl aria-label={`Ghi chú xử lý ${conflict.entity_id}`} value={note} onChange={(event) => setNote(event.target.value)} />
        </div>
      ) : null}
      <div className="conflict-actions">
        <ActionButton type="button" onClick={() => void onResolve(() => keepServerConflict(conflict))}>Giữ server</ActionButton>
        {canRetryTransaction ? <ActionButton type="button" onClick={() => void onResolve(() => editAndRetryTransactionConflict(conflict, { amount_vnd: Number(amount), note }))}>Sửa gửi lại</ActionButton> : null}
        <ActionButton type="button" onClick={() => void onResolve(() => discardLocalConflict(conflict))}>Bỏ offline</ActionButton>
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
        <ActionButton className="primary-cta login-button" type="button" onClick={() => { window.location.href = `${apiBaseURL()}/api/v1/auth/google`; }}>
          Đăng nhập bằng Google
        </ActionButton>
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
  return <main className="auth-gate"><section className="auth-gate-panel"><div className="brand-mark">MyPocket</div><h1>Quản lý tiền rõ ràng hơn</h1><p>Đăng nhập để xem ví, giao dịch và kế hoạch của bạn.</p><ActionButton className="primary-cta login-button" type="button" onClick={() => { window.location.href = `${apiBaseURL()}/api/v1/auth/google`; }}>Đăng nhập bằng Google</ActionButton></section></main>;
}

function ForbiddenState({ authState, onLogout }: { authState: Extract<AuthState, { status: "forbidden" }>; onLogout: () => void }) {
  return (
    <section className="content-stack">
      <section className="card auth-state-card">
        <h1>Không có quyền truy cập</h1>
        <p>Phiên hiện tại không thể mở dữ liệu này.</p>
        {authState.correlationID ? <small>{authState.correlationID}</small> : null}
        <ActionButton className="wide-pill destructive" type="button" onClick={onLogout}>Đăng xuất</ActionButton>
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
    <ActionButton className={cx("tab-button", active && "active")} aria-label={tab.label} type="button" onClick={onClick}>
      <Icon size={27} strokeWidth={2.4} />
      <span>{tab.label}</span>
    </ActionButton>
  );
}

function EventSheet({ event, transactions, onSaved, onArchived, onClose }: { event?: EventSummary; transactions: Transaction[]; onSaved: () => void; onArchived?: () => void; onClose: () => void }) {
  const [name, setName] = useState(event?.name ?? "");
  const [startsOn, setStartsOn] = useState(event?.starts_on ?? calendarDateInHoChiMinh());
  const [endsOn, setEndsOn] = useState(event?.ends_on ?? calendarDateInHoChiMinh());
  const [note, setNote] = useState(event?.note ?? "");
  const [transactionID, setTransactionID] = useState("");
  const [saving, setSaving] = useState(false);
  const canSave = name.trim() !== "" && startsOn !== "" && endsOn !== "";
  async function save() {
    if (!canSave || saving) return;
    setSaving(true);
    try {
      const input: EventInput = { name, starts_on: startsOn, ends_on: endsOn, note };
      const saved = event ? await updateEvent(event.id, input, event.version) : await createEvent(input);
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
      await archiveEvent(event.id, event.version);
      onArchived?.();
    } finally {
      setSaving(false);
    }
  }
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label={event ? "Sửa Sự Kiện" : "Tạo Sự Kiện"}>
        <header>
          <ActionButton className="pill-button" type="button" onClick={onClose}>Hủy</ActionButton>
          <h2>{event ? "Sửa Sự Kiện" : "Tạo Sự Kiện"}</h2>
          <span />
        </header>
        <label className="sheet-row"><MapPin /><InputControl aria-label="Tên sự kiện" value={name} onChange={(change) => setName(change.target.value)} placeholder="Tên sự kiện" /></label>
        <div className="manager-form two">
          <InputControl aria-label="Ngày bắt đầu sự kiện" type="date" value={startsOn} onChange={(change) => setStartsOn(change.target.value)} />
          <InputControl aria-label="Ngày kết thúc sự kiện" type="date" value={endsOn} onChange={(change) => setEndsOn(change.target.value)} />
        </div>
        <label className="sheet-row"><List /><InputControl aria-label="Ghi chú sự kiện" value={note} onChange={(change) => setNote(change.target.value)} placeholder="Ghi chú" /></label>
        <label className="sheet-row"><Wallet /><Select aria-label="Giao dịch sự kiện" value={transactionID} onChange={(change) => setTransactionID(change.target.value)}><option value="">Không gắn giao dịch</option>{transactions.map((transaction) => <option key={transaction.id} value={transaction.id}>{transaction.note || transaction.type} · {formatVND(transaction.amount_vnd)}</option>)}</Select></label>
        <div className="sheet-actions">
          {event ? <ActionButton className="wide-pill destructive" type="button" disabled={saving} onClick={() => void archive()}>Lưu trữ</ActionButton> : null}
          <ActionButton className="primary-cta" type="button" disabled={!canSave || saving} onClick={() => void save()}>{saving ? "Đang lưu" : "Lưu"}</ActionButton>
        </div>
      </section>
    </div>
  );
}

function ObligationSheet({ obligation, transactions, onSaved, onArchived, onClose }: { obligation?: ObligationSummary; transactions: Transaction[]; onSaved: () => void; onArchived?: () => void; onClose: () => void }) {
  const [counterparty, setCounterparty] = useState(obligation?.counterparty ?? "");
  const [direction, setDirection] = useState<ObligationDirection>(obligation?.direction ?? "borrowed");
  const [principal, setPrincipal] = useState(String(obligation?.principal_vnd ?? ""));
  const [dueOn, setDueOn] = useState(obligation?.due_on ?? calendarDateInHoChiMinh());
  const [note, setNote] = useState(obligation?.note ?? "");
  const [transactionID, setTransactionID] = useState("");
  const [saving, setSaving] = useState(false);
  const canSave = counterparty.trim() !== "" && Number(principal) > 0 && dueOn !== "";
  async function save() {
    if (!canSave || saving) return;
    setSaving(true);
    try {
      const input: ObligationInput = { direction, principal_vnd: Number(principal), counterparty, due_on: dueOn, note };
      const saved = obligation ? await updateObligation(obligation.id, input, obligation.version) : await createObligation(input);
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
      await archiveObligation(obligation.id, obligation.version);
      onArchived?.();
    } finally {
      setSaving(false);
    }
  }
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label={obligation ? "Sửa Khoản Nợ" : "Tạo Khoản Nợ"}>
        <header>
          <ActionButton className="pill-button" type="button" onClick={onClose}>Hủy</ActionButton>
          <h2>{obligation ? "Sửa Khoản Nợ" : "Tạo Khoản Nợ"}</h2>
          <span />
        </header>
        <div className="segmented sheet-segmented"><ActionButton type="button" className={direction === "borrowed" ? "active" : ""} onClick={() => setDirection("borrowed")}>Tôi vay</ActionButton><ActionButton type="button" className={direction === "lent" ? "active" : ""} onClick={() => setDirection("lent")}>Tôi cho vay</ActionButton></div>
        <label className="sheet-row"><Users /><InputControl aria-label="Đối tác" value={counterparty} onChange={(change) => setCounterparty(change.target.value)} placeholder="Người liên quan" /></label>
        <label className="amount-row"><span>VND</span><InputControl aria-label="Số tiền gốc" inputMode="numeric" value={principal} onChange={(change) => setPrincipal(change.target.value.replace(/\D/g, ""))} placeholder="0" /></label>
        <label className="sheet-row"><CalendarDays /><InputControl aria-label="Ngày đến hạn" type="date" value={dueOn} onChange={(change) => setDueOn(change.target.value)} /></label>
        <label className="sheet-row"><List /><InputControl aria-label="Ghi chú khoản nợ" value={note} onChange={(change) => setNote(change.target.value)} placeholder="Ghi chú" /></label>
        <label className="sheet-row"><Wallet /><Select aria-label="Giao dịch trả nợ" value={transactionID} onChange={(change) => setTransactionID(change.target.value)}><option value="">Không gắn trả nợ</option>{transactions.map((transaction) => <option key={transaction.id} value={transaction.id}>{transaction.note || transaction.type} · {formatVND(transaction.amount_vnd)}</option>)}</Select></label>
        {obligation ? <p className="sheet-meta">Còn {formatVND(obligation.remaining_vnd)} · Đã trả {formatVND(obligation.repaid_vnd)}</p> : null}
        <div className="sheet-actions">
          {obligation ? <ActionButton className="wide-pill destructive" type="button" disabled={saving} onClick={() => void archive()}>Lưu trữ</ActionButton> : null}
          <ActionButton className="primary-cta" type="button" disabled={!canSave || saving} onClick={() => void save()}>{saving ? "Đang lưu" : "Lưu"}</ActionButton>
        </div>
      </section>
    </div>
  );
}

function ScheduleSheet({ schedule, wallets, categories, budgets, onSaved, onArchived, onClose }: { schedule?: RecurringSchedule; wallets: WalletSummary[]; categories: CategorySummary[]; budgets: BudgetProgress[]; onSaved: () => void; onArchived?: () => void; onClose: () => void }) {
  const [name, setName] = useState(schedule?.name ?? "");
  const [frequency, setFrequency] = useState<RecurrenceFrequency>(schedule?.frequency ?? "monthly");
  const [amount, setAmount] = useState(String(schedule?.amount_vnd ?? ""));
  const [type, setType] = useState<"income" | "expense" | "transfer">(schedule?.type ?? "expense");
  const [sourceWalletID, setSourceWalletID] = useState(schedule?.source_wallet_id ?? wallets[0]?.id ?? "");
  const [destinationWalletID, setDestinationWalletID] = useState(schedule?.destination_wallet_id ?? wallets.find((wallet) => wallet.id !== sourceWalletID)?.id ?? "");
  const [categoryID, setCategoryID] = useState(schedule?.category_id ?? "");
  const [budgetID, setBudgetID] = useState(schedule?.budget_id ?? "");
  const [startsOn, setStartsOn] = useState(schedule ? calendarDateInHoChiMinh(new Date(schedule.starts_at)) : calendarDateInHoChiMinh());
  const [endsOn, setEndsOn] = useState(schedule?.ends_at ? calendarDateInHoChiMinh(new Date(schedule.ends_at)) : "");
  const [postingMode, setPostingMode] = useState<RecurringPostingMode>(schedule?.posting_mode ?? "draft");
  const [note, setNote] = useState(schedule?.note ?? "");
  const [saving, setSaving] = useState(false);
  const selectableCategories = categories.filter((category) => category.kind === type);
  const activeBudgets = budgets.map((row) => row.budget);
  const chosenCategoryID = type === "transfer" ? "" : categoryID || selectableCategories[0]?.id || "";
  const canSave = name.trim() !== "" && Number(amount) > 0 && sourceWalletID !== "" && startsOn !== "" && (type === "transfer" ? destinationWalletID !== "" && destinationWalletID !== sourceWalletID : chosenCategoryID !== "");
  useEffect(() => {
    if (!sourceWalletID && wallets[0]) setSourceWalletID(wallets[0].id);
    if (type === "transfer" && (!destinationWalletID || destinationWalletID === sourceWalletID)) {
      setDestinationWalletID(wallets.find((wallet) => wallet.id !== sourceWalletID)?.id ?? "");
    }
  }, [destinationWalletID, sourceWalletID, type, wallets]);
  async function save() {
    if (!canSave || saving) return;
    setSaving(true);
    try {
      const input: RecurringScheduleInput = {
        name,
        frequency,
        timezone: "Asia/Ho_Chi_Minh",
        starts_at: `${startsOn}T09:00:00+07:00`,
        ends_at: endsOn ? `${endsOn}T23:59:59+07:00` : undefined,
        posting_mode: postingMode,
        type,
        source_wallet_id: sourceWalletID,
        destination_wallet_id: type === "transfer" ? destinationWalletID : undefined,
        category_id: type === "transfer" ? undefined : chosenCategoryID,
        budget_id: type === "expense" && budgetID ? budgetID : undefined,
        amount_vnd: Number(amount),
        note,
      };
      if (schedule) {
        await updateRecurringSchedule(schedule.id, input, schedule.version);
      } else {
        await createRecurringSchedule(input);
      }
      onSaved();
    } finally {
      setSaving(false);
    }
  }
  async function togglePaused() {
    if (!schedule || saving) return;
    setSaving(true);
    try {
      if (schedule.paused_at) {
        await resumeRecurringSchedule(schedule.id, schedule.version);
      } else {
        await pauseRecurringSchedule(schedule.id, schedule.version);
      }
      onSaved();
    } finally {
      setSaving(false);
    }
  }
  async function archive() {
    if (!schedule || saving) return;
    setSaving(true);
    try {
      await archiveRecurringSchedule(schedule.id, schedule.version);
      onArchived?.();
    } finally {
      setSaving(false);
    }
  }
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label={schedule ? "Sửa Lịch Lặp" : "Tạo Lịch Lặp"}>
        <header>
          <ActionButton className="pill-button" type="button" onClick={onClose}>Hủy</ActionButton>
          <h2>{schedule ? "Sửa Lịch Lặp" : "Tạo Lịch Lặp"}</h2>
          <span />
        </header>
        {schedule ? <p className="sheet-meta">{schedule.paused_at ? "Đang tạm dừng" : "Đang hoạt động"} · phiên bản {schedule.version}</p> : null}
        <label className="sheet-row"><CalendarDays /><InputControl aria-label="Tên lịch lặp" value={name} onChange={(change) => setName(change.target.value)} placeholder="Tên lịch lặp" /></label>
        <label className="amount-row"><span>VND</span><InputControl aria-label="Số tiền lịch lặp" inputMode="numeric" value={amount} onChange={(change) => setAmount(change.target.value.replace(/\D/g, ""))} placeholder="0" /></label>
        <label className="sheet-row"><CalendarDays /><Select aria-label="Chu kỳ lặp" value={frequency} onChange={(change) => setFrequency(change.target.value as RecurrenceFrequency)}><option value="daily">Hàng ngày</option><option value="weekly">Hàng tuần</option><option value="monthly">Hàng tháng</option></Select></label>
        <label className="sheet-row"><CalendarDays /><InputControl aria-label="Ngày bắt đầu lịch lặp" type="date" value={startsOn} onChange={(change) => setStartsOn(change.target.value)} /></label>
        <label className="sheet-row"><CalendarDays /><InputControl aria-label="Ngày kết thúc lịch lặp" type="date" value={endsOn} onChange={(change) => setEndsOn(change.target.value)} /></label>
        <label className="sheet-row"><CalendarDays /><Select aria-label="Cách ghi lịch lặp" value={postingMode} onChange={(change) => setPostingMode(change.target.value as RecurringPostingMode)}><option value="draft">Tạo nháp để duyệt</option><option value="auto_post">Tự ghi giao dịch</option></Select></label>
        <div className="segmented sheet-segmented">
          {(["expense", "income", "transfer"] as const).map((option) => <ActionButton className={type === option ? "active" : ""} type="button" key={option} onClick={() => setType(option)}>{transactionTypeLabel(option)}</ActionButton>)}
        </div>
        <label className="sheet-row"><Wallet /><Select aria-label="Ví lịch lặp" value={sourceWalletID} onChange={(change) => setSourceWalletID(change.target.value)}>{wallets.map((wallet) => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</Select></label>
        {type === "transfer" ? <label className="sheet-row"><Wallet /><Select aria-label="Ví nhận lịch lặp" value={destinationWalletID} onChange={(change) => setDestinationWalletID(change.target.value)}>{wallets.filter((wallet) => wallet.id !== sourceWalletID).map((wallet) => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</Select></label> : null}
        {type !== "transfer" ? <label className="sheet-row"><span className="dot-icon" /><Select aria-label="Nhóm lịch lặp" value={chosenCategoryID} onChange={(change) => setCategoryID(change.target.value)}>{selectableCategories.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}</Select></label> : null}
        {type === "expense" && activeBudgets.length > 0 ? <label className="sheet-row"><span className="dot-icon" /><Select aria-label="Ngân sách lịch lặp" value={budgetID} onChange={(change) => setBudgetID(change.target.value)}><option value="">Không gắn ngân sách</option>{activeBudgets.map((budget) => <option key={budget.id} value={budget.id}>{budget.name}</option>)}</Select></label> : null}
        <label className="sheet-row"><List /><InputControl aria-label="Ghi chú lịch lặp" value={note} onChange={(change) => setNote(change.target.value)} placeholder="Ghi chú" /></label>
        <div className="sheet-actions">
          {schedule ? <ActionButton className="wide-pill" type="button" disabled={saving} onClick={() => void togglePaused()}>{schedule.paused_at ? "Tiếp tục" : "Tạm dừng"}</ActionButton> : null}
          {schedule ? <ActionButton className="wide-pill destructive" type="button" disabled={saving} onClick={() => void archive()}>Lưu trữ</ActionButton> : null}
          <ActionButton className="primary-cta" type="button" disabled={!canSave || saving} onClick={() => void save()}>{saving ? "Đang lưu" : schedule ? "Lưu thay đổi" : "Lưu"}</ActionButton>
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
      if (editing) await updateBudget(editing.id, input(), editing.version);
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
      await archiveBudget(editing.id, editing.version);
      onArchived?.();
    } finally {
      setSaving(false);
    }
  }
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label={editing ? "Sửa Ngân Sách" : "Tạo Ngân Sách"}>
        <header>
          <ActionButton className="pill-button" type="button" onClick={onClose}>Hủy</ActionButton>
          <h2>{editing ? "Sửa Ngân Sách" : "Tạo Ngân Sách"}</h2>
          <span />
        </header>
        <label className="sheet-row"><List /><InputControl aria-label="Tên ngân sách" value={name} onChange={(event) => setName(event.target.value)} placeholder="Tên ngân sách" /></label>
        <label className="amount-row"><span>VND</span><InputControl aria-label="Số tiền ngân sách" inputMode="numeric" value={amount} onChange={(event) => setAmount(event.target.value.replace(/\D/g, ""))} placeholder="0" /></label>
        <label className="sheet-row"><CalendarDays /><Select aria-label="Kỳ ngân sách" value={periodType} onChange={(event) => setPeriodType(event.target.value as BudgetPeriodType)}>{budgetPeriods.map((period) => <option key={period} value={period}>{budgetPeriodLabel(period)}</option>)}</Select></label>
        {periodType === "custom" ? (
          <div className="manager-form two">
            <InputControl aria-label="Ngày bắt đầu" type="date" value={customStart} onChange={(event) => setCustomStart(event.target.value)} />
            <InputControl aria-label="Ngày kết thúc" type="date" value={customEnd} onChange={(event) => setCustomEnd(event.target.value)} />
          </div>
        ) : null}
        <ActionButton className={allCategories ? "toggle-row active" : "toggle-row"} type="button" onClick={() => setAllCategories((current) => !current)}>Tất cả nhóm chi<span /></ActionButton>
        {!allCategories ? (
          <div className="category-picker">
            {expenseCategories.map((category) => <ActionButton className={categoryIDs.includes(category.id) ? "mini-toggle active" : "mini-toggle"} type="button" key={category.id} onClick={() => toggleCategory(category.id)}>{category.name}</ActionButton>)}
          </div>
        ) : null}
        <div className="sheet-actions">
          {editing ? <ActionButton className="wide-pill destructive" type="button" disabled={saving} onClick={() => void archive()}>Lưu trữ</ActionButton> : null}
          <ActionButton className="primary-cta" type="button" disabled={!canSave || saving} onClick={() => void save()}>{saving ? "Đang lưu" : "Lưu"}</ActionButton>
        </div>
      </section>
    </div>
  );
}

function AssetRow({ asset, privacyMasked, online, onChanged, onArchived }: { asset: AssetPosition; privacyMasked: boolean; online: boolean; onChanged: (asset: AssetPosition) => void; onArchived: (assetID: string) => void }) {
  const [mode, setMode] = useState<"idle" | "buy" | "sell" | "price">("idle");
  const marketValue = asset.summary.market_value_vnd;
  const pnl = asset.summary.unrealized_pnl_vnd;
  return (
    <article className="asset-row">
      <ActionButton type="button" className="asset-main" onClick={() => setMode((current) => current === "idle" ? "price" : "idle")}>
        <span className="category-dot" />
        <div>
          <strong>{asset.name}</strong>
          <p>{assetLabel(asset)} · {asset.summary.quantity} {unitLabel(asset.unit)}</p>
        </div>
        <b>{privacyMasked ? "••••••" : marketValue == null ? "Chưa có giá" : formatVND(marketValue)}</b>
      </ActionButton>
      <div className="asset-meta">
        <span>{asset.pricing_mode === "automatic" ? "Tự động" : "Thủ công"}</span>
        <span>{asset.latest_price?.priced_at ? `Giá ${new Date(asset.latest_price.priced_at).toLocaleDateString("vi-VN")}` : "Chưa định giá"}</span>
        {pnl != null ? <strong className={pnl >= 0 ? "income" : "expense"}>{privacyMasked ? "••••••" : formatVND(pnl)}</strong> : null}
      </div>
      <div className="asset-actions">
        <ActionButton type="button" onClick={() => setMode("buy")}>Mua</ActionButton>
        <ActionButton type="button" onClick={() => setMode("sell")}>Bán</ActionButton>
        <ActionButton type="button" onClick={() => setMode("price")}>Giá</ActionButton>
        <ActionButton type="button" className="danger-text" onClick={() => void archiveAsset(asset.id, asset.version).then(() => onArchived(asset.id)).catch(() => undefined)}>Ẩn</ActionButton>
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
      {mode !== "price" ? <InputControl aria-label="Số lượng tài sản" inputMode="decimal" value={quantity} onChange={(event) => setQuantity(event.target.value.replace(/[^0-9.]/g, ""))} placeholder="Số lượng" /> : null}
      <InputControl aria-label="Giá VND" inputMode="numeric" value={price} onChange={(event) => setPrice(event.target.value.replace(/\D/g, ""))} placeholder="Giá VND" />
      {mode !== "price" ? <InputControl aria-label="Phí VND" inputMode="numeric" value={fee} onChange={(event) => setFee(event.target.value.replace(/\D/g, ""))} placeholder="Phí" /> : null}
      <ActionButton type="button" disabled={!canSave || busy} onClick={() => void save()}>{busy ? "Đang lưu" : "Lưu"}</ActionButton>
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
    <SheetFrame title="Thêm tài sản" label="Thêm tài sản" leading={<ActionButton type="button" onClick={onClose}>Đóng</ActionButton>} trailing={<Info size={22} />}>
      <section className="manager-section">
        <label className="sheet-row"><BriefcaseBusiness /><Select aria-label="Loại tài sản" value={type} onChange={(event) => setType(event.target.value as AssetType)}>{assetTypes.map((item) => <option key={item} value={item}>{assetTypeLabel(item)}</option>)}</Select></label>
        <label className="sheet-row"><List /><InputControl aria-label="Tên tài sản" value={name} onChange={(event) => setName(event.target.value)} placeholder="Tên tài sản" /></label>
        <label className="sheet-row"><Search /><InputControl aria-label="Mã tài sản" value={symbol} onChange={(event) => setSymbol(event.target.value)} placeholder="Mã, ví dụ FPT hoặc BTC" /></label>
        <label className="sheet-row"><Wallet /><Select aria-label="Đơn vị tài sản" value={unit} onChange={(event) => setUnit(event.target.value)}>{unitsForAsset(type).map((item) => <option key={item} value={item}>{unitLabel(item)}</option>)}</Select></label>
        <div className="segmented sheet-segmented">
          <ActionButton type="button" className={pricingMode === "manual" ? "active" : ""} onClick={() => setPricingMode("manual")}>Thủ công</ActionButton>
          <ActionButton type="button" className={pricingMode === "automatic" ? "active" : ""} onClick={() => setPricingMode("automatic")}>Tự động</ActionButton>
        </div>
        {pricingMode === "automatic" ? (
          <>
            <label className="sheet-row"><List /><InputControl aria-label="Provider key" value={providerKey} onChange={(event) => setProviderKey(event.target.value)} placeholder="Provider" /></label>
            <label className="sheet-row"><Search /><InputControl aria-label="Provider symbol" value={providerSymbol} onChange={(event) => setProviderSymbol(event.target.value)} placeholder="Symbol provider" /></label>
          </>
        ) : null}
        <ActionButton className="primary-cta" type="button" disabled={!canSave || busy} onClick={() => void save()}>{busy ? "Đang lưu" : "Lưu"}</ActionButton>
      </section>
    </SheetFrame>
  );
}

function AddTransactionSheet({ categories, wallets, budgets, readOnly, onCreated, onDebtCreated, onClose }: { categories: CategorySummary[]; wallets: WalletSummary[]; budgets: BudgetProgress[]; readOnly: boolean; onCreated: (transaction: Transaction) => void; onDebtCreated: () => void; onClose: () => void }) {
  const [type, setType] = useState<"expense" | "income" | "debt" | "transfer">("expense");
  const [amount, setAmount] = useState("");
  const [note, setNote] = useState("");
  const [sourceWalletID, setSourceWalletID] = useState(wallets[0]?.id ?? "");
  const [destinationWalletID, setDestinationWalletID] = useState(wallets[1]?.id ?? "");
  const [categoryID, setCategoryID] = useState("");
  const [budgetID, setBudgetID] = useState("");
  const [excludedFromReports, setExcludedFromReports] = useState(false);
  const [debtDirection, setDebtDirection] = useState<ObligationDirection>("borrowed");
  const [counterparty, setCounterparty] = useState("");
  const [dueOn, setDueOn] = useState(() => calendarDateInHoChiMinh());
  const [occurredOn, setOccurredOn] = useState(() => calendarDateInHoChiMinh());
  const [saving, setSaving] = useState(false);
  const [receiptFile, setReceiptFile] = useState<File | null>(null);
  const receiptInput = useRef<HTMLInputElement>(null);
  const [operationError, setOperationError] = useState<OperationFailure | null>(null);
  const [completed, setCompleted] = useState(false);
  const [showDetails, setShowDetails] = useState(false);
  const filteredCategories = categories.filter((category) => category.kind === (type === "income" ? "income" : "expense"));
  const activeBudgets = budgets.map((row) => row.budget);
  const chosenCategoryID = type === "income" || type === "expense" ? categoryID || filteredCategories[0]?.id : "";
  const receiptLocked = readOnly || saving || completed;
  const canSave = !completed && !readOnly && wallets.length > 0 && Number.isSafeInteger(Number(amount)) && Number(amount) > 0 && (type === "debt" ? !receiptFile && counterparty.trim() !== "" && dueOn !== "" : Boolean(sourceWalletID)) && (type !== "transfer" || Boolean(destinationWalletID && destinationWalletID !== sourceWalletID));
  useEffect(() => {
    if (!sourceWalletID && wallets[0]) setSourceWalletID(wallets[0].id);
  }, [sourceWalletID, wallets]);
  async function save() {
    if (!canSave || saving) return;
    setSaving(true);
    setOperationError(null);
    let transactionSaved = false;
    try {
      if (type === "debt") {
        await createObligation({ direction: debtDirection, principal_vnd: Number(amount), counterparty: counterparty.trim(), due_on: dueOn, note });
        onDebtCreated();
        onClose();
        return;
      }
      const receipt = receiptFile && navigator.onLine ? await uploadFile(receiptFile) : undefined;
      const transaction = await createTransaction(buildTransactionInput({ type, amount, sourceWalletID, destinationWalletID, budgetID, categoryID: chosenCategoryID, note, excludedFromReports, occurredOn, receiptObjectID: receipt?.id }));
      transactionSaved = true;
      setCompleted(true);
      onCreated(transaction);
      if (receiptFile && !navigator.onLine) {
        await queueReceiptUpload({ transaction_id: transaction.id, file: receiptFile, filename: receiptFile.name, content_type: receiptFile.type });
      }
      onClose();
    } catch (error) {
      setOperationError(operationFailure(error, transactionSaved
        ? "Giao dịch đã lưu, nhưng chưa lưu được ảnh vào hàng đợi. Không tạo lại giao dịch; hãy giữ ảnh gốc để bổ sung sau."
        : "Chưa xác nhận được việc lưu giao dịch. Nội dung vẫn được giữ; hãy kiểm tra sổ giao dịch trước khi gửi lại."));
    } finally {
      setSaving(false);
    }
  }
  return (
    <SheetFrame title="Thêm Giao Dịch" leading={<PillButton onClick={onClose}>Hủy</PillButton>} trailing={<span />}
      footer={<div className="save-bar"><ActionButton disabled={!canSave || saving} onClick={() => void save()}>{saving ? "Đang lưu" : "Lưu"}</ActionButton><ActionButton className="receipt" aria-label="Đính kèm ảnh" disabled={receiptLocked || type === "debt"} onClick={() => receiptInput.current?.click()}><ImagePlus size={24} /></ActionButton></div>}
    >
        {readOnly ? <p className="offline-warning">Offline storage chưa sẵn sàng. Mở mạng lại để lưu giao dịch.</p> : null}
        <OperationError failure={operationError} />
        {completed ? <PillButton onClick={onClose}>Đóng</PillButton> : null}
        <div className="transaction-form-card">
          <div className="segmented sheet-segmented">
            {(["expense", "income", "debt", "transfer"] as const).map((option) => (
              <ActionButton className={type === option ? "active" : ""} disabled={receiptLocked} key={option} onClick={() => setType(option)}>{quickAddTypeLabel(option)}</ActionButton>
            ))}
          </div>
          <label className="amount-row"><span>VND</span><InputControl aria-label="Số tiền" type="text" inputMode="numeric" value={amount} onChange={(event) => setAmount(event.target.value.replace(/\D/g, ""))} placeholder="0" /></label>
          {type === "debt" ? <div className="segmented debt-segmented"><ActionButton type="button" className={debtDirection === "borrowed" ? "active" : ""} onClick={() => setDebtDirection("borrowed")}>Tôi vay</ActionButton><ActionButton type="button" className={debtDirection === "lent" ? "active" : ""} onClick={() => setDebtDirection("lent")}>Tôi cho vay</ActionButton></div> : null}
          {type !== "debt" ? <label className="sheet-row"><Wallet /><Select aria-label="Ví nguồn" value={sourceWalletID} disabled={receiptLocked} onChange={(event) => setSourceWalletID(event.target.value)}>{wallets.map((wallet) => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</Select></label> : <label className="sheet-row"><Users /><InputControl aria-label="Đối tác" value={counterparty} onChange={(event) => setCounterparty(event.target.value)} placeholder="Người liên quan" /></label>}
          {type === "transfer" ? <>
            <label className="sheet-row"><Wallet /><Select aria-label="Ví đích" value={destinationWalletID} disabled={receiptLocked} onChange={(event) => setDestinationWalletID(event.target.value)}><option value="">Chọn ví nhận</option>{wallets.map(wallet => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</Select></label>
            <p className="px-4 py-2 text-sm">Chuyển giữa hai ví khác nhau, không tính vào thu/chi báo cáo.</p>
          </> : type !== "debt" ? <label className="sheet-row"><span className="dot-icon" /><Select aria-label="Nhóm" value={chosenCategoryID} onChange={(event) => setCategoryID(event.target.value)}><option value="">Chọn nhóm</option>{filteredCategories.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}</Select></label> : <label className="sheet-row"><CalendarDays /><InputControl aria-label="Ngày đến hạn" type="date" value={dueOn} onChange={(event) => setDueOn(event.target.value)} /></label>}
          {type === "expense" && activeBudgets.length > 0 ? <label className="sheet-row"><span className="dot-icon" /><Select aria-label="Ngân sách giao dịch" value={budgetID} disabled={receiptLocked} onChange={(event) => setBudgetID(event.target.value)}><option value="">Không gắn ngân sách</option>{activeBudgets.map((budget) => <option key={budget.id} value={budget.id}>{budget.name}</option>)}</Select></label> : null}
          {type !== "debt" ? <label className="sheet-row"><List /><InputControl aria-label="Ghi chú" value={note} onChange={(event) => setNote(event.target.value)} placeholder="Ghi chú" /></label> : null}
          {type !== "debt" ? <label className="date-row"><CalendarDays /><InputControl aria-label="Ngày giao dịch" type="date" value={occurredOn} onChange={(event) => setOccurredOn(event.target.value)} /></label> : null}
          {type !== "debt" ? <label className="exclude-row"><InputControl type="checkbox" checked={excludedFromReports} onChange={(event) => setExcludedFromReports(event.target.checked)} /><span>Không tính vào báo cáo</span></label> : null}
          <ActionButton className="details-trigger" onClick={() => setShowDetails((current) => !current)} aria-expanded={showDetails}>{showDetails ? "Ẩn chi tiết" : "Thêm chi tiết"}</ActionButton>
          {showDetails ? <div className="details-panel">
            <label className="sheet-row"><List /><InputControl aria-label="Ghi chú chi tiết" value={note} onChange={(event) => setNote(event.target.value)} placeholder="Ghi chú" /></label>
            <UnavailableAction icon={<Users />} label="Với" reason="Chưa hỗ trợ trong form này" />
            <UnavailableAction icon={<MapPin />} label="Đặt vị trí" reason="Chưa hỗ trợ trong form này" />
            <UnavailableAction icon={<BriefcaseBusiness />} label="Chọn sự kiện" reason="Chưa hỗ trợ trong form này" />
            <UnavailableAction icon={<Bell />} label="Đặt nhắc nhở" reason="Chưa hỗ trợ trong form này" />
            <ActionButton className="image-row" disabled={receiptLocked || type === "debt"} onClick={() => receiptInput.current?.click()}><ImagePlus size={22} /><span>{receiptFile ? "Đổi ảnh" : "Thêm Hình Ảnh"}</span></ActionButton>
          </div> : null}
          <FilePickerInput id="receipt-image" ref={receiptInput} aria-label="Ảnh đính kèm" accept="image/jpeg,image/png,image/webp" hidden disabled={receiptLocked || type === "debt"} onFileSelected={setReceiptFile} />
          {receiptFile ? <div className="px-4 py-2">
            <p role="status" className="break-all text-sm">Ảnh đã chọn: {receiptFile.name}</p>
            <PillButton disabled={receiptLocked} onClick={() => setReceiptFile(null)}>Bỏ ảnh</PillButton>
          </div> : null}
          {type === "debt" ? <p className="px-4 py-2 text-sm">Ảnh chỉ hỗ trợ cho khoản thu/chi. {receiptFile ? "Bỏ ảnh hoặc quay lại khoản thu/chi trước khi lưu." : "Vay/nợ chưa hỗ trợ ảnh đính kèm."}</p> : null}
        </div>
    </SheetFrame>
  );
}

function EditTransactionSheet({ categories, wallets, budgets, readOnly, transaction, onChanged, onArchived, onClose }: { categories: CategorySummary[]; wallets: WalletSummary[]; budgets: BudgetProgress[]; readOnly: boolean; transaction: Transaction; onChanged: (transaction: Transaction) => void; onArchived: () => void; onClose: () => void }) {
  const [amount, setAmount] = useState(String(transaction.amount_vnd));
  const [note, setNote] = useState(transaction.note);
  const [budgetID, setBudgetID] = useState(transaction.budget_id ?? "");
  const [excludedFromReports, setExcludedFromReports] = useState(transaction.excluded_from_reports);
  const [saving, setSaving] = useState(false);
  const [operationError, setOperationError] = useState<OperationFailure | null>(null);
  const category = categories.find((item) => item.id === transaction.category_id);
  const sourceWallet = wallets.find((item) => item.id === transaction.source_wallet_id);
  const activeBudgets = budgets.map((row) => row.budget);
  async function save() {
    if (readOnly || Number(amount) <= 0 || saving) return;
    setSaving(true);
    setOperationError(null);
    try {
      const next = await updateTransaction(transaction.id, {
        type: transaction.type,
        source_wallet_id: transaction.source_wallet_id,
        destination_wallet_id: transaction.destination_wallet_id,
        category_id: transaction.category_id,
        budget_id: transaction.type === "expense" && budgetID ? budgetID : undefined,
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
    } catch (error) {
      setOperationError(operationFailure(error, "Chưa xác nhận được việc sửa giao dịch. Nội dung vẫn được giữ; hãy kiểm tra lại dữ liệu trước khi gửi lại."));
    } finally {
      setSaving(false);
    }
  }
  async function archive() {
    if (readOnly || saving) return;
    setSaving(true);
    setOperationError(null);
    try {
      await archiveTransaction(transaction.id, transaction.version);
      onArchived();
    } catch (error) {
      setOperationError(operationFailure(error, "Chưa xác nhận được việc lưu trữ giao dịch. Hãy kiểm tra lại sổ giao dịch trước khi thử lại."));
    } finally {
      setSaving(false);
    }
  }
  return (
    <div className="sheet-backdrop">
      <section className="transaction-sheet" role="dialog" aria-modal="true" aria-label="Sửa Giao Dịch">
        <header>
          <ActionButton className="pill-button" type="button" onClick={onClose}>Hủy</ActionButton>
          <h2>Sửa Giao Dịch</h2>
          <span />
        </header>
        {readOnly ? <p className="offline-warning">Offline storage chưa sẵn sàng. Mở mạng lại để sửa giao dịch.</p> : null}
        <OperationError failure={operationError} />
        <p className="sheet-meta">{transactionTypeLabel(transaction.type)} · {sourceWallet?.name ?? "Ví"} · {category?.name ?? "Không nhóm"}</p>
        <label className="amount-row"><span>VND</span><InputControl aria-label="Số tiền" inputMode="numeric" value={amount} onChange={(event) => setAmount(event.target.value.replace(/\D/g, ""))} placeholder="0" disabled={readOnly} /></label>
        {transaction.type === "expense" && activeBudgets.length > 0 ? <label className="sheet-row"><span className="dot-icon" /><Select aria-label="Ngân sách giao dịch" value={budgetID} disabled={readOnly} onChange={(event) => setBudgetID(event.target.value)}><option value="">Không gắn ngân sách</option>{activeBudgets.map((budget) => <option key={budget.id} value={budget.id}>{budget.name}</option>)}</Select></label> : null}
        <label className="sheet-row"><List /><InputControl aria-label="Ghi chú" value={note} onChange={(event) => setNote(event.target.value)} placeholder="Ghi chú" disabled={readOnly} /></label>
        <ActionButton className={excludedFromReports ? "toggle-row active" : "toggle-row"} type="button" disabled={readOnly} onClick={() => setExcludedFromReports((current) => !current)}>Không tính vào báo cáo<span /></ActionButton>
        <div className="sheet-actions">
          <ActionButton className="wide-pill destructive" disabled={readOnly || saving} onClick={() => void archive()}>Lưu trữ</ActionButton>
          <ActionButton className="primary-cta" disabled={readOnly || saving || Number(amount) <= 0} onClick={() => void save()}>{saving ? "Đang lưu" : "Lưu thay đổi"}</ActionButton>
        </div>
      </section>
    </div>
  );
}

function WalletManagerSheet({
  wallets,
  categories,
  online,
  readOnly,
  onWalletChanged,
  onWalletArchived,
  onCategoryChanged,
  onChanged,
  onClose,
}: {
  wallets: WalletSummary[];
  categories: CategorySummary[];
  online: boolean;
  readOnly: boolean;
  onWalletChanged: (wallet: WalletSummary) => void;
  onWalletArchived: (walletID: string) => void;
  onCategoryChanged: (category: CategorySummary) => void;
  onChanged: () => void | Promise<void>;
  onClose: () => void;
}) {
  const [walletName, setWalletName] = useState("");
  const [walletType, setWalletType] = useState<WalletType>("basic");
  const [goalTarget, setGoalTarget] = useState("");
  const [goalDeadline, setGoalDeadline] = useState("");
  const [creatingWallet, setCreatingWallet] = useState(false);
  const [categoryName, setCategoryName] = useState("");
  const [categoryKind, setCategoryKind] = useState<"income" | "expense">("expense");
  const [categoryParentID, setCategoryParentID] = useState("");
  const [managerError, setManagerError] = useState<OperationFailure | null>(null);
  const [editingWallets, setEditingWallets] = useState(false);
  const [busy, setBusy] = useState(false);
  const includedWallets = wallets.filter((wallet) => wallet.include_in_total);
  const excludedWallets = wallets.filter((wallet) => !wallet.include_in_total);
  const totalVND = totalIncludedVND(wallets);
  async function run(action: () => Promise<unknown>, failureMessage = "Chưa lưu được thay đổi. Dữ liệu đã xác nhận vẫn được giữ lại.", afterSuccess?: () => void) {
    if (busy) return false;
    setBusy(true);
    setManagerError(null);
    try {
      await action();
      await onChanged();
      afterSuccess?.();
      return true;
    } catch (error) {
      setManagerError(operationFailure(error, failureMessage));
      return false;
    } finally {
      setBusy(false);
    }
  }
  const createParentOptions = categories.filter((category) => category.kind === categoryKind);
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
          <ActionButton className="wallet-action-row" type="button" disabled={readOnly} onClick={() => setCreatingWallet((current) => !current)}>
            <span className="wallet-action-icon"><Plus size={28} /></span>
            <strong>Thêm ví</strong>
          </ActionButton>
          <ActionButton className="wallet-action-row" type="button" disabled>
            <span className="wallet-action-icon"><Link size={24} /></span>
            <strong>Liên kết dịch vụ</strong>
          </ActionButton>
        </section>

        {creatingWallet ? (
          <section className="manager-section wallet-create-panel">
            <h3>Thêm ví</h3>
            <div className="manager-form">
              <InputControl aria-label="Tên ví mới" value={walletName} onChange={(event) => setWalletName(event.target.value)} placeholder="Tên ví mới" disabled={readOnly} />
              <Select aria-label="Loại ví" value={walletType} onChange={(event) => setWalletType(event.target.value as WalletType)} disabled={readOnly}>{walletTypes.map((type) => <option key={type} value={type}>{walletTypeLabel(type)}</option>)}</Select>
              {walletType === "goal" ? (
                <>
                  <InputControl aria-label="Mục tiêu số tiền" inputMode="numeric" value={goalTarget} onChange={(event) => setGoalTarget(event.target.value)} placeholder="Mục tiêu VND" disabled={readOnly} />
                  <InputControl aria-label="Ngày hoàn thành mục tiêu" type="date" value={goalDeadline} onChange={(event) => setGoalDeadline(event.target.value)} disabled={readOnly} />
                </>
              ) : null}
              <ActionButton type="button" disabled={readOnly || !walletName.trim() || busy} onClick={() => void run(async () => {
                const wallet = await createWallet({
                  name: walletName,
                  type: walletType,
                  ...(walletType === "goal" && Number(goalTarget) > 0 ? { goal_target_vnd: Number(goalTarget) } : {}),
                  ...(walletType === "goal" && goalDeadline ? { goal_deadline_on: goalDeadline } : {}),
                });
                if (wallet) onWalletChanged(wallet);
                setWalletName("");
                setGoalTarget("");
                setGoalDeadline("");
                setCreatingWallet(false);
              })}>Tạo ví</ActionButton>
            </div>
          </section>
        ) : null}

        <section className="manager-section" aria-label="Tạo nhóm giao dịch">
          <h3>Nhóm giao dịch</h3>
          <div className="manager-form">
            <InputControl aria-label="Tên nhóm mới" value={categoryName} onChange={(event) => setCategoryName(event.target.value)} disabled={readOnly || busy} />
            <Select aria-label="Loại nhóm" value={categoryKind} onChange={(event) => { setCategoryKind(event.target.value as "income" | "expense"); setCategoryParentID(""); }} disabled={readOnly || busy}>
              <option value="expense">Chi</option><option value="income">Thu</option>
            </Select>
            <Select aria-label="Nhóm cha mới" value={categoryParentID} onChange={(event) => setCategoryParentID(event.target.value)} disabled={readOnly || busy}>
              <option value="">Không có nhóm cha</option>
              {createParentOptions.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}
            </Select>
            <ActionButton type="button" disabled={readOnly || busy || !categoryName.trim()} onClick={() => void run(async () => {
              const category = await createCategory({ name: categoryName, kind: categoryKind, ...(categoryParentID ? { parent_id: categoryParentID } : {}) });
              onCategoryChanged(category);
            }, "Chưa lưu được thay đổi nhóm. Nội dung và nhóm cha vẫn được giữ để bạn kiểm tra.", () => { setCategoryName(""); setCategoryParentID(""); })}>Tạo nhóm</ActionButton>
          </div>
          {categories.filter((category) => !category.is_system).map((category) => (
            <CategoryEditor key={category.id} category={category} categories={categories} readOnly={readOnly} busy={busy} onCategoryChanged={onCategoryChanged} onSave={run} />
          ))}
          {categories.filter((category) => category.is_system).map((category) => <p key={category.id}>{category.name} · Nhóm hệ thống</p>)}
        </section>
        <OperationError failure={managerError} />

        <WalletCategorySettingsPanel wallets={wallets} online={online} readOnly={readOnly} />

    </SheetFrame>
  );
}

function CategoryEditor({ category, categories, readOnly, busy, onCategoryChanged, onSave }: {
  category: CategorySummary;
  categories: CategorySummary[];
  readOnly: boolean;
  busy: boolean;
  onCategoryChanged: (category: CategorySummary) => void;
  onSave: (action: () => Promise<unknown>, failureMessage?: string) => Promise<boolean>;
}) {
  const [name, setName] = useState(category.name);
  const [parentID, setParentID] = useState(category.parent_id ?? "");
  const blockedParentIDs = categoryDescendantIDs(category.id, categories);
  blockedParentIDs.add(category.id);
  const parentOptions = categories.filter((candidate) => candidate.kind === category.kind && !blockedParentIDs.has(candidate.id));
  return (
    <div className="manager-row">
      <InputControl aria-label={`Tên nhóm ${category.name}`} value={name} onChange={(event) => setName(event.target.value)} disabled={readOnly || busy} />
      <Select aria-label={`Nhóm cha ${category.name}`} value={parentID} onChange={(event) => setParentID(event.target.value)} disabled={readOnly || busy}>
        <option value="">Không có nhóm cha</option>
        {parentOptions.map((candidate) => <option key={candidate.id} value={candidate.id}>{candidate.name}</option>)}
      </Select>
      <ActionButton type="button" aria-label={`Lưu nhóm ${category.name}`} disabled={readOnly || busy || !name.trim()} onClick={() => void onSave(async () => {
        const updated = await updateCategory(category.id, { name, parent_id: parentID || null, base_version: category.version, current_category: category });
        onCategoryChanged(updated);
      }, "Chưa lưu được thay đổi nhóm. Nội dung và nhóm cha vẫn được giữ để bạn kiểm tra.")}>Lưu</ActionButton>
    </div>
  );
}

function WalletCategorySettingsPanel({ wallets, online, readOnly }: { wallets: WalletSummary[]; online: boolean; readOnly: boolean }) {
  const [walletID, setWalletID] = useState(wallets[0]?.id ?? "");
  const [settings, setSettings] = useState<WalletCategorySetting[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [busyCategoryID, setBusyCategoryID] = useState<string | null>(null);
  const [failure, setFailure] = useState<OperationFailure | null>(null);
  const generation = useRef(0);

  async function reload() {
    if (!walletID || !online) return;
    const currentGeneration = ++generation.current;
    setLoading(true);
    setFailure(null);
    try {
      const next = await loadWalletCategorySettings(walletID);
      if (currentGeneration === generation.current) setSettings(next);
    } catch (error) {
      if (currentGeneration === generation.current) setFailure(operationFailure(error, "Chưa tải được cài đặt nhóm theo ví. Trạng thái đã xác nhận vẫn được giữ lại."));
    } finally {
      if (currentGeneration === generation.current) setLoading(false);
    }
  }

  useEffect(() => {
    if (!online || !walletID) {
      ++generation.current;
      setLoading(false);
      return;
    }
    void reload();
    return () => { ++generation.current; };
  }, [walletID, online]);

  async function toggle(setting: WalletCategorySetting) {
    if (!online || readOnly || loading || busyCategoryID) return;
    setBusyCategoryID(setting.id);
    setFailure(null);
    try {
      await setWalletCategoryActive(walletID, setting.id, !setting.active);
      await reload();
    } catch (error) {
      setFailure(operationFailure(error, "Chưa cập nhật được cài đặt nhóm. Trạng thái đã xác nhận vẫn được giữ; hãy tải lại."));
    } finally {
      setBusyCategoryID(null);
    }
  }

  return (
    <section className="manager-section" aria-label="Cài đặt nhóm theo ví">
      <h3>Nhóm theo ví</h3>
      {wallets.length > 0 ? (
        <Select aria-label="Ví cài đặt nhóm" value={walletID} onChange={(event) => { setWalletID(event.target.value); setSettings(null); setFailure(null); }} disabled={loading || busyCategoryID != null}>
          {wallets.map((wallet) => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}
        </Select>
      ) : <p className="empty-state">Tạo ví trước khi cài đặt nhóm</p>}
      {!online ? <p className="offline-warning neutral">Cần online để tải và thay đổi cài đặt nhóm theo ví.</p> : null}
      {loading ? <p role="status">Đang tải cài đặt nhóm…</p> : null}
      <OperationError failure={failure} onRetry={() => { void reload(); }} retryLabel="Tải lại cài đặt nhóm" busy={loading || busyCategoryID != null} />
      {settings?.map((setting) => (
        <div className="manager-row" key={setting.id}>
          <span>{setting.name}</span>
          <ActionButton
            type="button"
            className={setting.active ? "mini-toggle active" : "mini-toggle"}
            aria-label={`${setting.name} đang ${setting.active ? "bật" : "tắt"}`}
            aria-pressed={setting.active}
            disabled={!online || readOnly || loading || busyCategoryID != null}
            onClick={() => void toggle(setting)}
          >{busyCategoryID === setting.id ? "Đang lưu" : setting.active ? "Bật" : "Tắt"}</ActionButton>
        </div>
      ))}
    </section>
  );
}

function categoryDescendantIDs(categoryID: string, categories: CategorySummary[]) {
  const descendants = new Set<string>();
  const pending = [categoryID];
  while (pending.length > 0) {
    const parentID = pending.shift();
    for (const category of categories) {
      if (category.parent_id === parentID && !descendants.has(category.id)) {
        descendants.add(category.id);
        pending.push(category.id);
      }
    }
  }
  return descendants;
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
  onChanged: (action: () => Promise<unknown>) => Promise<unknown>;
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
      <InputControl aria-label={`Tên ví ${wallet.name}`} value={name} onChange={(event) => setName(event.target.value)} disabled={readOnly} />
      <ActionButton type="button" disabled={readOnly} className={include ? "mini-toggle active" : "mini-toggle"} onClick={() => setInclude((current) => !current)}>{include ? "Tổng" : "Ẩn"}</ActionButton>
      <ActionButton type="button" disabled={readOnly || busy || !name.trim()} onClick={() => void onChanged(async () => { const next = await updateWallet(wallet.id, { name, include_in_total: include, base_version: wallet.version, current_wallet: wallet }); if (next) onWalletChanged(next); })}>Lưu</ActionButton>
      <ActionButton type="button" disabled={readOnly || busy} onClick={() => void onChanged(async () => { await setDefaultAIWallet(wallet.id, wallets, wallet.version); wallets.forEach((item) => onWalletChanged({ ...item, is_default_ai: item.id === wallet.id })); })}>{wallet.is_default_ai ? "AI" : "Đặt AI"}</ActionButton>
      <ActionButton type="button" className="danger-text" disabled={readOnly || busy} onClick={() => void onChanged(async () => { await archiveWallet(wallet.id, wallet.version); onWalletArchived(wallet.id); })}>Ẩn</ActionButton>
    </div>
  );
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
    case "goal":
      return "◇";
    case "basic":
    default:
      return "₫";
  }
}

const walletTypes: WalletType[] = ["basic", "goal", "credit"];
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

function quickAddTypeLabel(type: "expense" | "income" | "debt" | "transfer") {
  switch (type) {
    case "transfer":
      return "Chuyển ví";
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
    case "credit":
      return "Tín dụng";
    case "goal":
      return "Mục tiêu";
    case "basic":
    default:
      return "Cơ bản";
  }
}

export default App;
