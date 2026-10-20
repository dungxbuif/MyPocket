import * as React from "react";
import { cn } from "../../lib/utils";

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "secondary" | "destructive" | "pill" | "outline" | "ghost" | "icon";
  size?: "default" | "sm" | "lg" | "icon";
  loading?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "default", loading, disabled, children, ...props }, ref) => {
    const baseStyles =
      "inline-flex items-center justify-center font-medium transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 select-none active:scale-[0.98]";

    const variantStyles: Record<string, string> = {
      default: "bg-black text-white hover:bg-neutral-800 shadow-none",
      secondary: "bg-neutral-100 text-black hover:bg-neutral-200",
      destructive: "bg-white text-[#9f1d1d] border border-[#9f1d1d] hover:bg-[#fff5f5] shadow-none",
      pill: "bg-black text-white rounded-full hover:bg-neutral-800",
      outline: "border border-neutral-300 bg-white text-black hover:bg-neutral-100",
      ghost: "text-black hover:bg-neutral-100",
      icon: "rounded-full p-2 text-neutral-600 hover:text-black hover:bg-neutral-100 active:scale-95",
    };

    const sizeStyles: Record<string, string> = {
      default: "h-12 px-5 py-2 rounded-full text-base font-semibold",
      sm: "h-9 px-3.5 py-1.5 rounded-full text-sm font-medium",
      lg: "h-14 px-8 py-3 rounded-full text-lg font-bold",
      icon: "h-10 w-10 p-0 rounded-full",
    };

    return (
      <button
        ref={ref}
        type={props.type || "button"}
        disabled={disabled || loading}
        className={cn(baseStyles, variantStyles[variant], sizeStyles[size], className)}
        {...props}
      >
        {loading ? (
          <span className="inline-block w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin mr-2" />
        ) : null}
        {children}
      </button>
    );
  }
);
Button.displayName = "Button";
