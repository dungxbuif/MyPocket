import { Banknote, Check, ChevronRight, CreditCard, PiggyBank, WalletCards } from "lucide-react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseSwitch } from "../atoms/BaseSwitch";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { Divider } from "../atoms/Divider";
import { IconBadge } from "../atoms/IconBadge";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import type { WalletFormState } from "./WalletEditorForm";
import type { WalletType } from "../../services/wallets";

const TYPES = {
  basic: { label: "Ví thường", description: "Tiền mặt hoặc tài khoản thanh toán", icon: WalletCards },
  goal: { label: "Ví tiết kiệm", description: "Theo dõi mục tiêu tiết kiệm", icon: PiggyBank },
  credit: { label: "Ví tín dụng", description: "Theo dõi ví có hạn mức tín dụng", icon: CreditCard },
} as const;

export function canCreateWallet(state: WalletFormState): boolean {
  const integer = (value: string) => /^-?\d+$/.test(value) && Number.isSafeInteger(Number(value));
  return Boolean(state.name.trim()) && integer(state.openingBalance)
    && (state.type !== "goal" || (integer(state.targetAmount) && Number(state.targetAmount) > 0))
    && (state.type !== "credit" || (integer(state.creditLimit) && Number(state.creditLimit) > 0));
}

export function WalletTypePicker({ value, onSelect }: { value: WalletType; onSelect: (value: WalletType) => void }) {
  return <SurfaceCard padding="sm"><div className="space-y-1">{(Object.keys(TYPES) as WalletType[]).map((type) => <BaseButton key={type} variant="row" size="row" aria-pressed={value === type} onClick={() => onSelect(type)}>
    <IconBadge icon={TYPES[type].icon} shape="circle" tone="categoryOrange" />
    <span className="min-w-0 flex-1"><Text as="span" weight="semibold" className="block">{TYPES[type].label}</Text><Text as="span" size="xs" tone="secondary" className="block">{TYPES[type].description}</Text></span>
    {value === type ? <Check size={20} aria-hidden="true" /> : null}
  </BaseButton>)}</div></SurfaceCard>;
}

export function WalletCreateForm({ state, onChange, onSelectType, onSubmit, saving }: { state: WalletFormState; onChange: (state: WalletFormState) => void; onSelectType: () => void; onSubmit: () => void; saving: boolean }) {
  return <form id="wallet-create-form" className="space-y-5" onSubmit={(event) => { event.preventDefault(); if (!saving && canCreateWallet(state)) onSubmit(); }}>
    <SurfaceCard padding="md">
      <div className="flex items-center gap-3 pb-3"><IconBadge icon={TYPES[state.type].icon} shape="circle" tone="categoryOrange" size="lg" /><BaseTextInput aria-label="Tên ví" placeholder="Tên" variant="title" required disabled={saving} value={state.name} onChange={(event) => onChange({ ...state, name: event.target.value })} /></div>
      <Divider />
      <div className="flex items-center gap-3 py-3"><IconBadge icon={Banknote} shape="circle" tone="neutral" /><div><Text size="xs" tone="secondary">Đơn vị tiền tệ</Text><Text weight="medium">VND — Đồng Việt Nam</Text></div></div>
      {state.type !== "goal" ? <><Divider /><div className="pt-3"><FormField label="Số tiền hiện có"><BaseTextInput variant="title" inputMode="numeric" aria-label="Số tiền hiện có" required disabled={saving} value={state.openingBalance} onChange={(event) => onChange({ ...state, openingBalance: event.target.value })} /></FormField></div></> : null}
    </SurfaceCard>
    <SurfaceCard padding="md"><FormField label="Loại ví"><BaseSelect disabled={saving} value={state.type} onChange={(event) => onChange({ ...state, type: event.target.value as WalletType })}>{(Object.keys(TYPES) as WalletType[]).map(type => <option key={type} value={type}>{TYPES[type].label}</option>)}</BaseSelect></FormField>
      {state.type !== "basic" ? <div className="px-2 pb-2"><FormField label={state.type === "goal" ? "Mục tiêu tiết kiệm (VND)" : "Hạn mức tín dụng (VND)"}><BaseTextInput inputMode="numeric" required disabled={saving} value={state.type === "goal" ? state.targetAmount : state.creditLimit} onChange={(event) => onChange({ ...state, [state.type === "goal" ? "targetAmount" : "creditLimit"]: event.target.value })} /></FormField></div> : null}
      {state.type === "goal" ? <div className="space-y-3 px-2"><Divider /><FormField label="Số tiền ban đầu"><BaseTextInput variant="title" inputMode="numeric" required disabled={saving} value={state.openingBalance} onChange={event => onChange({ ...state, openingBalance:event.target.value })} /></FormField><Divider /><FormField label="Ngày kết thúc"><BaseTextInput type="date" disabled={saving} value={state.targetDate} onChange={event => onChange({ ...state, targetDate:event.target.value })} /></FormField></div> : null}
    </SurfaceCard>
    {state.type === "goal" ? <div className="space-y-2"><SurfaceCard padding="md"><BaseSwitch label="Bật thông báo" checked={false} disabled /></SurfaceCard><Text size="xs" tone="secondary">Thông báo giao dịch chưa được hỗ trợ.</Text></div> : <div className="space-y-2"><SurfaceCard padding="sm"><BaseButton variant="ghost" disabled className="w-full">Liên kết dịch vụ</BaseButton></SurfaceCard><Text size="xs" tone="secondary" className="text-center">Liên kết dịch vụ chưa được hỗ trợ.</Text></div>}
    <div className="space-y-2"><SurfaceCard padding="md"><BaseSwitch label="Không tính vào tổng" disabled={saving} checked={!state.isInTotal} onChange={(event) => onChange({ ...state, isInTotal: !event.target.checked })} /></SurfaceCard><Text size="xs" tone="secondary" className="px-3">Bỏ qua ví này và số dư khỏi “Tổng”.</Text></div>
  </form>;
}
