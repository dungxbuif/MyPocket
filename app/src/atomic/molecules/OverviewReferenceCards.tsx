import { useState } from "react";
import { BarChart3, Info, ListRestart, type LucideIcon } from "lucide-react";

import { BaseBarChart } from "./BaseCharts";
import { BaseButton } from "../atoms/BaseButton";
import { SectionTitle } from "../atoms/SectionTitle";
import { SegmentedControl } from "../atoms/SegmentedControl";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import type { InsiderSummary } from "../../services/insider";
import type { MonthSummary } from "../../services/months";
import { formatVND } from "../utils/format";

export function OverviewMoneyInsiderCard({ value, masked, onOpen }: { value: InsiderSummary; masked: boolean; onOpen: () => void }) {
  const topCategory = value.top_categories[0];
  const change = value.expense_change_bps == null ? "Chưa có kỳ gốc" : `${value.expense_change_bps >= 0 ? "+" : ""}${(value.expense_change_bps / 100).toFixed(1)}%`;
  return (
    <SurfaceCard radius="reference" padding="lg" className="overflow-hidden">
      <SectionTitle title="Money Insider" action="↻" onAction={onOpen} />
      <div className="mt-5 flex items-center gap-2">
        <Text size="lg" weight="bold" tone="heading">{topCategory?.name ?? "Chi tiêu"}</Text>
        <span className="grid h-6 w-6 place-items-center rounded-full border border-line text-secondary" aria-label="Thông tin phân tích"><Info size={15} /></span>
      </div>
      <div className="mt-3 flex items-end justify-between gap-3 border-b border-line pb-4">
        <div><Text size="sm" tone="secondary">Tổng đã chi</Text><Text numeric size="2xl" weight="extrabold" tone="danger" className="mt-1">{masked ? "••••••" : formatVND(topCategory?.amount ?? value.expense)}</Text></div>
        <Text size="sm" tone="secondary" className="max-w-[10rem] text-right">{change} so với kỳ trước</Text>
      </div>
      <div className="mt-4 grid grid-cols-2 gap-4">
        <div><Text size="sm" weight="bold" tone="heading">Tháng này</Text><Text size="sm" tone="secondary" className="mt-2">Trung bình</Text><Text numeric size="lg" weight="extrabold" tone="heading" className="mt-1">{masked ? "••••••" : `${formatVND(value.average_daily_expense)}/ngày`}</Text></div>
        <div className="flex items-end justify-end"><Text size="sm" tone="secondary" className="text-right">{value.spending_income_ratio_bps == null ? "Chưa có thu nhập để tính tỷ lệ" : `Chi/thu ${(value.spending_income_ratio_bps / 100).toFixed(1)}%`}</Text></div>
      </div>
      <div className="mt-5 grid grid-cols-2 gap-3">
        <BaseButton variant="secondary" size="md" onClick={onOpen}>Xem chi tiết</BaseButton>
        <BaseButton variant="primary" size="md" onClick={onOpen}>Mở báo cáo</BaseButton>
      </div>
    </SurfaceCard>
  );
}

export function OverviewReportCard({ summary, week, masked, onOpen }: { summary: MonthSummary; week: { income: number; expense: number }; masked: boolean; onOpen: () => void }) {
  const [period, setPeriod] = useState<"week" | "month">("month");
  const values = period === "month" ? { income: summary.income, expense: summary.expense } : week;
  const max = Math.max(values.income, values.expense, 1);
  return (
    <SurfaceCard radius="reference" padding="lg">
      <SectionTitle title="Báo cáo tháng này" action="Xem báo cáo" onAction={onOpen} />
      <div className="mt-4"><SegmentedControl value={period} options={[{ value: "week", label: "Tuần" }, { value: "month", label: "Tháng" }]} onChange={setPeriod} /></div>
      <BaseBarChart values={[Math.round(values.income / max * 100), Math.round(values.expense / max * 100)]} dangerFrom={1} label="Thu và chi" />
      <div className="mt-3 grid grid-cols-2 gap-3"><ReportMetric icon={BarChart3} label="Thu" amount={values.income} masked={masked} tone="action" /><ReportMetric icon={ListRestart} label="Chi" amount={values.expense} masked={masked} tone="danger" /></div>
    </SurfaceCard>
  );
}

function ReportMetric({ icon: Icon, label, amount, masked, tone }: { icon: LucideIcon; label: string; amount: number; masked: boolean; tone: "action" | "danger" }) {
  return <SurfaceCard padding="sm" tone="muted" elevation="flat" radius="md"><div className="flex items-center gap-2"><Icon size={15} className={tone === "action" ? "text-action" : "text-danger"} /><Text size="xs" tone="secondary">{label}</Text></div><Text numeric weight="bold" tone={tone} className="mt-1">{masked ? "••••••" : formatVND(amount)}</Text></SurfaceCard>;
}
