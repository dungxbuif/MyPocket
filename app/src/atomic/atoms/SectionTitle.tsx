export function SectionTitle({ title, action }: { title: string; action?: string }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <h2 className="text-base font-bold text-[#1b1c1c]">{title}</h2>
      {action ? <button className="text-sm font-semibold text-[#006e1c]">{action}</button> : null}
    </div>
  );
}

