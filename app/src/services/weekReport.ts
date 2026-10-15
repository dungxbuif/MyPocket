import { dateKeyAt, todayDateKey, weekDateRange } from "./accountTime";

export type WeekReportRow = {
  type: string;
  amount: number;
  occurred_at: string;
  included_in_reports: boolean;
};

export type WeekReportTotals = { income: number; expense: number };

export function buildWeekReportTotals(
  transactions: WeekReportRow[],
  timezone: string,
  now = new Date(),
): WeekReportTotals {
  const today = todayDateKey(timezone, now);
  const range = weekDateRange(today, 0);
  return transactions.reduce<WeekReportTotals>((totals, transaction) => {
    if (!transaction.included_in_reports || (transaction.type !== "income" && transaction.type !== "expense")) return totals;
    const day = dateKeyAt(transaction.occurred_at, timezone);
    if (!day || day < range.start || day > range.end) return totals;
    totals[transaction.type] += transaction.amount;
    return totals;
  }, { income: 0, expense: 0 });
}
