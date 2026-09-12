import * as React from "react";
import { cn } from "../../lib/utils";

export interface TooltipPopoverProps {
  isOpen: boolean;
  text: string;
  onClose?: () => void;
  className?: string;
}

export function TooltipPopover({ isOpen, text, onClose, className }: TooltipPopoverProps) {
  if (!isOpen) return null;

  return (
    <div
      role="tooltip"
      onClick={onClose}
      className={cn(
        "absolute z-50 px-3 py-2 text-xs font-normal text-white bg-[#22242a] rounded-xl shadow-lg max-w-[240px] animate-in fade-in duration-150 select-none",
        className
      )}
    >
      {text}
    </div>
  );
}
