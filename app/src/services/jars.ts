import { apiRequest } from "./api";
import { getStoredToken } from "./auth";

export type JarItem = {
  jar_id: string;
  name: string;
  allocation_mode: "none" | "fixed" | "percent";
  allocation_amount?: number;
  allocation_percent?: number;
  calculated_allocation?: number;
  spent: number;
  active: boolean;
};

export type JarMonthSummary = {
  month: string;
  timezone: string;
  actual_income: number;
  total_spent: number;
  unassigned_spent: number;
  total_allocated: number;
  total_allocation_percent: number;
  over_income: boolean;
  over_one_hundred_percent: boolean;
  jars: { jar_id: string; name: string }[];
  items: JarItem[];
};

export type JarConfigInput = {
  month: string;
  name: string;
  allocation_mode: JarItem["allocation_mode"];
  allocation_amount?: number;
  allocation_percent?: number;
};

export type JarCumulativeMonth = { month: string; name?: string; spent: number; calculated_allocation?: number; allocation_covered: boolean };
export type JarCumulative = {
  jar_id: string;
  name: string;
  from_month: string;
  to_month: string;
  total_spent: number;
  total_allocated: number;
  allocation_months: number;
  months_in_range: number;
  allocation_variance?: number;
  months: JarCumulativeMonth[];
};

const PATH = "/api/v1/jars";
const token = () => getStoredToken();

export const fetchJarMonth = (month: string) => apiRequest<JarMonthSummary>(`${PATH}?month=${encodeURIComponent(month)}`, {}, token());
export const createJar = (input: JarConfigInput) => apiRequest<JarItem>(PATH, { method: "POST", body: JSON.stringify(input) }, token());
export const updateJarMonth = (jarID: string, input: JarConfigInput) => apiRequest<JarItem>(`${PATH}/${encodeURIComponent(jarID)}/months/${encodeURIComponent(input.month)}`, { method: "PUT", body: JSON.stringify(input) }, token());
export const removeJarMonth = (jarID: string, month: string) => apiRequest<void>(`${PATH}/${encodeURIComponent(jarID)}/months/${encodeURIComponent(month)}`, { method: "DELETE" }, token());
export const fetchJarCumulative = (jarID: string, to: string, from?: string) => {
  const query = new URLSearchParams({ to });
  if (from) query.set("from", from);
  return apiRequest<JarCumulative>(`${PATH}/${encodeURIComponent(jarID)}/report?${query}`, {}, token());
};
