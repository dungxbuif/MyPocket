import * as React from "react";
import { cn } from "../../lib/utils";

export interface AmountInputHeroProps {
  value: number;
  currency?: string;
  onChange: (value: number) => void;
  placeholder?: string;
  autoFocus?: boolean;
  className?: string;
  readOnly?: boolean;
}

export function AmountInputHero({
  value,
  currency = "VND",
  onChange,
  placeholder = "0",
  autoFocus,
  className,
  readOnly,
}: AmountInputHeroProps) {
  const formattedDisplay = value ? new Intl.NumberFormat("vi-VN").format(value) : "";

  function handleInputChange(e: React.ChangeEvent<HTMLInputElement>) {
    const rawVal = e.target.value.replace(/\D/g, "");
    const parsed = parseInt(rawVal, 10);
    onChange(isNaN(parsed) ? 0 : parsed);
  }

  return (
    <div className={cn("flex items-center gap-3 py-3 px-2 w-full", className)}>
      <span className="h-7 px-2.5 rounded-full bg-[#eef0f4] text-xs font-bold text-[#111111] flex items-center justify-center shrink-0 select-none">
        {currency}
      </span>
      <div className="flex-1 relative flex items-center">
        <input
          type="text"
          inputMode="numeric"
          pattern="[0-9]*"
          readOnly={readOnly}
          autoFocus={autoFocus}
          value={formattedDisplay}
          placeholder={placeholder}
          onChange={handleInputChange}
          className="w-full bg-transparent text-[32px] sm:text-[36px] font-bold text-[#111111] placeholder:text-[#8e8e93] focus:outline-none tracking-tight tabular-nums"
        />
      </div>
    </div>
  );
}
