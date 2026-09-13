import type { ReactNode } from "react";
import { SURFACE_CARD_ELEVATIONS, SURFACE_CARD_PADDING, SURFACE_CARD_RADIUS } from "./tokens";

const SURFACE_CARD_BASE = "border border-slate-100 bg-white shadow-[0_4px_20px_rgb(0_0_0/0.03)]";

export function SurfaceCard({ children, className = "", radius = "lg", padding = "none", elevation = "subtle" }: { children: ReactNode; className?: string; radius?: keyof typeof SURFACE_CARD_RADIUS; padding?: keyof typeof SURFACE_CARD_PADDING; elevation?: keyof typeof SURFACE_CARD_ELEVATIONS }) {
  return <section className={`${SURFACE_CARD_BASE} ${SURFACE_CARD_RADIUS[radius]} ${SURFACE_CARD_PADDING[padding]} ${SURFACE_CARD_ELEVATIONS[elevation]} ${className}`}>{children}</section>;
}
