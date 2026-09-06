import * as React from "react";
import { cn } from "../../lib/utils";

export interface ProgressBarWithMarkerProps {
  spentPercent: number; // 0 to 100
  dayProgressPercent: number; // 0 to 100
  spentAmountFormatted: string;
  remainingAmountFormatted: string;
  color?: string;
  className?: string;
}

export function ProgressBarWithMarker({
  spentPercent,
  dayProgressPercent,
  spentAmountFormatted,
  remainingAmountFormatted,
  color = "#2dbd4f",
  className,
}: ProgressBarWithMarkerProps) {
  const safeSpent = Math.min(100, Math.max(0, spentPercent));
  const safeDay = Math.min(100, Math.max(0, dayProgressPercent));

  return (
    <div className={cn("w-full flex flex-col gap-1.5 py-1 select-none", className)}>
      {/* Top Labels */}
      <div className="flex items-center justify-between text-xs font-semibold">
        <span className="text-[#111111] tabular-nums">{spentAmountFormatted}</span>
        <span className="text-[#8e8e93] tabular-nums">{remainingAmountFormatted}</span>
      </div>

      {/* Progress Bar with "Hôm nay" Marker */}
      <div className="relative w-full h-2.5 rounded-full bg-[#e9eaef] overflow-visible my-2">
        {/* Fill Track */}
        <div
          className="h-full rounded-full transition-all duration-300"
          style={{ width: `${safeSpent}%`, backgroundColor: color }}
        />

        {/* Floating "Hôm nay" Flag Marker */}
        <div
          className="absolute -top-3.5 -bottom-2 flex flex-col items-center pointer-events-none transition-all duration-300 -translate-x-1/2"
          style={{ left: `${safeDay}%` }}
        >
          <span className="px-1.5 py-0.5 rounded-full bg-[#111111] text-[9px] font-bold text-white shadow-xs leading-none">
            Hôm nay
          </span>
          <div className="w-0.5 flex-1 bg-[#111111]/70 mt-0.5" />
        </div>
      </div>
    </div>
  );
}
