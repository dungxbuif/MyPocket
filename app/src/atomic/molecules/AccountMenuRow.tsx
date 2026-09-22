import type { LucideIcon } from "lucide-react";
import { ChevronRight } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { IconBadge } from "../atoms/IconBadge";

export function AccountMenuRow({ to, icon, label }: { to: "/account/wallets" | "/account/groups" | "/account/feedback" | "/account/budgets"; icon: LucideIcon; label: string }) {
  return <Link to={to} className="flex min-h-12 cursor-pointer items-center justify-between px-4 py-3.5 transition hover:bg-row"><span className="flex items-center gap-3"><IconBadge icon={icon} size="sm" /><span className="text-sm font-semibold text-ink">{label}</span></span><ChevronRight size={17} className="text-secondary" /></Link>;
}
