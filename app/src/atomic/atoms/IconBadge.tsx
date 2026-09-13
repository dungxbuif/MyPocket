import type { LucideIcon } from "lucide-react";
import { ICON_BADGE_SHAPES, ICON_BADGE_SIZES, ICON_BADGE_TONES } from "./tokens";
export function IconBadge({ icon: Icon, size = "md", shape = "rounded", tone = "brand", className = "" }: { icon: LucideIcon; size?: keyof typeof ICON_BADGE_SIZES; shape?: keyof typeof ICON_BADGE_SHAPES; tone?: keyof typeof ICON_BADGE_TONES; className?: string }) {
  const iconSize = size === "sm" ? 16 : size === "lg" ? 22 : 18;
  return <span className={`grid shrink-0 place-items-center ${ICON_BADGE_SIZES[size]} ${ICON_BADGE_SHAPES[shape]} ${ICON_BADGE_TONES[tone]} ${className}`}><Icon size={iconSize} /></span>;
}
