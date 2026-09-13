export function SectionTitle({ title, action, onAction }: { title: string; action?: string; onAction?: () => void }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <h2 className="text-base font-bold text-[#1b1c1c]">{title}</h2>
      {action ? <button type="button" onClick={onAction} className="cursor-pointer rounded-full px-2 py-1 text-sm font-semibold text-[#006e1c] transition hover:bg-[#ecfdf5]">{action}</button> : null}
    </div>
  );
}
