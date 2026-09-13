export const UI_TOKENS = {
  primary: "#006e1c",
  primaryDark: "#005313",
  surface: "#fbf9f9",
  surfaceLow: "#f5f3f3",
  surfaceContainer: "#efeded",
  onSurface: "#1b1c1c",
  onVariant: "#3f4a3c",
  outline: "#6f7a6b",
  outlineVariant: "#becab9",
  error: "#ba1a1a",
} as const;

export const UI_CLASSES = {
  focus: "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#006e1c]",
  transition: "transition duration-150 ease-out",
  interactive: "cursor-pointer transition duration-150 ease-out",
} as const;

export const SURFACE_CARD_RADIUS = {
  sm: "rounded-xl",
  md: "rounded-2xl",
  lg: "rounded-3xl",
} as const;

export const ICON_BADGE_SIZES = {
  sm: "h-8 w-8",
  md: "h-10 w-10",
  lg: "h-12 w-12",
} as const;

export const ICON_BADGE_SHAPES = {
  rounded: "rounded-2xl",
  circle: "rounded-full",
} as const;

export const ICON_BADGE_TONES = {
  brand: "bg-[#ecfdf5] text-[#059669] border border-[#d1fae5]",
  neutral: "bg-[#f5f3f3] text-[#556158] border border-[#e3e2e2]",
  success: "bg-[#d9e6da] text-[#006e1c] border border-[#bdcabe]",
  danger: "bg-[#ffdad6] text-[#93000a] border border-[#ffb4a9]",
  warning: "bg-[#fff4e5] text-[#bb5b00] border border-[#ffd8a8]",
} as const;

export const BUTTON_VARIANTS = {
  primary: "bg-[#006e1c] text-white hover:bg-[#005313]",
  secondary: "bg-[#d9e6da] text-[#006e1c] hover:bg-[#bdcabe]",
  ghost: "bg-transparent text-[#006e1c] hover:bg-[#f5f3f3]",
  danger: "bg-[#ba1a1a] text-white hover:bg-[#93000a]",
} as const;

export const SURFACE_CARD_ELEVATIONS = {
  flat: "shadow-none",
  subtle: "shadow-[0_4px_20px_rgb(0_0_0/0.03)]",
  raised: "shadow-[0_10px_25px_rgb(0_0_0/0.08)]",
} as const;

export const SURFACE_CARD_PADDING = {
  none: "p-0",
  sm: "p-3",
  md: "p-4",
  lg: "p-5",
} as const;

export const BUTTON_SIZES = {
  sm: "min-h-9 px-3 text-xs",
  md: "min-h-12 px-4 text-sm",
  lg: "min-h-14 px-5 text-base",
} as const;

export const WALLET_BADGE_CLASSES = {
  cash: "bg-amber-50 text-amber-600 border border-amber-100",
  bank: "bg-red-50 text-red-600 border border-red-100",
  credit: "bg-teal-50 text-teal-600 border border-teal-100",
  goal: "bg-emerald-50 text-emerald-600 border border-emerald-100",
} as const;

export const TRANSACTION_BADGE_CLASSES = {
  expense: "bg-orange-50 text-orange-500 border border-orange-100",
  income: "bg-emerald-50 text-emerald-600 border border-emerald-100",
  transfer: "bg-slate-50 text-slate-600 border border-slate-100",
  debt: "bg-amber-50 text-amber-500 border border-amber-100",
} as const;

export const CATEGORY_TREE_CLASSES = {
  card: "relative overflow-hidden rounded-3xl border border-[#e3e2e2] bg-white p-3 shadow-[0_4px_20px_-2px_rgb(0_0_0/0.04)]",
  parent: "relative z-10 flex cursor-pointer items-center justify-between rounded-2xl p-1.5 transition hover:bg-[#f5f3f3]",
  parentIcon: "grid h-10 w-10 shrink-0 place-items-center rounded-full bg-[#d9e6da] text-[#006e1c]",
  children: "relative ml-10 mt-1.5 space-y-0.5 border-l-2 border-[#e3e2e2] pl-4",
  child: "relative flex cursor-pointer items-center justify-between rounded-2xl px-2.5 py-1.5 transition hover:bg-[#f5f3f3]",
  childIcon: "grid h-8 w-8 shrink-0 place-items-center rounded-full bg-[#d9e6da] text-[#006e1c]",
} as const;

export const PROFILE_CARD_CLASSES = {
  card: "relative overflow-hidden rounded-3xl border border-slate-100 bg-white p-5 text-center shadow-[0_2px_10px_rgb(0_0_0/0.03)]",
  avatar: "grid h-20 w-20 place-items-center rounded-full border-2 border-white bg-[#48a855] text-4xl font-bold text-white shadow-inner ring-4 ring-[#ecfdf5]",
  row: "mt-4 flex w-full cursor-pointer items-center justify-between border-t border-[#e3e2e2] pt-3 text-left transition hover:bg-[#f5f3f3]",
} as const;
