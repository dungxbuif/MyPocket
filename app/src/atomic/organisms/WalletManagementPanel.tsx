import { useEffect, useState } from "react";
import { ArrowLeft, CreditCard, Landmark, PiggyBank, Plus, Trash2, WalletCards } from "lucide-react";

import { BaseButton } from "../atoms/BaseButton";
import { BaseLink } from "../atoms/BaseLink";
import { IconBadge } from "../atoms/IconBadge";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { EMPTY_WALLET_FORM, type WalletFormState, WalletEditorForm, walletFormToInput } from "../molecules/WalletEditorForm";
import { formatVND } from "../utils/format";
import { fetchTransactions, type Transaction } from "../../services/transactions";
import { createWallet, deleteWallet, fetchWallets, updateWallet, WALLET_TYPES, type Wallet } from "../../services/wallets";

const COPY = { title: "Ví của tôi", subtitle: "Tạo, sửa hoặc xóa ví", add: "Thêm ví", loading: "Đang tải ví...", empty: "Chưa có ví nào.", error: "Không thể lưu thay đổi. Vui lòng thử lại.", edit: "Sửa", delete: "Xóa" } as const;
const WALLET_ICONS = { [WALLET_TYPES.basic]: WalletCards, [WALLET_TYPES.goal]: PiggyBank, [WALLET_TYPES.credit]: CreditCard } as const;

function toForm(wallet: Wallet): WalletFormState {
  return { name: wallet.name, type: wallet.type, openingBalance: String(wallet.opening_balance), isInTotal: wallet.is_in_total, description: wallet.description ?? "", targetAmount: wallet.target_amount ? String(wallet.target_amount) : "", creditLimit: wallet.credit_limit ? String(wallet.credit_limit) : "" };
}

function deleteWarning(transactionCount: number): string {
  const impact = transactionCount > 0 ? ` ${transactionCount} giao dịch thuộc ví cũng sẽ bị xóa và báo cáo sẽ được tính lại.` : " Ví chưa có giao dịch.";
  return `Xóa vĩnh viễn ví này?${impact} Thao tác không thể hoàn tác.`;
}

export function WalletManagementPanel({ onChanged, refreshKey = 0 }: { onChanged: () => void; refreshKey?: number }) {
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [form, setForm] = useState<WalletFormState | null>(null);
  const [editingID, setEditingID] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const load = async () => {
    try {
      setLoading(true);
      const [nextWallets, nextTransactions] = await Promise.all([fetchWallets(), fetchTransactions()]);
      setWallets(nextWallets);
      setTransactions(nextTransactions);
    } catch {
      setError(COPY.error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void load(); }, [refreshKey]);

  const save = async () => {
    if (!form) return;
    try {
      setSaving(true);
      setError("");
      const input = walletFormToInput(form);
      if (editingID) await updateWallet(editingID, input);
      else await createWallet(input);
      setForm(null);
      setEditingID(null);
      await load();
      onChanged();
    } catch {
      setError(COPY.error);
    } finally {
      setSaving(false);
    }
  };

  const remove = async (id: string) => {
    const count = transactions.filter((transaction) => transaction.wallet_id === id).length;
    if (!window.confirm(deleteWarning(count))) return;
    try {
      await deleteWallet(id);
      await load();
      onChanged();
    } catch {
      setError(COPY.error);
    }
  };

  return (
    <section className="space-y-3">
      <div className="flex items-center gap-3">
        <BaseLink to="/account" label="Quay lại Tài khoản"><ArrowLeft size={18} /></BaseLink>
        <div className="min-w-0 flex-1"><Text as="h1" size="xl" weight="bold">{COPY.title}</Text><Text size="xs" tone="secondary">{COPY.subtitle}</Text></div>
        <BaseButton size="sm" onClick={() => { setEditingID(null); setForm(EMPTY_WALLET_FORM); }}><Plus size={16} /> {COPY.add}</BaseButton>
      </div>
      {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
      {form ? <SurfaceCard padding="md"><WalletEditorForm state={form} onChange={setForm} onSubmit={() => void save()} onCancel={() => { setForm(null); setEditingID(null); }} saving={saving} editing={Boolean(editingID)} /></SurfaceCard> : null}
      {loading ? <StatusMessage>{COPY.loading}</StatusMessage> : null}
      {!loading && wallets.length === 0 ? <StatusMessage>{COPY.empty}</StatusMessage> : null}
      {wallets.map((wallet) => {
        const Icon = WALLET_ICONS[wallet.type] ?? Landmark;
        return <SurfaceCard key={wallet.id} padding="md"><div className="flex items-center gap-3"><IconBadge icon={Icon} /><div className="min-w-0 flex-1"><Text weight="bold" className="truncate">{wallet.name}</Text><Text size="xs" tone="secondary">{wallet.type}</Text></div><Text numeric size="sm" weight="bold">{formatVND(wallet.current_balance)}</Text></div><div className="mt-3 flex gap-2"><BaseButton size="sm" variant="ghost" onClick={() => { setEditingID(wallet.id); setForm(toForm(wallet)); }}>{COPY.edit}</BaseButton><BaseButton size="sm" variant="danger" onClick={() => void remove(wallet.id)}><Trash2 size={15} /> {COPY.delete}</BaseButton></div></SurfaceCard>;
      })}
    </section>
  );
}
