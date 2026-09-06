import * as React from "react";
import { cn } from "../../lib/utils";
import { Lock, ChevronRight } from "lucide-react";

export interface CategoryTreeRowProps {
  id: string;
  name: string;
  icon?: React.ReactNode;
  isChild?: boolean;
  walletCount?: number;
  isSystemLocked?: boolean;
  onLockedClick?: () => void;
  onClick?: () => void;
  className?: string;
}

export function CategoryTreeRow({
  name,
  icon,
  isChild,
  walletCount = 3,
  isSystemLocked,
  onLockedClick,
  onClick,
  className,
}: CategoryTreeRowProps) {
  return (
    <div
      onClick={onClick}
      role={onClick ? "button" : undefined}
      className={cn(
        "relative w-full flex items-center justify-between py-3 px-2 transition-colors outline-none",
        isChild ? "pl-10" : "pl-2",
        onClick && "hover:bg-[#f8f9fa] active:bg-[#f0f1f4] cursor-pointer rounded-2xl",
        className
      )}
    >
      {/* L-shaped Branch Line for Child Nodes */}
      {isChild && (
        <div className="absolute left-4 top-0 bottom-1/2 w-4 border-b-2 border-l-2 border-[#d1d1d6] rounded-bl-sm pointer-events-none" />
      )}

      <div className="flex items-center gap-3 min-w-0 pr-2">
        <div className="w-9 h-9 rounded-full bg-[#29495a] text-white flex items-center justify-center shrink-0 font-bold text-sm shadow-xs">
          {icon || name.slice(0, 1).toUpperCase()}
        </div>

        <div className="flex flex-col min-w-0">
          <div className="flex items-center gap-1.5">
            <span className="text-[15px] font-semibold text-[#111111] truncate">{name}</span>
            {isSystemLocked && (
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  onLockedClick?.();
                }}
                className="text-[#8e8e93] hover:text-[#111111]"
              >
                <Lock className="w-3.5 h-3.5" />
              </button>
            )}
          </div>
          <span className="text-xs text-[#8e8e93]">Hoạt động trong {walletCount} ví</span>
        </div>
      </div>

      <ChevronRight className="w-4 h-4 text-[#c7c7cc] shrink-0" />
    </div>
  );
}
