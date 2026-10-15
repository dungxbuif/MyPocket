import { accountMonthKey, dateKeyAt } from "./accountTime";

export type ReportTrendRow = {
  type: "income" | "expense" | "adjustment" | string;
  amount: number;
  occurred_at: string;
  included_in_reports: boolean;
};

export type ReportTrendPoint = {
  day: string;
  income: number;
  expense: number;
};

export function buildReportTrend(transactions: ReportTrendRow[], month: string, timezone: string): ReportTrendPoint[] {
  const totals = new Map<string, ReportTrendPoint>();
  for (const transaction of transactions) {
    if (!transaction.included_in_reports || (transaction.type !== "income" && transaction.type !== "expense")) continue;
    if (accountMonthKey(transaction.occurred_at, timezone) !== month) continue;
    const day = dateKeyAt(transaction.occurred_at, timezone);
    if (!day) continue;
    const current = totals.get(day) ?? { day, income: 0, expense: 0 };
    current[transaction.type] += transaction.amount;
    totals.set(day, current);
  }
  return [...totals.values()].sort((left, right) => left.day.localeCompare(right.day));
}
