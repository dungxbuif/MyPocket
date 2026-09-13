import type { ReactNode } from "react";
export function Chip({ children, className = "" }: { children: ReactNode; className?: string }) {
  return <span className={`inline-flex shrink-0 items-center gap-2 rounded-full bg-card px-4 py-2 text-sm font-semibold text-secondary ${className}`}>{children}</span>;
}
