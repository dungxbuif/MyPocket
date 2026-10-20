import * as React from "react";
import { cn } from "../../lib/utils";

export interface BarItem {
  label: string;
  amount: number;
  amountFormatted: string;
  isCurrent?: boolean;
}

export interface ComparisonBarChartProps {
  data: BarItem[];
  height?: number;
  className?: string;
}

export function ComparisonBarChart({ data, height = 150, className }: ComparisonBarChartProps) {
  const maxAmount = Math.max(...data.map((d) => d.amount), 1);

  return (
    <div className={cn("w-full min-w-0 flex flex-col select-none py-2", className)}>
      <div className="grid items-end gap-2 px-2" style={{ height, gridTemplateColumns: `repeat(${Math.max(data.length, 1)}, minmax(0, 1fr))` }}>
        {data.map((item, idx) => {
          const heightPercent = Math.max(12, Math.round((item.amount / maxAmount) * 100));

          return (
            <div key={idx} className="min-w-0 w-full flex flex-col items-center gap-1.5 h-full justify-end">
              <span title={item.amountFormatted} className="max-w-full text-[11px] font-bold text-[#111111] tabular-nums truncate">
                {item.amountFormatted}
              </span>
              <div
                className={cn(
                  "w-full rounded-t-xl transition-all duration-300",
                  item.isCurrent ? "bg-[#9f1d1d]" : "bg-[#ffd2d6]"
                )}
                style={{ height: `${heightPercent}%` }}
              />
            </div>
          );
        })}
      </div>

      {/* Baseline Divider & Labels */}
      <div className="w-full min-w-0 border-t border-[#e8e8ec] mt-1 pt-1.5 grid gap-2 px-2" style={{ gridTemplateColumns: `repeat(${Math.max(data.length, 1)}, minmax(0, 1fr))` }}>
        {data.map((item, idx) => (
          <span
            key={idx}
            title={item.label}
            className={cn(
              "min-w-0 truncate text-xs font-semibold text-center",
              item.isCurrent ? "text-[#111111] font-bold" : "text-[#8e8e93]"
            )}
          >
            {item.label}
          </span>
        ))}
      </div>
    </div>
  );
}
