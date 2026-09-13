import { useId } from "react";
export function Progress({ value, label, danger = false, className = "" }: { value: number; label: string; danger?: boolean; className?: string }) {
  const percent = Math.max(0, Math.min(100, Number.isFinite(value) ? value : 0));
  return <div role="progressbar" aria-label={label} aria-valuemin={0} aria-valuemax={100} aria-valuenow={percent} className={`h-2 overflow-hidden rounded-full bg-line ${className}`}><div className={`h-full rounded-full ${danger ? "bg-danger" : "bg-accent"}`} style={{ width: `${percent}%` }} /></div>;
}

export function BudgetGauge({ value, label }: { value: number; label: string }) {
  const title = useId();
  const percent = Math.max(0, Math.min(100, Number.isFinite(value) ? value : 0));
  return <svg viewBox="0 0 240 130" role="img" aria-labelledby={title} className="mx-auto mt-5 w-56"><title id={title}>{`${label}: ${value}%`}</title><path d="M 12 118 A 108 108 0 0 1 228 118" fill="none" className="stroke-line" strokeWidth="16" strokeLinecap="round" /><path d="M 12 118 A 108 108 0 0 1 228 118" fill="none" className={value > 100 ? "stroke-danger" : "stroke-accent"} strokeWidth="16" strokeLinecap="round" pathLength="100" strokeDasharray={`${percent} 100`} /><text x="120" y="112" textAnchor="middle" className="fill-ink text-xl font-bold">{value}%</text></svg>;
}
