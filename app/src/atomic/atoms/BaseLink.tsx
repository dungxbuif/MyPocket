import { Link } from "@tanstack/react-router";
import type { ReactNode } from "react";
import { UI_CLASSES } from "./tokens";

export function BaseLink({ to, label, children, className = "" }: { to: string; label?: string; children: ReactNode; className?: string }) {
  return <Link to={to} aria-label={label} className={`${UI_CLASSES.focus} inline-flex min-h-11 min-w-11 items-center justify-center gap-2 rounded-full border border-line bg-card px-3 text-sm font-semibold text-action shadow-control hover:bg-row ${className}`}>{children}</Link>;
}
