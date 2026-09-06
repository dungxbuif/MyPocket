import * as React from "react";
import { cn } from "../../lib/utils";

export interface NoticeBannerProps {
  message: string;
  dismissLabel?: string;
  onDismiss?: () => void;
  className?: string;
}

export function NoticeBanner({
  message,
  dismissLabel = "Tắt",
  onDismiss,
  className,
}: NoticeBannerProps) {
  return (
    <div
      className={cn(
        "w-full flex items-center justify-between gap-3 px-4 py-2.5 rounded-2xl bg-[#2f80ed] text-white text-xs font-medium shadow-sm select-none",
        className
      )}
    >
      <span className="truncate leading-tight">{message}</span>
      {onDismiss && (
        <button
          type="button"
          onClick={onDismiss}
          className="px-2.5 py-1 rounded-full bg-white/20 hover:bg-white/30 text-white font-bold text-xs shrink-0 active:scale-95 transition-all outline-none"
        >
          {dismissLabel}
        </button>
      )}
    </div>
  );
}
