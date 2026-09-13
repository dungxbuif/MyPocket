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
