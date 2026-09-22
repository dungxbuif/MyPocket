import { apiRequest } from "./api";
import { getStoredToken } from "./auth";

export { categoryAppliesToTransaction, signedTransactionAmount } from "./transactionLogic";

export type TransactionType = "income" | "expense";

export type Transaction = {
  id: string;
  owner_id: string;
  wallet_id: string;
  category_id?: string | null;
  jar_id?: string | null;
  jar_name?: string;
  type: TransactionType;
  amount: number;
  occurred_at: string;
  note?: string | null;
  included_in_reports: boolean;
  transfer_id?: string | null;
  created_at: string;
  updated_at: string;
};

export type TransactionInput = {
  wallet_id: string;
  category_id?: string;
  jar_id?: string | null;
  type: TransactionType;
  amount: number;
  occurred_at: string;
  note?: string;
  included_in_reports: boolean;
};

export type TransferInput = {
  source_wallet_id: string;
  destination_wallet_id: string;
  amount: number;
  occurred_at: string;
  note?: string;
};

const TRANSACTION_API_PATH = "/api/v1/transactions";

export function fetchTransactions(): Promise<Transaction[]> {
  return apiRequest<Transaction[]>(TRANSACTION_API_PATH, {}, getStoredToken());
}

export function createTransaction(input: TransactionInput): Promise<Transaction> {
  return apiRequest<Transaction>(TRANSACTION_API_PATH, { method: "POST", body: JSON.stringify(input) }, getStoredToken());
}

export function updateTransaction(id: string, input: TransactionInput): Promise<Transaction> {
  return apiRequest<Transaction>(`${TRANSACTION_API_PATH}/${id}`, { method: "PATCH", body: JSON.stringify(input) }, getStoredToken());
}

export function deleteTransaction(id: string): Promise<void> {
  return apiRequest<void>(`${TRANSACTION_API_PATH}/${id}`, { method: "DELETE" }, getStoredToken());
}

export function createTransfer(input: TransferInput): Promise<Transaction[]> {
  return apiRequest<Transaction[]>(`${TRANSACTION_API_PATH}/transfer`, { method: "POST", body: JSON.stringify(input) }, getStoredToken());
}
