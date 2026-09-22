import { useEffect, useMemo, useState } from "react";
import { ImagePlus, Trash2 } from "lucide-react";

import { BaseButton } from "../atoms/BaseButton";
import { TransactionFields, type TransactionFormState } from "../molecules/TransactionFields";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { SegmentedControl } from "../atoms/SegmentedControl";
import { StatusMessage } from "../atoms/StatusMessage";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { fetchCategories, type Category } from "../../services/categories";
import { fetchWallets, type Wallet } from "../../services/wallets";
import { instantFromLocalDateTime, localDateTimeAt } from "../../services/accountTime";
import { useAccountTimezone } from "../../services/AccountTimezoneContext";
import { fetchJarMonth, type JarMonthSummary } from "../../services/jars";
import { isMonthKey, jarAssignmentNeedsSelection } from "../../services/monthJarLogic";
import {
  categoryAppliesToTransaction,
  createTransaction,
  deleteTransaction,
  updateTransaction,
  type Transaction,
  type TransactionInput,
  type TransactionType,
} from "../../services/transactions";

type EditorState = TransactionFormState;

const COPY = {
  addTitle: "Thêm giao dịch",
  editTitle: "Sửa giao dịch",
  close: "Đóng trình sửa giao dịch",
  expense: "Khoản chi",
  income: "Khoản thu",
  amount: "Số tiền (VND)",
  wallet: "Ví",
  category: "Nhóm",
  categoryOptional: "Không chọn nhóm",
  occurredAt: "Ngày và giờ",
  note: "Ghi chú",
  reports: "Tính vào báo cáo",
  save: "Lưu giao dịch",
  delete: "Xóa giao dịch",
  noWallet: "Bạn cần tạo ít nhất một ví trước khi thêm giao dịch.",
  loading: "Đang tải dữ liệu giao dịch...",
  loadError: "Không thể tải Ví và Nhóm. Vui lòng thử lại.",
  validationError: "Hãy chọn ví và nhập số tiền nguyên lớn hơn 0.",
  saveError: "Không thể lưu giao dịch. Dữ liệu bạn nhập vẫn được giữ lại.",
  deleteError: "Không thể xóa giao dịch. Vui lòng thử lại.",
  confirmDelete: "Xóa giao dịch này? Số dư ví và báo cáo sẽ được tính lại.",
} as const;

function localDateTimeValue(value:Date,timezone:string):string {
  return localDateTimeAt(value,timezone);
}

function initialState(transaction:Transaction|undefined,timezone:string): EditorState {
  if (!transaction) {
    return { type: "expense", amount: "", walletID: "", categoryID: "", jarID: "", occurredAt: localDateTimeValue(new Date(),timezone), note: "", includedInReports: true };
  }
  return {
    type: transaction.type,
    amount: String(transaction.amount),
    walletID: transaction.wallet_id,
    categoryID: transaction.category_id ?? "",
    jarID: transaction.jar_id ?? "",
    occurredAt: localDateTimeValue(new Date(transaction.occurred_at),timezone),
    note: transaction.note ?? "",
    includedInReports: transaction.included_in_reports,
  };
}

export function QuickAddSheet({ onClose, onSaved, onAiEntry, transaction, initialWalletID }: { onClose: () => void; onSaved: () => void; onAiEntry?: () => void; transaction?: Transaction; initialWalletID?: string }) {
  const timezone=useAccountTimezone();
  const [state, setState] = useState(() => ({...initialState(transaction,timezone), walletID:transaction?.wallet_id ?? initialWalletID ?? ""}));
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [jarSummary, setJarSummary] = useState<JarMonthSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const transactionMonth = state.occurredAt.slice(0, 7);
  useEffect(() => {
    if (state.type !== "expense" || !isMonthKey(transactionMonth)) {
      setJarSummary(null);
      return;
    }
    let cancelled = false;
    fetchJarMonth(transactionMonth).then(value => { if (!cancelled) setJarSummary(value); }).catch(() => { if (!cancelled) setJarSummary(null); });
    return () => { cancelled = true; };
  }, [state.type, transactionMonth]);


  useEffect(() => {
    let cancelled = false;
    Promise.all([fetchWallets(), fetchCategories()])
      .then(([nextWallets, nextCategories]) => {
        if (cancelled) return;
        const ledgerWallets = nextWallets.filter((wallet) => wallet.type !== "credit");
        setWallets(ledgerWallets);
        setCategories(nextCategories);
        setState((current) => ({ ...current, walletID: current.walletID || ledgerWallets[0]?.id || "" }));
      })
      .catch(() => { if (!cancelled) setError(COPY.loadError); })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, []);

  const applicableCategories = useMemo(
    () => categories.filter((category) => categoryAppliesToTransaction(category, state.type, state.walletID, wallets.find(wallet => wallet.id === state.walletID)?.type)),
    [categories, state.type, state.walletID, wallets],
  );
  const jarOptions = useMemo(() => {
    const configured = (jarSummary?.items ?? []).filter(item => item.active || item.jar_id === state.jarID).map(item => ({ jar_id: item.jar_id, name: item.name, active: item.active }));
    const stableJarName = jarSummary?.jars.find(item => item.jar_id === state.jarID)?.name;
    if (state.jarID && !configured.some(item => item.jar_id === state.jarID)) configured.push({ jar_id: state.jarID, name: transaction?.jar_name ?? stableJarName ?? "Hũ không có cấu hình tháng", active: false });
    return configured;
  }, [jarSummary, state.jarID, transaction?.jar_name]);

  const changeScope = (updates: Partial<Pick<EditorState, "type" | "walletID">>) => {
    setState((current) => {
      const next = { ...current, ...updates };
      const selected = categories.find((category) => category.id === next.categoryID);
      if (selected && !categoryAppliesToTransaction(selected, next.type, next.walletID, wallets.find(wallet => wallet.id === next.walletID)?.type)) next.categoryID = "";
      return next;
    });
  };

  const submit = async () => {
    if (saving || loading) return;
    const amount = Number(state.amount);
    if (!wallets.some(wallet => wallet.id === state.walletID) || !Number.isSafeInteger(amount) || amount <= 0 || !Number.isFinite(Date.parse(state.occurredAt))) {
      setError(COPY.validationError);
      return;
    }
    if (state.categoryID && !applicableCategories.some(category => category.id === state.categoryID)) {
      setError("Nhóm cũ không phù hợp với ví này. Hãy chọn lại hoặc bỏ chọn nhóm.");
      return;
    }
    let occurredAt:string;
    try { occurredAt=instantFromLocalDateTime(state.occurredAt,timezone); }
    catch { setError("Giờ này không tồn tại trong múi giờ đã chọn. Hãy chọn thời điểm khác."); return; }
    if (jarAssignmentNeedsSelection(state.jarID, jarOptions, transaction?.occurred_at, occurredAt)) {
      setError("Hũ đã gỡ khỏi tháng này. Hãy bỏ chọn hoặc chọn một hũ đang hoạt động cho tháng giao dịch.");
      return;
    }
    const input: TransactionInput = {
      wallet_id: state.walletID,
      category_id: state.categoryID || undefined,
      jar_id: state.jarID || null,
      type: state.type,
      amount,
      occurred_at: occurredAt,
      note: state.note.trim() || undefined,
      included_in_reports: state.includedInReports,
    };
    try {
      setSaving(true);
      setError("");
      if (transaction) await updateTransaction(transaction.id, input);
      else await createTransaction(input);
      onSaved();
      onClose();
    } catch {
      setError(COPY.saveError);
    } finally {
      setSaving(false);
    }
  };

  const remove = async () => {
    if (saving || !transaction || !window.confirm(COPY.confirmDelete)) return;
    try {
      setSaving(true);
      setError("");
      await deleteTransaction(transaction.id);
      onSaved();
      onClose();
    } catch {
      setError(COPY.deleteError);
    } finally {
      setSaving(false);
    }
  };

  let validTime=true;
  try { instantFromLocalDateTime(state.occurredAt,timezone); } catch { validTime=false; }
  const valid = wallets.some(wallet => wallet.id === state.walletID) && Number.isSafeInteger(Number(state.amount)) && Number(state.amount) > 0 && validTime;
  return <BaseBottomSheet presentation="form" closingDisabled={saving} title={transaction ? COPY.editTitle : COPY.addTitle} closeLabel="Hủy" onClose={() => { if (!saving) onClose(); }}
    footer={<div className="flex gap-3"><BaseButton className="flex-1" size="lg" loading={saving} disabled={loading || !valid} onClick={() => void submit()}>Lưu</BaseButton>{transaction ? <BaseButton variant="danger" disabled={saving} aria-label={COPY.delete} onClick={() => void remove()}><Trash2 size={20} /></BaseButton> : onAiEntry ? <BaseButton disabled={saving} aria-label="Nhập bằng AI từ ảnh hoặc nội dung" onClick={onAiEntry}><ImagePlus size={22} /></BaseButton> : null}</div>}>
    <div className="space-y-3">
      {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
      {loading ? <StatusMessage>{COPY.loading}</StatusMessage> : null}
      {!loading && wallets.length === 0 ? <StatusMessage variant="plain">{COPY.noWallet}</StatusMessage> : null}
      <TransactionFields state={state} onChange={setState} wallets={wallets} categories={categories} jars={jarOptions} disabled={saving || loading} />
      {!transaction && onAiEntry ? <BaseButton variant="ghost" className="w-full" disabled={saving} onClick={onAiEntry}>Nhập bằng AI</BaseButton> : null}
    </div>
  </BaseBottomSheet>;
}
