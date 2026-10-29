import { useState } from "react";
import { Calculator, Delete } from "lucide-react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseTextInput } from "../atoms/FormField";
import { BaseModal } from "../atoms/BaseModal";
import { Text } from "../atoms/Text";
import { StatusMessage } from "../atoms/StatusMessage";
import { calculateAmount } from "../../services/amountCalculator";
const KEYS = ["C", "÷", "×", "⌫", "7", "8", "9", "−", "4", "5", "6", "+", "1", "2", "3", "=", "0", "000"];
export function AmountField({ value, onChange, disabled, label = "Số tiền" }: { value: string; onChange: (value: string) => void; disabled?: boolean; label?: string }) {
  const [open, setOpen] = useState(false), [expression, setExpression] = useState(""), [error, setError] = useState("");
  const finish = () => { const result = calculateAmount(expression); if (result === null) { setError("Nhập phép tính hợp lệ, kết quả nguyên đồng và không âm."); return; } onChange(result); setOpen(false); };
  const press = (key: string) => {
    setError("");
    if (key === "=") { const result = calculateAmount(expression); if (result === null) setError("Phép tính chưa hợp lệ."); else setExpression(result); return; }
    setExpression(current => key === "C" ? "" : key === "⌫" ? current.slice(0, -1) : current.length < 80 ? current + key : current);
  };
  return <div className="space-y-1 py-3">
    <Text size="xs" tone="secondary">{label}</Text>
    <div className="flex items-center gap-3"><Text as="span" tone="secondary" weight="medium">VND</Text>
      <BaseTextInput variant="amount" aria-label={label} inputMode="numeric" placeholder="0" disabled={disabled} value={value ? Number(value).toLocaleString("vi-VN") : ""} onChange={e => onChange(e.target.value.replace(/\D/g, ""))} />
      <BaseButton variant="ghost" size="sm" aria-label="Mở bàn phím số" disabled={disabled} onClick={() => { setExpression(value); setError(""); setOpen(true); }}><Calculator size={20} /></BaseButton>
    </div>
    {open ? <BaseModal label="Bàn phím số" onClose={() => setOpen(false)}><div className="space-y-3"><Text size="2xl" numeric className="break-all">{expression || "0"}</Text>{error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}<div className="grid grid-cols-4 gap-2">{KEYS.map(key => <BaseButton key={key} variant="key" size="lg" aria-label={key === "⌫" ? "Xóa số cuối" : key} onClick={() => press(key)}>{key === "⌫" ? <Delete size={20} /> : key}</BaseButton>)}<BaseButton className="col-span-2" onClick={finish}>Xong</BaseButton></div></div></BaseModal> : null}
  </div>;
}
