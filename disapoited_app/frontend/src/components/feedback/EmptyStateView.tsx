import * as React from "react";
import { cn } from "../../lib/utils";

export interface EmptyStateViewProps {
  illustration?: React.ReactNode;
  title: string;
  description?: string;
  actionButton?: {
    label: string;
    onClick: () => void;
  };
  className?: string;
}

export function EmptyStateView({
  illustration = <span className="text-4xl">🙌</span>,
  title,
  description,
  actionButton,
  className,
}: EmptyStateViewProps) {
  return (
    <div className={cn("w-full flex flex-col items-center justify-center py-8 px-4 text-center select-none", className)}>
      <div className="mb-3">{illustration}</div>
      <h4 className="text-[15px] font-bold text-[#111111] max-w-[260px] leading-snug">{title}</h4>
      {description && <p className="text-xs text-[#8e8e93] mt-1 max-w-[260px] leading-relaxed">{description}</p>}
      {actionButton && (
        <button
          type="button"
          onClick={actionButton.onClick}
          className="mt-4 px-5 py-2 rounded-full bg-[#111111] text-white text-xs font-bold hover:bg-[#242424] active:scale-95 transition-all shadow-xs"
        >
          {actionButton.label}
        </button>
      )}
    </div>
  );
}
