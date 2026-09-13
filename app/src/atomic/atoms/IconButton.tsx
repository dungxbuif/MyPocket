import type { ButtonHTMLAttributes, ReactNode } from "react";
import { BASE_COMPONENT_RADIUS, ICON_BUTTON_VARIANTS } from "./tokens";

export function IconButton({ label, children, className = "", disabled = false, variant = "surface", ...props }: ButtonHTMLAttributes<HTMLButtonElement> & { label: string; children: ReactNode; variant?: keyof typeof ICON_BUTTON_VARIANTS }) {
  return (
    <button
      type="button"
      aria-label={label}
      disabled={disabled}
      {...props}
      className={`grid h-10 w-10 cursor-pointer place-items-center ${BASE_COMPONENT_RADIUS} text-[#3f4a3c] transition active:scale-95 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#006e1c] disabled:cursor-not-allowed disabled:opacity-50 ${ICON_BUTTON_VARIANTS[variant]} ${className}`}
    >
      {children}
    </button>
  );
}
