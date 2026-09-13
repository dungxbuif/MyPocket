import { useEffect, useMemo, useState } from "react";
import { Trash2 } from "lucide-react";

import { BaseButton } from "../atoms/BaseButton";
import { BaseCheckbox } from "../atoms/BaseCheckbox";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { SegmentedControl } from "../atoms/SegmentedControl";
import { StatusMessage } from "../atoms/StatusMessage";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
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

export function QuickAddSheet({ onClose, onSaved, transaction }: { onClose: () => void; onSaved: () => void; transaction?: Transaction }) {
  const [state, setState] = useState(() => initialState(transaction));
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

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
    () => categories.filter((category) => categoryAppliesToTransaction(category, state.type, state.walletID)),
    [categories, state.type, state.walletID],
  );

  const changeScope = (updates: Partial<Pick<EditorState, "type" | "walletID">>) => {
    setState((current) => {
      const next = { ...current, ...updates };
      const selected = categories.find((category) => category.id === next.categoryID);
      if (selected && !categoryAppliesToTransaction(selected, next.type, next.walletID)) next.categoryID = "";
      return next;
    });
  };

  const submit = async () => {
    const amount = Number(state.amount);
    if (!state.walletID || !Number.isSafeInteger(amount) || amount <= 0 || !state.occurredAt) {
      setError(COPY.validationError);
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
    if (!transaction || !window.confirm(COPY.confirmDelete)) return;
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
    <BaseBottomSheet title={transaction ? COPY.editTitle : COPY.addTitle} closeLabel={COPY.close} onClose={onClose}>
      <div className="space-y-3">
        {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
        {loading ? <StatusMessage>{COPY.loading}</StatusMessage> : null}
        {!loading && wallets.length === 0 ? <StatusMessage>{COPY.noWallet}</StatusMessage> : null}
        <SegmentedControl value={state.type} onChange={(type) => changeScope({ type })} options={[{ value: "expense", label: COPY.expense }, { value: "income", label: COPY.income }]} />
        <FormField label={COPY.amount}><BaseTextInput autoFocus inputMode="numeric" required disabled={saving || loading} value={state.amount} onChange={(event) => setState({ ...state, amount: event.target.value.replace(/\D/g, "") })} /></FormField>
        <FormField label={COPY.wallet}><BaseSelect required disabled={saving || loading || wallets.length === 0} value={state.walletID} onChange={(event) => changeScope({ walletID: event.target.value })}><option value="">Chọn ví</option>{wallets.map((wallet) => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</BaseSelect></FormField>
        <FormField label={COPY.category}><BaseSelect disabled={saving || loading} value={state.categoryID} onChange={(event) => setState({ ...state, categoryID: event.target.value })}><option value="">{COPY.categoryOptional}</option>{applicableCategories.map((category) => <option key={category.id} value={category.id}>{category.name}</option>)}</BaseSelect></FormField>
        <FormField label={COPY.occurredAt}><BaseTextInput type="datetime-local" required disabled={saving || loading} value={state.occurredAt} onChange={(event) => setState({ ...state, occurredAt: event.target.value })} /></FormField>
        <FormField label={COPY.note}><BaseTextInput disabled={saving || loading} value={state.note} onChange={(event) => setState({ ...state, note: event.target.value })} /></FormField>
        <BaseCheckbox label={COPY.reports} disabled={saving || loading} checked={state.includedInReports} onChange={(event) => setState({ ...state, includedInReports: event.target.checked })}>{COPY.reports}</BaseCheckbox>
        <div className="flex gap-2"><BaseButton className="flex-1" loading={saving} disabled={loading || wallets.length === 0} onClick={() => void submit()}>{COPY.save}</BaseButton>{transaction ? <BaseButton variant="danger" disabled={saving} onClick={() => void remove()}><Trash2 size={16} />{COPY.delete}</BaseButton> : null}</div>
      </div>
    </BaseBottomSheet>
  );
}
