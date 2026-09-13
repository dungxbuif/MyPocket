export function SectionTitle({ title, action, onAction }: { title: string; action?: string; onAction?: () => void }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <h2 className="text-base font-bold text-ink">{title}</h2>
      {action ? <button type="button" onClick={onAction} className="cursor-pointer rounded-full px-2 py-1 text-sm font-semibold text-action transition hover:bg-success-soft">{action}</button> : null}
    </div>
  );
}
