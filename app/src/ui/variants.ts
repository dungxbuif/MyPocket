export const UI_CLASSES = {
  focus: "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-action",
  interactive: "cursor-pointer transition-colors duration-150 disabled:cursor-not-allowed disabled:opacity-50",
  stack: "flex flex-col gap-3", row: "flex items-center gap-3",
  title: "text-base font-bold text-heading", body: "text-sm text-secondary",
  amount: "money text-sm font-bold tracking-tight",
} as const;
export const SURFACE_CARD_BASE = "border border-line bg-card";
export const SURFACE_CARD_RADIUS = { sm: "rounded-control", md: "rounded-badge", lg: "rounded-card" } as const;
export const SURFACE_CARD_PADDING = { none: "p-0", sm: "p-3", md: "p-4", lg: "p-5" } as const;
export const SURFACE_CARD_ELEVATIONS = { flat: "shadow-none", subtle: "shadow-card", raised: "shadow-raised" } as const;
export const BUTTON_SIZES = { sm: "min-h-11 px-3 text-xs", md: "min-h-12 px-4 text-sm", lg: "min-h-14 px-5 text-base" } as const;
export const BUTTON_VARIANTS = {
  primary: "bg-brand text-card enabled:hover:bg-brand-hover",
  secondary: "bg-success-soft text-action enabled:hover:bg-success-line",
  ghost: "bg-transparent text-action enabled:hover:bg-row",
  danger: "bg-danger text-card enabled:hover:bg-danger-hover",
} as const;
export const ICON_BADGE_SIZES = { sm: "h-8 w-8", md: "h-10 w-10", lg: "h-12 w-12" } as const;
export const ICON_BADGE_ICON_SIZES = { sm: 16, md: 20, lg: 24 } as const;
export const ICON_BADGE_SHAPES = { rounded: "rounded-badge", compact: "rounded-control", circle: "rounded-full" } as const;
export const ICON_BADGE_TONES = {
  brand: "bg-success-soft text-action border-success-line",
  success: "bg-success-soft text-action border-success-line",
  neutral: "bg-row text-secondary border-line",
  danger: "bg-danger-soft text-danger border-danger-line",
  warning: "bg-warning-soft text-warning border-warning-line",
  teal: "bg-teal-soft text-teal border-teal-line",
  red: "bg-red-soft text-red border-red-line",
  orange: "bg-orange-soft text-orange border-orange-line",
} as const;
export type BadgeTone = keyof typeof ICON_BADGE_TONES;
export const ICON_BUTTON_VARIANTS = {
  surface: "border border-line bg-card text-secondary shadow-sm enabled:hover:bg-row",
  ghost: "border border-transparent bg-transparent text-muted enabled:hover:text-ink",
} as const;
export const SEGMENT_CLASSES = {
  group: "flex rounded-full bg-line p-1", item: "min-h-11 flex-1 rounded-full px-3 text-sm font-semibold",
  selected: "bg-card text-action shadow-sm", idle: "bg-transparent text-secondary",
} as const;
export const METRIC_CLASSES = {
  neutral: "text-heading", success: "text-action", danger: "text-danger",
  body: "flex h-full flex-col justify-between gap-2", header: "flex items-center justify-between gap-2",
  label: "text-[11px] font-medium text-muted", value: "money text-base font-extrabold",
  footnote: "text-[10px] text-muted", badge: "rounded-full bg-danger-line px-2 py-0.5 text-[11px] font-bold text-danger",
} as const;
export const PAGE_CLASSES = {
  root: "min-h-screen bg-canvas px-4 py-6 text-ink",
  content: "mx-auto flex w-full max-w-phone flex-col gap-4",
} as const;
export const LIST_CLASSES = {
  row: "flex w-full items-center gap-3 rounded-control px-2 py-3.5 text-left",
  interactive: "hover:bg-row", content: "min-w-0 flex-1",
  title: "truncate text-sm font-semibold text-ink", subtitle: "truncate text-[11px] text-muted",
  divider: "divide-y divide-row", header: "flex items-center justify-between gap-3 border-b border-line pb-3",
} as const;
