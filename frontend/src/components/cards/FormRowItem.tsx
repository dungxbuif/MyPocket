import * as React from "react";
import { cn } from "../../lib/utils";
import { ChevronRight } from "lucide-react";

export interface FormRowItemProps {
  leadingIcon?: React.ReactNode;
  label: string;
  description?: string;
  value?: string | React.ReactNode;
  showChevron?: boolean;
  onClick?: () => void;
  disabled?: boolean;
  className?: string;
}

export function FormRowItem({
  leadingIcon,
  label,
  description,
  value,
  showChevron = true,
  onClick,
  disabled,
  className,
}: FormRowItemProps) {
  const Component = onClick ? "button" : "div";

  return (
    <Component
      type={onClick ? "button" : undefined}
      onClick={disabled ? undefined : onClick}
      disabled={disabled}
      className={cn(
        "w-full flex items-center justify-between py-3.5 px-2 text-left transition-colors outline-none",
        onClick && !disabled && "hover:bg-[#f8f9fa] active:bg-[#f0f1f4] cursor-pointer rounded-xl",
        disabled && "opacity-50 cursor-not-allowed",
        className
      )}
    >
      <div className="flex items-center gap-3 min-w-0 pr-2">
        {leadingIcon && (
          <div className="w-8 h-8 rounded-full bg-[#eef0f4] flex items-center justify-center shrink-0 text-[#29495a]">
            {leadingIcon}
          </div>
        )}
        <div className="flex flex-col min-w-0">
          <span className="text-[15px] font-semibold text-[#111111] truncate">{label}</span>
          {description && <span className="text-xs text-[#8e8e93] leading-tight truncate">{description}</span>}
        </div>
      </div>

      <div className="flex items-center gap-1.5 shrink-0">
        {value && <span className="text-[14px] text-[#6c6c70] font-medium">{value}</span>}
        {showChevron && <ChevronRight className="w-4 h-4 text-[#c7c7cc]" />}
      </div>
    </Component>
  );
}
