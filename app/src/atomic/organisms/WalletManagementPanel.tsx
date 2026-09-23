import { useEffect, useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { WalletSelectionList } from "../molecules/WalletSelectionList";
import { WalletDetailPanel } from "./SavingsWalletPanel";
import { ArrowLeft, CreditCard, Landmark, PiggyBank, Plus, Trash2, WalletCards } from "lucide-react";

import { BaseButton } from "../atoms/BaseButton";
import { BaseLink } from "../atoms/BaseLink";
import { IconBadge } from "../atoms/IconBadge";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { EMPTY_WALLET_FORM, type WalletFormState, WalletEditorForm, walletFormToInput } from "../molecules/WalletEditorForm";
import { formatVND } from "../utils/format";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
import { WalletCreateForm, WalletTypePicker, canCreateWallet } from "../molecules/WalletCreateForm";
import { fetchTransactions, type Transaction } from "../../services/transactions";
import { fetchCategories, type Category } from "../../services/categories";
import { createWallet, deleteWallet, fetchWallets, updateWallet, WALLET_TYPES, type Wallet } from "../../services/wallets";

const COPY = { title: "Ví của tôi", subtitle: "Tạo, sửa hoặc xóa ví", add: "Thêm ví", loading: "Đang tải ví...", empty: "Chưa có ví nào.", error: "Không thể lưu thay đổi. Vui lòng thử lại.", edit: "Sửa", delete: "Xóa" } as const;
const WALLET_ICONS = { [WALLET_TYPES.basic]: WalletCards, [WALLET_TYPES.goal]: PiggyBank, [WALLET_TYPES.credit]: CreditCard } as const;

function toForm(wallet: Wallet): WalletFormState {
  return { name: wallet.name, type: wallet.type, openingBalance: String(wallet.opening_balance), isInTotal: wallet.is_in_total, description: wallet.description ?? "", targetAmount: wallet.target_amount ? String(wallet.target_amount) : "", targetDate: wallet.target_date?.slice(0,10) ?? "", creditLimit: wallet.credit_limit ? String(wallet.credit_limit) : "" };
}

function deleteWarning(transactionCount: number): string {
  const impact = transactionCount > 0 ? ` ${transactionCount} giao dịch thuộc ví cũng sẽ bị xóa và báo cáo sẽ được tính lại.` : " Ví chưa có giao dịch.";
  return `Xóa vĩnh viễn ví này?${impact} Thao tác không thể hoàn tác.`;
}

export function WalletManagementPanel({ onChanged, refreshKey = 0 }: { onChanged: () => void; refreshKey?: number }) {
  const navigate = useNavigate();
  const [editingList, setEditingList] = useState(false);
  const [selectedID, setSelectedID] = useState<string | null>(null);
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [form, setForm] = useState<WalletFormState | null>(null);
  const [editingID, setEditingID] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [choosingType, setChoosingType] = useState(false);

  const load = async () => {
    try {
      setLoading(true);
      const [nextWallets, nextTransactions, nextCategories] = await Promise.all([fetchWallets(), fetchTransactions(), fetchCategories()]);
      setWallets(nextWallets);
      setTransactions(nextTransactions);
      setCategories(nextCategories);
      setError("");
    } catch {
      setError(COPY.error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void load(); }, [refreshKey]);

  const save = async () => {
    if (!form || saving || (!editingID && !canCreateWallet(form))) return;
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
    if (!window.confirm(deleteWarning(count))) return false;
    try {
      await deleteWallet(id);
      await load();
      onChanged();
      setForm(null);
      setEditingID(null);
      setSelectedID(current => current === id ? null : current);
      return true;
    } catch {
      setError(COPY.error);
      return false;
    }
  };

  const selectedWallet = wallets.find(wallet => wallet.id === selectedID);
  if (selectedWallet && !editingList) return <div className="space-y-3">{error ? <StatusMessage tone="danger">{error}<BaseButton variant="ghost" onClick={() => void load()}>Thử lại</BaseButton></StatusMessage> : null}{loading ? <StatusMessage>Đang cập nhật số dư...</StatusMessage> : null}<WalletDetailPanel wallet={selectedWallet} transactions={transactions} categories={categories} onBack={() => setSelectedID(null)} onChanged={onChanged} /></div>;

  return (
    <section className="space-y-3">
      <header className="mb-6 grid grid-cols-[1fr_auto_1fr] items-center"><BaseButton variant="chip" size="sm" className="justify-self-start" onClick={() => void navigate({ to: "/account" })}>Đóng</BaseButton><Text as="h1" size="lg" weight="bold">Chọn Ví</Text><BaseButton variant="chip" size="sm" className="justify-self-end" onClick={() => setEditingList(value => !value)}>{editingList ? "Xong" : "Sửa"}</BaseButton></header>
      {error && (!form || editingID) ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
      {form && editingID ? <BaseBottomSheet title="Sửa ví" closeLabel="Đóng" onClose={() => { if (!saving) { setForm(null); setEditingID(null); } }}><div className="space-y-3">{error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}<WalletEditorForm state={form} onChange={setForm} onSubmit={() => void save()} onCancel={() => { if (!saving) { setForm(null); setEditingID(null); } }} saving={saving} editing /><BaseButton variant="danger" disabled={saving} onClick={() => void remove(editingID)}><Trash2 size={16} />Xóa ví</BaseButton></div></BaseBottomSheet> : null}
      {form && !editingID ? <BaseBottomSheet title={choosingType ? "Chọn loại ví" : "Thêm Ví"} closeLabel={choosingType ? "Quay lại" : "Hủy"} presentation="form" closingDisabled={saving} onClose={() => { if (saving) return; if (choosingType) setChoosingType(false); else { setForm(null); setError(""); } }} headerAction={!choosingType ? <BaseButton type="submit" form="wallet-create-form" size="sm" loading={saving} loadingLabel="Đang lưu" disabled={!canCreateWallet(form)}>Lưu</BaseButton> : undefined}>
        {choosingType ? <WalletTypePicker value={form.type} onSelect={(type) => { setForm({ ...form, type }); setChoosingType(false); }} /> : <div className="space-y-3">{error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}<WalletCreateForm state={form} onChange={setForm} saving={saving} onSelectType={() => setChoosingType(true)} onSubmit={() => void save()} /></div>}
      </BaseBottomSheet> : null}
      {loading ? <StatusMessage>{COPY.loading}</StatusMessage> : null}
      {!loading ? <WalletSelectionList wallets={wallets} selectedID={selectedID} editing={editingList} onAdd={() => { setEditingID(null); setError(""); setForm(EMPTY_WALLET_FORM); }} onSelect={(id) => { if (editingList && id) { const wallet = wallets.find(item => item.id === id); if (wallet) { setError(""); setEditingID(id); setForm(toForm(wallet)); } } else setSelectedID(id); }} /> : null}
    </section>
  );
}
