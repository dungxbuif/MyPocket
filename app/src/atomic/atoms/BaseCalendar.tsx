import { useRef, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { BaseButton } from "./BaseButton";
import { BaseSelect, BaseTextInput, FormField } from "./FormField";
import { Text } from "./Text";
import { calendarLabel, dateKey, monthDays, parseDay, shiftDay } from "../../services/formDates";
const WEEKDAYS = ["T2", "T3", "T4", "T5", "T6", "T7", "CN"];
export function BaseCalendar({ value, onSelect, onCancel }: { value: string; onSelect: (value: string) => void; onCancel: () => void }) {
  const initial = value ? parseDay(value) : new Date();
  const [month, setMonth] = useState(new Date(initial.getFullYear(), initial.getMonth(), 1, 12));
  const [focused, setFocused] = useState(value.slice(0, 10) || dateKey(initial));
  const [jump, setJump] = useState(false);
  const today = dateKey(new Date());
  const grid = useRef<HTMLDivElement>(null);
  const move = (days: number) => {
    const next = shiftDay(focused, days), date = parseDay(next);
    setFocused(next); setMonth(new Date(date.getFullYear(), date.getMonth(), 1, 12));
    requestAnimationFrame(() => grid.current?.querySelector<HTMLButtonElement>(`[data-day="${next}"]`)?.focus());
  };
  const moveMonth = (step: number, focus = false) => {
    const next = new Date(month.getFullYear(), month.getMonth() + step, 1, 12);
    setMonth(next); setFocused(dateKey(next));
    if (focus) requestAnimationFrame(() => grid.current?.querySelector<HTMLButtonElement>(`[data-day="${dateKey(next)}"]`)?.focus());
  };
  return <div className="space-y-3">
    <div className="flex items-center justify-between gap-1">
      <BaseButton variant="chip" size="sm" aria-label="Tháng trước" onClick={() => moveMonth(-1)}><ChevronLeft size={18} /></BaseButton>
      <BaseButton variant="ghost" onClick={() => setJump(!jump)}>{new Intl.DateTimeFormat("vi-VN", { month: "short", year: "numeric" }).format(month)}</BaseButton>
      <BaseButton variant="chip" size="sm" aria-label="Tháng sau" onClick={() => moveMonth(1)}><ChevronRight size={18} /></BaseButton>
    </div>
    {jump ? <div className="grid grid-cols-2 gap-2">
      <FormField label="Tháng"><BaseSelect value={month.getMonth()} onChange={e => { const next = new Date(month.getFullYear(), Number(e.target.value), 1, 12); setMonth(next); setFocused(dateKey(next)); }}>{Array.from({ length: 12 }, (_, i) => <option key={i} value={i}>{i + 1}</option>)}</BaseSelect></FormField>
      <FormField label="Năm"><BaseTextInput type="number" min="1900" max="2100" value={month.getFullYear()} onChange={e => { const year = Number(e.target.value); if (year >= 1900 && year <= 2100) { const next = new Date(year, month.getMonth(), 1, 12); setMonth(next); setFocused(dateKey(next)); } }} /></FormField>
    </div> : null}
    <div className="grid grid-cols-7">{WEEKDAYS.map(day => <Text key={day} size="xs" className="py-2 text-center">{day}</Text>)}</div>
    <div ref={grid} role="grid" aria-label="Lịch chọn ngày" className="grid grid-cols-7" onKeyDown={e => {
      const offset = ({ ArrowLeft: -1, ArrowRight: 1, ArrowUp: -7, ArrowDown: 7 } as Record<string, number>)[e.key];
      if (offset) { e.preventDefault(); move(offset); }
      else if (e.key === "PageUp" || e.key === "PageDown") { e.preventDefault(); moveMonth(e.key === "PageUp" ? -1 : 1, true); }
      else if (e.key === "Home" || e.key === "End") { e.preventDefault(); const weekday = (parseDay(focused).getDay() + 6) % 7; move(e.key === "Home" ? -weekday : 6 - weekday); }
    }}>
      {monthDays(month.getFullYear(), month.getMonth()).map(day => <button key={day} type="button" role="gridcell" data-day={day} aria-label={calendarLabel(day)} aria-selected={value.slice(0, 10) === day} tabIndex={focused === day ? 0 : -1} onFocus={() => setFocused(day)} onClick={() => onSelect(day)} className={`min-h-11 min-w-0 rounded-full text-sm focus-visible:outline-2 focus-visible:outline-action ${day === value.slice(0, 10) ? "bg-action font-bold text-card" : parseDay(day).getMonth() === month.getMonth() ? "text-ink hover:bg-row" : "text-muted hover:bg-row"}`}>{parseDay(day).getDate()}</button>)}
    </div>
    <div className="flex items-center justify-between gap-2">
      <BaseButton variant="ghost" onClick={() => { const next = parseDay(today); setFocused(today); setMonth(new Date(next.getFullYear(), next.getMonth(), 1, 12)); onSelect(today); }}>Hôm nay</BaseButton>
      <BaseButton variant="ghost" onClick={onCancel}>Hủy</BaseButton>
    </div>
  </div>;
}
