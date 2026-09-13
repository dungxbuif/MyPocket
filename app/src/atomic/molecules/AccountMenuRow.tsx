import type { LucideIcon } from "lucide-react";
import { ChevronRight } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { IconBadge } from "../atoms/IconBadge";

export function AccountMenuRow({ to, icon, label }: { to: "/account/wallets" | "/account/groups"; icon: LucideIcon; label: string }) {
  return <Link to={to} className="flex min-h-12 cursor-pointer items-center justify-between px-4 py-3.5 transition hover:bg-[#f5f3f3]"><span className="flex items-center gap-3"><IconBadge icon={icon} size="sm" /><span className="text-sm font-semibold text-[#1b1c1c]">{label}</span></span><ChevronRight size={17} className="text-[#6f7a6b]" /></Link>;
}
