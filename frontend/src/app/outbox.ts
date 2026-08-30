import type { Transaction } from "./finance";

const OUTBOX_KEY = "mypocket.transaction-outbox.v1";
export type QueuedTransaction = { id: string; input: Record<string, unknown>; created_at: string };

export function queueTransaction(input: Record<string, unknown>): Transaction {
  const id = globalThis.crypto?.randomUUID?.() ?? `offline-${Date.now()}`;
  const items = readOutbox();
  items.push({ id, input, created_at: new Date().toISOString() });
  localStorage.setItem(OUTBOX_KEY, JSON.stringify(items));
  return { id, type: String(input.type) as Transaction["type"], source_wallet_id: String(input.source_wallet_id), category_id: input.category_id ? String(input.category_id) : undefined, amount_vnd: Number(input.amount_vnd), balance_after_vnd: 0, occurred_at: String(input.occurred_at), note: String(input.note ?? ""), with_person: "", event_ref: "", excluded_from_reports: false, version: 0 };
}

export function readOutbox(): QueuedTransaction[] {
  try {
    const parsed = JSON.parse(localStorage.getItem(OUTBOX_KEY) ?? "[]");
    return Array.isArray(parsed) ? parsed as QueuedTransaction[] : [];
  } catch { return []; }
}

export async function drainOutbox(send: (input: Record<string, unknown>) => Promise<unknown>) {
  const pending = readOutbox();
  const completed: string[] = [];
  for (const item of pending) {
    try { await send(item.input); completed.push(item.id); } catch { break; }
  }
  if (completed.length > 0) localStorage.setItem(OUTBOX_KEY, JSON.stringify(readOutbox().filter((item) => !completed.includes(item.id))));
  return completed.length;
}
