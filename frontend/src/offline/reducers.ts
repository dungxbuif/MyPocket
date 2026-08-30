import type { Transaction } from "../app/finance";

export function upsertByID<T extends { id: string }>(items: T[], next: T) {
  return [next, ...items.filter((item) => item.id !== next.id)];
}

export function removeByID<T extends { id: string }>(items: T[], id: string) {
  return items.filter((item) => item.id !== id);
}

export function mergePendingTransactions(cached: Transaction[], pending: Transaction[]) {
  return pending.reduce((items, transaction) => upsertByID(items, transaction), cached);
}
