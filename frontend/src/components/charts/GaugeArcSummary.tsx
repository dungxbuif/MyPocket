import * as React from "react";
import { cn } from "../../lib/utils";

export interface GaugeArcSummaryProps {
  availableAmount: number;
  totalBudget: number;
  totalSpent: number;
  daysRemaining: number;
  currency?: string;
  className?: string;
}

export function GaugeArcSummary({
  availableAmount,
  totalBudget,
  totalSpent,
  daysRemaining,
  className,
}: GaugeArcSummaryProps) {
  // SVG Arc calculation for 200° gauge
  const radius = 70;
  const strokeWidth = 12;
  const circumference = 2 * Math.PI * radius;
  const arcLength = (circumference * 200) / 360; // 200 degrees of the circle
  const spentRatio = totalBudget > 0 ? Math.max(0, totalSpent / totalBudget) : 0;
  const fillLength = arcLength * (1 - Math.min(1, spentRatio));

  const overspent = availableAmount < 0;
  const formattedAvailable = new Intl.NumberFormat("vi-VN").format(Math.abs(availableAmount));
  const formattedBudgetM = `${(totalBudget / 1_000_000).toFixed(totalBudget % 1_000_000 === 0 ? 0 : 2)} M đ`;
  const formattedSpentM = `${(totalSpent / 1_000_000).toFixed(totalSpent % 1_000_000 === 0 ? 0 : 2)} M đ`;

  return (
    <div className={cn("w-full flex flex-col items-center select-none py-2", className)}>
      <div className="relative w-48 h-32 flex items-center justify-center">
        <svg className="w-48 h-48 -rotate-[190deg]" viewBox="0 0 160 160">
          {/* Background Track */}
          <circle
            cx="80"
            cy="80"
            r={radius}
            fill="none"
            stroke="#e9eaef"
            strokeWidth={strokeWidth}
            strokeDasharray={`${arcLength} ${circumference}`}
            strokeLinecap="round"
          />
          {/* Active Fill Track */}
          <circle
            cx="80"
            cy="80"
            r={radius}
            fill="none"
            stroke={spentRatio > 1 ? "#ff5a66" : spentRatio > 0.8 ? "#ff8800" : "#2dbd4f"}
            strokeWidth={strokeWidth}
            strokeDasharray={`${fillLength} ${circumference}`}
            strokeLinecap="round"
            className="transition-all duration-500 ease-out"
          />
        </svg>

        {/* Center Text */}
        <div className="absolute top-12 flex flex-col items-center text-center">
          <span className="text-xs font-semibold text-[#8e8e93]">
            {overspent ? "Vượt ngân sách" : "Số tiền bạn có thể chi"}
          </span>
          <span className="text-xl sm:text-2xl font-bold text-[#111111] tabular-nums mt-0.5">
            {formattedAvailable} đ
          </span>
        </div>
      </div>

      {/* 3 Metrics Footer */}
      <div className="w-full grid grid-cols-3 divide-x divide-[#e8e8ec] mt-2 pt-3 border-t border-[#e8e8ec]/80 text-center">
        <div className="flex flex-col px-1">
          <span className="text-xs font-bold text-[#111111] tabular-nums">{formattedBudgetM}</span>
          <span className="text-[11px] text-[#8e8e93] mt-0.5">Tổng ngân sách</span>
        </div>
        <div className="flex flex-col px-1">
          <span className="text-xs font-bold text-[#111111] tabular-nums">{formattedSpentM}</span>
          <span className="text-[11px] text-[#8e8e93] mt-0.5">Tổng đã chi</span>
        </div>
        <div className="flex flex-col px-1">
          <span className="text-xs font-bold text-[#111111] tabular-nums">{daysRemaining} ngày</span>
          <span className="text-[11px] text-[#8e8e93] mt-0.5">Còn lại trong kỳ</span>
        </div>
      </div>
    </div>
  );
}
