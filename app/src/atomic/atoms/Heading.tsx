import type { HTMLAttributes, ReactNode } from "react";
import { HEADING_SIZES } from "./tokens";

export function Heading({ children, as: Component = "h2", size = "section", className = "", ...props }: HTMLAttributes<HTMLElement> & { children: ReactNode; as?: "h1" | "h2" | "h3" | "span"; size?: keyof typeof HEADING_SIZES }) {
  return <Component {...props} className={`${HEADING_SIZES[size]} ${className}`}>{children}</Component>;
}
