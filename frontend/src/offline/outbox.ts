import type { CategorySummary, Transaction, TransactionInput, WalletSummary } from "../app/finance";
import type { AssetPosition } from "../app/portfolio";
import {
  archiveOfflineAsset,
  archiveOfflineCategory,
  archiveOfflineTransaction,
  archiveOfflineWallet,
  enqueueMutation,
  markMutationsSynced,
  readPendingMutations,
  saveOfflineConflict,
  upsertOfflineCategory,
  upsertOfflineAsset,
  upsertOfflineTransaction,
  upsertOfflineWallet,
} from "./db";
import { submitSyncMutations, type SyncMutationResult } from "./syncApi";
import type { CategoryCreateInput, CategoryUpdateInput, OfflineMutation, WalletCreateInput, WalletUpdateInput } from "./types";

const NOW_VERSION = 0;

export async function queueWalletCreate(input: WalletCreateInput): Promise<WalletSummary> {
  const wallet: WalletSummary = {
    id: randomID("wallet"),
    name: input.name,
    type: input.type,
    balance_vnd: input.balance_vnd ?? 0,
    include_in_total: input.include_in_total ?? true,
    is_default_ai: false,
    version: NOW_VERSION,
  };
  await enqueueMutation({
    entity_type: "wallet",
    entity_id: wallet.id,
    operation: "create",
    base_version: NOW_VERSION,
    payload: {
      name: input.name,
      type: input.type,
      balance_vnd: input.balance_vnd ?? 0,
      include_in_total: input.include_in_total ?? true,
    },
  });
  await upsertOfflineWallet(wallet);
  return wallet;
}

export async function queueWalletUpdate(current: WalletSummary, input: WalletUpdateInput): Promise<WalletSummary> {
  const baseVersion = input.base_version ?? current.version ?? NOW_VERSION;
  const wallet = { ...current, name: input.name, include_in_total: input.include_in_total, version: baseVersion } satisfies WalletSummary;
  await enqueueMutation({
    entity_type: "wallet",
    entity_id: current.id,
    operation: "update",
    base_version: baseVersion,
    payload: { name: input.name, include_in_total: input.include_in_total },
  });
  await upsertOfflineWallet(wallet);
  return wallet;
}

export async function queueWalletArchive(id: string, baseVersion = NOW_VERSION) {
  await enqueueMutation({
    entity_type: "wallet",
    entity_id: id,
    operation: "archive",
    base_version: baseVersion,
    payload: {},
  });
  await archiveOfflineWallet(id, baseVersion);
}

export async function queueWalletDefaultAI(id: string, wallets: WalletSummary[], baseVersion = NOW_VERSION) {
  await enqueueMutation({
    entity_type: "wallet",
    entity_id: id,
    operation: "set_default_ai",
    base_version: baseVersion,
    payload: {},
  });
  await Promise.all(wallets.map((wallet) => upsertOfflineWallet({ ...wallet, is_default_ai: wallet.id === id })));
}

export async function queueCategoryCreate(input: CategoryCreateInput): Promise<CategorySummary> {
  const category: CategorySummary = {
    id: randomID("category"),
    kind: input.kind,
    name: input.name,
    is_system: false,
    version: NOW_VERSION,
  };
  await enqueueMutation({
    entity_type: "category",
    entity_id: category.id,
    operation: "create",
    base_version: NOW_VERSION,
    payload: input,
  });
  await upsertOfflineCategory(category);
  return category;
}

export async function queueCategoryUpdate(current: CategorySummary, input: CategoryUpdateInput): Promise<CategorySummary> {
  const baseVersion = input.base_version ?? current.version ?? NOW_VERSION;
  const category = { ...current, name: input.name } satisfies CategorySummary;
  await enqueueMutation({
    entity_type: "category",
    entity_id: current.id,
    operation: "update",
    base_version: baseVersion,
    payload: { name: input.name },
  });
  await upsertOfflineCategory(category);
  return category;
}

export async function queueCategoryArchive(id: string, baseVersion = NOW_VERSION) {
  await enqueueMutation({
    entity_type: "category",
    entity_id: id,
    operation: "archive",
    base_version: baseVersion,
    payload: {},
  });
  await archiveOfflineCategory(id, baseVersion);
}

export async function queueWalletCategoryActive(walletID: string, categoryID: string, active: boolean) {
  await enqueueMutation({
    entity_type: "category",
    entity_id: categoryID,
    operation: "set_category_active",
    base_version: NOW_VERSION,
    payload: { wallet_id: walletID, active },
  });
}

export async function queueTransactionCreate(input: TransactionInput): Promise<Transaction> {
  const transaction = optimisticTransaction(input);
  await enqueueMutation({
    entity_type: "transaction",
    entity_id: transaction.id,
    operation: "create",
    base_version: NOW_VERSION,
    payload: input as unknown as Record<string, unknown>,
  });
  await upsertOfflineTransaction(transaction);
  return transaction;
}

export async function queueTransactionUpdate(id: string, input: TransactionInput, baseVersion: number): Promise<Transaction> {
  const transaction = { ...optimisticTransaction(input, id), version: baseVersion, sync_state: "pending" } as Transaction;
  await enqueueMutation({
    entity_type: "transaction",
    entity_id: id,
    operation: "update",
    base_version: baseVersion,
    payload: input as unknown as Record<string, unknown>,
  });
  await upsertOfflineTransaction(transaction);
  return transaction;
}

export async function queueTransactionArchive(id: string, baseVersion: number) {
  await enqueueMutation({
    entity_type: "transaction",
    entity_id: id,
    operation: "archive",
    base_version: baseVersion,
    payload: {},
  });
  await archiveOfflineTransaction(id, baseVersion);
}

export async function drainLegacyCompatibleOutbox(send: (input: Record<string, unknown>) => Promise<unknown>) {
  const pending = await readPendingMutations();
  if (pending.length > 0) {
    try {
      return await drainSyncOutbox(pending);
    } catch {
      return drainCreateOnlyFallback(pending, send);
    }
  }
  return 0;
}

async function drainSyncOutbox(pending: OfflineMutation[]) {
  const results = await submitSyncMutations(pending);
  const completed: string[] = [];
  for (const result of results) {
    if (result.state === "applied" || result.state === "replayed") {
      await applyServerResult(result);
      completed.push(result.mutation_id);
    }
    if (result.state === "conflict") {
      await saveConflictResult(result);
      break;
    }
    if (result.state === "rejected") break;
  }
  await markMutationsSynced(completed);
  return completed.length;
}

async function drainCreateOnlyFallback(pending: OfflineMutation[], send: (input: Record<string, unknown>) => Promise<unknown>) {
  const completed: string[] = [];
  for (const item of pending) {
    if (item.entity_type !== "transaction" || item.operation !== "create") break;
    try {
      await send(item.payload);
      completed.push(item.mutation_id);
    } catch {
      break;
    }
  }
  await markMutationsSynced(completed);
  return completed.length;
}

export async function listOutbox(): Promise<OfflineMutation[]> {
  return readPendingMutations();
}

function optimisticTransaction(input: TransactionInput, id = randomID()): Transaction {
  return {
    id,
    type: input.type,
    source_wallet_id: input.source_wallet_id,
    destination_wallet_id: input.destination_wallet_id,
    category_id: input.category_id,
    amount_vnd: input.amount_vnd,
    balance_after_vnd: 0,
    occurred_at: input.occurred_at,
    note: input.note ?? "",
    with_person: input.with_person ?? "",
    event_ref: input.event_ref ?? "",
    excluded_from_reports: input.excluded_from_reports ?? false,
    version: input.base_version ?? NOW_VERSION,
  };
}

async function applyServerResult(result: SyncMutationResult) {
  if (result.operation === "archive") {
    if (result.entity_type === "wallet") await archiveOfflineWallet(result.entity_id, result.version ?? 0);
    if (result.entity_type === "category") await archiveOfflineCategory(result.entity_id, result.version ?? 0);
    if (result.entity_type === "transaction") await archiveOfflineTransaction(result.entity_id, result.version ?? 0);
    if (result.entity_type === "asset") await archiveOfflineAsset(result.entity_id, result.version ?? 0);
    return;
  }
  if (!result.payload) return;
  if (result.entity_type === "wallet") await upsertOfflineWallet(result.payload as unknown as WalletSummary);
  if (result.entity_type === "category") await upsertOfflineCategory(result.payload as unknown as CategorySummary);
  if (result.entity_type === "transaction") await upsertOfflineTransaction(result.payload as unknown as Transaction);
  if (result.entity_type === "asset") await upsertOfflineAsset(result.payload as unknown as AssetPosition);
}

async function saveConflictResult(result: SyncMutationResult) {
  if (!result.conflict) return;
  await saveOfflineConflict({
    conflict_id: result.mutation_id,
    mutation_id: result.mutation_id,
    entity_type: result.conflict.entity_type,
    entity_id: result.conflict.entity_id,
    operation: result.conflict.operation,
    base_version: result.conflict.base_version,
    server_version: result.conflict.server_version,
    local_payload: result.conflict.local_payload,
    server_payload: result.conflict.server_payload,
    reason: result.reason,
    status: "open",
    created_at: new Date().toISOString(),
  });
}

function randomID(prefix = "offline") {
  return globalThis.crypto?.randomUUID?.() ?? `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}
