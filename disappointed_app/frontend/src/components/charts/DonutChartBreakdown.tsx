import * as React from "react";
import { cn } from "../../lib/utils";

export interface DonutSlice {
  key: string;
  label: string;
  percentage: number;
  color: string;
}

export interface DonutChartBreakdownProps {
  slices: DonutSlice[];
  centerIcon?: React.ReactNode;
  centerBadgeText?: string;
  size?: number;
  className?: string;
}

export function DonutChartBreakdown({
  slices,
  centerIcon,
  centerBadgeText = "100%",
  size = 180,
  className,
}: DonutChartBreakdownProps) {
  const radius = 65;
  const strokeWidth = 16;
  const circumference = 2 * Math.PI * radius;

  let currentOffset = 0;

  return (
    <div className={cn("relative flex items-center justify-center select-none py-2", className)}>
      <svg className="w-44 h-44 -rotate-90" viewBox="0 0 160 160">
        {/* Background Circle */}
        <circle
          cx="80"
          cy="80"
          r={radius}
          fill="none"
          stroke="#e9eaef"
          strokeWidth={strokeWidth}
        />

        {/* Slices */}
        {slices.map((slice) => {
          const sliceStroke = (slice.percentage / 100) * circumference;
          const strokeDashoffset = -currentOffset;
          currentOffset += sliceStroke;

          return (
            <circle
              key={slice.key}
              cx="80"
              cy="80"
              r={radius}
              fill="none"
              stroke={slice.color}
              strokeWidth={strokeWidth}
              strokeDasharray={`${sliceStroke} ${circumference}`}
              strokeDashoffset={strokeDashoffset}
              className="transition-all duration-300"
            />
          );
        })}
      </svg>

      {/* Center Icon & Badge */}
      <div className="absolute flex flex-col items-center justify-center">
        {centerIcon && (
          <div className="w-10 h-10 rounded-full bg-[#333333] text-white flex items-center justify-center font-bold text-sm shadow-xs mb-1">
            {centerIcon}
          </div>
        )}
        <span className="px-2 py-0.5 rounded-full bg-[#eef0f4] text-xs font-bold text-[#111111]">
          {centerBadgeText}
        </span>
      </div>
    </div>
  );
}
