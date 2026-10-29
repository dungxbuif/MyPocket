import { Text } from "../atoms/Text";
import { useMemo, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseSelect, BaseTextInput } from "../atoms/FormField";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { IconBadge } from "../atoms/IconBadge";
import { categoryPresentationFor } from "../atoms/categoryPresentation";
import { CategoryIconPicker } from "./CategoryIconPicker";
import { ApplicableWalletsCard } from "./ApplicableWalletsCard";
import { InlineControlRow } from "./InlineControlRow";
import type { Category, CategoryInput } from "../../services/categories";
import type { Wallet } from "../../services/wallets";

const FORM_TEXT = {
  name: "Tên nhóm",
  kind: "Loại nhóm",
  parent: "Nhóm cha",
  parentPlaceholder: "Chọn nhóm cha",
  expense: "Khoản chi",
  income: "Khoản thu",
  debt: "Vay/Nợ",
  save: "Lưu nhóm",
  delete: "Xóa nhóm",
  cancel: "Hủy",
  required: "Nhập tên nhóm để tiếp tục.",
  systemHint: "Đây là nhóm của hệ thống. Bạn chỉ có thể chọn ví áp dụng.",
  saveWallets: "Lưu ví áp dụng",
  typeSymbol: "⁺⁄₋",
} as const;

const CATEGORY_KINDS = [
  { value: "expense", label: FORM_TEXT.expense },
  { value: "income", label: FORM_TEXT.income },
  { value: "debt", label: FORM_TEXT.debt },
] as const;

export function CategoryEditForm({ category, categories, wallets, saving, onSave, onCancel, readOnly = false }: { category: Category | null; categories: Category[]; wallets: Wallet[]; saving: boolean; onSave: (input: CategoryInput) => Promise<void>; onCancel: () => void; readOnly?: boolean }) {
  const [name, setName] = useState(category?.name ?? "");
  const [kind, setKind] = useState(category?.kind ?? new URLSearchParams(window.location.search).get("kind") ?? "expense");
  const [parentID, setParentID] = useState(category?.parent_id ?? "");
  const [iconKey, setIconKey] = useState(category?.icon_key ?? "tag");
  const [walletIDs, setWalletIDs] = useState(category?.wallet_ids ?? []);
  const [error, setError] = useState("");
  const parents = useMemo(() => categories.filter((item) => !item.parent_id && item.kind === kind && item.id !== category?.id), [categories, category?.id, kind]);
  const toggleWallet = (walletID: string) => setWalletIDs((current) => current.includes(walletID) ? current.filter((item) => item !== walletID) : [...current, walletID]);
  const submit = async () => { if (!readOnly && !name.trim()) { setError(FORM_TEXT.required); return; } setError(""); await onSave({ name: name.trim(), kind, parent_id: parentID || null, wallet_ids: walletIDs, icon_key: iconKey }); };
  const presentation = categoryPresentationFor(iconKey);
  return <div className="space-y-5">
    {readOnly ? <CategoryReadonlyMetadata icon={presentation.icon} tone={presentation.tone} name={name} kind={kind} /> : <SurfaceCard padding="none" className="divide-y divide-line p-2">
      <InlineControlRow leading={<CategoryIconPicker compact value={iconKey} onChange={setIconKey} />}><BaseTextInput aria-label={FORM_TEXT.name} placeholder={FORM_TEXT.name} value={name} variant="title" onChange={(event) => setName(event.target.value)} /></InlineControlRow>
      <InlineControlRow leading={<span className="grid h-11 w-11 place-items-center text-2xl font-semibold leading-none text-ink">{FORM_TEXT.typeSymbol}</span>}><BaseSelect aria-label={FORM_TEXT.kind} value={kind} variant="inline" onChange={(event) => { setKind(event.target.value); setParentID(""); }}>{CATEGORY_KINDS.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</BaseSelect></InlineControlRow>
      <InlineControlRow leading={<span className="grid h-11 w-11 place-items-center text-lg font-semibold leading-none text-ink">↳</span>}><BaseSelect aria-label={FORM_TEXT.parent} value={parentID} variant="inline" onChange={(event) => setParentID(event.target.value)}><option value="">{FORM_TEXT.parentPlaceholder}</option>{parents.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</BaseSelect></InlineControlRow>
    </SurfaceCard>}
    {readOnly ? <Text size="xs" tone="secondary" className="px-1">{FORM_TEXT.systemHint}</Text> : null}
    <ApplicableWalletsCard wallets={wallets} selectedWalletIDs={walletIDs} onToggle={toggleWallet} />
    {error ? <Text size="sm" tone="danger" role="alert" className="">{error}</Text> : null}
    <div className="flex gap-2"><BaseButton className="flex-1" loading={saving} onClick={submit}>{readOnly ? FORM_TEXT.saveWallets : FORM_TEXT.save}</BaseButton><BaseButton variant="secondary" disabled={saving} onClick={onCancel}>{FORM_TEXT.cancel}</BaseButton></div>
  </div>;
}

function CategoryReadonlyMetadata({ icon, tone, name, kind }: { icon: ReturnType<typeof categoryPresentationFor>["icon"]; tone: ReturnType<typeof categoryPresentationFor>["tone"]; name: string; kind: string }) {
  const kindLabel = CATEGORY_KINDS.find((item) => item.value === kind)?.label ?? kind;
  return <SurfaceCard padding="none" className="divide-y divide-line p-2"><div className="flex items-center gap-4 px-3 py-3.5"><IconBadge icon={icon} tone={tone} size="md" shape="circle" /><span className="text-xl font-bold tracking-tight text-ink">{name}</span></div><ReadonlyRow icon={FORM_TEXT.typeSymbol} label={kindLabel} /></SurfaceCard>;
}

function ReadonlyRow({ icon, label }: { icon: string; label: string }) {
  return <div className="flex items-center gap-4 px-3 py-3.5"><span className="grid h-11 w-11 place-items-center text-2xl font-semibold leading-none text-ink">{icon}</span><span className="text-base font-medium tracking-tight text-ink">{label}</span></div>;
}
