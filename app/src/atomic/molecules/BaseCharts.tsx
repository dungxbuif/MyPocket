import { Text } from "../atoms/Text";

export function BaseBarChart({ values, dangerFrom = Infinity, label }: { values: number[]; dangerFrom?: number; label: string }) {
  return <div role="img" aria-label={`${label}: ${values.join(', ')}`} className="mt-4 flex h-32 items-end gap-2">
    {values.map((value, index) => <div key={index} className="flex h-full flex-1 items-end"><div className="w-full rounded-t-full bg-line" style={{ height: `${Math.max(0, Math.min(100, value))}%` }}><div className={`h-full rounded-t-full ${index >= dangerFrom ? "bg-danger" : "bg-action"}`} /></div></div>)}
  </div>;
}

export function BaseDonutChart({ shares, center }: { shares: { name: string; value: number; color: string }[]; center: string }) {
  const total = shares.reduce((sum, share) => sum + Math.max(0, share.value), 0);
  let offset = 0;
  const segments = shares.map(share => { const start = offset; offset += total ? Math.max(0, share.value) / total * 100 : 0; return `${share.color} ${start}% ${offset}%`; });
  return <div className="mt-4 flex items-center gap-5">
    <div role="img" aria-label={shares.map(share => `${share.name}: ${share.value}%`).join(', ')} className="grid h-32 w-32 shrink-0 place-items-center rounded-full" style={{ background: total ? `conic-gradient(${segments.join(', ')})` : 'var(--color-line)' }}><div className="grid h-20 w-20 place-items-center rounded-full bg-card text-center"><Text as="span" size="xs" weight="bold" tone="secondary">Top</Text><Text as="strong">{center}</Text></div></div>
    <div className="min-w-0 flex-1 space-y-2">{shares.map(share => <div key={share.name} className="flex items-center gap-2"><span aria-hidden className="h-3 w-3 rounded-full" style={{ background: share.color }} /><Text as="span" className="min-w-0 flex-1 truncate">{share.name}</Text><Text as="strong">{share.value}%</Text></div>)}</div>
  </div>;
}

type TrendPoint = { day: string; income: number; expense: number };

export function BaseTrendChart({ points }: { points: TrendPoint[] }) {
  if (points.length === 0) {
    return <div role="status" aria-label="Xu hướng chi tiêu" className="mt-4"><Text size="sm" tone="secondary">Chưa có giao dịch được tính vào xu hướng.</Text></div>;
  }
  const maximum = Math.max(1, ...points.flatMap(point => [point.income, point.expense]));
  const polyline = (key: "income" | "expense") => points.map((point, index) => {
    const x = points.length === 1 ? 0 : index / (points.length - 1) * 100;
    const y = 40 - Math.min(1, Math.max(0, point[key] / maximum)) * 40;
    return `${x},${y}`;
  }).join(" ");
  return <div role="img" aria-label="Xu hướng thu chi" className="mt-4 space-y-2">
    <div className="flex items-center gap-4 text-xs text-secondary"><span className="inline-flex items-center gap-1"><span aria-hidden className="h-2 w-2 rounded-full bg-action" />Thu</span><span className="inline-flex items-center gap-1"><span aria-hidden className="h-2 w-2 rounded-full bg-danger" />Chi</span></div>
    <svg viewBox="0 0 100 40" preserveAspectRatio="none" className="h-36 w-full" aria-hidden>
      <line x1="0" y1="0" x2="100" y2="0" stroke="var(--color-line)" strokeOpacity="0.45" />
      <line x1="0" y1="20" x2="100" y2="20" stroke="var(--color-line)" strokeOpacity="0.45" />
      <line x1="0" y1="40" x2="100" y2="40" stroke="var(--color-line)" strokeOpacity="0.45" />
      <polyline points={polyline("income")} fill="none" stroke="var(--color-action)" strokeWidth="1.4" vectorEffect="non-scaling-stroke" strokeLinecap="round" strokeLinejoin="round" />
      <polyline points={polyline("expense")} fill="none" stroke="var(--color-danger)" strokeWidth="1.4" vectorEffect="non-scaling-stroke" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
    <div className="flex items-center justify-between gap-2 text-xs text-secondary" aria-hidden>{points.map(point => <span key={point.day}>{point.day.slice(-2)}</span>)}</div>
  </div>;
}
