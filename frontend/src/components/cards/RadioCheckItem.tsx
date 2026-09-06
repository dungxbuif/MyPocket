import * as React from "react";
import { cn } from "../../lib/utils";
import { Check } from "lucide-react";

export interface RadioCheckItemProps {
  leadingIcon?: React.ReactNode;
  label: string;
  sublabel?: string;
  selected: boolean;
  onSelect: () => void;
  className?: string;
}

export function RadioCheckItem({
  leadingIcon,
  label,
  sublabel,
  selected,
  onSelect,
  className,
}: RadioCheckItemProps) {
  return (
    <button
      type="button"
      onClick={onSelect}
      className={cn(
        "w-full flex items-center justify-between py-3.5 px-2 text-left hover:bg-[#f8f9fa] active:bg-[#f0f1f4] rounded-xl transition-colors outline-none",
        className
      )}
    >
      <div className="flex items-center gap-3">
        {leadingIcon && (
          <div className="w-8 h-8 rounded-full bg-[#eef0f4] flex items-center justify-center text-[#29495a]">
            {leadingIcon}
          </div>
        )}
        <div className="flex flex-col">
          <span className="text-[15px] font-semibold text-[#111111]">{label}</span>
          {sublabel && <span className="text-xs text-[#8e8e93]">{sublabel}</span>}
        </div>
      </div>
      {selected ? <Check className="w-5 h-5 text-[#2dbd4f] stroke-[2.5]" /> : null}
    </button>
  );
}
