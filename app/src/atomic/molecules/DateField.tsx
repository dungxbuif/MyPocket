import { useState } from "react";
import { CalendarDays, ChevronLeft, ChevronRight } from "lucide-react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseCalendar } from "../atoms/BaseCalendar";
import { BaseModal } from "../atoms/BaseModal";
import { Text } from "../atoms/Text";
import { calendarLabel, shiftDay } from "../../services/formDates";
export function DateField({ value, onChange, disabled, label = "Ngày giao dịch", stepper = true }: { value: string; onChange: (value: string) => void; disabled?: boolean; label?: string; stepper?: boolean }) {
  const [open, setOpen] = useState(false);
  return <>
    <div className="flex items-center gap-2 py-3">
      <CalendarDays size={22} aria-hidden className="shrink-0" />
      {stepper ? <BaseButton variant="chip" size="sm" disabled={disabled || !value} aria-label="Ngày trước" onClick={() => onChange(shiftDay(value, -1))}><ChevronLeft size={18} /></BaseButton> : null}
      <BaseButton variant={stepper ? "secondary" : "row"} className="min-w-0 flex-1" disabled={disabled} aria-label={label} onClick={() => setOpen(true)}>
        <span><Text as="span" size="xs" tone="secondary" className="block">{label}</Text><Text as="span" size="sm" tone={stepper ? "action" : "ink"}>{calendarLabel(value)}</Text></span>
      </BaseButton>
      {stepper ? <BaseButton variant="chip" size="sm" disabled={disabled || !value} aria-label="Ngày sau" onClick={() => onChange(shiftDay(value, 1))}><ChevronRight size={18} /></BaseButton> : null}
    </div>
    {open ? <BaseModal label="Chọn ngày" onClose={() => setOpen(false)}><BaseCalendar value={value} onCancel={() => setOpen(false)} onSelect={day => { onChange(day + value.slice(10)); setOpen(false); }} /></BaseModal> : null}
  </>;
}
