import { useEffect, useState } from "react";
import { ArrowLeft, CreditCard, Landmark, PiggyBank, Plus, Trash2, WalletCards } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { BaseButton } from "../atoms/BaseButton";
import { IconBadge } from "../atoms/IconBadge";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { EMPTY_WALLET_FORM, type WalletFormState, WalletEditorForm, walletFormToInput } from "../molecules/WalletEditorForm";
import { createWallet, deleteWallet, fetchWallets, updateWallet, WALLET_TYPES, type Wallet } from "../../services/wallets";
import { formatVND } from "../utils/format";

const COPY = { title: "Ví của tôi", subtitle: "Tạo, sửa hoặc xóa ví", add: "Thêm ví", loading: "Đang tải ví...", empty: "Chưa có ví nào.", error: "Không thể lưu thay đổi. Vui lòng thử lại.", confirmDelete: "Xóa ví này? Thao tác không thể hoàn tác.", edit: "Sửa", delete: "Xóa" } as const;

const WALLET_ICONS = { [WALLET_TYPES.basic]: WalletCards, [WALLET_TYPES.goal]: PiggyBank, [WALLET_TYPES.credit]: CreditCard } as const;

function toForm(wallet: Wallet): WalletFormState { return { name: wallet.name, type: wallet.type, openingBalance: String(wallet.opening_balance), isInTotal: wallet.is_in_total, description: wallet.description ?? "" }; }

export function WalletManagementPanel() {
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [form, setForm] = useState<WalletFormState | null>(null);
  const [editingID, setEditingID] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const load = async () => { try { setLoading(true); setWallets(await fetchWallets()); } catch { setError(COPY.error); } finally { setLoading(false); } };
  useEffect(() => { void load(); }, []);
  const save = async () => { if (!form) return; try { setSaving(true); setError(""); const input = walletFormToInput(form); if (editingID) await updateWallet(editingID, input); else await createWallet(input); setForm(null); setEditingID(null); await load(); } catch { setError(COPY.error); } finally { setSaving(false); } };
  const remove = async (id: string) => { if (!window.confirm(COPY.confirmDelete)) return; try { await deleteWallet(id); await load(); } catch { setError(COPY.error); } };
  return <section className="space-y-3"><div className="flex items-center gap-3"><Link to="/account" className="grid h-10 w-10 place-items-center rounded-full border border-[#e3e2e2] bg-white"><ArrowLeft size={18} /></Link><div className="min-w-0 flex-1"><h1 className="text-xl font-bold">{COPY.title}</h1><p className="text-xs text-[#6f7a6b]">{COPY.subtitle}</p></div><BaseButton size="sm" onClick={() => { setEditingID(null); setForm(EMPTY_WALLET_FORM); }}><Plus size={16} /> {COPY.add}</BaseButton></div>{error ? <p className="rounded-2xl bg-[#ffdad6] p-3 text-sm text-[#93000a]">{error}</p> : null}{form ? <SurfaceCard className="p-4"><WalletEditorForm state={form} onChange={setForm} onSubmit={() => void save()} onCancel={() => { setForm(null); setEditingID(null); }} saving={saving} /></SurfaceCard> : null}{loading ? <SurfaceCard className="p-5 text-sm text-[#6f7a6b]">{COPY.loading}</SurfaceCard> : null}{!loading && wallets.length === 0 ? <SurfaceCard className="p-5 text-sm text-[#6f7a6b]">{COPY.empty}</SurfaceCard> : null}{wallets.map((wallet) => { const Icon = WALLET_ICONS[wallet.type] ?? Landmark; return <SurfaceCard key={wallet.id} className="p-4"><div className="flex items-center gap-3"><IconBadge icon={Icon} /><div className="min-w-0 flex-1"><p className="truncate font-bold">{wallet.name}</p><p className="text-xs text-[#6f7a6b]">{wallet.type}</p></div><p className="money text-sm font-bold">{formatVND(wallet.opening_balance)}</p></div><div className="mt-3 flex gap-2"><BaseButton size="sm" variant="ghost" onClick={() => { setEditingID(wallet.id); setForm(toForm(wallet)); }}>{COPY.edit}</BaseButton><BaseButton size="sm" variant="danger" onClick={() => void remove(wallet.id)}><Trash2 size={15} /> {COPY.delete}</BaseButton></div></SurfaceCard>; })}</section>;
}
