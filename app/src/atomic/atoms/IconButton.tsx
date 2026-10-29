import type { ButtonHTMLAttributes, ReactNode } from "react";
import { BASE_COMPONENT_RADIUS, ICON_BUTTON_VARIANTS } from "./tokens";

export function IconButton({ label, children, className = "", disabled = false, variant = "surface", shape = "circle", selected = false, ...props }: ButtonHTMLAttributes<HTMLButtonElement> & { label: string; children: ReactNode; variant?: keyof typeof ICON_BUTTON_VARIANTS; shape?: "circle" | "control"; selected?: boolean }) {
  return (
    <button
      type="button"
      aria-label={label}
      disabled={disabled}
      {...props}
      aria-pressed={selected || undefined}
      className={`grid min-h-11 min-w-11 cursor-pointer place-items-center ${shape === "circle" ? "rounded-full" : BASE_COMPONENT_RADIUS} text-secondary transition enabled:active:scale-95 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-action disabled:cursor-not-allowed disabled:opacity-50 ${selected ? "ring-2 ring-accent ring-offset-2" : ""} ${ICON_BUTTON_VARIANTS[variant]} ${className}`}
    >
      {children}
    </button>
  );
}
