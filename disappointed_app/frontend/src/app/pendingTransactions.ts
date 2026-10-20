import type { Transaction } from "./finance";
import type { QueuedTransaction } from "./outbox";

export function mergePendingTransactions(confirmed: Transaction[], queued: QueuedTransaction[]): Transaction[] {
  const rows = new Map(confirmed.map((row) => [row.id, row]));
  for (const item of queued) {
    if (item.entity_type && item.entity_type !== "transaction") continue;
    const id = item.entity_id ?? item.id;
    if (item.operation === "archive") { rows.delete(id); continue; }
    if ((!item.operation || item.operation === "create") && rows.has(id)) continue;
    const existing = rows.get(id);
    rows.set(id, {
      balance_after_vnd: 0, occurred_at: String(item.input.occurred_at),
      note: "", with_person: "", event_ref: "", excluded_from_reports: false, version: 0,
      ...existing, ...item.input, id,
      amount_vnd: Number(item.input.amount_vnd ?? existing?.amount_vnd ?? 0),
    } as Transaction);
  }
  return [...rows.values()];
}
