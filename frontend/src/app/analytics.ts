import { apiFetch } from "./apiClient";
import { readUserDataCache, writeUserDataCache } from "./userDataCache";

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

export async function loadDashboard(userID: string) {
  try { const response = await apiFetch<{ report: Dashboard }>("/api/v1/dashboard"); writeUserDataCache(userID, "dashboard", response.report); return response.report; }
  catch (error) { if (!(error instanceof TypeError)) throw error; const cached = readUserDataCache<Dashboard>(userID, "dashboard"); if (cached) return cached; throw error; }
}
export async function loadReport(userID: string, kind: "cash-flow" | "categories" | "daily" | "comparison" | "cumulative") {
  const resource = `report:${kind}`;
  try { const response = await apiFetch<{ report: Report }>(`/api/v1/reports/${kind}`); writeUserDataCache(userID, resource, response.report); return response.report; }
  catch (error) { if (!(error instanceof TypeError)) throw error; const cached = readUserDataCache<Report>(userID, resource); if (cached) return cached; throw error; }
}
export async function loadInsider(userID: string) {
  try {
    const response = await apiFetch<{ report: InsiderReport }>("/api/v1/reports/insider");
    writeUserDataCache(userID, "report:insider", response.report);
    return response.report;
  } catch (error) {
    if (!(error instanceof TypeError)) throw error;
    const cached = readUserDataCache<InsiderReport>(userID, "report:insider");
    if (cached) return cached;
    throw error;
  }
}
export async function searchRecords(userID: string, query: string) {
  const resource = `search:${query.trim().toLowerCase()}`;
  try { const response = await apiFetch<{ results: SearchResult[] }>(`/api/v1/search?q=${encodeURIComponent(query)}`); writeUserDataCache(userID, resource, response.results); return response.results; }
  catch (error) { if (!(error instanceof TypeError)) throw error; const cached = readUserDataCache<SearchResult[]>(userID, resource); if (cached) return cached; throw error; }
}
export async function loadWalletDetail(id: string): Promise<WalletDetail> {
  // Existing API serializes an empty Go slice as null; keep component state array-shaped.
  const response = await apiFetch<Omit<WalletDetail, "transactions"> & { transactions: WalletDetail["transactions"] | null }>(`/api/v1/wallets/${encodeURIComponent(id)}/detail`);
  return { ...response, transactions: response.transactions ?? [] };
}
