import { useEffect, useState } from "react";
import { BaseBarChart, BaseDonutChart, BaseTrendChart } from "../molecules/BaseCharts";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { SectionTitle } from "../atoms/SectionTitle";
import { BaseButton } from "../atoms/BaseButton";
import { StatusMessage } from "../atoms/StatusMessage";
import { fetchAdvisorOverview } from "../../services/aiAdvisor";
import { useAccountTimezone } from "../../services/AccountTimezoneContext";
import { accountMonthKey } from "../../services/accountTime";
import { formatVND } from "../utils/format";
import { fetchMonthSummary, type MonthCategoryTotal } from "../../services/months";
import { fetchMoneyInsider, type InsiderSummary } from "../../services/insider";
import { fetchTransactions, type Transaction } from "../../services/transactions";
import { buildReportTrend } from "../../services/reportTrend";

type Summary = { income: number; expense: number; net: number; count: number };

function readSummary(value: unknown): Summary | null {
  if (!value || typeof value !== "object") return null;
  const row = value as Record<string, unknown>;
  if (!["income", "expense", "net", "count"].every(key => typeof row[key] === "number")) return null;
  return { income: row.income as number, expense: row.expense as number, net: row.net as number, count: row.count as number };
}

export function ReportsPanel({ masked }: { masked: boolean }) {
  const timezone = useAccountTimezone();
  const month = accountMonthKey(new Date(), timezone);
  const [summary, setSummary] = useState<Summary | null>(null);
  const [categories, setCategories] = useState<MonthCategoryTotal[]>([]);
  const [insider, setInsider] = useState<InsiderSummary | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [retry, setRetry] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(false);
    Promise.allSettled([fetchAdvisorOverview(month), fetchMonthSummary(month), fetchMoneyInsider(month), fetchTransactions()]).then(([overviewResult, monthResult, insiderResult, transactionResult]) => {
      const next = overviewResult.status === "fulfilled" ? readSummary(overviewResult.value.view) : null;
      if (!cancelled) {
        setSummary(next);
        setCategories(monthResult.status === "fulfilled" ? monthResult.value.categories : []);
        setInsider(insiderResult.status === "fulfilled" ? insiderResult.value : null);
        setTransactions(transactionResult.status === "fulfilled" ? transactionResult.value : []);
        setError(!next);
      }
    }).finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [month, retry]);

  if (loading) return <StatusMessage>Đang tải báo cáo tháng...</StatusMessage>;
  if (error || !summary) return <div className="space-y-2"><StatusMessage tone="danger">Không thể tải báo cáo từ dữ liệu thật.</StatusMessage><BaseButton variant="ghost" onClick={() => setRetry(value => value + 1)}>Thử lại</BaseButton></div>;
  const max = Math.max(summary.income, summary.expense, 1);
  const bars = [Math.round(summary.income / max * 100), Math.round(summary.expense / max * 100)];
  return <>
    <SurfaceCard padding="md">
      <SectionTitle title="Tổng quan báo cáo" action={month} />
      <BaseBarChart values={bars} dangerFrom={1} label="Thu và chi trong tháng" />
      <div className="mt-4 grid grid-cols-3 gap-2 text-center">
        <Metric label="Chi tiêu" amount={summary.expense} tone="danger" masked={masked} />
        <Metric label="Thu nhập" amount={summary.income} tone="action" masked={masked} />
        <Metric label="Chênh lệch" amount={summary.net} tone="ink" masked={masked} />
      </div>
    </SurfaceCard>
    <SurfaceCard padding="md">
      <SectionTitle title="Xu hướng chi tiêu" action={month} />
      <BaseTrendChart points={buildReportTrend(transactions, month, timezone)} />
    </SurfaceCard>
    <SurfaceCard padding="md">
      <SectionTitle title="Theo nhóm" />
      {categories.length === 0 ? <StatusMessage variant="plain">Chưa có giao dịch được tính vào báo cáo.</StatusMessage> : <div className="mt-3 space-y-3">{categories.map(category => <div key={`${category.type}:${category.category_id ?? category.name}`} className="flex items-center justify-between gap-3"><div className="min-w-0"><Text>{category.name}</Text><Text size="xs" tone="secondary">{category.type === "income" ? "Thu" : "Chi"} · {category.count} giao dịch</Text></div><Text numeric weight="semibold" tone={category.type === "expense" ? "danger" : "action"}>{masked ? "••••••" : formatVND(category.amount)}</Text></div>)}</div>}
    </SurfaceCard>
    <SurfaceCard padding="md">
      <SectionTitle title="Hoạt động trong tháng" />
      <Text size="sm" tone="secondary">{summary.count} giao dịch được tính vào báo cáo. Chuyển ví và điều chỉnh số dư không làm tăng thu/chi.</Text>
      <Text size="xs" tone="secondary" className="mt-3">Số liệu được tính lại từ ledger theo múi giờ tài khoản; giao dịch chuyển ví và điều chỉnh số dư luôn bị loại khỏi tổng.</Text>
    </SurfaceCard>
    {insider ? <InsiderCard value={insider} masked={masked} /> : null}
  </>;
}

function InsiderCard({ value, masked }: { value: InsiderSummary; masked: boolean }) {
  const change = value.expense_change_bps == null ? "chưa có kỳ gốc" : `${value.expense_change_bps >= 0 ? "+" : ""}${(value.expense_change_bps / 100).toFixed(1)}%`;
  const ratio = value.spending_income_ratio_bps == null ? "Chưa có thu nhập để tính tỷ lệ" : `Chi/thu ${(value.spending_income_ratio_bps / 100).toFixed(1)}%`;
  const changePercentage = value.expense_change_bps == null ? null : value.expense_change_bps / 100;
  const changeColor = value.expense_change_bps == null ? "var(--color-line)" : value.expense_change_bps > 0 ? "var(--color-danger)" : value.expense_change_bps < 0 ? "var(--color-action)" : "var(--color-secondary)";
  return <SurfaceCard padding="md"><SectionTitle title="Money Insider" action={value.estimated ? "Ước tính" : "Từ ledger thật"} /><Text size="sm" tone="secondary" className="mt-2">{value.expense_delta > 0 ? `Chi tiêu tăng ${change} so với kỳ trước.` : value.expense_delta < 0 ? `Chi tiêu giảm ${change.replace("+", "")} so với kỳ trước.` : "Chi tiêu không đổi so với kỳ trước."} Trung bình {masked ? "••••••" : formatVND(value.average_daily_expense)}/ngày trong {value.days_considered} ngày.</Text><div className="mt-3 flex items-start justify-between gap-4"><div className="min-w-0 flex-1"><Text size="xs" tone="secondary">So với kỳ trước</Text><Text size="sm" tone="secondary" className="mt-1">{ratio}</Text></div>{changePercentage == null ? null : <BaseDonutChart shares={[{ name: "So với kỳ trước", value: Math.min(100, Math.abs(changePercentage)), color: changeColor }]} center={`${changePercentage.toFixed(1)}%`} />}</div><div className="mt-3 space-y-2"><Text weight="semibold" size="sm">Top danh mục</Text>{value.top_categories.length === 0 ? <Text size="sm" tone="secondary">Chưa có khoản chi.</Text> : value.top_categories.map(category => <div key={`${category.category_id ?? "none"}:${category.name}`} className="flex items-center justify-between gap-3"><Text size="sm">{category.name}</Text><Text numeric size="sm" weight="semibold">{masked ? "••••••" : formatVND(category.amount)}</Text></div>)}</div><div className="mt-3 space-y-2"><Text weight="semibold" size="sm">Top khoản chi</Text>{value.top_expenses.length === 0 ? <Text size="sm" tone="secondary">Chưa có khoản chi.</Text> : value.top_expenses.map(row => <div key={row.id} className="flex items-center justify-between gap-3"><div className="min-w-0"><Text size="sm" className="truncate">{row.category_name || "Khoản chi"}</Text><Text size="xs" tone="secondary" className="truncate">{row.wallet_name || "Ví"}</Text></div><Text numeric size="sm" weight="semibold">{masked ? "••••••" : formatVND(row.amount)}</Text></div>)}</div></SurfaceCard>;
}

function Metric({ label, amount, tone, masked }: { label: string; amount: number; tone: "danger" | "action" | "ink"; masked: boolean }) {
  return <SurfaceCard padding="sm" tone="muted" elevation="flat"><Text size="xs" tone="secondary">{label}</Text><Text numeric size="sm" weight="bold" tone={tone} className="mt-1">{masked ? "••••••" : formatVND(amount)}</Text></SurfaceCard>;
}
