import { useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
import { DateField } from "../molecules/DateField";
import { AmountField } from "../molecules/AmountField";
import { SegmentedControl } from "../atoms/SegmentedControl";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Divider } from "../atoms/Divider";
import type { Wallet } from "../../services/wallets";
import { createAdjustment, type AdjustmentInput } from "../../services/transactions";
import { instantFromLocalDateTime, localDateTimeAt } from "../../services/accountTime";
import { useAccountTimezone } from "../../services/AccountTimezoneContext";

type Direction = "increase" | "decrease";

export function WalletAdjustmentSheet({ wallets, onClose, onSaved }: { wallets: Wallet[]; onClose: () => void; onSaved: () => void }) {
  const timezone = useAccountTimezone();
  const ledgerWallets = wallets.filter(wallet => wallet.type !== "credit");
  const [walletID, setWalletID] = useState(ledgerWallets[0]?.id ?? "");
  const [direction, setDirection] = useState<Direction>("increase");
  const [amount, setAmount] = useState("");
  const [occurredAt, setOccurredAt] = useState(localDateTimeAt(new Date(), timezone));
  const [note, setNote] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const submit = async () => {
    const value = Number(amount);
    if (!walletID || !Number.isSafeInteger(value) || value <= 0) {
      setError("Hãy chọn ví và nhập số tiền nguyên lớn hơn 0.");
      return;
    }
    let occurred: string;
    try { occurred = instantFromLocalDateTime(occurredAt, timezone); } catch { setError("Thời điểm điều chỉnh không hợp lệ."); return; }
    const input: AdjustmentInput = { wallet_id: walletID, amount: value, direction, occurred_at: occurred, note: note.trim() || undefined };
    try {
      setSaving(true);
      setError("");
      await createAdjustment(input);
      onSaved();
      onClose();
    } catch { setError("Không thể lưu điều chỉnh số dư. Vui lòng thử lại."); } finally { setSaving(false); }
  };

  return <BaseBottomSheet presentation="form" title="Điều chỉnh số dư" closeLabel="Hủy" closingDisabled={saving} onClose={onClose}
    footer={<BaseButton className="w-full" size="lg" loading={saving} disabled={!walletID || !amount} onClick={() => void submit()}>Lưu điều chỉnh</BaseButton>}>
    <div className="space-y-3">
      {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
      <SurfaceCard tone="form" padding="md">
        <FormField label="Ví"><BaseSelect aria-label="Chọn ví điều chỉnh" value={walletID} disabled={saving} onChange={event => setWalletID(event.target.value)}>{ledgerWallets.map(wallet => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</BaseSelect></FormField>
        <Divider />
        <SegmentedControl value={direction} disabled={saving} options={[{ value: "increase", label: "Tăng số dư" }, { value: "decrease", label: "Giảm số dư" }]} onChange={setDirection} />
        <AmountField value={amount} label="Số tiền điều chỉnh" disabled={saving} onChange={setAmount} />
        <Divider />
        <div className="py-3"><BaseTextInput variant="inline" aria-label="Ghi chú điều chỉnh" placeholder="Ghi chú" value={note} disabled={saving} onChange={event => setNote(event.target.value)} /></div>
        <Divider />
        <DateField label="Ngày điều chỉnh" value={occurredAt} disabled={saving} onChange={value => setOccurredAt(value.includes("T") ? value : `${value}T00:00`)} />
      </SurfaceCard>
    </div>
  </BaseBottomSheet>;
}
