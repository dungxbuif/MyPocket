import { apiFetch } from "./apiClient";
import { queueTransaction } from "./outbox";

export type WalletType = "cash" | "bank" | "credit" | "e_wallet" | "savings" | "debt";

export type WalletSummary = {
  id: string;
  name: string;
  type: WalletType;
  balance_vnd: number;
  include_in_total: boolean;
  is_default_ai: boolean;
  version: number;
};

export type CategorySummary = {
  id: string;
  parent_id?: string;
  kind: "expense" | "income" | "debt";
  name: string;
  system_key?: string;
  is_system: boolean;
};

export type TransactionType = "income" | "expense" | "transfer" | "adjustment";

export type Transaction = {
  id: string;
  type: TransactionType;
  source_wallet_id: string;
  destination_wallet_id?: string;
  category_id?: string;
  amount_vnd: number;
  balance_after_vnd: number;
  occurred_at: string;
  note: string;
  with_person: string;
  event_ref: string;
  excluded_from_reports: boolean;
  version: number;
};

export type TransactionFilters = { query?: string };

export async function loadWallets() {
  const response = await apiFetch<{ wallets: WalletSummary[] }>("/api/v1/wallets");
  return response.wallets;
}

export async function loadCategories() {
  const response = await apiFetch<{ categories: CategorySummary[] }>("/api/v1/categories");
  return response.categories;
}

export async function loadTransactions(filters: TransactionFilters = {}) {
  const query = filters.query ? `?q=${encodeURIComponent(filters.query)}` : "";
  const response = await apiFetch<{ transactions: Transaction[] }>(`/api/v1/transactions${query}`);
  return response.transactions;
}

export async function createTransaction(input: {
  type: TransactionType;
  source_wallet_id: string;
  category_id?: string;
  amount_vnd: number;
  occurred_at: string;
  note?: string;
}) {
  if (!navigator.onLine) return queueTransaction(input);
  const idempotencyKey = globalThis.crypto?.randomUUID?.() ?? `web-${Date.now()}`;
  const response = await apiFetch<{ transaction: Transaction }>("/api/v1/transactions", {
    method: "POST",
    headers: { "Content-Type": "application/json", "Idempotency-Key": idempotencyKey },
    body: JSON.stringify(input),
  });
  return response.transaction;
}
