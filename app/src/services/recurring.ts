import { apiRequest } from "./api";
import { getStoredToken } from "./auth";
import type { TransactionType } from "./transactions";

export type RecurringFrequency = "daily" | "weekly" | "monthly" | "yearly";
export type RecurringSchedule = {
  id: string;
  owner_id: string;
  name: string;
  wallet_id: string;
  category_id?: string | null;
  type: TransactionType;
  amount: number;
  note?: string | null;
  frequency: RecurringFrequency;
  interval: number;
  next_run_at: string;
  ends_at?: string | null;
  active: boolean;
};

export type RecurringInput = Omit<RecurringSchedule, "id" | "owner_id" | "next_run_at" | "ends_at" | "category_id" | "active"> & { category_id?: string | null; next_run_at: string; ends_at?: string | null; active?: boolean };
const PATH = "/api/v1/recurring";
const token = () => getStoredToken();

export function fetchRecurring(): Promise<RecurringSchedule[]> { return apiRequest<RecurringSchedule[]>(PATH, {}, token()); }
export function createRecurring(input: RecurringInput): Promise<RecurringSchedule> { return apiRequest<RecurringSchedule>(PATH, { method: "POST", body: JSON.stringify(input) }, token()); }
export function updateRecurring(id: string, input: RecurringInput): Promise<RecurringSchedule> { return apiRequest<RecurringSchedule>(`${PATH}/${encodeURIComponent(id)}`, { method: "PATCH", body: JSON.stringify(input) }, token()); }
export function deleteRecurring(id: string): Promise<void> { return apiRequest<void>(`${PATH}/${encodeURIComponent(id)}`, { method: "DELETE" }, token()); }
export function runDueRecurring(): Promise<{ created: number }> { return apiRequest<{ created: number }>(`${PATH}/run-due`, { method: "POST" }, token()); }
