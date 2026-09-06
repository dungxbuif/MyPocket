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
    <div className={cn("w-full flex flex-col select-none py-2", className)}>
      <div className="flex items-end justify-around gap-6 px-4" style={{ height }}>
        {data.map((item, idx) => {
          const heightPercent = Math.max(12, Math.round((item.amount / maxAmount) * 100));

          return (
            <div key={idx} className="flex-1 flex flex-col items-center gap-1.5 h-full justify-end max-w-[90px]">
              <span className="text-[11px] font-bold text-[#111111] tabular-nums truncate">
                {item.amountFormatted}
              </span>
              <div
                className={cn(
                  "w-full rounded-t-xl transition-all duration-300",
                  item.isCurrent ? "bg-[#ff5a66]" : "bg-[#ffd2d6]"
                )}
                style={{ height: `${heightPercent}%` }}
              />
            </div>
          );
        })}
      </div>

      {/* Baseline Divider & Labels */}
      <div className="w-full border-t border-[#e8e8ec] mt-1 pt-1.5 flex justify-around px-4">
        {data.map((item, idx) => (
          <span
            key={idx}
            className={cn(
              "text-xs font-semibold text-center flex-1 max-w-[90px]",
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
