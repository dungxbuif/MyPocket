export const UI_TOKENS = {
  primary: "var(--color-action)",
  primaryDark: "var(--color-action-hover)",
  surface: "var(--color-canvas)",
  surfaceLow: "var(--color-row)",
  surfaceContainer: "var(--color-line)",
  onSurface: "var(--color-ink)",
  onVariant: "var(--color-secondary)",
  outline: "var(--color-secondary)",
  outlineVariant: "var(--color-line)",
  error: "var(--color-danger)",
} as const;

export const UI_CLASSES = {
  focus: "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-action",
  transition: "transition duration-150 ease-out",
  interactive: "cursor-pointer transition duration-150 ease-out",
} as const;

export const BASE_COMPONENT_RADIUS = "rounded-control";

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
  brand: "bg-success-soft text-action border border-success-line",
  neutral: "bg-row text-secondary border border-line",
  success: "bg-success-soft text-action border border-success-line",
  danger: "bg-danger-soft text-danger border border-danger-line",
  warning: "bg-warning-soft text-warning border border-warning-line",
  categoryOrange: "bg-orange-50 text-orange-600 border border-orange-100",
  categoryBlue: "bg-blue-50 text-blue-600 border border-blue-100",
  categoryViolet: "bg-violet-50 text-violet-600 border border-violet-100",
  categoryRose: "bg-danger-soft text-danger border border-danger-line",
  categorySky: "bg-sky-50 text-sky-600 border border-sky-100",
  categoryRed: "bg-red-50 text-red-600 border border-red-100",
  categoryIndigo: "bg-indigo-50 text-indigo-600 border border-indigo-100",
  categoryFuchsia: "bg-fuchsia-50 text-fuchsia-600 border border-fuchsia-100",
  categoryPink: "bg-pink-50 text-pink-600 border border-pink-100",
  categoryTeal: "bg-teal-50 text-teal-600 border border-teal-100",
  categoryEmerald: "bg-success-soft text-action border border-success-line",
  categoryCyan: "bg-cyan-50 text-cyan-600 border border-cyan-100",
  categoryLime: "bg-lime-50 text-lime-700 border border-lime-100",
  categoryAmber: "bg-amber-50 text-amber-600 border border-amber-100",
  categorySlate: "bg-row text-secondary border border-line",
  categoryIncome: "bg-success-soft text-action-hover border border-success-line",
  categoryDebt: "bg-sky-50 text-sky-700 border border-sky-100",
} as const;

export const BUTTON_VARIANTS = {
  primary: "bg-brand text-card enabled:hover:bg-brand-hover",
  secondary: "bg-success-soft text-action hover:bg-success-line",
  ghost: "bg-transparent text-action hover:bg-row",
  danger: "bg-danger text-card enabled:hover:bg-danger-hover",
  outline: "border border-line bg-card text-action shadow-card hover:bg-success-soft",
  row: "w-full bg-transparent text-ink text-left enabled:hover:bg-row",
  key: "bg-card text-heading shadow-control enabled:hover:bg-row",
  chip: "bg-card text-ink border border-line enabled:hover:bg-row",
} as const;

export const SURFACE_CARD_ELEVATIONS = {
  flat: "shadow-none",
  subtle: "shadow-card",
  raised: "shadow-raised",
} as const;

export const SURFACE_CARD_PADDING = {
  none: "p-0",
  sm: "p-3",
  md: "p-4",
  lg: "p-5",
} as const;

export const BUTTON_SIZES = {
  sm: "min-h-11 px-3 text-xs",
  md: "min-h-12 px-4 text-sm",
  lg: "min-h-14 px-5 text-base",
  row: "min-h-11 px-2 py-3 text-sm",
} as const;

export const ICON_BUTTON_VARIANTS = {
  surface: "border border-line bg-card shadow-sm hover:bg-row",
  bare: "border border-transparent bg-transparent shadow-none hover:bg-transparent active:scale-100",
  brand: "border border-action bg-brand text-card shadow-raised hover:bg-brand-hover",
} as const;

export const HEADING_SIZES = {
  screen: "text-xl font-extrabold tracking-tight text-heading",
  section: "text-base font-bold text-ink",
  field: "text-sm font-semibold text-ink",
} as const;

export const FORM_CONTROL_CLASS = `mt-1.5 w-full ${BASE_COMPONENT_RADIUS} border border-line bg-card px-3 py-2.5 text-sm text-heading outline-none transition focus:border-accent focus:ring-2 focus:ring-success-line disabled:cursor-not-allowed disabled:bg-row disabled:text-muted`;

export const FORM_CONTROL_VARIANTS = {
  amount: "w-full min-w-0 min-h-14 border-0 bg-transparent text-4xl font-medium text-heading outline-none focus-visible:ring-2 focus-visible:ring-success-line disabled:text-muted",
  default: FORM_CONTROL_CLASS,
  inline: "w-full min-h-11 rounded-control border border-transparent bg-transparent px-0 text-base font-medium text-ink focus-visible:outline-2 focus-visible:outline-action disabled:text-muted",
  title: "w-full min-h-11 rounded-control border border-transparent bg-transparent px-0 text-xl font-bold text-heading focus-visible:outline-2 focus-visible:outline-action disabled:text-muted",
} as const;

export const TEXTAREA_VARIANTS = {
  default: `${FORM_CONTROL_CLASS} min-h-28 resize-y`,
  composer: "min-h-24 w-full resize-y rounded-control border-0 bg-transparent px-1 py-1 text-base text-heading outline-none placeholder:text-muted focus-visible:ring-2 focus-visible:ring-success-line disabled:text-muted",
} as const;

export const CATEGORY_TREE_LAYOUTS = {
  nested: {
    children: "space-y-0.5 pb-1 pt-1.5 pl-12 pr-0",
    connector: "absolute bottom-5 left-[37px] top-[70px] w-0.5 rounded-full bg-line",
    branch: "relative before:pointer-events-none before:absolute before:-left-[18px] before:top-1/2 before:h-3 before:w-5 before:-translate-y-[90%] before:rounded-bl-[7px] before:border-b-2 before:border-l-2 before:border-line",
  },
  line: {
    children: "space-y-1 pb-1 pt-1.5 pl-12 pr-0",
    connector: "absolute bottom-5 left-[37px] top-[70px] w-0.5 rounded-full bg-line",
    branch: "relative before:pointer-events-none before:absolute before:-left-[18px] before:top-1/2 before:h-3 before:w-5 before:-translate-y-[90%] before:rounded-bl-[7px] before:border-b-2 before:border-l-2 before:border-line",
  },
} as const;

export const PROFILE_CARD_CLASSES = {
  card: "relative overflow-hidden p-5 text-center",
  avatar: "grid h-20 w-20 place-items-center rounded-full border-2 border-card bg-brand text-4xl font-bold text-card shadow-inner ring-4 ring-success-soft",
  row: "mt-4 flex w-full cursor-pointer items-center justify-between border-t border-line pt-3 text-left transition hover:bg-row",
} as const;

export type BadgeTone = keyof typeof ICON_BADGE_TONES;
