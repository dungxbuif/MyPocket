import { apiFetch } from "./apiClient";

export type BudgetPeriodType = "weekly" | "monthly" | "quarterly" | "yearly" | "custom";

export type BudgetSummary = {
  id: string;
  name: string;
  period_type: BudgetPeriodType;
  amount_vnd: number;
  category_ids: string[] | null;
  all_categories: boolean;
  custom_start?: string;
  custom_end?: string;
  version: number;
};

export type BudgetProgress = {
  budget: BudgetSummary;
  period_start: string;
  period_end: string;
  spent_vnd: number;
  remaining_vnd: number;
  percent: number;
  alert_80: boolean;
  alert_100: boolean;
};

export type BudgetInput = {
  name: string;
  period_type: BudgetPeriodType;
  amount_vnd: number;
  category_ids?: string[];
  custom_start?: string;
  custom_end?: string;
};

export async function loadBudgets() {
  const response = await apiFetch<{ budgets: BudgetProgress[] }>("/api/v1/budgets");
  return response.budgets;
}

export async function createBudget(input: BudgetInput) {
  const response = await apiFetch<{ budget: BudgetSummary }>("/api/v1/budgets", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.budget;
}

export async function updateBudget(id: string, input: BudgetInput) {
  const response = await apiFetch<{ budget: BudgetSummary }>(`/api/v1/budgets/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.budget;
}

export async function archiveBudget(id: string) {
  await apiFetch(`/api/v1/budgets/${encodeURIComponent(id)}/archive`, { method: "POST" });
}
