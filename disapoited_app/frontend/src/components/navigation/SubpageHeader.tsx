import * as React from "react";
import { cn } from "../../lib/utils";
import { ChevronLeft } from "lucide-react";

export interface SubpageHeaderProps {
  title: string;
  onBack: () => void;
  backLabel?: string;
  centerExtra?: React.ReactNode;
  rightAction?: {
    label?: string;
    icon?: React.ReactNode;
    onClick: () => void;
    disabled?: boolean;
  };
  className?: string;
}

export function SubpageHeader({
  title,
  onBack,
  backLabel = "Quay lại",
  centerExtra,
  rightAction,
  className,
}: SubpageHeaderProps) {
  return (
    <header className={cn("flex items-center justify-between h-12 w-full px-2 mb-2 select-none", className)}>
      <button
        type="button"
        onClick={onBack}
        className="flex items-center gap-1 px-3 py-1.5 rounded-full bg-[#eef0f4] text-[#111111] text-sm font-medium hover:bg-[#e2e4e9] active:scale-95 transition-all outline-none"
      >
        <ChevronLeft className="w-4 h-4 stroke-[2.5]" />
        <span>{backLabel}</span>
      </button>

      <div className="flex items-center gap-1.5 text-center">
        <h2 className="font-bold text-base text-[#111111] truncate">{title}</h2>
        {centerExtra}
      </div>

      <div className="min-w-[70px] flex justify-end">
        {rightAction ? (
          <button
            type="button"
            disabled={rightAction.disabled}
            onClick={rightAction.onClick}
            className="flex items-center justify-center p-2 rounded-full text-[#111111] hover:bg-[#eef0f4] active:scale-95 transition-all outline-none text-sm font-semibold disabled:opacity-50"
          >
            {rightAction.icon}
            {rightAction.label && <span className="ml-1">{rightAction.label}</span>}
          </button>
        ) : null}
      </div>
    </header>
  );
}
