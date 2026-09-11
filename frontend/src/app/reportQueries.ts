import { apiFetch } from "./apiClient";

export type ReportKind = "cash-flow" | "categories" | "daily" | "comparison" | "cumulative";
export type ReportFilters = { kind: ReportKind; from: string; to: string; walletID: string };
export type ReportSummary = {
  income_vnd: number; expense_vnd: number; net_income_vnd: number; daily_average_vnd: number;
  from: string; to: string; generated_at: string; timezone: string; data_version: number;
  not_comparable: boolean; income_change_percent?: number; expense_change_percent?: number;
};
export type DetailedReport = {
  summary: ReportSummary;
  categories?: Array<{ category_id?: string; category_name: string; amount_vnd: number; share_percent: number }>;
  daily?: Array<{ date: string; income_vnd: number; expense_vnd: number; net_income_vnd: number; cumulative_net_vnd: number }>;
  prior?: ReportSummary;
};

export async function queryReport(filters: ReportFilters, signal: AbortSignal): Promise<DetailedReport> {
  const query = new URLSearchParams({ from: filters.from, to: filters.to });
  if (filters.walletID) query.set("wallet_id", filters.walletID);
  const response = await apiFetch<{ report: DetailedReport }>(`/api/v1/reports/${filters.kind}?${query}`, { signal });
  if (!response.report?.summary) throw new Error("Invalid report response");
  return response.report;
}
