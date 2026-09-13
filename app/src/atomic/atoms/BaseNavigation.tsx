import type { ButtonHTMLAttributes, ReactNode } from "react";
import { UI_CLASSES } from "./tokens";

export function BaseNavigationItem({ active, children, className = "", ...props }: ButtonHTMLAttributes<HTMLButtonElement> & { active: boolean; children: ReactNode }) {
  return <button {...props} type="button" aria-current={active ? "page" : undefined} className={`${UI_CLASSES.focus} flex min-h-13 flex-col items-center justify-center gap-1 text-xs font-bold ${active ? "text-action" : "text-secondary"} ${className}`}>{children}</button>;
}
export function BaseFab({ children, className = "", ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button {...props} type="button" className={`${UI_CLASSES.focus} grid h-14 w-14 place-items-center rounded-full bg-brand text-card shadow-fab disabled:opacity-50 ${className}`}>{children}</button>;
}
