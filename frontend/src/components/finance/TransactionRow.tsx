import * as React from "react";
import { cn } from "../../lib/utils";

export interface TransactionRowProps {
  id: string;
  categoryName: string;
  categoryIcon?: React.ReactNode;
  walletName?: string;
  walletBadgeIcon?: React.ReactNode;
  note?: string;
  amount: number;
  currency?: string;
  dateFormatted?: string;
  onClick?: () => void;
  className?: string;
}

export function TransactionRow({
  categoryName,
  categoryIcon,
  walletName,
  note,
  amount,
  currency = "VND",
  dateFormatted,
  onClick,
  className,
}: TransactionRowProps) {
  const isExpense = amount < 0;
  const isIncome = amount > 0;
  const formattedAbs = new Intl.NumberFormat("vi-VN").format(Math.abs(amount));

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
      {/* Leading Category Icon with Mini Wallet Emblem */}
      <div className="flex items-center gap-3 min-w-0 pr-2">
        <div className="relative w-11 h-11 rounded-full bg-[#29495a] text-white flex items-center justify-center shrink-0 font-bold text-sm shadow-xs">
          {categoryIcon || categoryName.slice(0, 1).toUpperCase()}
          {walletName && (
            <span className="absolute -bottom-1 -right-1 w-4 h-4 rounded-full bg-white border border-[#e8e8ec] flex items-center justify-center text-[9px] text-[#29495a]">
              💼
            </span>
          )}
        </div>

        <div className="flex flex-col min-w-0">
          <span className="text-[15px] font-bold text-[#111111] truncate">{categoryName}</span>
          <div className="flex items-center gap-1.5 text-xs text-[#8e8e93] truncate mt-0.5">
            {note && <span>{note}</span>}
            {note && walletName && <span>•</span>}
            {walletName && <span>{walletName}</span>}
            {(note || walletName) && dateFormatted && <span>•</span>}
            {dateFormatted && <span>{dateFormatted}</span>}
          </div>
        </div>
      </div>

      {/* Trailing Amount */}
      <div className="shrink-0 text-right">
        <span
          className={cn(
            "text-[15px] font-bold tabular-nums",
            isExpense && "text-[#ff5a66]",
            isIncome && "text-[#32a9df]",
            !isExpense && !isIncome && "text-[#111111]"
          )}
        >
          {isExpense ? `-${formattedAbs} đ` : isIncome ? `+${formattedAbs} đ` : `0 đ`}
        </span>
      </div>
    </div>
  );
}
