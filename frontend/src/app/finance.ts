import { apiFetch } from "./apiClient";
import {
  queueCategoryArchive,
  queueCategoryCreate,
  queueCategoryUpdate,
  queueTransactionArchive,
  queueTransactionCreate,
  queueTransactionUpdate,
  queueWalletArchive,
  queueWalletCategoryActive,
  queueWalletCreate,
  queueWalletDefaultAI,
  queueWalletUpdate,
} from "../offline/outbox";

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
  version: number;
};

export type WalletCategorySetting = CategorySummary & { active: boolean };

export type TransactionType = "income" | "expense" | "transfer" | "adjustment";

export type Transaction = {
  id: string;
  type: TransactionType;
  source_wallet_id: string;
  destination_wallet_id?: string;
  category_id?: string;
  receipt_object_id?: string;
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
  receipt_object_id?: string;
  amount_vnd: number;
  target_balance_vnd?: number | null;
  occurred_at: string;
  note?: string;
  with_person?: string;
  event_ref?: string;
  excluded_from_reports?: boolean;
  base_version?: number;
};

export type WalletCreateInput = { name: string; type: WalletType };
export type WalletUpdateInput = { name: string; include_in_total: boolean; base_version?: number; current_wallet?: WalletSummary };
export type CategoryCreateInput = { kind: CategorySummary["kind"]; name: string; parent_id?: string };
export type CategoryUpdateInput = { name: string; parent_id: string | null; base_version?: number; current_category?: CategorySummary };

export async function loadWallets() {
  const response = await apiFetch<{ wallets: WalletSummary[] }>("/api/v1/wallets");
  return response.wallets;
}

export async function loadCategories() {
  const response = await apiFetch<{ categories: CategorySummary[] }>("/api/v1/categories");
  return response.categories;
}

export async function loadWalletCategorySettings(walletID: string) {
  const response = await apiFetch<{ categories: WalletCategorySetting[] }>(`/api/v1/wallets/${encodeURIComponent(walletID)}/category-settings`);
  return response.categories;
}

export async function loadTransactions(filters: TransactionFilters = {}) {
  const query = filters.query ? `?q=${encodeURIComponent(filters.query)}` : "";
  const response = await apiFetch<{ transactions: Transaction[] }>(`/api/v1/transactions${query}`);
  return response.transactions;
}

export async function createWallet(input: WalletCreateInput) {
  if (!navigator.onLine) return queueWalletCreate(input);
  const response = await apiFetch<{ wallet: WalletSummary }>("/api/v1/wallets", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.wallet;
}

export async function updateWallet(id: string, input: WalletUpdateInput) {
  if (!navigator.onLine && input.current_wallet) return queueWalletUpdate(input.current_wallet, input);
  const response = await apiFetch<{ wallet: WalletSummary }>(`/api/v1/wallets/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name: input.name, include_in_total: input.include_in_total, base_version: input.base_version ?? input.current_wallet?.version ?? 0 }),
  });
  return response.wallet;
}

export async function archiveWallet(id: string, baseVersion = 0) {
  if (!navigator.onLine) return queueWalletArchive(id, baseVersion);
  await apiFetch(`/api/v1/wallets/${encodeURIComponent(id)}/archive`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ base_version: baseVersion }),
  });
}

export async function setDefaultAIWallet(id: string, wallets: WalletSummary[] = [], baseVersion = 0) {
  if (!navigator.onLine) return queueWalletDefaultAI(id, wallets, baseVersion);
  await apiFetch(`/api/v1/wallets/${encodeURIComponent(id)}/default-ai`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ base_version: baseVersion }),
  });
}

export async function createCategory(input: CategoryCreateInput) {
  if (!navigator.onLine) return queueCategoryCreate(input);
  const response = await apiFetch<{ category: CategorySummary }>("/api/v1/categories", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.category;
}

export async function updateCategory(id: string, input: CategoryUpdateInput) {
  if (!navigator.onLine && input.current_category) return queueCategoryUpdate(input.current_category, input);
  const response = await apiFetch<{ category: CategorySummary }>(`/api/v1/categories/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name: input.name, parent_id: input.parent_id ?? null, base_version: input.base_version ?? input.current_category?.version ?? 0 }),
  });
  return response.category;
}

export async function archiveCategory(id: string, baseVersion: number) {
  if (!navigator.onLine) return queueCategoryArchive(id, baseVersion);
  await apiFetch(`/api/v1/categories/${encodeURIComponent(id)}/archive`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ base_version: baseVersion }) });
}

export async function setWalletCategoryActive(walletID: string, categoryID: string, active: boolean) {
  if (!navigator.onLine) return queueWalletCategoryActive(walletID, categoryID, active);
  await apiFetch(`/api/v1/wallets/${encodeURIComponent(walletID)}/categories/${encodeURIComponent(categoryID)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ active }),
  });
}

export async function createTransaction(input: TransactionInput) {
  if (!navigator.onLine) return queueTransactionCreate(input);
  const idempotencyKey = globalThis.crypto?.randomUUID?.() ?? `web-${Date.now()}`;
  const { base_version: _baseVersion, ...body } = input;
  const response = await apiFetch<{ transaction: Transaction }>("/api/v1/transactions", {
    method: "POST",
    headers: { "Content-Type": "application/json", "Idempotency-Key": idempotencyKey },
    body: JSON.stringify(body),
  });
  return response.transaction;
}

export async function updateTransaction(id: string, input: TransactionInput) {
  if (!navigator.onLine) return queueTransactionUpdate(id, input, input.base_version ?? 0);
  const response = await apiFetch<{ transaction: Transaction }>(`/api/v1/transactions/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.transaction;
}

export async function archiveTransaction(id: string, baseVersion = 0) {
  if (!navigator.onLine) {
    await queueTransactionArchive(id, baseVersion);
    return;
  }
  await apiFetch(`/api/v1/transactions/${encodeURIComponent(id)}/archive`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ base_version: baseVersion }),
  });
}
