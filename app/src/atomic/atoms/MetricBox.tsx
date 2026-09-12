export function MetricBox({ label, value, danger = false }: { label: string; value: string; danger?: boolean }) {
  return (
    <div className="rounded-2xl bg-[#f5f3f3] p-3">
      <p className="text-xs text-[#3f4a3c]">{label}</p>
      <p className={`money mt-1 text-sm font-bold ${danger ? "text-[#bb1614]" : "text-[#1b1c1c]"}`}>{value}</p>
    </div>
  );
}

