import { useState } from "react";
import { AlignLeft, Tags, WalletCards } from "lucide-react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { BaseSwitch } from "../atoms/BaseSwitch";
import { BaseModal } from "../atoms/BaseModal";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Divider } from "../atoms/Divider";
import { SegmentedControl } from "../atoms/SegmentedControl";
import { FormSelectorRow } from "./FormSelectorRow";
import { DateField } from "./DateField";
import { AmountField } from "./AmountField";
import { CategoryTreeSelector } from "./CategoryTreeSelector";
import { Text } from "../atoms/Text";
import { categoryPresentationFor } from "../atoms/categoryPresentation";
import { categoryAppliesToTransaction } from "../../services/transactionLogic";
import type { Category } from "../../services/categories";
import type { Wallet } from "../../services/wallets";
import type { TransactionType } from "../../services/transactions";
export type TransactionFormState = { type: TransactionType; amount: string; walletID: string; categoryID: string; jarID?: string; occurredAt: string; note: string; includedInReports: boolean };
export type TransactionJarOption = { jar_id: string; name: string; active: boolean };
export function TransactionFields({ state, onChange, wallets, categories, jars = [], disabled = false }: { state: TransactionFormState; onChange: (next: TransactionFormState) => void; wallets: Wallet[]; categories: Category[]; jars?: TransactionJarOption[]; disabled?: boolean }) {
  const [picker, setPicker] = useState<"wallet" | "category" | null>(null), [details, setDetails] = useState(false);
  const wallet = wallets.find(item => item.id === state.walletID);
  const applicable = categories.filter(category => categoryAppliesToTransaction(category, state.type, state.walletID, wallet?.type));
  const selectedCategory = categories.find(item => item.id === state.categoryID);
  const display = selectedCategory ? categoryPresentationFor(selectedCategory.system_key ?? selectedCategory.icon_key) : undefined;
  const change = (patch: Partial<TransactionFormState>) => {
    const next = { ...state, ...patch };
    if ("type" in patch || "walletID" in patch) {
      const c = categories.find(item => item.id === next.categoryID);
      if (c && !categoryAppliesToTransaction(c, next.type, next.walletID, wallets.find(item => item.id === next.walletID)?.type)) next.categoryID = "";
    }
    if (next.type === "income" || categories.find(item => item.id === next.categoryID)?.system_key === "expense_transfer_out") next.jarID = "";
    onChange(next);
  };
  return <div className="space-y-3">
    <SurfaceCard tone="form" padding="md">
      <SegmentedControl disabled={disabled} value={state.type} options={[{ value: "expense", label: "Khoản chi" }, { value: "income", label: "Khoản thu" }]} onChange={type => change({ type })} />
      <div className="mt-3"><FormSelectorRow label="Ví" value={wallet?.name ?? "Chọn ví"} placeholder={!wallet} icon={WalletCards} disabled={disabled} onClick={() => setPicker("wallet")} /></div>
      <Divider />
      <AmountField value={state.amount} disabled={disabled} onChange={amount => change({ amount })} />
      <Divider />
      <FormSelectorRow label="Nhóm" value={selectedCategory?.name ?? "Chọn nhóm"} placeholder={!selectedCategory} icon={display?.icon ?? Tags} tone={display?.tone ?? "neutral"} disabled={disabled} onClick={() => setPicker("category")} />
      <Divider />
      {state.type === "expense" && selectedCategory?.system_key !== "expense_transfer_out" && jars.length > 0 ? <><FormField label="Hũ (tùy chọn)"><BaseSelect aria-label="Chọn hũ cho giao dịch" disabled={disabled} value={state.jarID ?? ""} onChange={event => change({ jarID: event.target.value })}><option value="">Không gắn hũ</option>{jars.map(jar => <option key={jar.jar_id} value={jar.jar_id}>{jar.active ? jar.name : `${jar.name} (đã gỡ khỏi tháng)`}</option>)}</BaseSelect></FormField><Divider /></> : null}
      <div className="flex items-center gap-4 py-3"><AlignLeft size={22} aria-hidden className="shrink-0" /><BaseTextInput aria-label="Ghi chú" placeholder="Ghi chú" variant="inline" disabled={disabled} value={state.note} onChange={e => change({ note: e.target.value })} /></div>
      <Divider />
      <DateField value={state.occurredAt} disabled={disabled} onChange={day => change({ occurredAt: day.includes("T") ? day : day + "T00:00" })} />
    </SurfaceCard>
    <SurfaceCard tone="form" padding="sm"><BaseButton variant="ghost" disabled={disabled} className="w-full" aria-expanded={details} onClick={() => setDetails(!details)}>{details ? "Ẩn chi tiết" : "Thêm chi tiết"}</BaseButton>
      {details ? <div className="space-y-3 px-2"><BaseSwitch label="Tính vào báo cáo" disabled={disabled} checked={state.includedInReports} onChange={e => change({ includedInReports: e.target.checked })} /><div className="flex items-center gap-3"><Text as="span">Giờ giao dịch</Text><BaseTextInput type="time" aria-label="Giờ giao dịch" disabled={disabled || !state.occurredAt} value={state.occurredAt.slice(11, 16)} onChange={e => { if (e.target.value) change({ occurredAt: state.occurredAt.slice(0, 10) + "T" + e.target.value }); }} /></div></div> : null}
    </SurfaceCard>
    {picker ? <BaseModal label={picker === "wallet" ? "Chọn ví" : "Chọn nhóm"} onClose={() => setPicker(null)}>
      <div className="space-y-3"><Text size="lg" weight="bold">{picker === "wallet" ? "Chọn ví" : "Chọn nhóm"}</Text>
        {picker === "category" ? <CategoryTreeSelector categories={categories.filter(category => category.kind === state.type)} selectableIDs={applicable.map(category => category.id)} selectedID={state.categoryID} onSelect={categoryID => { change({ categoryID }); setPicker(null); }} /> : wallets.filter(item => item.type !== "credit").map(item => <FormSelectorRow key={item.id} label={item.name} value={item.name} icon={WalletCards} onClick={() => { change({ walletID: item.id }); setPicker(null); }} />)}
        <BaseButton variant="ghost" onClick={() => setPicker(null)}>Hủy</BaseButton>
      </div>
    </BaseModal> : null}
  </div>;
}
