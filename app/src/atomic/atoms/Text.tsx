import type { HTMLAttributes } from "react";

const SIZES = { micro: "text-[10px]", tiny: "text-[11px]", xs: "text-xs", sm: "text-sm", base: "text-base", lg: "text-lg", xl: "text-xl", "2xl": "text-2xl", "3xl": "text-3xl", "4xl": "text-4xl" } as const;
const WEIGHTS = { normal: "font-normal", medium: "font-medium", semibold: "font-semibold", bold: "font-bold", extrabold: "font-extrabold" } as const;
const TONES = { ink: "text-ink", heading: "text-heading", secondary: "text-secondary", muted: "text-muted", action: "text-action", danger: "text-danger", card: "text-card", inherit: "text-inherit" } as const;

export function Text({ as: Tag = "p", size = "sm", weight = "normal", tone = "ink", numeric = false, className = "", ...props }: HTMLAttributes<HTMLElement> & { as?: "p" | "span" | "h1" | "h2" | "h3" | "h4" | "strong" | "small"; size?: keyof typeof SIZES; weight?: keyof typeof WEIGHTS; tone?: keyof typeof TONES; numeric?: boolean }) {
  return <Tag {...props} className={`${SIZES[size]} ${WEIGHTS[weight]} ${TONES[tone]} ${numeric ? "money" : ""} ${className}`} />;
}
