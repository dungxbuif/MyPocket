import * as React from "react";
import { cn } from "../../lib/utils";

export interface SheetProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  ariaLabel?: string;
  header?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
  maxHeight?: string;
}

export function Sheet({
  isOpen,
  onClose,
  title,
  ariaLabel,
  header,
  children,
  className,
  maxHeight = "92vh",
}: SheetProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/45 backdrop-blur-sm animate-in fade-in duration-200" onClick={onClose}>
      <section
        role="dialog"
        aria-modal="true"
        aria-label={ariaLabel || title}
        className={cn(
          "w-full max-w-[440px] rounded-t-[28px] bg-[#f2f3f8] shadow-2xl flex flex-col overflow-hidden animate-in slide-in-from-bottom duration-250",
          className
        )}
        style={{ maxHeight }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="w-full flex justify-center py-2 shrink-0">
          <div className="w-9 h-1 rounded-full bg-[#d1d1d6]" />
        </div>
        {header}
        <div className="flex-1 overflow-y-auto px-4 pb-[calc(24px+env(safe-area-inset-bottom))]">
          {children}
        </div>
      </section>
    </div>
  );
}
