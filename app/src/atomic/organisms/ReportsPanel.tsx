import { useEffect, useState } from "react";
import { BaseBarChart } from "../molecules/BaseCharts";
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
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [retry, setRetry] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(false);
    Promise.allSettled([fetchAdvisorOverview(month), fetchMonthSummary(month)]).then(([overviewResult, monthResult]) => {
      const next = overviewResult.status === "fulfilled" ? readSummary(overviewResult.value.view) : null;
      if (!cancelled) {
        setSummary(next);
        setCategories(monthResult.status === "fulfilled" ? monthResult.value.categories : []);
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
      <SectionTitle title="Theo nhóm" />
      {categories.length === 0 ? <StatusMessage variant="plain">Chưa có giao dịch được tính vào báo cáo.</StatusMessage> : <div className="mt-3 space-y-3">{categories.map(category => <div key={`${category.type}:${category.category_id ?? category.name}`} className="flex items-center justify-between gap-3"><div className="min-w-0"><Text>{category.name}</Text><Text size="xs" tone="secondary">{category.type === "income" ? "Thu" : "Chi"} · {category.count} giao dịch</Text></div><Text numeric weight="semibold" tone={category.type === "expense" ? "danger" : "action"}>{masked ? "••••••" : formatVND(category.amount)}</Text></div>)}</div>}
    </SurfaceCard>
    <SurfaceCard padding="md">
      <SectionTitle title="Hoạt động trong tháng" />
      <Text size="sm" tone="secondary">{summary.count} giao dịch được tính vào báo cáo. Chuyển ví và điều chỉnh số dư không làm tăng thu/chi.</Text>
      <Text size="xs" tone="secondary" className="mt-3">Số liệu được tính lại từ ledger theo múi giờ tài khoản; giao dịch chuyển ví và điều chỉnh số dư luôn bị loại khỏi tổng.</Text>
    </SurfaceCard>
  </>;
}

function Metric({ label, amount, tone, masked }: { label: string; amount: number; tone: "danger" | "action" | "ink"; masked: boolean }) {
  return <SurfaceCard padding="sm" tone="muted" elevation="flat"><Text size="xs" tone="secondary">{label}</Text><Text numeric size="sm" weight="bold" tone={tone} className="mt-1">{masked ? "••••••" : formatVND(amount)}</Text></SurfaceCard>;
}
