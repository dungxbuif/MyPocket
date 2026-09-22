import { BaseButton } from "../atoms/BaseButton";
import { Heading } from "../atoms/Heading";
import { Progress } from "../atoms/Progress";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";

type UnknownRecord = Record<string, unknown>;

function record(value: unknown): UnknownRecord | null {
  return value !== null && typeof value === "object" && !Array.isArray(value) ? value as UnknownRecord : null;
}

function numberValue(value: unknown): number | null {
  return typeof value === "number" && Number.isSafeInteger(value) ? value : null;
}

function money(value: unknown, masked: boolean): string {
  if (masked) return "••••";
  const amount = numberValue(value);
  return amount === null ? "Chưa có dữ liệu" : `${new Intl.NumberFormat("vi-VN").format(amount)} ₫`;
}

function count(value: unknown, masked: boolean): string {
  if (masked) return "••••";
  const amount = numberValue(value);
  return amount === null ? "Chưa có dữ liệu" : new Intl.NumberFormat("vi-VN").format(amount);
}

function normalizeType(type: string): string {
  const aliases: Record<string, string> = {
    finance_summary: "metric_group",
    spending_comparison: "period_comparison",
    transaction_search: "transaction_list",
    budget_progress: "budget_status",
  };
  return aliases[type] ?? type;
}

function sourceLabel(data: UnknownRecord): string {
  const source = record(data.source);
  if (!source) return "Nguồn dữ liệu tài chính";
  const timezone = typeof source.timezone === "string" ? source.timezone : "";
  return timezone ? `Nguồn dữ liệu · ${timezone}` : "Nguồn dữ liệu tài chính";
}

export function FinanceAssistantPart({ type, data, masked = false }: { type: string; data: unknown; masked?: boolean }) {
  const payload = record(data);
  const view = payload ? record(payload.view) : null;
  if (!payload || !view) return <SurfaceCard tone="danger" padding="sm"><Text tone="danger">Nội dung trợ lý chưa được hỗ trợ.</Text></SurfaceCard>;
  const normalizedType = normalizeType(type);
  if (normalizedType === "metric_group") {
    return <SurfaceCard padding="sm" className="space-y-2"><Heading as="h3" size="field">Tổng quan kỳ chọn</Heading><div className="grid grid-cols-2 gap-2"><Metric label="Thu nhập" value={view.income} masked={masked} /><Metric label="Chi tiêu" value={view.expense} masked={masked} /><Metric label="Thu trừ chi" value={view.net} masked={masked} /><Metric label="Số giao dịch" value={view.count} masked={masked} format="count" /></div><Text size="micro" tone="secondary">{sourceLabel(payload)}</Text></SurfaceCard>;
  }
  if (normalizedType === "period_comparison") {
    return <SurfaceCard padding="sm" className="space-y-2"><Heading as="h3" size="field">So sánh chi tiêu</Heading><div className="grid grid-cols-3 gap-2"><Metric label="Hiện tại" value={view.current_expense} masked={masked} /><Metric label="Kỳ trước" value={view.previous_expense} masked={masked} /><Metric label="Chênh lệch" value={view.delta} masked={masked} /></div><Text size="micro" tone="secondary">Phần trăm thay đổi chỉ hiển thị khi kỳ trước có số liệu.</Text></SurfaceCard>;
  }
  if (normalizedType === "transaction_list") {
    const items = Array.isArray(view.items) ? view.items.slice(0, 5).map(record).filter((item): item is UnknownRecord => !!item) : [];
    return <SurfaceCard padding="sm" className="space-y-2"><div className="flex items-center justify-between gap-2"><Heading as="h3" size="field">Giao dịch</Heading><Text size="micro" tone="secondary">{String(view.total_count ?? items.length)} dòng</Text></div><div className="space-y-1">{items.map(item => <div className="flex items-center justify-between gap-2 border-b border-line py-2 last:border-0" key={String(item.id)}><Text size="xs">{String(item.note || item.type || "Giao dịch")}</Text><Text size="xs" numeric>{money(item.amount, masked)}</Text></div>)}</div><Text size="micro" tone="secondary">{sourceLabel(payload)}</Text></SurfaceCard>;
  }
  if (normalizedType === "transaction_detail") {
    return <SurfaceCard padding="sm" className="space-y-2"><Heading as="h3" size="field">Chi tiết giao dịch</Heading><Text>{String(view.wallet_name || "Ví")}{view.category_name ? ` · ${String(view.category_name)}` : ""}</Text><Text size="xl" weight="bold" numeric>{money(view.amount, masked)}</Text><Text size="micro" tone="secondary">{view.included_in_reports === false ? "Không tính vào báo cáo" : "Tính vào báo cáo"}{view.transfer_id ? " · Chuyển ví" : ""}</Text></SurfaceCard>;
  }
  if (normalizedType === "wallet_balances" || normalizedType === "goal_progress") {
    const items = Array.isArray(view.items) ? view.items.map(record).filter((item): item is UnknownRecord => !!item) : [];
    const isGoal = normalizedType === "goal_progress";
    return <SurfaceCard padding="sm" className="space-y-2"><Heading as="h3" size="field">{isGoal ? "Tiến độ mục tiêu" : "Số dư các ví"}</Heading><div className="space-y-2">{items.slice(0, 8).map(item => <div className="rounded-lg bg-row p-2" key={String(item.id)}><div className="flex items-center justify-between gap-2"><Text weight="semibold">{String(item.name || "Ví")}</Text><Text numeric weight="bold">{money(item.current_balance, masked)}</Text></div>{isGoal && numberValue(item.target_amount) !== null ? <><Progress value={progress(item.current_balance, item.target_amount)} label={`Tiến độ ${String(item.name || "mục tiêu")}`} className="mt-2" /><Text size="micro" tone="secondary">{count(progress(item.current_balance, item.target_amount), false)}% mục tiêu</Text></> : null}</div>)}</div><Text size="micro" tone="secondary">{sourceLabel(payload)}</Text></SurfaceCard>;
  }
  if (normalizedType === "budget_status") {
    const items = Array.isArray(view.items) ? view.items.map(record).filter((item): item is UnknownRecord => !!item) : [];
    return <SurfaceCard padding="sm" className="space-y-2"><Heading as="h3" size="field">Tiến độ ngân sách</Heading><div className="space-y-2">{items.slice(0, 8).map(item => { const spent = numberValue(item.spent) ?? 0; const limit = numberValue(item.limit_amount) ?? 0; const percent = progress(spent, limit); return <div className="rounded-lg bg-row p-2" key={String(item.id)}><div className="flex items-center justify-between gap-2"><Text weight="semibold">{String(item.name || "Ngân sách")}</Text><Text numeric weight="bold" tone={spent > limit ? "danger" : "ink"}>{count(percent, false)}%</Text></div><Text size="xs" numeric tone="secondary">{money(spent, masked)} / {money(limit, masked)}</Text><Progress value={percent} label={String(item.name || "Ngân sách")} danger={spent > limit} className="mt-2" /></div>; })}</div><Text size="micro" tone="secondary">Tổng: {money(view.spent, masked)} / {money(view.limit_amount, masked)}</Text></SurfaceCard>;
  }
  if (normalizedType === "jar_progress") {
    const items = Array.isArray(view.items) ? view.items.map(record).filter((item): item is UnknownRecord => !!item) : [];
    return <SurfaceCard padding="sm" className="space-y-2"><Heading as="h3" size="field">Tiến độ hũ chi tiêu</Heading><div className="space-y-2">{items.slice(0, 8).map(item => { const spent = numberValue(item.spent) ?? 0; const allocation = numberValue(item.allocation_amount); const percent = allocation === null ? null : progress(spent, allocation); return <div className="rounded-lg bg-row p-2" key={String(item.jar_id)}><div className="flex items-center justify-between gap-2"><Text weight="semibold">{String(item.name || "Hũ")}</Text><Text numeric weight="bold">{money(spent, masked)}</Text></div>{allocation !== null && percent !== null ? <><Text size="xs" numeric tone="secondary">/ {money(allocation, masked)}</Text><Progress value={percent} label={String(item.name || "Hũ")} danger={spent > allocation} className="mt-2" /></> : null}</div>; })}</div><Text size="micro" tone="secondary">Kỳ {String(view.month || "đang chọn")}</Text></SurfaceCard>;
  }
  return <SurfaceCard tone="muted" padding="sm"><Text size="xs">Dữ liệu {type} đã có nhưng giao diện hiện tại chưa hỗ trợ hiển thị.</Text></SurfaceCard>;
}

function progress(value: unknown, target: unknown): number {
  const current = numberValue(value);
  const limit = numberValue(target);
  if (current === null || limit === null || limit <= 0) return 0;
  return Math.round((current / limit) * 100);
}

function Metric({ label, value, masked, format = "money" }: { label: string; value: unknown; masked: boolean; format?: "money" | "count" }) {
  return <div className="rounded-lg bg-row p-2"><Text size="micro" tone="secondary">{label}</Text><Text size="sm" weight="bold" numeric>{format === "count" ? count(value, masked) : money(value, masked)}{format === "money" && value !== null ? "" : ""}</Text></div>;
}

export function FinanceAssistantPartAction({ label, onClick }: { label: string; onClick?: () => void }) {
  return <BaseButton variant="secondary" size="sm" onClick={onClick}>{label}</BaseButton>;
}
