import { useMemo, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseCheckbox } from "../atoms/BaseCheckbox";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { CategoryIconPicker } from "./CategoryIconPicker";
import type { Category, CategoryInput } from "../../services/categories";
import type { Wallet } from "../../services/wallets";

const FORM_TEXT = {
  name: "Tên nhóm",
  kind: "Loại nhóm",
  parent: "Nhóm cha",
  icon: "Biểu tượng",
  noParent: "Không có nhóm cha",
  expense: "Khoản chi",
  income: "Khoản thu",
  debt: "Vay/Nợ",
  save: "Lưu nhóm",
  delete: "Xóa nhóm",
  cancel: "Hủy",
  required: "Nhập tên nhóm để tiếp tục.",
  wallets: "Ví áp dụng",
  walletHint: "Không chọn ví nghĩa là chưa giới hạn ví áp dụng.",
} as const;

const CATEGORY_KINDS = [
  { value: "expense", label: FORM_TEXT.expense },
  { value: "income", label: FORM_TEXT.income },
  { value: "debt", label: FORM_TEXT.debt },
] as const;

export function CategoryEditForm({ category, categories, wallets, saving, onSave, onDelete, onCancel }: { category: Category | null; categories: Category[]; wallets: Wallet[]; saving: boolean; onSave: (input: CategoryInput) => Promise<void>; onDelete?: () => Promise<void>; onCancel: () => void }) {
  const [name, setName] = useState(category?.name ?? "");
  const [kind, setKind] = useState(category?.kind ?? "expense");
  const [parentID, setParentID] = useState(category?.parent_id ?? "");
  const [iconKey, setIconKey] = useState(category?.icon_key ?? "tag");
  const [walletIDs, setWalletIDs] = useState(category?.wallet_ids ?? []);
  const [error, setError] = useState("");
  const parents = useMemo(() => categories.filter((item) => !item.parent_id && item.kind === kind && item.id !== category?.id), [categories, category?.id, kind]);
  const toggleWallet = (walletID: string) => setWalletIDs((current) => current.includes(walletID) ? current.filter((item) => item !== walletID) : [...current, walletID]);
  const submit = async () => { if (!name.trim()) { setError(FORM_TEXT.required); return; } setError(""); await onSave({ name: name.trim(), kind, parent_id: parentID || null, wallet_ids: walletIDs, icon_key: iconKey }); };
  return <div className="space-y-3">
    <SurfaceCard padding="md" className="space-y-4">
      <FormField label={FORM_TEXT.name}><BaseTextInput value={name} onChange={(event) => setName(event.target.value)} /></FormField>
      <FormField label={FORM_TEXT.icon}><CategoryIconPicker value={iconKey} onChange={setIconKey} /></FormField>
      <FormField label={FORM_TEXT.kind}><BaseSelect value={kind} onChange={(event) => { setKind(event.target.value); setParentID(""); }}>{CATEGORY_KINDS.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</BaseSelect></FormField>
      <FormField label={FORM_TEXT.parent}><BaseSelect value={parentID} onChange={(event) => setParentID(event.target.value)}><option value="">{FORM_TEXT.noParent}</option>{parents.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</BaseSelect></FormField>
    </SurfaceCard>
    <SurfaceCard padding="md"><FormField label={FORM_TEXT.wallets}><p className="mt-1 text-xs text-slate-500">{FORM_TEXT.walletHint}</p><div className="mt-1 divide-y divide-slate-100">{wallets.map((wallet) => <BaseCheckbox key={wallet.id} label={wallet.name} checked={walletIDs.includes(wallet.id)} onChange={() => toggleWallet(wallet.id)}>{wallet.name}</BaseCheckbox>)}</div></FormField></SurfaceCard>
    {error ? <p role="alert" className="text-sm text-rose-600">{error}</p> : null}
    <div className="flex gap-2"><BaseButton className="flex-1" loading={saving} onClick={submit}>{FORM_TEXT.save}</BaseButton><BaseButton variant="secondary" disabled={saving} onClick={onCancel}>{FORM_TEXT.cancel}</BaseButton></div>
    {category && onDelete ? <BaseButton variant="danger" className="w-full" disabled={saving} onClick={() => void onDelete()}>{FORM_TEXT.delete}</BaseButton> : null}
  </div>;
}
