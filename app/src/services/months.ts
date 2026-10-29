import { apiRequest } from "./api";
import { getStoredToken } from "./auth";
import type { JarMonthSummary } from "./jars";

export type MonthCategoryTotal = { category_id?: string; name: string; type: "income" | "expense"; amount: number; count: number };
export type MonthSummary = {
  month: string;
  timezone: string;
  start_at: string;
  next_start_at: string;
  is_current: boolean;
  is_complete: boolean;
  income: number;
  expense: number;
  net: number;
  transaction_count: number;
  categories: MonthCategoryTotal[];
  note: string;
  jar_summary?: JarMonthSummary;
  calculated_at: string;
};

const path = (month: string) => `/api/v1/months/${encodeURIComponent(month)}`;
export const fetchMonthSummary = (month: string) => apiRequest<MonthSummary>(path(month), {}, getStoredToken());
export const saveMonthNote = (month: string, note: string) => apiRequest<void>(`${path(month)}/note`, { method: "PUT", body: JSON.stringify({ note }) }, getStoredToken());
export const deleteMonthNote = (month: string) => apiRequest<void>(`${path(month)}/note`, { method: "DELETE" }, getStoredToken());
