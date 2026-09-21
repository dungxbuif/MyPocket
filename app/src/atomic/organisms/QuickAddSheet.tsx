import { useEffect, useMemo, useState } from "react";
import { Trash2 } from "lucide-react";

import { BaseButton } from "../atoms/BaseButton";
import { BaseCheckbox } from "../atoms/BaseCheckbox";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { SegmentedControl } from "../atoms/SegmentedControl";
import { StatusMessage } from "../atoms/StatusMessage";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
import { CategorySelectionList } from "../molecules/CategorySelectionList";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { fetchCategories, type Category } from "../../services/categories";
import { fetchWallets, type Wallet } from "../../services/wallets";
import {
  categoryAppliesToTransaction,
  createTransaction,
  deleteTransaction,
  updateTransaction,
  type Transaction,
  type TransactionInput,
  type TransactionType,
} from "../../services/transactions";

type EditorState = {
  type: TransactionType;
  amount: string;
  walletID: string;
  categoryID: string;
  occurredAt: string;
  note: string;
  includedInReports: boolean;
};

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

function localDateTimeValue(value = new Date()): string {
  const local = new Date(value.getTime() - value.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(0, 16);
}

function initialState(transaction?: Transaction): EditorState {
  if (!transaction) {
    return { type: "expense", amount: "", walletID: "", categoryID: "", occurredAt: localDateTimeValue(), note: "", includedInReports: true };
  }
  return {
    type: transaction.type,
    amount: String(transaction.amount),
    walletID: transaction.wallet_id,
    categoryID: transaction.category_id ?? "",
    occurredAt: localDateTimeValue(new Date(transaction.occurred_at)),
    note: transaction.note ?? "",
    includedInReports: transaction.included_in_reports,
  };
}

export function QuickAddSheet({ onClose, onSaved, onAiEntry, transaction, initialWalletID }: { onClose: () => void; onSaved: () => void; onAiEntry?: () => void; transaction?: Transaction; initialWalletID?: string }) {
  const [state, setState] = useState(() => ({...initialState(transaction), walletID:transaction?.wallet_id ?? initialWalletID ?? ""}));
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [choosingCategory, setChoosingCategory] = useState(false);

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
    const input: TransactionInput = {
      wallet_id: state.walletID,
      category_id: state.categoryID || undefined,
      type: state.type,
      amount,
      occurred_at: new Date(state.occurredAt).toISOString(),
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

  return (
    <BaseBottomSheet presentation="form" closingDisabled={saving} title={choosingCategory ? "Chọn nhóm" : transaction ? COPY.editTitle : COPY.addTitle} closeLabel={choosingCategory ? "Quay lại" : "Hủy"} onClose={() => { if (saving) return; if (choosingCategory) setChoosingCategory(false); else onClose(); }}>
      {choosingCategory ? <div className="space-y-4"><SegmentedControl value={state.type} onChange={(type) => changeScope({type})} options={[{value:"expense",label:COPY.expense},{value:"income",label:COPY.income}]} /><CategorySelectionList categories={applicableCategories} onSelect={(categoryID) => {setState(current => ({...current,categoryID}));setChoosingCategory(false);}} /></div> :
      <div className="space-y-3">
        {!transaction && onAiEntry ? <BaseButton variant="secondary" disabled={saving} onClick={onAiEntry}>Nhập bằng AI</BaseButton> : null}
        {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
        {loading ? <StatusMessage>{COPY.loading}</StatusMessage> : null}
        {!loading && wallets.length === 0 ? <StatusMessage>{COPY.noWallet}</StatusMessage> : null}
        <SegmentedControl value={state.type} onChange={(type) => changeScope({ type })} options={[{ value: "expense", label: COPY.expense }, { value: "income", label: COPY.income }]} />
        <SurfaceCard padding="md" className="space-y-3">
        <FormField label={COPY.wallet}><BaseSelect required disabled={saving || loading || wallets.length === 0} value={state.walletID} onChange={(event) => changeScope({ walletID: event.target.value })}><option value="">Chọn ví</option>{wallets.map((wallet) => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</BaseSelect></FormField>
        <FormField label={COPY.amount}><BaseTextInput variant="title" autoFocus inputMode="numeric" required disabled={saving || loading} value={state.amount} onChange={(event) => setState({ ...state, amount: event.target.value.replace(/\D/g, "") })} /></FormField>
        <BaseButton variant="row" disabled={saving || loading} onClick={() => setChoosingCategory(true)}>{categories.find(category => category.id === state.categoryID)?.name ?? "Chọn nhóm"}</BaseButton>
        <FormField label={COPY.note}><BaseTextInput variant="inline" disabled={saving || loading} value={state.note} onChange={(event) => setState({ ...state, note: event.target.value })} /></FormField>
        </SurfaceCard>
        <SurfaceCard padding="md">
        <FormField label={COPY.occurredAt}><BaseTextInput type="datetime-local" required disabled={saving || loading} value={state.occurredAt} onChange={(event) => setState({ ...state, occurredAt: event.target.value })} /></FormField>
        </SurfaceCard>
        <BaseCheckbox label={COPY.reports} disabled={saving || loading} checked={state.includedInReports} onChange={(event) => setState({ ...state, includedInReports: event.target.checked })}>{COPY.reports}</BaseCheckbox>
        <div className="flex gap-2"><BaseButton className="flex-1" loading={saving} disabled={loading || wallets.length === 0} onClick={() => void submit()}>{COPY.save}</BaseButton>{transaction ? <BaseButton variant="danger" disabled={saving} onClick={() => void remove()}><Trash2 size={16} />{COPY.delete}</BaseButton> : null}</div>
      </div>}
    </BaseBottomSheet>
  );
}
