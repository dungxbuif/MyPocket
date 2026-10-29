import type { ButtonHTMLAttributes, ReactNode } from "react";
import { BASE_COMPONENT_RADIUS, BUTTON_SIZES, BUTTON_VARIANTS, UI_CLASSES } from "./tokens";
export function BaseButton({ children, variant = "primary", size = "md", loading = false, loadingLabel = "Đang xử lý...", className = "", disabled, ...props }: ButtonHTMLAttributes<HTMLButtonElement> & { children: ReactNode; variant?: keyof typeof BUTTON_VARIANTS; size?: keyof typeof BUTTON_SIZES; loading?: boolean; loadingLabel?: string }) {
  return <button {...props} type={props.type ?? "button"} disabled={disabled || loading} aria-busy={loading || undefined} className={`${UI_CLASSES.interactive} ${UI_CLASSES.focus} inline-flex items-center gap-2 ${variant === "row" ? `justify-start ${BASE_COMPONENT_RADIUS}` : "justify-center rounded-full font-semibold"} disabled:cursor-not-allowed disabled:opacity-50 ${BUTTON_SIZES[size]} ${BUTTON_VARIANTS[variant]} ${className}`}>{loading ? loadingLabel : children}</button>;
}
