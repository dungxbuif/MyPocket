export type AdvisorHistoryMessage = { id: string; seq: number };

export function mergeAdvisorMessages<T extends AdvisorHistoryMessage>(current: T[], incoming: T[]): T[] {
  const byID = new Map<string, T>();
  for (const message of incoming) byID.set(message.id, message);
  for (const message of current) byID.set(message.id, message);
  return [...byID.values()].sort((left, right) => left.seq - right.seq || left.id.localeCompare(right.id));
}
