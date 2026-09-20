import { BaseButton } from "../atoms/BaseButton";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { BaseCheckbox } from "../atoms/BaseCheckbox";
import { WALLET_TYPES, type WalletInput, type WalletType } from "../../services/wallets";

const COPY = {
  name: "Tên ví",
  openingBalance: "Số dư đầu kỳ (VND)",
  type: "Loại ví",
  includeTotal: "Tính vào tổng số dư",
  description: "Ghi chú",
  save: "Lưu ví",
  cancel: "Hủy",
  basic: "Ví thường",
  goal: "Ví tiết kiệm",
  credit: "Ví tín dụng",
} as const;

export type WalletFormState = {
  name: string;
  type: WalletType;
  openingBalance: string;
  isInTotal: boolean;
  description: string;
  targetAmount: string;
  targetDate: string;
  creditLimit: string;
};

export const EMPTY_WALLET_FORM: WalletFormState = {
  name: "",
  type: WALLET_TYPES.basic,
  openingBalance: "0",
  isInTotal: true,
  description: "",
  targetAmount: "",
  targetDate: "",
  creditLimit: "",
};

const TYPE_LABELS: Record<WalletType, string> = {
  [WALLET_TYPES.basic]: COPY.basic,
  [WALLET_TYPES.goal]: COPY.goal,
  [WALLET_TYPES.credit]: COPY.credit,
};

export function walletFormToInput(state: WalletFormState): WalletInput {
  return {
    name: state.name.trim(),
    type: state.type,
    opening_balance: Number(state.openingBalance) || 0,
    is_in_total: state.isInTotal,
    description: state.description.trim() || undefined,
    target_amount: state.type === WALLET_TYPES.goal ? Number(state.targetAmount) : undefined,
    target_date: state.type === WALLET_TYPES.goal ? state.targetDate : undefined,
    credit_limit: state.type === WALLET_TYPES.credit ? Number(state.creditLimit) : undefined,
  };
}

export function WalletEditorForm({ state, onChange, onSubmit, onCancel, saving, editing = false }: { state: WalletFormState; onChange: (next: WalletFormState) => void; onSubmit: () => void; onCancel: () => void; saving: boolean; editing?: boolean }) {
  return <form className="space-y-3" onSubmit={(event) => { event.preventDefault(); onSubmit(); }}>
    <FormField label={COPY.name}><BaseTextInput required disabled={saving} value={state.name} onChange={(event) => onChange({ ...state, name: event.target.value })} /></FormField>
    <FormField label={editing ? "Loại ví (không đổi sau khi tạo)" : COPY.type}><BaseSelect disabled={saving || editing} value={state.type} onChange={(event) => onChange({ ...state, type: event.target.value as WalletType })}>{Object.values(WALLET_TYPES).map((type) => <option key={type} value={type}>{TYPE_LABELS[type]}</option>)}</BaseSelect></FormField>
    <FormField label={editing ? "Số dư đầu kỳ (không đổi khi sửa)" : COPY.openingBalance}><BaseTextInput disabled={saving || editing} inputMode="numeric" value={state.openingBalance} onChange={(event) => onChange({ ...state, openingBalance: event.target.value.replace(/[^0-9-]/g, "") })} /></FormField>
    {state.type === WALLET_TYPES.goal ? <FormField label="Mục tiêu tiết kiệm (VND)"><BaseTextInput required disabled={saving} inputMode="numeric" value={state.targetAmount} onChange={(event) => onChange({ ...state, targetAmount: event.target.value.replace(/\D/g, "") })} /></FormField> : null}
    {state.type === WALLET_TYPES.credit ? <FormField label="Hạn mức tín dụng (VND)"><BaseTextInput required disabled={saving} inputMode="numeric" value={state.creditLimit} onChange={(event) => onChange({ ...state, creditLimit: event.target.value.replace(/\D/g, "") })} /></FormField> : null}
    {state.type === WALLET_TYPES.goal ? <FormField label="Ngày kết thúc"><BaseTextInput type="date" disabled={saving} value={state.targetDate} onChange={(event) => onChange({ ...state, targetDate: event.target.value })} /></FormField> : null}
    <BaseCheckbox label={COPY.includeTotal} disabled={saving} checked={state.isInTotal} onChange={(event) => onChange({ ...state, isInTotal: event.target.checked })}>{COPY.includeTotal}</BaseCheckbox>
    {editing ? <FormField label={COPY.description}><BaseTextInput disabled={saving} value={state.description} onChange={(event) => onChange({ ...state, description: event.target.value })} /></FormField> : null}
    <div className="flex gap-2"><BaseButton type="submit" loading={saving} className="flex-1">{COPY.save}</BaseButton><BaseButton type="button" variant="ghost" disabled={saving} onClick={onCancel}>{COPY.cancel}</BaseButton></div>
  </form>;
}
