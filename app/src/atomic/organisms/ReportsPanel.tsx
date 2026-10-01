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
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [retry, setRetry] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(false);
    fetchAdvisorOverview(month).then(result => {
      const next = readSummary(result.view);
      if (!cancelled) {
        setSummary(next);
        setError(!next);
      }
    }).catch(() => { if (!cancelled) setError(true); }).finally(() => { if (!cancelled) setLoading(false); });
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
      <SectionTitle title="Hoạt động trong tháng" />
      <Text size="sm" tone="secondary">{summary.count} giao dịch được tính vào báo cáo. Chuyển ví và điều chỉnh số dư không làm tăng thu/chi.</Text>
      <Text size="xs" tone="secondary" className="mt-3">Báo cáo danh mục chi tiết sẽ dùng cùng finance query layer khi endpoint drill-down được bật; không hiển thị dữ liệu mẫu.</Text>
    </SurfaceCard>
  </>;
}

function Metric({ label, amount, tone, masked }: { label: string; amount: number; tone: "danger" | "action" | "ink"; masked: boolean }) {
  return <SurfaceCard padding="sm" tone="muted" elevation="flat"><Text size="xs" tone="secondary">{label}</Text><Text numeric size="sm" weight="bold" tone={tone} className="mt-1">{masked ? "••••••" : formatVND(amount)}</Text></SurfaceCard>;
}
