import * as React from "react";
import { cn } from "../../lib/utils";

export interface DestructiveActionRowProps {
  label: string;
  onClick: () => void;
  variant?: "row" | "pill";
  className?: string;
}

export function DestructiveActionRow({
  label,
  onClick,
  variant = "row",
  className,
}: DestructiveActionRowProps) {
  if (variant === "pill") {
    return (
      <button
        type="button"
        onClick={onClick}
        className={cn(
          "w-full h-12 rounded-full bg-white border border-[#ffd2d6] text-[#9f1d1d] font-semibold text-[15px] hover:bg-[#ffebee] active:scale-[0.98] transition-all flex items-center justify-center outline-none shadow-sm",
          className
        )}
      >
        {label}
      </button>
    );
  }

  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "w-full flex items-center justify-start py-3.5 px-2 text-[15px] font-semibold text-[#9f1d1d] hover:bg-[#ffebee]/40 active:bg-[#ffebee] rounded-xl transition-colors outline-none",
        className
      )}
    >
      {label}
    </button>
  );
}
