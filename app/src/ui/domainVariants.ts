import type { BadgeTone } from "./variants";
export const WALLET_BADGE_TONES = { cash: "warning", bank: "categoryRed", credit: "categoryTeal", goal: "success" } as const satisfies Record<string, BadgeTone>;
export const TRANSACTION_BADGE_TONES = { expense: "categoryOrange", income: "success", transfer: "neutral", debt: "warning" } as const satisfies Record<string, BadgeTone>;
