import * as React from "react";
import { cn } from "../../lib/utils";
import { Calendar, ChevronLeft, ChevronRight } from "lucide-react";

export interface DateNavigationRowProps {
  currentDate: string; // YYYY-MM-DD
  onDateChange: (dateStr: string) => void;
  onOpenCalendarModal?: () => void;
  className?: string;
}

export function DateNavigationRow({
  currentDate,
  onDateChange,
  className,
}: DateNavigationRowProps) {
  // Parse YYYY-MM-DD
  const dateObj = new Date(currentDate + "T00:00:00");
  const isValid = !isNaN(dateObj.getTime());

  function shiftDay(days: number) {
    const base = isValid ? new Date(dateObj) : new Date();
    base.setDate(base.getDate() + days);
    const y = base.getFullYear();
    const m = String(base.getMonth() + 1).padStart(2, "0");
    const d = String(base.getDate()).padStart(2, "0");
    onDateChange(`${y}-${m}-${d}`);
  }

  // Format localized date: "Chủ Nhật, 23/08/2026"
  const weekdayNames = ["Chủ Nhật", "Thứ Hai", "Thứ Ba", "Thứ Tư", "Thứ Năm", "Thứ Sáu", "Thứ Bảy"];
  const weekday = isValid ? weekdayNames[dateObj.getDay()] : "";
  const dayStr = isValid ? String(dateObj.getDate()).padStart(2, "0") : "";
  const monthStr = isValid ? String(dateObj.getMonth() + 1).padStart(2, "0") : "";
  const yearStr = isValid ? String(dateObj.getFullYear()) : "";
  const formattedLabel = isValid ? `${weekday}, ${dayStr}/${monthStr}/${yearStr}` : currentDate;

  return (
    <div className={cn("flex items-center justify-between py-3 px-2 w-full", className)}>
      <div className="flex items-center gap-2">
        <Calendar className="w-5 h-5 text-[#333333]" />
        <span className="text-sm font-semibold text-[#111111]">Ngày</span>
      </div>

      <div className="flex items-center gap-1.5">
        <button
          type="button"
          onClick={() => shiftDay(-1)}
          className="w-7 h-7 rounded-full bg-[#eef0f4] hover:bg-[#e2e4e9] text-[#111111] flex items-center justify-center active:scale-90 transition-all outline-none"
        >
          <ChevronLeft className="w-4 h-4" />
        </button>

        <span className="px-3 py-1 rounded-full bg-[#eef0f4] text-xs font-semibold text-[#111111] select-none">
          {formattedLabel}
        </span>

        <button
          type="button"
          onClick={() => shiftDay(1)}
          className="w-7 h-7 rounded-full bg-[#eef0f4] hover:bg-[#e2e4e9] text-[#111111] flex items-center justify-center active:scale-90 transition-all outline-none"
        >
          <ChevronRight className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}
