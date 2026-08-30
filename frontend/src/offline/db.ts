import type { CategorySummary, Transaction, WalletSummary } from "../app/finance";
import { migrateLegacyLocalStorageOutbox } from "./migrations";
import { stores, type OfflineConflict, type OfflineMeta, type OfflineMutation, type OfflineSnapshot, type OfflineTombstone } from "./types";

const DB_NAME = "mypocket.offline.v1";
const DB_VERSION = 1;
const META_CURSOR = "sync_cursor";
const META_DEVICE_ID = "device_id";
const META_SEQUENCE = "mutation_sequence";

export type OfflineDB = IDBDatabase;

export async function openOfflineDatabase(factory?: IDBFactory): Promise<OfflineDB> {
  const databaseFactory = factory ?? globalThis.indexedDB;
  if (!databaseFactory) throw new Error("IndexedDB unavailable");
  return new Promise((resolve, reject) => {
    const request = databaseFactory.open(DB_NAME, DB_VERSION);
    request.onupgradeneeded = () => {
      const db = request.result;
      const tx = request.transaction;
      ensureStore(db, stores.wallets, "id", tx);
      ensureStore(db, stores.categories, "id", tx);
      ensureStore(db, stores.transactions, "id", tx);
      const outbox = ensureStore(db, stores.outbox, "mutation_id", request.transaction);
      if (!outbox.indexNames.contains("by_state_sequence")) outbox.createIndex("by_state_sequence", ["state", "sequence"]);
      ensureStore(db, stores.conflicts, "conflict_id", tx);
      ensureStore(db, stores.tombstones, "id", tx);
      ensureStore(db, stores.meta, "key", tx);
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error ?? new Error("Cannot open offline database"));
    request.onblocked = () => reject(new Error("Offline database upgrade is blocked"));
  });
}

export async function initializeOfflineStore(factory?: IDBFactory): Promise<OfflineSnapshot> {
  let db: IDBDatabase | undefined;
  try {
    db = await openOfflineDatabase(factory);
    await ensureDeviceID(db);
    await migrateLegacyLocalStorageOutbox(db);
    return readOfflineSnapshot(db);
  } catch (error) {
    return emptyDegradedSnapshot(error instanceof Error ? error.message : "IndexedDB unavailable");
  } finally {
    db?.close();
  }
}

export async function readOfflineSnapshot(db?: IDBDatabase): Promise<OfflineSnapshot> {
  const database = db ?? await openOfflineDatabase();
  const [wallets, categories, transactions, outbox, conflicts, tombstones, cursor] = await Promise.all([
    readAll<WalletSummary>(database, stores.wallets),
    readAll<CategorySummary>(database, stores.categories),
    readAll<Transaction>(database, stores.transactions),
    readAll<OfflineMutation>(database, stores.outbox),
    readAll<OfflineConflict>(database, stores.conflicts),
    readAll<OfflineTombstone>(database, stores.tombstones),
    readMetaNumber(database, META_CURSOR, 0),
  ]);
  const snapshot = {
    wallets,
    categories,
    transactions: applyPendingArchives(transactions, tombstones),
    outbox: outbox.filter((item) => item.state !== "synced").sort((a, b) => a.sequence - b.sequence),
    conflicts,
    tombstones,
    cursor,
    mode: "ready",
  };
  if (!db) database.close();
  return snapshot;
}

export async function saveFinanceMirror(input: {
  wallets?: WalletSummary[];
  categories?: CategorySummary[];
  transactions?: Transaction[];
  cursor?: number;
}, db?: IDBDatabase) {
  const database = db ?? await openOfflineDatabase();
  const names: string[] = [];
  if (input.wallets) names.push(stores.wallets);
  if (input.categories) names.push(stores.categories);
  if (input.transactions) names.push(stores.transactions);
  if (typeof input.cursor === "number") names.push(stores.meta);
  if (names.length === 0) {
    if (!db) database.close();
    return;
  }
  const tx = database.transaction(names, "readwrite");
  if (input.wallets) replaceStore(tx.objectStore(stores.wallets), input.wallets);
  if (input.categories) replaceStore(tx.objectStore(stores.categories), input.categories);
  if (input.transactions) replaceStore(tx.objectStore(stores.transactions), input.transactions);
  if (typeof input.cursor === "number") tx.objectStore(stores.meta).put({ key: META_CURSOR, value: input.cursor } satisfies OfflineMeta);
  await transactionDone(tx);
  if (!db) database.close();
}

export async function enqueueMutation(input: Omit<OfflineMutation, "mutation_id" | "device_id" | "sequence" | "attempts" | "state" | "created_at" | "updated_at">, db?: IDBDatabase) {
  const database = db ?? await openOfflineDatabase();
  const now = new Date().toISOString();
  const [deviceID, sequence] = await Promise.all([getDeviceID(database), reserveSequence(database)]);
  const mutation: OfflineMutation = {
    ...input,
    mutation_id: randomID("mut"),
    device_id: deviceID,
    sequence,
    attempts: 0,
    state: "pending",
    created_at: now,
    updated_at: now,
  };
  const tx = database.transaction(stores.outbox, "readwrite");
  tx.objectStore(stores.outbox).put(mutation);
  await transactionDone(tx);
  if (!db) database.close();
  return mutation;
}

export async function upsertOfflineTransaction(transaction: Transaction, db?: IDBDatabase) {
  const database = db ?? await openOfflineDatabase();
  const tx = database.transaction(stores.transactions, "readwrite");
  tx.objectStore(stores.transactions).put(transaction);
  await transactionDone(tx);
  if (!db) database.close();
}

export async function upsertOfflineWallet(wallet: WalletSummary, db?: IDBDatabase) {
  const database = db ?? await openOfflineDatabase();
  const tx = database.transaction(stores.wallets, "readwrite");
  tx.objectStore(stores.wallets).put(wallet);
  await transactionDone(tx);
  if (!db) database.close();
}

export async function upsertOfflineCategory(category: CategorySummary, db?: IDBDatabase) {
  const database = db ?? await openOfflineDatabase();
  const tx = database.transaction(stores.categories, "readwrite");
  tx.objectStore(stores.categories).put(category);
  await transactionDone(tx);
  if (!db) database.close();
}

export async function archiveOfflineTransaction(id: string, baseVersion = 0, db?: IDBDatabase) {
  await archiveEntity("transaction", id, baseVersion, stores.transactions, db);
}

export async function archiveOfflineWallet(id: string, baseVersion = 0, db?: IDBDatabase) {
  await archiveEntity("wallet", id, baseVersion, stores.wallets, db);
}

export async function archiveOfflineCategory(id: string, baseVersion = 0, db?: IDBDatabase) {
  await archiveEntity("category", id, baseVersion, stores.categories, db);
}

async function archiveEntity(entityType: OfflineTombstone["entity_type"], id: string, baseVersion: number, storeName: string, db?: IDBDatabase) {
  const database = db ?? await openOfflineDatabase();
  const now = new Date().toISOString();
  const tx = database.transaction([storeName, stores.tombstones], "readwrite");
  tx.objectStore(storeName).delete(id);
  tx.objectStore(stores.tombstones).put({ id: `${entityType}:${id}`, entity_type: entityType, entity_id: id, operation: "archive", version: baseVersion, created_at: now } satisfies OfflineTombstone);
  await transactionDone(tx);
  if (!db) database.close();
}

export async function readPendingMutations(db?: IDBDatabase) {
  const database = db ?? await openOfflineDatabase();
  const outbox = await readAll<OfflineMutation>(database, stores.outbox);
  if (!db) database.close();
  return outbox.filter((item) => item.state === "pending" || item.state === "retryable").sort((a, b) => a.sequence - b.sequence);
}

export async function markMutationsSynced(ids: string[], db?: IDBDatabase) {
  if (ids.length === 0) return;
  const database = db ?? await openOfflineDatabase();
  const tx = database.transaction(stores.outbox, "readwrite");
  const store = tx.objectStore(stores.outbox);
  ids.forEach((id) => store.delete(id));
  await transactionDone(tx);
  if (!db) database.close();
}

export async function clearOfflineStore(db?: IDBDatabase) {
  const database = db ?? await openOfflineDatabase();
  const tx = database.transaction(Object.values(stores), "readwrite");
  Object.values(stores).forEach((name) => tx.objectStore(name).clear());
  await transactionDone(tx);
  if (!db) database.close();
}

function ensureStore(db: IDBDatabase, name: string, keyPath: string, tx?: IDBTransaction | null) {
  if (db.objectStoreNames.contains(name)) {
    if (!tx) throw new Error(`Store already exists outside upgrade transaction: ${name}`);
    return tx.objectStore(name);
  }
  return db.createObjectStore(name, { keyPath });
}

function replaceStore<T>(store: IDBObjectStore, records: T[]) {
  store.clear();
  records.forEach((record) => store.put(record));
}

function readAll<T>(db: IDBDatabase, name: string): Promise<T[]> {
  const tx = db.transaction(name, "readonly");
  const request = tx.objectStore(name).getAll();
  return requestToPromise<T[]>(request);
}

async function readMetaNumber(db: IDBDatabase, key: string, fallback: number) {
  const tx = db.transaction(stores.meta, "readonly");
  const record = await requestToPromise<OfflineMeta | undefined>(tx.objectStore(stores.meta).get(key));
  return typeof record?.value === "number" ? record.value : fallback;
}

async function getDeviceID(db: IDBDatabase) {
  const tx = db.transaction(stores.meta, "readonly");
  const record = await requestToPromise<OfflineMeta | undefined>(tx.objectStore(stores.meta).get(META_DEVICE_ID));
  if (typeof record?.value === "string" && record.value) return record.value;
  return ensureDeviceID(db);
}

async function ensureDeviceID(db: IDBDatabase) {
  const existing = await getMetaString(db, META_DEVICE_ID);
  if (existing) return existing;
  const deviceID = randomID("device");
  const tx = db.transaction(stores.meta, "readwrite");
  tx.objectStore(stores.meta).put({ key: META_DEVICE_ID, value: deviceID } satisfies OfflineMeta);
  await transactionDone(tx);
  return deviceID;
}

async function getMetaString(db: IDBDatabase, key: string) {
  const tx = db.transaction(stores.meta, "readonly");
  const record = await requestToPromise<OfflineMeta | undefined>(tx.objectStore(stores.meta).get(key));
  return typeof record?.value === "string" ? record.value : "";
}

async function reserveSequence(db: IDBDatabase) {
  const tx = db.transaction(stores.meta, "readwrite");
  const store = tx.objectStore(stores.meta);
  const record = await requestToPromise<OfflineMeta | undefined>(store.get(META_SEQUENCE));
  const next = (typeof record?.value === "number" ? record.value : 0) + 1;
  store.put({ key: META_SEQUENCE, value: next } satisfies OfflineMeta);
  await transactionDone(tx);
  return next;
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

function applyPendingArchives(transactions: Transaction[], tombstones: OfflineTombstone[]) {
  const archived = new Set(tombstones.filter((item) => item.entity_type === "transaction").map((item) => item.entity_id));
  return transactions.filter((transaction) => !archived.has(transaction.id));
}

function emptyDegradedSnapshot(reason: string): OfflineSnapshot {
  return { wallets: [], categories: [], transactions: [], outbox: [], conflicts: [], tombstones: [], cursor: 0, mode: "degraded", reason };
}

function randomID(prefix: string) {
  return globalThis.crypto?.randomUUID?.() ?? `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}
