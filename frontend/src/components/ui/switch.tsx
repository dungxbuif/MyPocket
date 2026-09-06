import * as React from "react";
import { cn } from "../../lib/utils";

export interface SwitchProps {
  checked: boolean;
  onCheckedChange: (checked: boolean) => void;
  disabled?: boolean;
  className?: string;
  id?: string;
}

export function Switch({ checked, onCheckedChange, disabled, className, id }: SwitchProps) {
  return (
    <button
      id={id}
      type="button"
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      onClick={() => !disabled && onCheckedChange(!checked)}
      className={cn(
        "relative inline-flex h-[31px] w-[51px] shrink-0 cursor-pointer rounded-full transition-colors duration-200 ease-in-out focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
        checked ? "bg-[#2dbd4f]" : "bg-[#e5e5ea]",
        className
      )}
    >
      <span
        className={cn(
          "pointer-events-none inline-block h-[27px] w-[27px] rounded-full bg-white shadow-md transform transition duration-200 ease-in-out mt-[2px] ml-[2px]",
          checked ? "translate-x-[20px]" : "translate-x-0"
        )}
      />
    </button>
  );
}
