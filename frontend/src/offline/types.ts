import type { CategorySummary, Transaction, TransactionInput, WalletSummary } from "../app/finance";
import type { AssetPosition } from "../app/portfolio";

export const stores = {
  wallets: "wallets",
  categories: "categories",
  transactions: "transactions",
  assets: "assets",
  outbox: "outbox",
  conflicts: "conflicts",
  tombstones: "tombstones",
  meta: "meta",
  receipts: "receipts",
} as const;

export type OfflineEntityType = "wallet" | "category" | "transaction" | "asset";
export type OfflineOperation = "create" | "update" | "archive" | "set_default_ai" | "set_category_active" | "add_trade" | "update_trade" | "archive_trade" | "add_price";
export type OfflineMutationState = "pending" | "sending" | "retryable" | "conflict" | "synced" | "quarantined";
export type OfflineStoreMode = "ready" | "degraded";

export type OfflineMutation = {
  mutation_id: string;
  device_id: string;
  sequence: number;
  entity_type: OfflineEntityType;
  entity_id: string;
  operation: OfflineOperation;
  base_version: number;
  payload: Record<string, unknown>;
  attempts: number;
  state: OfflineMutationState;
  created_at: string;
  updated_at: string;
};

export type OfflineConflict = {
  conflict_id: string;
  mutation_id: string;
  entity_type: OfflineEntityType;
  entity_id: string;
  operation: OfflineOperation;
  base_version: number;
  server_version: number;
  local_payload: Record<string, unknown>;
  server_payload: Record<string, unknown>;
  reason?: string;
  status: "open" | "resolved";
  created_at: string;
};

export type OfflineTombstone = {
  id: string;
  entity_type: OfflineEntityType;
  entity_id: string;
  operation: "archive";
  version: number;
  created_at: string;
};

export type OfflineMeta = {
  key: string;
  value: string | number | boolean;
};

export type OfflineReceiptUpload = {
  id: string;
  transaction_id: string;
  file: Blob;
  filename: string;
  content_type: string;
  created_at: string;
};

export type OfflineSnapshot = {
  wallets: WalletSummary[];
  categories: CategorySummary[];
  transactions: Transaction[];
  assets: AssetPosition[];
  outbox: OfflineMutation[];
  conflicts: OfflineConflict[];
  tombstones: OfflineTombstone[];
  cursor: number;
  mode: OfflineStoreMode;
  reason?: string;
};

export type LegacyQueuedTransaction = {
  id: string;
  input: TransactionInput;
  created_at: string;
};

export type WalletCreateInput = { name: string; type: WalletSummary["type"]; balance_vnd?: number; include_in_total?: boolean };
export type WalletUpdateInput = { name: string; include_in_total: boolean; base_version?: number };
export type CategoryCreateInput = { kind: CategorySummary["kind"]; name: string };
export type CategoryUpdateInput = { name: string; base_version?: number };
