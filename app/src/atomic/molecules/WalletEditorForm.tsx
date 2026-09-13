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
};

export const EMPTY_WALLET_FORM: WalletFormState = {
  name: "",
  type: WALLET_TYPES.basic,
  openingBalance: "0",
  isInTotal: true,
  description: "",
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
  };
}

export function WalletEditorForm({ state, onChange, onSubmit, onCancel, saving }: { state: WalletFormState; onChange: (next: WalletFormState) => void; onSubmit: () => void; onCancel: () => void; saving: boolean }) {
  return <form className="space-y-3" onSubmit={(event) => { event.preventDefault(); onSubmit(); }}>
    <FormField label={COPY.name}><BaseTextInput required disabled={saving} value={state.name} onChange={(event) => onChange({ ...state, name: event.target.value })} /></FormField>
    <FormField label={COPY.type}><BaseSelect disabled={saving} value={state.type} onChange={(event) => onChange({ ...state, type: event.target.value as WalletType })}>{Object.values(WALLET_TYPES).map((type) => <option key={type} value={type}>{TYPE_LABELS[type]}</option>)}</BaseSelect></FormField>
    <FormField label={COPY.openingBalance}><BaseTextInput disabled={saving} inputMode="numeric" value={state.openingBalance} onChange={(event) => onChange({ ...state, openingBalance: event.target.value.replace(/[^0-9-]/g, "") })} /></FormField>
    <BaseCheckbox label={COPY.includeTotal} disabled={saving} checked={state.isInTotal} onChange={(event) => onChange({ ...state, isInTotal: event.target.checked })}>{COPY.includeTotal}</BaseCheckbox>
    <FormField label={COPY.description}><BaseTextInput disabled={saving} value={state.description} onChange={(event) => onChange({ ...state, description: event.target.value })} /></FormField>
    <div className="flex gap-2"><BaseButton type="submit" loading={saving} className="flex-1">{COPY.save}</BaseButton><BaseButton type="button" variant="ghost" disabled={saving} onClick={onCancel}>{COPY.cancel}</BaseButton></div>
  </form>;
}
