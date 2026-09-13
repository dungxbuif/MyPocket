import { BaseButton } from "../atoms/BaseButton";
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
    <label className="block text-sm font-semibold text-[#1b1c1c]">{COPY.name}<input required value={state.name} onChange={(event) => onChange({ ...state, name: event.target.value })} className="mt-1 w-full rounded-2xl border border-[#e3e2e2] bg-white px-3 py-3 font-normal" /></label>
    <label className="block text-sm font-semibold text-[#1b1c1c]">{COPY.type}<select value={state.type} onChange={(event) => onChange({ ...state, type: event.target.value as WalletType })} className="mt-1 w-full rounded-2xl border border-[#e3e2e2] bg-white px-3 py-3 font-normal">{Object.values(WALLET_TYPES).map((type) => <option key={type} value={type}>{TYPE_LABELS[type]}</option>)}</select></label>
    <label className="block text-sm font-semibold text-[#1b1c1c]">{COPY.openingBalance}<input inputMode="numeric" value={state.openingBalance} onChange={(event) => onChange({ ...state, openingBalance: event.target.value.replace(/[^0-9-]/g, "") })} className="mt-1 w-full rounded-2xl border border-[#e3e2e2] bg-white px-3 py-3 font-normal" /></label>
    <label className="flex cursor-pointer items-center gap-3 text-sm font-semibold text-[#1b1c1c]"><input type="checkbox" checked={state.isInTotal} onChange={(event) => onChange({ ...state, isInTotal: event.target.checked })} />{COPY.includeTotal}</label>
    <label className="block text-sm font-semibold text-[#1b1c1c]">{COPY.description}<input value={state.description} onChange={(event) => onChange({ ...state, description: event.target.value })} className="mt-1 w-full rounded-2xl border border-[#e3e2e2] bg-white px-3 py-3 font-normal" /></label>
    <div className="flex gap-2"><BaseButton type="submit" loading={saving} className="flex-1">{COPY.save}</BaseButton><BaseButton type="button" variant="ghost" onClick={onCancel}>{COPY.cancel}</BaseButton></div>
  </form>;
}
