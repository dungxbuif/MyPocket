import * as React from "react";
import { cn } from "../../lib/utils";

export interface ModalHeaderProps {
  title: string;
  onDismiss: () => void;
  dismissLabel?: string;
  rightAction?: {
    label?: string;
    icon?: React.ReactNode;
    onClick: () => void;
    isPrimary?: boolean;
    disabled?: boolean;
  };
  className?: string;
}

export function ModalHeader({
  title,
  onDismiss,
  dismissLabel = "Huỷ",
  rightAction,
  className,
}: ModalHeaderProps) {
  return (
    <header className={cn("flex items-center justify-between h-14 w-full px-4 border-b border-[#e8e8ec]/60 select-none", className)}>
      <button
        type="button"
        onClick={onDismiss}
        className="px-3.5 py-1.5 rounded-full bg-[#eef0f4] text-[#111111] text-sm font-semibold hover:bg-[#e2e4e9] active:scale-95 transition-all outline-none"
      >
        {dismissLabel}
      </button>

      <h2 className="font-bold text-base text-[#111111] text-center">{title}</h2>

      <div className="min-w-[60px] flex justify-end">
        {rightAction ? (
          <button
            type="button"
            disabled={rightAction.disabled}
            onClick={rightAction.onClick}
            className={cn(
              "px-3 py-1.5 rounded-full text-sm font-semibold active:scale-95 transition-all outline-none disabled:opacity-50",
              rightAction.isPrimary
                ? "text-[#2dbd4f] hover:bg-[#e8f7ed]"
                : "text-[#111111] hover:bg-[#eef0f4]"
            )}
          >
            {rightAction.icon}
            {rightAction.label}
          </button>
        ) : null}
      </div>
    </header>
  );
}
