import { AlignLeft, ArrowDownToLine, ArrowUpFromLine, WalletCards } from "lucide-react";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Divider } from "../atoms/Divider";
import { DateField } from "./DateField";
import { AmountField } from "./AmountField";
import type { Wallet } from "../../services/wallets";

export type TransferFormState = {
  sourceWalletID: string;
  destinationWalletID: string;
  amount: string;
  occurredAt: string;
  note: string;
};

export function TransferFields({ state, onChange, wallets, disabled = false }: { state: TransferFormState; onChange: (next: TransferFormState) => void; wallets: Wallet[]; disabled?: boolean }) {
  const change = (patch: Partial<TransferFormState>) => onChange({ ...state, ...patch });
  return <SurfaceCard tone="form" padding="md">
    <FormField label="Ví chuyển đi"><div className="relative"><ArrowUpFromLine aria-hidden size={20} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-secondary" /><BaseSelect aria-label="Ví chuyển đi" disabled={disabled} value={state.sourceWalletID} onChange={event => change({ sourceWalletID: event.target.value })} className="pl-10"><option value="">Chọn ví nguồn</option>{wallets.map(wallet => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</BaseSelect></div></FormField>
    <Divider />
    <FormField label="Ví nhận"><div className="relative"><ArrowDownToLine aria-hidden size={20} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-secondary" /><BaseSelect aria-label="Ví nhận" disabled={disabled} value={state.destinationWalletID} onChange={event => change({ destinationWalletID: event.target.value })} className="pl-10"><option value="">Chọn ví nhận</option>{wallets.map(wallet => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</BaseSelect></div></FormField>
    <Divider />
    <AmountField value={state.amount} disabled={disabled} onChange={amount => change({ amount })} />
    <Divider />
    <div className="flex items-center gap-4 py-3"><AlignLeft size={22} aria-hidden className="shrink-0" /><BaseTextInput aria-label="Ghi chú chuyển ví" placeholder="Ghi chú (tùy chọn)" variant="inline" disabled={disabled} value={state.note} onChange={event => change({ note: event.target.value })} /></div>
    <Divider />
    <DateField value={state.occurredAt} disabled={disabled} onChange={day => change({ occurredAt: day.includes("T") ? day : day + "T00:00" })} />
    <div className="mt-3 flex items-center gap-2 text-secondary"><WalletCards size={18} aria-hidden /><span className="text-sm">Chuyển ví không tính vào báo cáo và không gắn hũ.</span></div>
  </SurfaceCard>;
}
