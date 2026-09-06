import * as React from "react";
import { cn } from "../../lib/utils";

export interface GroupedCardProps {
  title?: string;
  headerAction?: {
    label: string;
    onClick: () => void;
  };
  children: React.ReactNode;
  className?: string;
}

export function GroupedCard({ title, headerAction, children, className }: GroupedCardProps) {
  return (
    <div className={cn("w-full mb-4", className)}>
      {(title || headerAction) && (
        <div className="flex items-center justify-between px-3 pb-1.5 select-none">
          {title && <span className="text-xs font-bold uppercase tracking-wider text-[#8e8e93]">{title}</span>}
          {headerAction && (
            <button
              type="button"
              onClick={headerAction.onClick}
              className="text-xs font-semibold text-[#2dbd4f] hover:underline"
            >
              {headerAction.label}
            </button>
          )}
        </div>
      )}
      <div className="rounded-[24px] bg-white p-3.5 shadow-[0_2px_10px_rgba(0,0,0,0.03)] divide-y divide-[#e8e8ec]">
        {children}
      </div>
    </div>
  );
}
