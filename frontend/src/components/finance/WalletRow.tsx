import * as React from "react";
import { cn } from "../../lib/utils";
import { CreditCard, Wallet as WalletIcon, Landmark, Check } from "lucide-react";

export interface WalletRowProps {
  id: string;
  name: string;
  type?: "cash" | "bank" | "credit" | string;
  balance: number;
  currency?: string;
  selected?: boolean;
  privacyMasked?: boolean;
  onClick?: () => void;
  className?: string;
}

export function WalletRow({
  name,
  type = "bank",
  balance,
  selected,
  privacyMasked,
  onClick,
  className,
}: WalletRowProps) {
  const isNegative = balance < 0;
  const formattedBalance = privacyMasked
    ? "•••••• đ"
    : `${isNegative ? "-" : ""}${new Intl.NumberFormat("vi-VN").format(Math.abs(balance))} đ`;

  function renderIcon() {
    switch (type) {
      case "credit":
        return <CreditCard className="w-5 h-5 text-white" />;
      case "cash":
        return <WalletIcon className="w-5 h-5 text-white" />;
      default:
        return <Landmark className="w-5 h-5 text-white" />;
    }
  }

  return (
    <div
      onClick={onClick}
      role={onClick ? "button" : undefined}
      tabIndex={onClick ? 0 : undefined}
      className={cn(
        "w-full flex items-center justify-between py-3 px-2 transition-colors outline-none",
        onClick && "hover:bg-[#f8f9fa] active:bg-[#f0f1f4] cursor-pointer rounded-2xl",
        className
      )}
    >
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 rounded-full bg-[#29495a] flex items-center justify-center shrink-0 shadow-xs">
          {renderIcon()}
        </div>
        <div className="flex flex-col">
          <span className="text-[15px] font-semibold text-[#111111]">{name}</span>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <span
          className={cn(
            "text-[15px] font-bold tabular-nums",
            isNegative ? "text-[#ff5a66]" : "text-[#111111]"
          )}
        >
          {formattedBalance}
        </span>
        {selected ? <Check className="w-5 h-5 text-[#2dbd4f] stroke-[2.5]" /> : null}
      </div>
    </div>
  );
}
