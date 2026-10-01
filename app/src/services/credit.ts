import { apiRequest } from "./api";
import { getStoredToken } from "./auth";
import type { Transaction } from "./transactions";

export type CreditKind = "purchase" | "refund" | "fee" | "interest";
export type CreditEntryInput = { kind: CreditKind; category_id?: string; amount: number; occurred_at?: string; note?: string };
export type CreditPaymentInput = { source_wallet_id: string; amount: number; occurred_at?: string; note?: string };
export type CreditStatement = { wallet_id: string; credit_limit: number; balance: number; available_credit: number; items: Transaction[] };

function path(walletID: string, suffix: string) { return `/api/v1/credit/wallets/${walletID}/${suffix}`; }

export function fetchCreditStatement(walletID: string, from?: string, to?: string): Promise<CreditStatement> {
  const params = new URLSearchParams();
  if (from) params.set("from", from);
  if (to) params.set("to", to);
  const query = params.toString();
  return apiRequest<CreditStatement>(`${path(walletID, "statement")}${query ? `?${query}` : ""}`, {}, getStoredToken());
}

export function createCreditEntry(walletID: string, input: CreditEntryInput): Promise<Transaction> {
  return apiRequest<Transaction>(path(walletID, "entries"), { method: "POST", body: JSON.stringify(input) }, getStoredToken());
}

export function createCreditPayment(walletID: string, input: CreditPaymentInput): Promise<Transaction[]> {
  return apiRequest<Transaction[]>(path(walletID, "payments"), { method: "POST", body: JSON.stringify(input) }, getStoredToken());
}
