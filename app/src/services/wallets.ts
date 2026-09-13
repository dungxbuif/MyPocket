import { getStoredToken } from "./auth";
import { apiRequest } from "./api";

export const WALLET_TYPES = {
  basic: "basic",
  goal: "goal",
  credit: "credit",
} as const;

export type WalletType = keyof typeof WALLET_TYPES;

export type Wallet = {
  id: string;
  name: string;
  type: WalletType;
  currency: "VND";
  opening_balance: number;
  current_balance: number;
  is_in_total: boolean;
  description?: string | null;
  target_amount?: number | null;
  credit_limit?: number | null;
};

export type WalletInput = {
  name: string;
  type: WalletType;
  opening_balance: number;
  is_in_total: boolean;
  description?: string;
  target_amount?: number;
  credit_limit?: number;
};

const WALLET_API_PATH = "/api/v1/wallets";

export function fetchWallets(): Promise<Wallet[]> {
  return apiRequest<Wallet[]>(WALLET_API_PATH, {}, getStoredToken());
}

export function createWallet(input: WalletInput): Promise<Wallet> {
  return apiRequest<Wallet>(WALLET_API_PATH, { method: "POST", body: JSON.stringify(input) }, getStoredToken());
}

export function updateWallet(id: string, input: WalletInput): Promise<Wallet> {
  return apiRequest<Wallet>(`${WALLET_API_PATH}/${id}`, { method: "PATCH", body: JSON.stringify(input) }, getStoredToken());
}

export function deleteWallet(id: string): Promise<void> {
  return apiRequest<void>(`${WALLET_API_PATH}/${id}`, { method: "DELETE" }, getStoredToken());
}
