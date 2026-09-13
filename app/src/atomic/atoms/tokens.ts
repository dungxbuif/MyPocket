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

export const BASE_COMPONENT_RADIUS = "rounded-xl";

export const SURFACE_CARD_RADIUS = {
  sm: BASE_COMPONENT_RADIUS,
  md: BASE_COMPONENT_RADIUS,
  lg: BASE_COMPONENT_RADIUS,
} as const;

export const ICON_BADGE_SIZES = {
  xs: "h-7 w-7",
  sm: "h-8 w-8",
  md: "h-10 w-10",
  lg: "h-12 w-12",
} as const;

export const ICON_BADGE_SHAPES = {
  rounded: BASE_COMPONENT_RADIUS,
  circle: "rounded-full",
} as const;

export const ICON_BADGE_TONES = {
  brand: "bg-[#ecfdf5] text-[#059669] border border-[#d1fae5]",
  neutral: "bg-[#f5f3f3] text-[#556158] border border-[#e3e2e2]",
  success: "bg-[#d9e6da] text-[#006e1c] border border-[#bdcabe]",
  danger: "bg-[#ffdad6] text-[#93000a] border border-[#ffb4a9]",
  warning: "bg-[#fff4e5] text-[#bb5b00] border border-[#ffd8a8]",
  categoryOrange: "bg-orange-50 text-orange-600 border border-orange-100",
  categoryBlue: "bg-blue-50 text-blue-600 border border-blue-100",
  categoryViolet: "bg-violet-50 text-violet-600 border border-violet-100",
  categoryRose: "bg-rose-50 text-rose-600 border border-rose-100",
  categorySky: "bg-sky-50 text-sky-600 border border-sky-100",
  categoryRed: "bg-red-50 text-red-600 border border-red-100",
  categoryIndigo: "bg-indigo-50 text-indigo-600 border border-indigo-100",
  categoryFuchsia: "bg-fuchsia-50 text-fuchsia-600 border border-fuchsia-100",
  categoryPink: "bg-pink-50 text-pink-600 border border-pink-100",
  categoryTeal: "bg-teal-50 text-teal-600 border border-teal-100",
  categoryEmerald: "bg-emerald-50 text-emerald-600 border border-emerald-100",
  categoryCyan: "bg-cyan-50 text-cyan-600 border border-cyan-100",
  categoryLime: "bg-lime-50 text-lime-700 border border-lime-100",
  categoryAmber: "bg-amber-50 text-amber-600 border border-amber-100",
  categorySlate: "bg-slate-50 text-slate-600 border border-slate-200",
  categoryIncome: "bg-emerald-50 text-emerald-700 border border-emerald-100",
  categoryDebt: "bg-sky-50 text-sky-700 border border-sky-100",
} as const;

export const BUTTON_VARIANTS = {
  primary: "bg-[#006e1c] text-white hover:bg-[#005313]",
  secondary: "bg-[#d9e6da] text-[#006e1c] hover:bg-[#bdcabe]",
  ghost: "bg-transparent text-[#006e1c] hover:bg-[#f5f3f3]",
  danger: "bg-[#ba1a1a] text-white hover:bg-[#93000a]",
  outline: "border border-slate-100 bg-white text-emerald-600 shadow-[0_4px_20px_rgb(0_0_0/0.03)] hover:bg-emerald-50",
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

export const ICON_BUTTON_VARIANTS = {
  surface: "border border-slate-100 bg-white shadow-sm hover:bg-[#f5f3f3]",
  bare: "border border-transparent bg-transparent shadow-none hover:bg-transparent active:scale-100",
} as const;

export const HEADING_SIZES = {
  screen: "text-xl font-extrabold tracking-tight text-slate-900",
  section: "text-base font-bold text-slate-800",
  field: "text-sm font-semibold text-slate-800",
} as const;

export const FORM_CONTROL_CLASS = `mt-1.5 w-full ${BASE_COMPONENT_RADIUS} border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-900 outline-none transition focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100 disabled:cursor-not-allowed disabled:bg-slate-50 disabled:text-slate-400`;

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

export const CATEGORY_TREE_LAYOUTS = {
  nested: {
    children: "space-y-0.5 pb-1 pt-1.5 pl-12 pr-0",
    connector: "absolute bottom-5 left-[37px] top-[70px] w-0.5 rounded-full bg-slate-200",
    branch: "relative before:pointer-events-none before:absolute before:-left-[18px] before:top-1/2 before:h-3 before:w-5 before:-translate-y-[90%] before:rounded-bl-[7px] before:border-b-2 before:border-l-2 before:border-slate-200",
  },
  line: {
    children: "space-y-1 pb-1 pt-1.5 pl-12 pr-0",
    connector: "absolute bottom-5 left-[37px] top-[70px] w-0.5 rounded-full bg-slate-200",
    branch: "relative before:pointer-events-none before:absolute before:-left-[18px] before:top-1/2 before:h-3 before:w-5 before:-translate-y-[90%] before:rounded-bl-[7px] before:border-b-2 before:border-l-2 before:border-slate-200",
  },
} as const;

export const PROFILE_CARD_CLASSES = {
  card: "relative overflow-hidden rounded-2xl border border-slate-100 bg-white p-5 text-center shadow-[0_2px_10px_rgb(0_0_0/0.03)]",
  avatar: "grid h-20 w-20 place-items-center rounded-full border-2 border-white bg-[#48a855] text-4xl font-bold text-white shadow-inner ring-4 ring-[#ecfdf5]",
  row: "mt-4 flex w-full cursor-pointer items-center justify-between border-t border-[#e3e2e2] pt-3 text-left transition hover:bg-[#f5f3f3]",
} as const;
