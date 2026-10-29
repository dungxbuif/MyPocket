import { apiRequest } from "./api";

export type InsiderCategoryTotal = { category_id?: string; name: string; amount: number; count: number };
export type InsiderExpenseRow = { id: string; wallet_id: string; wallet_name?: string; category_id?: string; category_name?: string; amount: number; occurred_at: string; note?: string };
export type InsiderSummary = {
  month: string;
  timezone: string;
  start_at: string;
  next_start_at: string;
  is_current: boolean;
  income: number;
  expense: number;
  net: number;
  previous_expense: number;
  expense_delta: number;
  expense_change_bps?: number;
  average_daily_expense: number;
  days_considered: number;
  spending_income_ratio_bps?: number;
  top_categories: InsiderCategoryTotal[];
  top_expenses: InsiderExpenseRow[];
  estimated: boolean;
  wallet_filter?: string;
  category_filter?: string;
};

export const fetchMoneyInsider = (month: string, filters: { walletID?: string; categoryID?: string } = {}) => {
  const query = new URLSearchParams({ month });
  if (filters.walletID) query.set("wallet_id", filters.walletID);
  if (filters.categoryID) query.set("category_id", filters.categoryID);
  return apiRequest<InsiderSummary>(`/api/v1/reports/insider?${query.toString()}`);
};
