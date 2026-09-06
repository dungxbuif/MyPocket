import * as React from "react";
import { cn } from "../../lib/utils";
import { Globe, ChevronsUpDown } from "lucide-react";

export interface WalletFilterChipProps {
  walletName: string;
  icon?: React.ReactNode;
  onClick?: () => void;
  className?: string;
}

export function WalletFilterChip({
  walletName,
  icon = <Globe className="w-3.5 h-3.5 text-[#29495a]" />,
  onClick,
  className,
}: WalletFilterChipProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "inline-flex items-center gap-1.5 h-8 px-3 rounded-full bg-[#eef0f4] hover:bg-[#e2e4e9] text-xs font-semibold text-[#111111] active:scale-95 transition-all select-none outline-none border border-transparent",
        className
      )}
    >
      {icon}
      <span>{walletName}</span>
      <ChevronsUpDown className="w-3 h-3 text-[#8e8e93]" />
    </button>
  );
}
