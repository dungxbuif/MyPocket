import { apiFetch } from "./apiClient";

export type Dashboard = { net_worth_vnd: number; wallet_net_worth_vnd?: number; investment_market_value_vnd?: number; combined_net_worth_vnd?: number; missing_asset_price_count?: number; summary: { income_vnd: number; expense_vnd: number; net_income_vnd: number; generated_at: string; timezone: string; from: string; to: string; data_version: number }; wallets: Array<{ id: string; name: string; balance_vnd: number; include_in_total: boolean }>; recent_transactions: Array<{ id: string; type: string; amount_vnd: number; note: string; occurred_at: string }> };
export type Report = { summary: Dashboard["summary"]; categories?: Array<{ category_name: string; amount_vnd: number; share_percent: number }>; daily?: Array<{ date: string; income_vnd: number; expense_vnd: number; net_income_vnd: number; cumulative_net_vnd: number }> };
export type SearchResult = { kind: string; id: string; label: string; detail?: string };
export type WalletDetail = { wallet: { id: string; name: string; balance_vnd: number; include_in_total: boolean }; transactions: Dashboard["recent_transactions"] };
export type InsiderReport = {
  selected_category?: { id?: string; name: string; transaction_count: number };
  spent_vnd: number;
  average_daily_vnd: number;
  prior_average_daily_vnd: number;
  change_percent?: number;
  not_comparable: boolean;
  elapsed_days: number;
  generated_at: string;
  timezone: string;
  from: string;
  to: string;
  data_version: number;
};

export async function loadDashboard() {
  try { const response = await apiFetch<{ report: Dashboard }>("/api/v1/dashboard"); localStorage.setItem("mypocket:dashboard-cache", JSON.stringify(response.report)); return response.report; }
  catch (error) { const cached = localStorage.getItem("mypocket:dashboard-cache"); if (cached) return JSON.parse(cached) as Dashboard; throw error; }
}
export async function loadReport(kind: "cash-flow" | "categories" | "daily" | "comparison" | "cumulative") {
  try { const response = await apiFetch<{ report: Report }>(`/api/v1/reports/${kind}`); localStorage.setItem(`mypocket:report-cache:${kind}`, JSON.stringify(response.report)); return response.report; }
  catch (error) { const cached = localStorage.getItem(`mypocket:report-cache:${kind}`); if (cached) return JSON.parse(cached) as Report; throw error; }
}
export async function loadInsider() {
  try {
    const response = await apiFetch<{ report: InsiderReport }>("/api/v1/reports/insider");
    localStorage.setItem("mypocket:report-cache:insider", JSON.stringify(response.report));
    return response.report;
  } catch (error) {
    const cached = localStorage.getItem("mypocket:report-cache:insider");
    if (cached) return JSON.parse(cached) as InsiderReport;
    throw error;
  }
}
export async function searchRecords(query: string) {
  const key = `mypocket:search-cache:${query.trim().toLowerCase()}`;
  try { const response = await apiFetch<{ results: SearchResult[] }>(`/api/v1/search?q=${encodeURIComponent(query)}`); localStorage.setItem(key, JSON.stringify(response.results)); return response.results; }
  catch (error) { const cached = localStorage.getItem(key); if (cached) return JSON.parse(cached) as SearchResult[]; throw error; }
}
export async function loadWalletDetail(id: string) { const response = await apiFetch<WalletDetail>(`/api/v1/wallets/${encodeURIComponent(id)}/detail`); return response; }
