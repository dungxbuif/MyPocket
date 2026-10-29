import * as React from "react";
import { Sheet } from "../../components/ui/sheet";
import { ModalHeader } from "../../components/navigation/ModalHeader";
import { SegmentedControl } from "../../components/ui/segmented-control";
import { DonutChartBreakdown } from "../../components/charts/DonutChartBreakdown";
import { ComparisonBarChart } from "../../components/charts/ComparisonBarChart";

export interface ExpenseDetailSheetProps {
  isOpen: boolean;
  onClose: () => void;
  categoryName?: string;
  totalSpent?: number;
}

export function ExpenseDetailSheet({
  isOpen,
  onClose,
  categoryName = "Ăn uống",
  totalSpent = 250000,
}: ExpenseDetailSheetProps) {
  const [tab, setTab] = React.useState<"detail" | "trend">("detail");

  const dailyBars = [
    { label: "10", amount: 0, amountFormatted: "0" },
    { label: "11", amount: 0, amountFormatted: "0" },
    { label: "12", amount: 0, amountFormatted: "0" },
    { label: "13", amount: 90000, amountFormatted: "90.000 đ", isCurrent: true },
    { label: "14", amount: 160000, amountFormatted: "160.000 đ", isCurrent: true },
    { label: "15", amount: 0, amountFormatted: "0" },
    { label: "16", amount: 0, amountFormatted: "0" },
  ];

  const donutSlices = [
    { key: "food", label: categoryName, percentage: 100, color: "#111111" },
  ];

  const formattedTotal = new Intl.NumberFormat("vi-VN").format(totalSpent);
  const formattedDailyAvg = new Intl.NumberFormat("vi-VN").format(Math.round(totalSpent / 7));

  return (
    <Sheet
      isOpen={isOpen}
      onClose={onClose}
      title="Chi tiết khoản chi"
      header={
        <ModalHeader
          title="Chi tiết khoản chi"
          dismissLabel="Đóng"
          onDismiss={onClose}
        />
      }
    >
      <div className="flex flex-col gap-4 py-2">
        {/* Date / Week Range Header */}
        <div className="flex items-center justify-between text-xs font-semibold text-[#8e8e93] px-1">
          <span>03/08/2026 - 09/08/2026</span>
          <div className="flex items-center gap-2">
            <span className="text-[#111111] font-bold">TUẦN NÀY</span>
          </div>
        </div>

        {/* Aggregated Numbers */}
        <div className="rounded-2xl bg-white p-4 text-center shadow-xs flex flex-col gap-1">
          <span className="text-xs text-[#8e8e93]">Tổng cộng</span>
          <span className="text-2xl font-bold text-[#9f1d1d] tabular-nums">
            {formattedTotal} đ
          </span>
          <span className="text-xs text-[#8e8e93] mt-1">
            Trung bình hàng ngày: <strong className="text-[#111111]">{formattedDailyAvg} đ</strong>
          </span>
        </div>

        {/* Segmented Control */}
        <SegmentedControl
          options={[
            { key: "detail", label: "Chi tiết" },
            { key: "trend", label: "Xu hướng" },
          ]}
          selectedKey={tab}
          onChange={(key) => setTab(key as "detail" | "trend")}
        />

        {/* Tab 1: Chi tiết (Donut Chart) */}
        {tab === "detail" ? (
          <div className="rounded-2xl bg-white p-4 shadow-xs flex flex-col items-center">
            <DonutChartBreakdown
              slices={donutSlices}
              centerBadgeText="100%"
            />
            <div className="w-full mt-4 border-t border-[#e8e8ec] pt-3 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span className="w-3 h-3 rounded-full bg-[#111111]" />
                <span className="text-sm font-semibold text-[#111111]">{categoryName}</span>
              </div>
              <span className="text-sm font-bold text-[#111111] tabular-nums">{formattedTotal} đ</span>
            </div>
          </div>
        ) : (
          /* Tab 2: Xu hướng (Daily Spending Bars) */
          <div className="rounded-2xl bg-white p-4 shadow-xs flex flex-col">
            <ComparisonBarChart data={dailyBars} height={140} />
            <div className="mt-4 border-t border-[#e8e8ec] pt-2 flex flex-col divide-y divide-[#e8e8ec]/60">
              <div className="flex justify-between py-2 text-xs">
                <span className="text-[#8e8e93]">Thứ Năm, 13/08</span>
                <span className="font-bold text-[#9f1d1d]">90.000 đ</span>
              </div>
              <div className="flex justify-between py-2 text-xs">
                <span className="text-[#8e8e93]">Thứ Sáu, 14/08</span>
                <span className="font-bold text-[#9f1d1d]">160.000 đ</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </Sheet>
  );
}
