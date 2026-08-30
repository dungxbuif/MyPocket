import { apiFetch } from "./apiClient";

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

export async function loadWallets() {
  const response = await apiFetch<{ wallets: WalletSummary[] }>("/api/v1/wallets");
  return response.wallets;
}

export async function loadCategories() {
  const response = await apiFetch<{ categories: CategorySummary[] }>("/api/v1/categories");
  return response.categories;
}
