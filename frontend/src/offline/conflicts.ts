import type { CategorySummary, Transaction, TransactionInput, WalletSummary } from "../app/finance";
import type { AssetPosition } from "../app/portfolio";
import {
  archiveOfflineAsset,
  archiveOfflineCategory,
  archiveOfflineTransaction,
  archiveOfflineWallet,
  listOpenConflicts,
  readServerEpoch,
  removeOfflineMutation,
  resolveOfflineConflict,
  saveFinanceMirror,
  upsertOfflineAsset,
  upsertOfflineCategory,
  upsertOfflineTransaction,
  upsertOfflineWallet,
} from "./db";
import { queueTransactionUpdate } from "./outbox";
import { fetchAuthoritativeSnapshot } from "./syncApi";
import type { OfflineConflict } from "./types";

export { listOpenConflicts };

export const REQUIRED_SERVER_EPOCH = "atomic-sync-v1";

export async function keepServerConflict(conflict: OfflineConflict) {
  await mirrorConflictServerPayload(conflict);
  await removeOfflineMutation(conflict.mutation_id);
  await resolveOfflineConflict(conflict.conflict_id);
}

export async function discardLocalConflict(conflict: OfflineConflict) {
  await mirrorConflictServerPayload(conflict);
  await removeOfflineMutation(conflict.mutation_id);
  await resolveOfflineConflict(conflict.conflict_id);
}

export async function editAndRetryTransactionConflict(conflict: OfflineConflict, input: { amount_vnd: number; note: string }) {
  const transaction = conflict.server_payload as Partial<Transaction>;
  if (conflict.entity_type !== "transaction" || !transaction.id || !transaction.type || !transaction.source_wallet_id || !transaction.occurred_at) {
    throw new Error("Conflict cannot be retried as a transaction");
  }
  const next: TransactionInput = {
    type: transaction.type,
    source_wallet_id: transaction.source_wallet_id,
    destination_wallet_id: transaction.destination_wallet_id,
    category_id: transaction.category_id,
    amount_vnd: input.amount_vnd,
    occurred_at: transaction.occurred_at,
    note: input.note,
    with_person: transaction.with_person ?? "",
    event_ref: transaction.event_ref ?? "",
    excluded_from_reports: Boolean(transaction.excluded_from_reports),
    base_version: conflict.server_version,
  };
  await removeOfflineMutation(conflict.mutation_id);
  await resolveOfflineConflict(conflict.conflict_id);
  await queueTransactionUpdate(conflict.entity_id, next, conflict.server_version);
}

export async function fullResync() {
  const snapshot = await fetchAuthoritativeSnapshot();
  await saveAuthoritativeSnapshot(snapshot);
}

export async function reconcileServerEpoch(requiredServerEpoch = REQUIRED_SERVER_EPOCH) {
  if (await readServerEpoch() === requiredServerEpoch) return false;
  const snapshot = await fetchAuthoritativeSnapshot();
  if (snapshot.server_epoch !== requiredServerEpoch) {
    throw new Error(`Unsupported sync epoch: ${snapshot.server_epoch || "missing"}`);
  }
  await saveAuthoritativeSnapshot(snapshot);
  return true;
}

async function saveAuthoritativeSnapshot(snapshot: Awaited<ReturnType<typeof fetchAuthoritativeSnapshot>>) {
  await saveFinanceMirror({
    wallets: snapshot.wallets as WalletSummary[],
    categories: snapshot.categories as CategorySummary[],
    transactions: snapshot.transactions as Transaction[],
    assets: snapshot.assets as AssetPosition[],
    cursor: snapshot.next_cursor,
    serverEpoch: snapshot.server_epoch,
  });
}

async function mirrorConflictServerPayload(conflict: OfflineConflict) {
  if (conflict.operation === "archive") {
    if (conflict.entity_type === "wallet") await archiveOfflineWallet(conflict.entity_id, conflict.server_version);
    if (conflict.entity_type === "category") await archiveOfflineCategory(conflict.entity_id, conflict.server_version);
    if (conflict.entity_type === "transaction") await archiveOfflineTransaction(conflict.entity_id, conflict.server_version);
    if (conflict.entity_type === "asset") await archiveOfflineAsset(conflict.entity_id, conflict.server_version);
    return;
  }
  if (conflict.entity_type === "wallet") await upsertOfflineWallet(conflict.server_payload as WalletSummary);
  if (conflict.entity_type === "category") await upsertOfflineCategory(conflict.server_payload as CategorySummary);
  if (conflict.entity_type === "transaction") await upsertOfflineTransaction(conflict.server_payload as Transaction);
  if (conflict.entity_type === "asset") await upsertOfflineAsset(conflict.server_payload as AssetPosition);
}
