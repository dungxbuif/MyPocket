import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { SURFACE_CARD_ELEVATIONS, SURFACE_CARD_PADDING, SURFACE_CARD_RADIUS } from "./tokens";

const SURFACE_CARD_BASE = "border border-line bg-card";

export function SurfaceCard({ children, className = "", radius = "lg", padding = "none", elevation = "subtle", tone = "default", ...props }: ComponentPropsWithoutRef<"section"> & { children: ReactNode; radius?: keyof typeof SURFACE_CARD_RADIUS; padding?: keyof typeof SURFACE_CARD_PADDING; elevation?: keyof typeof SURFACE_CARD_ELEVATIONS; tone?: "default" | "muted" | "danger" }) {
  const surface = tone === "muted" ? "border border-line bg-row" : tone === "danger" ? "border border-danger-line bg-danger-soft text-danger" : SURFACE_CARD_BASE;
  return <section {...props} className={`${surface} ${SURFACE_CARD_RADIUS[radius]} ${SURFACE_CARD_PADDING[padding]} ${SURFACE_CARD_ELEVATIONS[elevation]} ${className}`}>{children}</section>;
}
