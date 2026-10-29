import type { TransactionInput } from "../app/finance";
import { stores, type LegacyQueuedTransaction, type OfflineMeta, type OfflineMutation } from "./types";

export const LEGACY_OUTBOX_KEY = "mypocket.transaction-outbox.v1";
const MIGRATION_META_KEY = "legacy_localstorage_outbox_migrated";
const MUTATION_SEQUENCE_META_KEY = "mutation_sequence";
const QUARANTINE_KEY = "mypocket.transaction-outbox.quarantine.v1";

export async function migrateLegacyLocalStorageOutbox(db: IDBDatabase) {
  if (await wasMigrated(db)) return;
  const legacy = readLegacyOutbox();
  const valid = legacy.filter(isLegacyQueuedTransaction);
  const invalid = legacy.filter((item) => !isLegacyQueuedTransaction(item));
  const deviceID = await ensureMigrationDeviceID(db);
  const now = new Date().toISOString();
  const existing = await readExistingOutbox(db);
  const startSequence = existing.reduce((max, item) => Math.max(max, item.sequence), 0);
  const tx = db.transaction([stores.outbox, stores.meta], "readwrite");
  const outbox = tx.objectStore(stores.outbox);
  valid.forEach((item, index) => {
    outbox.put({
      mutation_id: item.id,
      device_id: deviceID,
      sequence: startSequence + index + 1,
      entity_type: "transaction",
      entity_id: item.id,
      operation: "create",
      base_version: 0,
      payload: item.input as unknown as Record<string, unknown>,
      attempts: 0,
      state: "pending",
      created_at: item.created_at,
      updated_at: now,
    } satisfies OfflineMutation);
  });
  tx.objectStore(stores.meta).put({ key: MIGRATION_META_KEY, value: true } satisfies OfflineMeta);
  tx.objectStore(stores.meta).put({ key: MUTATION_SEQUENCE_META_KEY, value: startSequence + valid.length } satisfies OfflineMeta);
  await transactionDone(tx);
  if (invalid.length > 0) localStorage.setItem(QUARANTINE_KEY, JSON.stringify(invalid));
  localStorage.removeItem(LEGACY_OUTBOX_KEY);
}

function readLegacyOutbox(): unknown[] {
  try {
    const parsed = JSON.parse(localStorage.getItem(LEGACY_OUTBOX_KEY) ?? "[]");
    return Array.isArray(parsed) ? parsed : [];
  } catch (error) {
    localStorage.setItem(QUARANTINE_KEY, JSON.stringify([{ reason: "parse_failed", message: error instanceof Error ? error.message : "invalid json" }]));
    return [];
  }
}

function isLegacyQueuedTransaction(value: unknown): value is LegacyQueuedTransaction {
  const item = value as Partial<LegacyQueuedTransaction>;
  const input = item.input as Partial<TransactionInput> | undefined;
  return Boolean(
    item &&
      typeof item.id === "string" &&
      typeof item.created_at === "string" &&
      input &&
      typeof input.type === "string" &&
      typeof input.source_wallet_id === "string" &&
      typeof input.amount_vnd === "number" &&
      typeof input.occurred_at === "string",
  );
}

async function wasMigrated(db: IDBDatabase) {
  const tx = db.transaction(stores.meta, "readonly");
  const record = await requestToPromise<OfflineMeta | undefined>(tx.objectStore(stores.meta).get(MIGRATION_META_KEY));
  return record?.value === true;
}

async function ensureMigrationDeviceID(db: IDBDatabase) {
  const key = "device_id";
  const readTx = db.transaction(stores.meta, "readonly");
  const existing = await requestToPromise<OfflineMeta | undefined>(readTx.objectStore(stores.meta).get(key));
  if (typeof existing?.value === "string") return existing.value;
  const deviceID = globalThis.crypto?.randomUUID?.() ?? `device-${Date.now()}`;
  const writeTx = db.transaction(stores.meta, "readwrite");
  writeTx.objectStore(stores.meta).put({ key, value: deviceID } satisfies OfflineMeta);
  await transactionDone(writeTx);
  return deviceID;
}

function readExistingOutbox(db: IDBDatabase): Promise<OfflineMutation[]> {
  const tx = db.transaction(stores.outbox, "readonly");
  return requestToPromise(tx.objectStore(stores.outbox).getAll());
}

function requestToPromise<T>(request: IDBRequest<T>): Promise<T> {
  return new Promise((resolve, reject) => {
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error ?? new Error("IndexedDB request failed"));
  });
}

function transactionDone(tx: IDBTransaction): Promise<void> {
  return new Promise((resolve, reject) => {
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error ?? new Error("IndexedDB transaction failed"));
    tx.onabort = () => reject(tx.error ?? new Error("IndexedDB transaction aborted"));
  });
}
