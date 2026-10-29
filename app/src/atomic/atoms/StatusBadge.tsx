import type { HTMLAttributes } from "react";

export type StatusBadgeTone = "success" | "muted" | "danger";

const toneClass: Record<StatusBadgeTone, string> = {
  success: "bg-successSoft text-action",
  muted: "bg-line text-secondary",
  danger: "bg-dangerSoft text-danger",
};

/** Compact semantic status surface. The caller resolves business state/copy. */
export function StatusBadge({ label, tone = "muted", className = "", ...props }: { label: string; tone?: StatusBadgeTone; className?: string } & Omit<HTMLAttributes<HTMLSpanElement>, "children">) {
  return <span role="status" data-testid="status-badge" className={`inline-flex min-h-7 items-center rounded-full px-2.5 py-1 text-xs font-semibold ${toneClass[tone]} ${className}`} {...props}>{label}</span>;
}
