import { useState } from "react";
import { Globe, Tags, CalendarDays } from "lucide-react";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Divider } from "../atoms/Divider";
import { BaseTextInput } from "../atoms/FormField";
import { BaseButton } from "../atoms/BaseButton";
import { BaseSwitch } from "../atoms/BaseSwitch";
import { BaseModal } from "../atoms/BaseModal";
import { Text } from "../atoms/Text";
import { FormSelectorRow } from "./FormSelectorRow";
import { AmountField } from "./AmountField";
import { DateField } from "./DateField";
import { CategoryTreeSelector } from "./CategoryTreeSelector";
import type { Wallet } from "../../services/wallets";
import type { Category } from "../../services/categories";
export type BudgetDraft = { name: string; amount: string; wallet: string; category: string; start: string; end: string };
export function BudgetEditorForm({ draft, onChange, wallets, categories, disabled }: { draft: BudgetDraft; onChange: (draft: BudgetDraft) => void; wallets: Wallet[]; categories: Category[]; disabled: boolean }) {
  const [picker, setPicker] = useState<"wallet" | "category" | "period" | null>(null);
  const category = categories.find(item => item.id === draft.category);
  const change = (patch: Partial<BudgetDraft>) => onChange({ ...draft, ...patch });
  return <div className="space-y-5">
    <SurfaceCard tone="form" padding="md">
      <FormSelectorRow label="Nhóm chi" value={category?.name ?? "Tất cả nhóm chi"} icon={Tags} tone="neutral" disabled={disabled} onClick={() => setPicker("category")} />
      <Divider />
      <AmountField label="Hạn mức" disabled={disabled} value={draft.amount} onChange={amount => change({ amount })} />
      <Divider />
      <FormSelectorRow label="Khoảng thời gian" value={draft.start.split("-").reverse().join("/") + " – " + draft.end.split("-").reverse().join("/")} icon={CalendarDays} tone="neutral" disabled={disabled} onClick={() => setPicker("period")} />
      <Divider />
      <FormSelectorRow label="Ví áp dụng" value={wallets.find(item => item.id === draft.wallet)?.name ?? "Tổng cộng"} icon={Globe} disabled={disabled} onClick={() => setPicker("wallet")} />
      <Divider />
      <BaseTextInput variant="inline" aria-label="Tên ngân sách" placeholder="Tên ngân sách" disabled={disabled} value={draft.name} maxLength={200} onChange={e => change({ name: e.target.value })} />
    </SurfaceCard>
    <SurfaceCard tone="form" padding="md"><BaseSwitch label="Lặp lại ngân sách này" checked={false} disabled /><Text size="sm" tone="secondary">Ngân sách lặp lại chưa được hỗ trợ.</Text></SurfaceCard>
    {picker ? <BaseModal label={picker === "period" ? "Khoảng thời gian" : picker === "wallet" ? "Ví áp dụng" : "Nhóm chi"} onClose={() => setPicker(null)}>
      <div className="space-y-3">
        {picker === "category" ? <CategoryTreeSelector categories={categories.filter(item => item.kind === "expense")} selectedID={draft.category} clearLabel="Tất cả nhóm chi" onSelect={id => { change({ category: id }); setPicker(null); }} /> :
         picker === "wallet" ? <><FormSelectorRow label="Tổng cộng" value="Tổng cộng" icon={Globe} onClick={() => { change({ wallet: "" }); setPicker(null); }} />{wallets.map(wallet => <FormSelectorRow key={wallet.id} label={wallet.name} value={wallet.name} icon={Globe} onClick={() => { change({ wallet: wallet.id }); setPicker(null); }} />)}</> :
         <><Text weight="bold">Khoảng thời gian</Text><DateField label="Từ ngày" stepper={false} value={draft.start} onChange={start => change({ start })} /><Divider /><DateField label="Đến hết ngày" stepper={false} value={draft.end} onChange={end => change({ end })} /></>}
        <BaseButton variant="ghost" onClick={() => setPicker(null)}>Xong</BaseButton>
      </div>
    </BaseModal> : null}
  </div>;
}
