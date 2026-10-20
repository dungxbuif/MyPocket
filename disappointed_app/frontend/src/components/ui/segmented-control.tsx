import * as React from "react";
import { cn } from "../../lib/utils";

export interface SegmentOption<T extends string = string> {
  key: T;
  label: string;
}

export interface SegmentedControlProps<T extends string = string> {
  options: SegmentOption<T>[];
  selectedKey: T;
  onChange: (key: T) => void;
  className?: string;
  size?: "sm" | "md" | "lg";
}

export function SegmentedControl<T extends string = string>({
  options,
  selectedKey,
  onChange,
  className,
  size = "md",
}: SegmentedControlProps<T>) {
  const sizeStyles = {
    sm: "h-8 p-0.5 text-xs",
    md: "h-10 p-1 text-sm",
    lg: "h-12 p-1.5 text-base",
  };

  return (
    <div
      role="tablist"
      className={cn(
        "flex w-full items-center justify-between rounded-full bg-[#e5e6eb] select-none",
        sizeStyles[size],
        className
      )}
    >
      {options.map((opt) => {
        const isSelected = opt.key === selectedKey;
        return (
          <button
            key={opt.key}
            type="button"
            role="tab"
            aria-selected={isSelected}
            onClick={() => onChange(opt.key)}
            className={cn(
              "flex-1 h-full rounded-full font-semibold transition-all duration-200 flex items-center justify-center outline-none",
              isSelected
                ? "bg-white text-[#111111] shadow-[0_2px_6px_rgba(0,0,0,0.08)]"
                : "text-[#6c6c70] hover:text-[#111111]"
            )}
          >
            {opt.label}
          </button>
        );
      })}
    </div>
  );
}
