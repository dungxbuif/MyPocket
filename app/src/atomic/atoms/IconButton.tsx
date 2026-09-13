import type { ReactNode } from "react";

export function IconButton({ label, children, onClick, disabled = false }: { label: string; children: ReactNode; onClick?: () => void; disabled?: boolean }) {
  return (
    <button
      type="button"
      aria-label={label}
      onClick={onClick}
      disabled={disabled}
      className="grid h-10 w-10 cursor-pointer place-items-center rounded-full border border-slate-100 bg-white text-[#3f4a3c] shadow-sm transition hover:bg-[#f5f3f3] active:scale-95 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#006e1c] disabled:cursor-not-allowed disabled:opacity-50"
    >
      {children}
    </button>
  );
}
