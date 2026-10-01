import { useEffect, useMemo, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { AmountField } from "../molecules/AmountField";
import { Divider } from "../atoms/Divider";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { formatVND } from "../utils/format";
import { createCreditEntry, createCreditPayment, fetchCreditStatement, type CreditKind, type CreditStatement } from "../../services/credit";
import type { Category } from "../../services/categories";
import type { Wallet } from "../../services/wallets";

const entryKinds: Array<{ value: CreditKind; label: string }> = [
  { value: "purchase", label: "Mua hàng" },
  { value: "refund", label: "Hoàn tiền" },
  { value: "fee", label: "Phí" },
  { value: "interest", label: "Lãi" },
];

export function CreditWalletPanel({ wallet, wallets, categories, onChanged }: { wallet: Wallet; wallets: Wallet[]; categories: Category[]; onChanged: () => void }) {
  const [statement, setStatement] = useState<CreditStatement | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [action, setAction] = useState<"entry" | "payment" | null>(null);
  const [saving, setSaving] = useState(false);
  const [kind, setKind] = useState<CreditKind>("purchase");
  const [amount, setAmount] = useState("");
  const [sourceWalletID, setSourceWalletID] = useState("");
  const [note, setNote] = useState("");
  const sourceWallets = useMemo(() => wallets.filter(item => item.type !== "credit" && item.id !== wallet.id), [wallets, wallet.id]);
  const categoriesForKind = useMemo(() => categories.filter(category => category.kind === (kind === "refund" ? "income" : "expense")), [categories, kind]);
  const [categoryID, setCategoryID] = useState("");

  const load = () => {
    setLoading(true); setError("");
    fetchCreditStatement(wallet.id).then(setStatement).catch(() => setError("Không thể tải sổ tín dụng.")).finally(() => setLoading(false));
  };
  useEffect(load, [wallet.id]);
  useEffect(() => { if (!sourceWalletID && sourceWallets[0]) setSourceWalletID(sourceWallets[0].id); }, [sourceWalletID, sourceWallets]);
  useEffect(() => { if (!categoryID || !categoriesForKind.some(category => category.id === categoryID)) setCategoryID(categoriesForKind[0]?.id ?? ""); }, [categoryID, categoriesForKind]);

  const close = () => { if (!saving) { setAction(null); setAmount(""); setNote(""); } };
  const save = async () => {
    const value = Number(amount);
    if (!Number.isFinite(value) || value <= 0) return;
    try {
      setSaving(true);
      if (action === "entry") await createCreditEntry(wallet.id, { kind, amount: value, category_id: categoryID || undefined, note: note || undefined, occurred_at: new Date().toISOString() });
      if (action === "payment") await createCreditPayment(wallet.id, { source_wallet_id: sourceWalletID, amount: value, note: note || undefined, occurred_at: new Date().toISOString() });
      close(); load(); onChanged();
    } catch { setError("Không thể lưu thao tác tín dụng."); } finally { setSaving(false); }
  };
  const balanceLabel = statement ? formatVND(Math.abs(statement.balance)) : "—";
  return <div className="space-y-3">
    {loading ? <StatusMessage>Đang tải sổ tín dụng...</StatusMessage> : null}
    {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
    {statement ? <SurfaceCard tone="form" padding="md"><div className="grid grid-cols-2 gap-3"><div><Text size="xs" tone="secondary">Dư nợ hiện tại</Text><Text size="xl" weight="bold" numeric>{balanceLabel}</Text></div><div><Text size="xs" tone="secondary">Còn khả dụng</Text><Text size="xl" weight="bold" numeric tone={statement.available_credit < 0 ? "danger" : "action"}>{formatVND(statement.available_credit)}</Text></div></div><Divider /><div className="flex justify-between"><Text size="sm" tone="secondary">Hạn mức</Text><Text size="sm" numeric>{formatVND(statement.credit_limit)}</Text></div>{statement.payment_status !== "not_configured" ? <><div className="mt-2 flex justify-between"><Text size="sm" tone="secondary">Kỳ này cần trả</Text><Text size="sm" numeric tone={statement.payment_status === "overdue" ? "danger" : "ink"}>{formatVND(statement.amount_due)}</Text></div><Text size="xs" tone={statement.payment_status === "overdue" ? "danger" : "secondary"}>{statement.payment_status === "paid" ? "Đã thanh toán" : statement.payment_status === "partial" ? "Đã trả một phần" : statement.payment_status === "overdue" ? "Đã quá hạn" : "Đang đến hạn"}{statement.payment_due_at ? ` · hạn ${new Date(statement.payment_due_at).toLocaleDateString("vi-VN")}` : ""}</Text></> : <Text size="xs" tone="secondary" className="mt-2">Chưa cấu hình dư nợ sao kê và ngày đến hạn.</Text>}</SurfaceCard> : null}
    <div className="grid grid-cols-2 gap-2"><BaseButton onClick={() => setAction("entry")}>Thêm giao dịch</BaseButton><BaseButton variant="chip" onClick={() => setAction("payment")}>Thanh toán thẻ</BaseButton></div>
    {statement && statement.items.length === 0 ? <StatusMessage variant="plain">Chưa có giao dịch tín dụng.</StatusMessage> : null}
    {statement?.items.map(item => <SurfaceCard key={item.id} tone="form" padding="sm"><div className="flex items-center justify-between gap-3"><div><Text size="sm">{item.credit_kind === "purchase" ? "Mua hàng" : item.credit_kind === "refund" ? "Hoàn tiền" : item.credit_kind === "fee" ? "Phí" : item.credit_kind === "payment" ? "Thanh toán" : "Lãi"}</Text><Text size="xs" tone="secondary">{item.note || new Date(item.occurred_at).toLocaleDateString("vi-VN")}</Text></div><Text numeric tone={item.type === "income" ? "action" : "danger"}>{item.type === "income" ? "+" : "−"}{formatVND(item.amount)}</Text></div></SurfaceCard>)}
    {action ? <BaseBottomSheet presentation="form" title={action === "payment" ? "Thanh toán thẻ" : "Thêm giao dịch tín dụng"} closeLabel="Hủy" onClose={close} closingDisabled={saving} footer={<BaseButton className="w-full" size="lg" loading={saving} disabled={!amount || (action === "payment" && !sourceWalletID)} onClick={() => void save()}>Lưu</BaseButton>}><div className="space-y-3"><SurfaceCard tone="form" padding="md">{action === "entry" ? <><FormField label="Loại giao dịch"><BaseSelect value={kind} disabled={saving} onChange={event => setKind(event.target.value as CreditKind)}>{entryKinds.map(item => <option key={item.value} value={item.value}>{item.label}</option>)}</BaseSelect></FormField><Divider /><FormField label="Nhóm"><BaseSelect value={categoryID} disabled={saving} onChange={event => setCategoryID(event.target.value)}><option value="">Không chọn nhóm</option>{categoriesForKind.map(category => <option key={category.id} value={category.id}>{category.name}</option>)}</BaseSelect></FormField></> : <FormField label="Ví thanh toán"><BaseSelect value={sourceWalletID} disabled={saving} onChange={event => setSourceWalletID(event.target.value)}>{sourceWallets.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}</BaseSelect></FormField>}<AmountField value={amount} disabled={saving} onChange={setAmount} /><FormField label="Ghi chú"><BaseTextInput variant="inline" value={note} disabled={saving} onChange={event => setNote(event.target.value)} /></FormField></SurfaceCard></div></BaseBottomSheet> : null}
  </div>;
}
