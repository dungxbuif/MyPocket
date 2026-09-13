import { BASE_COMPONENT_RADIUS } from "./tokens";

export function MetricBox({ label, value, danger = false, className = "" }: { label: string; value: string; danger?: boolean; className?: string }) {
  return (
    <div className={`${BASE_COMPONENT_RADIUS} bg-row p-3 ${className}`}>
      <p className="text-xs text-secondary">{label}</p>
      <p className={`money mt-1 text-sm font-bold ${danger ? "text-danger" : "text-ink"}`}>{value}</p>
    </div>
  );
}
