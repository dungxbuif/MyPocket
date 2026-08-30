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

export type TransactionInput = {
  type: TransactionType;
  source_wallet_id: string;
  destination_wallet_id?: string;
  category_id?: string;
  amount_vnd: number;
  target_balance_vnd?: number | null;
  occurred_at: string;
  note?: string;
  with_person?: string;
  event_ref?: string;
  excluded_from_reports?: boolean;
};

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

export async function createWallet(input: { name: string; type: WalletType }) {
  const response = await apiFetch<{ wallet: WalletSummary }>("/api/v1/wallets", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.wallet;
}

export async function updateWallet(id: string, input: { name: string; include_in_total: boolean }) {
  const response = await apiFetch<{ wallet: WalletSummary }>(`/api/v1/wallets/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.wallet;
}

export async function archiveWallet(id: string) {
  await apiFetch(`/api/v1/wallets/${encodeURIComponent(id)}/archive`, { method: "POST" });
}

export async function setDefaultAIWallet(id: string) {
  await apiFetch(`/api/v1/wallets/${encodeURIComponent(id)}/default-ai`, { method: "POST" });
}

export async function createCategory(input: { kind: CategorySummary["kind"]; name: string }) {
  const response = await apiFetch<{ category: CategorySummary }>("/api/v1/categories", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.category;
}

export async function updateCategory(id: string, input: { name: string }) {
  const response = await apiFetch<{ category: CategorySummary }>(`/api/v1/categories/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.category;
}

export async function archiveCategory(id: string) {
  await apiFetch(`/api/v1/categories/${encodeURIComponent(id)}/archive`, { method: "POST" });
}

export async function setWalletCategoryActive(walletID: string, categoryID: string, active: boolean) {
  await apiFetch(`/api/v1/wallets/${encodeURIComponent(walletID)}/categories/${encodeURIComponent(categoryID)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ active }),
  });
}

export async function createTransaction(input: TransactionInput) {
  if (!navigator.onLine) return queueTransaction(input);
  const idempotencyKey = globalThis.crypto?.randomUUID?.() ?? `web-${Date.now()}`;
  const response = await apiFetch<{ transaction: Transaction }>("/api/v1/transactions", {
    method: "POST",
    headers: { "Content-Type": "application/json", "Idempotency-Key": idempotencyKey },
    body: JSON.stringify(input),
  });
  return response.transaction;
}

export async function updateTransaction(id: string, input: TransactionInput) {
  const response = await apiFetch<{ transaction: Transaction }>(`/api/v1/transactions/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.transaction;
}

export async function archiveTransaction(id: string) {
  await apiFetch(`/api/v1/transactions/${encodeURIComponent(id)}/archive`, { method: "POST" });
}
