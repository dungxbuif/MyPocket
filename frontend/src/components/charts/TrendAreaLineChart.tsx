import * as React from "react";
import { cn } from "../../lib/utils";

export interface TrendPoint {
  dateLabel: string;
  actualAmount: number;
  baselineAmount: number;
}

export interface TrendAreaLineChartProps {
  points: TrendPoint[];
  height?: number;
  className?: string;
}

export function TrendAreaLineChart({ points, height = 140, className }: TrendAreaLineChartProps) {
  if (!points || points.length === 0) return null;

  const maxVal = Math.max(
    ...points.map((p) => Math.max(p.actualAmount, p.baselineAmount)),
    1000
  );

  const width = 300;
  const paddingX = 20;
  const paddingY = 15;
  const chartW = width - paddingX * 2;
  const chartH = height - paddingY * 2;

  // Generate SVG path for actual points
  const coords = points.map((p, idx) => {
    const x = paddingX + (idx / Math.max(1, points.length - 1)) * chartW;
    const y = height - paddingY - (p.actualAmount / maxVal) * chartH;
    return { x, y, ...p };
  });

  const pathD = coords.reduce((acc, c, idx) => {
    return idx === 0 ? `M ${c.x} ${c.y}` : `${acc} L ${c.x} ${c.y}`;
  }, "");

  // Baseline curve
  const baselineCoords = points.map((p, idx) => {
    const x = paddingX + (idx / Math.max(1, points.length - 1)) * chartW;
    const y = height - paddingY - (p.baselineAmount / maxVal) * chartH;
    return { x, y };
  });

  const baselineD = baselineCoords.reduce((acc, c, idx) => {
    return idx === 0 ? `M ${c.x} ${c.y}` : `${acc} L ${c.x} ${c.y}`;
  }, "");

  return (
    <div className={cn("w-full flex flex-col items-center select-none py-2", className)}>
      <svg className="w-full h-auto max-h-[140px]" viewBox={`0 0 ${width} ${height}`}>
        {/* Baseline (3-month historical average) */}
        <path
          d={baselineD}
          fill="none"
          stroke="#d1d1d6"
          strokeWidth="2"
          strokeDasharray="4 4"
        />

        {/* Actual Spend Cumulative Curve */}
        <path
          d={pathD}
          fill="none"
          stroke="#ff5a66"
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
        />

        {/* Milestone Nodes */}
        {coords.map((c, idx) => (
          <circle
            key={idx}
            cx={c.x}
            cy={c.y}
            r="4"
            fill="#ffffff"
            stroke="#ff5a66"
            strokeWidth="2.5"
          />
        ))}
      </svg>

      {/* Legend & Date Labels */}
      <div className="w-full flex justify-between px-4 text-[11px] text-[#8e8e93] mt-1 border-t border-[#e8e8ec] pt-1">
        <span>{points[0]?.dateLabel}</span>
        <div className="flex items-center gap-3">
          <span className="flex items-center gap-1">
            <span className="w-2 h-0.5 bg-[#ff5a66] inline-block" /> Tháng này
          </span>
          <span className="flex items-center gap-1">
            <span className="w-2 h-0.5 border-t border-dashed border-[#8e8e93] inline-block" /> Trung bình 3T
          </span>
        </div>
        <span>{points[points.length - 1]?.dateLabel}</span>
      </div>
    </div>
  );
}
