import * as React from "react";
import { cn } from "../../lib/utils";

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "secondary" | "destructive" | "outline" | "success" | "warning";
}

export function Badge({ className, variant = "default", ...props }: BadgeProps) {
  const variantStyles: Record<string, string> = {
    default: "bg-[#111111] text-white",
    secondary: "bg-[#eef0f4] text-[#111111]",
    destructive: "bg-[#ffebee] text-[#9f1d1d]",
    outline: "border border-[#e8e8ec] text-[#8e8e93]",
    success: "bg-[#f1f1ee] text-[#111111]",
    warning: "bg-[#fff3e0] text-[#6b4f00]",
  };

  return (
    <div
      className={cn(
        "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold transition-colors select-none",
        variantStyles[variant],
        className
      )}
      {...props}
    />
  );
}
