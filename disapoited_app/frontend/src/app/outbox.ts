import type { Transaction } from "./finance";
import { drainLegacyCompatibleOutbox, listOutbox, queueTransactionCreate } from "../offline/outbox";
import { LEGACY_OUTBOX_KEY } from "../offline/migrations";
import type { OfflineEntityType, OfflineOperation } from "../offline/types";

export type QueuedTransaction = { id: string; entity_id?: string; entity_type?: OfflineEntityType; operation?: OfflineOperation; input: Record<string, unknown>; created_at: string };

export async function queueTransaction(input: Record<string, unknown>): Promise<Transaction> {
  return queueTransactionCreate(input as Parameters<typeof queueTransactionCreate>[0]);
}

export async function readOutbox(): Promise<QueuedTransaction[]> {
  if ("indexedDB" in globalThis) {
    return (await listOutbox()).map((item) => ({ id: item.mutation_id, entity_id: item.entity_id, entity_type: item.entity_type, operation: item.operation, input: item.payload, created_at: item.created_at }));
  }
  try {
    const parsed = JSON.parse(localStorage.getItem(LEGACY_OUTBOX_KEY) ?? "[]");
    return Array.isArray(parsed) ? parsed as QueuedTransaction[] : [];
  } catch { return []; }
}

export async function drainOutbox(send: (input: Record<string, unknown>) => Promise<unknown>) {
  return drainLegacyCompatibleOutbox(send);
}
