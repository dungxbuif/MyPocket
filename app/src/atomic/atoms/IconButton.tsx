import type { ReactNode } from "react";

export function IconButton({ label, children, onClick }: { label: string; children: ReactNode; onClick?: () => void }) {
  return (
    <button
      type="button"
      aria-label={label}
      onClick={onClick}
      className="grid h-10 w-10 place-items-center rounded-full bg-white text-[#3f4a3c] shadow-sm transition active:scale-95"
    >
      {children}
    </button>
  );
}

