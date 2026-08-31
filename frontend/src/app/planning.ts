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

export type EventSummary = {
  id: string;
  name: string;
  starts_on: string;
  ends_on: string;
  note: string;
  total_vnd: number;
  transaction_count: number;
  version: number;
};

export type EventInput = {
  name: string;
  starts_on: string;
  ends_on?: string;
  note?: string;
};

export type ObligationDirection = "borrowed" | "lent";

export type ObligationSummary = {
  id: string;
  direction: ObligationDirection;
  principal_vnd: number;
  counterparty: string;
  due_on: string;
  note: string;
  repaid_vnd: number;
  remaining_vnd: number;
  version: number;
};

export type ObligationInput = {
  direction: ObligationDirection;
  principal_vnd: number;
  counterparty: string;
  due_on: string;
  note?: string;
};

export type RecurrenceFrequency = "daily" | "weekly" | "monthly";

export type RecurringSchedule = {
  id: string;
  name: string;
  frequency: RecurrenceFrequency;
  timezone: string;
  starts_at: string;
  next_occurs_at: string;
  type: "income" | "expense" | "transfer";
  source_wallet_id: string;
  destination_wallet_id?: string;
  category_id?: string;
  amount_vnd: number;
  note: string;
  version: number;
};

export type RecurringScheduleInput = {
  name: string;
  frequency: RecurrenceFrequency;
  timezone: string;
  starts_at: string;
  type: "income" | "expense" | "transfer";
  source_wallet_id: string;
  destination_wallet_id?: string;
  category_id?: string;
  amount_vnd: number;
  note?: string;
};

export type TransactionDraft = {
  id: string;
  schedule_id?: string;
  occurrence_key: string;
  type: "income" | "expense" | "transfer";
  source_wallet_id: string;
  destination_wallet_id?: string;
  category_id?: string;
  amount_vnd: number;
  occurred_at: string;
  note: string;
  status: string;
  version: number;
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

export async function loadEvents() {
  const response = await apiFetch<{ events: EventSummary[] }>("/api/v1/events");
  return response.events;
}

export async function createEvent(input: EventInput) {
  const response = await apiFetch<{ event: EventSummary }>("/api/v1/events", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.event;
}

export async function updateEvent(id: string, input: EventInput) {
  const response = await apiFetch<{ event: EventSummary }>(`/api/v1/events/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.event;
}

export async function archiveEvent(id: string) {
  await apiFetch(`/api/v1/events/${encodeURIComponent(id)}/archive`, { method: "POST" });
}

export async function linkEventTransaction(eventID: string, transactionID: string) {
  await apiFetch(`/api/v1/events/${encodeURIComponent(eventID)}/transactions/${encodeURIComponent(transactionID)}`, { method: "POST" });
}

export async function loadObligations() {
  const response = await apiFetch<{ obligations: ObligationSummary[] }>("/api/v1/obligations");
  return response.obligations;
}

export async function createObligation(input: ObligationInput) {
  const response = await apiFetch<{ obligation: ObligationSummary }>("/api/v1/obligations", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.obligation;
}

export async function updateObligation(id: string, input: ObligationInput) {
  const response = await apiFetch<{ obligation: ObligationSummary }>(`/api/v1/obligations/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.obligation;
}

export async function archiveObligation(id: string) {
  await apiFetch(`/api/v1/obligations/${encodeURIComponent(id)}/archive`, { method: "POST" });
}

export async function linkObligationRepayment(obligationID: string, transactionID: string) {
  await apiFetch(`/api/v1/obligations/${encodeURIComponent(obligationID)}/repayments/${encodeURIComponent(transactionID)}`, { method: "POST" });
}

export async function loadRecurringSchedules() {
  const response = await apiFetch<{ schedules: RecurringSchedule[] }>("/api/v1/recurring-schedules");
  return response.schedules;
}

export async function createRecurringSchedule(input: RecurringScheduleInput) {
  const response = await apiFetch<{ schedule: RecurringSchedule }>("/api/v1/recurring-schedules", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
  return response.schedule;
}

export async function archiveRecurringSchedule(id: string) {
  await apiFetch(`/api/v1/recurring-schedules/${encodeURIComponent(id)}/archive`, { method: "POST" });
}

export async function loadTransactionDrafts() {
  const response = await apiFetch<{ drafts: TransactionDraft[] }>("/api/v1/transaction-drafts");
  return response.drafts;
}
