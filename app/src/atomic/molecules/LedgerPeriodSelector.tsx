import { ChevronLeft, ChevronRight } from "lucide-react";
import { BaseButton } from "../atoms/BaseButton";
import { SegmentedControl } from "../atoms/SegmentedControl";
import { Text } from "../atoms/Text";

export type LedgerPeriodMode = "week" | "custom";

function displayDate(key: string): string {
  if (!key) return "";
  const [year, month, day] = key.split("-");
  return `${day}/${month}/${year}`;
}

export function LedgerPeriodSelector({ mode, range, weekOffset, onModeChange, onPreviousWeek, onNextWeek, onEditCustom }: {
  mode: LedgerPeriodMode;
  range: { start: string; end: string } | null;
  weekOffset: number;
  onModeChange: (mode: LedgerPeriodMode) => void;
  onPreviousWeek: () => void;
  onNextWeek: () => void;
  onEditCustom: () => void;
}) {
  const options: ReadonlyArray<{ value: LedgerPeriodMode; label: string }> = [
    { value: "week", label: "Theo tuần" },
    { value: "custom", label: "Tùy chọn" },
  ];
  return <div className="space-y-2">
    <SegmentedControl value={mode} options={options} onChange={onModeChange} />
    {mode === "week" && range ? <div className="flex items-center justify-between gap-2">
      <BaseButton variant="chip" size="sm" aria-label="Tuần trước" onClick={onPreviousWeek}><ChevronLeft size={18} /></BaseButton>
      <Text size="sm" weight="bold" className="text-center" aria-live="polite">{displayDate(range.start)} – {displayDate(range.end)}</Text>
      <BaseButton variant="chip" size="sm" aria-label="Tuần sau" onClick={onNextWeek} disabled={weekOffset >= 0}><ChevronRight size={18} /></BaseButton>
    </div> : null}
    {mode === "custom" ? <div className="flex items-center justify-between gap-2">
      <Text size="sm" tone="secondary">{range?.start ? displayDate(range.start) : "Từ đầu"} – {range?.end ? displayDate(range.end) : "Hiện tại"}</Text>
      <BaseButton variant="chip" size="sm" onClick={onEditCustom}>Đổi ngày</BaseButton>
    </div> : null}
  </div>;
}
