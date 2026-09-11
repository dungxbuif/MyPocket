import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";

type Draft = {
  id: string;
  schedule_id: string;
  occurrence_key: string;
  type: "expense" | "transfer";
  source_wallet_id: string;
  destination_wallet_id?: string;
  category_id?: string;
  amount_vnd: number;
  occurred_at: string;
  note: string;
  status: "pending" | "confirmed" | "rejected";
  confirmed_transaction_id?: string;
  version: number;
};

const expenseDraft: Draft = {
  id: "draft_expense",
  schedule_id: "schedule_expense",
  occurrence_key: "recurring:schedule_expense:2026-09-01T02:00:00Z",
  type: "expense",
  source_wallet_id: "wallet_main",
  category_id: "cat_home",
  amount_vnd: 3_500_000,
  occurred_at: "2026-09-01T02:00:00Z",
  note: "Tiền nhà",
  status: "pending",
  version: 1,
};

const transferDraft: Draft = {
  id: "draft_transfer",
  schedule_id: "schedule_transfer",
  occurrence_key: "recurring:schedule_transfer:2026-09-02T02:00:00Z",
  type: "transfer",
  source_wallet_id: "wallet_main",
  destination_wallet_id: "wallet_savings",
  amount_vnd: 1_000_000,
  occurred_at: "2026-09-02T02:00:00Z",
  note: "Gửi tiết kiệm",
  status: "pending",
  version: 4,
};

describe("mounted recurring draft decisions", () => {
  afterEach(() => {
    setNavigatorOnline(true);
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("shows expense and transfer details and confirms the edited expense through the real fetch boundary", async () => {
    let drafts: Draft[] = [expenseDraft, transferDraft];
    const fetchMock = installFetch(async (path, options) => {
      if (path === "/api/v1/transaction-drafts/draft_expense/confirm" && options?.method === "POST") {
        const body = JSON.parse(String(options.body));
        const confirmed = { ...expenseDraft, ...body, status: "confirmed" as const, confirmed_transaction_id: "tx_confirmed_1", version: 2 };
        drafts = [confirmed, transferDraft];
        return ok({
          draft: confirmed,
          transaction: transaction("tx_confirmed_1", confirmed.amount_vnd, confirmed.note),
        });
      }
      return defaultResponse(path, drafts);
    });

    render(<App />);
    await openPlanning();

    expect(screen.getByText("Chi · Ví chính")).toBeInTheDocument();
    expect(screen.getByText("Chuyển · Ví chính → Ví tiết kiệm")).toBeInTheDocument();
    const amount = screen.getByLabelText("Số tiền bản nháp Tiền nhà");
    const note = screen.getByLabelText("Ghi chú bản nháp Tiền nhà");
    await userEvent.clear(amount);
    await userEvent.type(amount, "3600000");
    await userEvent.clear(note);
    await userEvent.type(note, "Tiền nhà tháng 9");
    await userEvent.click(screen.getByRole("button", { name: "Xác nhận Tiền nhà" }));

    await screen.findByText("Đã xác nhận");
    expect(screen.getByText("Giao dịch: tx_confirmed_1")).toBeInTheDocument();
    const confirmCall = fetchMock.mock.calls.find(([input, options]) => pathOf(input) === "/api/v1/transaction-drafts/draft_expense/confirm" && options?.method === "POST");
    expect(confirmCall).toBeTruthy();
    expect(JSON.parse(String(confirmCall?.[1]?.body))).toEqual({ version: 1, amount_vnd: 3_600_000, note: "Tiền nhà tháng 9" });
    expect(new Headers(confirmCall?.[1]?.headers).get("Idempotency-Key")).toBeTruthy();
    await waitFor(() => {
      expect(fetchMock.mock.calls.filter(([input, options]) => pathOf(input) === "/api/v1/transaction-drafts" && !options?.method).length).toBeGreaterThan(1);
    });
  });

  it("retains edits after failure and explicitly retries confirmation with the same idempotency key", async () => {
    let attempts = 0;
    let drafts: Draft[] = [expenseDraft];
    const fetchMock = installFetch(async (path, options) => {
      if (path === "/api/v1/transaction-drafts/draft_expense/confirm" && options?.method === "POST") {
        attempts += 1;
        if (attempts === 1) {
          return error("INTERNAL_RETRYABLE", "RAW_DATABASE_DETAIL", "req_draft_retry", 503);
        }
        const body = JSON.parse(String(options.body));
        const confirmed = { ...expenseDraft, ...body, status: "confirmed" as const, confirmed_transaction_id: "tx_retry", version: 2 };
        drafts = [confirmed];
        return ok({ draft: confirmed, transaction: transaction("tx_retry", confirmed.amount_vnd, confirmed.note) });
      }
      return defaultResponse(path, drafts);
    });

    render(<App />);
    await openPlanning();
    await userEvent.clear(screen.getByLabelText("Số tiền bản nháp Tiền nhà"));
    await userEvent.type(screen.getByLabelText("Số tiền bản nháp Tiền nhà"), "3650000");
    await userEvent.clear(screen.getByLabelText("Ghi chú bản nháp Tiền nhà"));
    await userEvent.type(screen.getByLabelText("Ghi chú bản nháp Tiền nhà"), "Giữ nội dung này");
    await userEvent.click(screen.getByRole("button", { name: "Xác nhận Tiền nhà" }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("Chưa xác nhận được bản nháp");
    expect(alert).toHaveTextContent("req_draft_retry");
    expect(alert).not.toHaveTextContent("RAW_DATABASE_DETAIL");
    expect(screen.getByLabelText("Số tiền bản nháp Tiền nhà")).toHaveValue("3650000");
    expect(screen.getByLabelText("Ghi chú bản nháp Tiền nhà")).toHaveValue("Giữ nội dung này");
    expect(attempts).toBe(1);

    await userEvent.click(screen.getByRole("button", { name: "Thử xác nhận" }));
    expect(await screen.findByText("Giao dịch: tx_retry")).toBeInTheDocument();
    const confirmationCalls = fetchMock.mock.calls.filter(([input, options]) => pathOf(input) === "/api/v1/transaction-drafts/draft_expense/confirm" && options?.method === "POST");
    expect(confirmationCalls).toHaveLength(2);
    expect(new Headers(confirmationCalls[0][1]?.headers).get("Idempotency-Key"))
      .toBe(new Headers(confirmationCalls[1][1]?.headers).get("Idempotency-Key"));
    expect(String(confirmationCalls[0][1]?.body)).toBe(String(confirmationCalls[1][1]?.body));
  });

  it("clears an uncertain confirm retry and mints a new key when the user edits its payload", async () => {
    let attempts = 0;
    let drafts: Draft[] = [expenseDraft];
    const fetchMock = installFetch(async (path, options) => {
      if (path === "/api/v1/transaction-drafts/draft_expense/confirm" && options?.method === "POST") {
        attempts += 1;
        if (attempts === 1) return error("INTERNAL_RETRYABLE", "temporary", "req_edit_retry", 503);
        const body = JSON.parse(String(options.body));
        const confirmed = { ...expenseDraft, ...body, status: "confirmed" as const, confirmed_transaction_id: "tx_edited_retry", version: 2 };
        drafts = [confirmed];
        return ok({ draft: confirmed, transaction: transaction("tx_edited_retry", confirmed.amount_vnd, confirmed.note) });
      }
      return defaultResponse(path, drafts);
    });

    render(<App />);
    await openPlanning();
    await userEvent.click(screen.getByRole("button", { name: "Xác nhận Tiền nhà" }));
    await screen.findByRole("button", { name: "Thử xác nhận" });

    await userEvent.clear(screen.getByLabelText("Số tiền bản nháp Tiền nhà"));
    await userEvent.type(screen.getByLabelText("Số tiền bản nháp Tiền nhà"), "3700000");
    await userEvent.clear(screen.getByLabelText("Ghi chú bản nháp Tiền nhà"));
    await userEvent.type(screen.getByLabelText("Ghi chú bản nháp Tiền nhà"), "Nội dung mới");
    expect(screen.queryByRole("button", { name: "Thử xác nhận" })).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Xác nhận Tiền nhà" }));

    expect(await screen.findByText("Giao dịch: tx_edited_retry")).toBeInTheDocument();
    const calls = fetchMock.mock.calls.filter(([input, options]) => pathOf(input) === "/api/v1/transaction-drafts/draft_expense/confirm" && options?.method === "POST");
    expect(calls).toHaveLength(2);
    expect(new Headers(calls[1][1]?.headers).get("Idempotency-Key"))
      .not.toBe(new Headers(calls[0][1]?.headers).get("Idempotency-Key"));
    expect(JSON.parse(String(calls[1][1]?.body))).toEqual({ version: 1, amount_vnd: 3_700_000, note: "Nội dung mới" });
  });

  it("prevents duplicate draft decisions while confirmation is pending", async () => {
    let finish!: (response: Response) => void;
    const decision = new Promise<Response>((resolve) => { finish = resolve; });
    let drafts: Draft[] = [expenseDraft];
    const fetchMock = installFetch(async (path, options) => {
      if (path === "/api/v1/transaction-drafts/draft_expense/confirm" && options?.method === "POST") return decision;
      return defaultResponse(path, drafts);
    });

    render(<App />);
    await openPlanning();
    const confirm = screen.getByRole("button", { name: "Xác nhận Tiền nhà" });
    fireEvent.click(confirm);
    fireEvent.click(confirm);

    expect(confirm).toBeDisabled();
    expect(screen.getByRole("button", { name: "Từ chối Tiền nhà" })).toBeDisabled();
    expect(fetchMock.mock.calls.filter(([input]) => pathOf(input) === "/api/v1/transaction-drafts/draft_expense/confirm")).toHaveLength(1);

    const confirmed = { ...expenseDraft, status: "confirmed" as const, confirmed_transaction_id: "tx_once", version: 2 };
    drafts = [confirmed];
    await act(async () => {
      finish(ok({ draft: confirmed, transaction: transaction("tx_once", confirmed.amount_vnd, confirmed.note) }));
    });
    expect(await screen.findByText("Giao dịch: tx_once")).toBeInTheDocument();
  });

  it("rejects a pending transfer with its current version and no offline-style idempotency header", async () => {
    let drafts: Draft[] = [transferDraft];
    const fetchMock = installFetch(async (path, options) => {
      if (path === "/api/v1/transaction-drafts/draft_transfer/reject" && options?.method === "POST") {
        const rejected = { ...transferDraft, status: "rejected" as const, version: 5 };
        drafts = [rejected];
        return ok({ draft: rejected });
      }
      return defaultResponse(path, drafts);
    });

    render(<App />);
    await openPlanning();
    await userEvent.click(screen.getByRole("button", { name: "Từ chối Gửi tiết kiệm" }));

    expect(await screen.findByText("Đã từ chối")).toBeInTheDocument();
    const rejectCall = fetchMock.mock.calls.find(([input, options]) => pathOf(input) === "/api/v1/transaction-drafts/draft_transfer/reject" && options?.method === "POST");
    expect(JSON.parse(String(rejectCall?.[1]?.body))).toEqual({ version: 4 });
    expect(new Headers(rejectCall?.[1]?.headers).get("Idempotency-Key")).toBeNull();
  });

  it("renders confirmed and rejected terminal drafts without decision controls", async () => {
    installFetch(async (path) => defaultResponse(path, [
      { ...expenseDraft, status: "confirmed", confirmed_transaction_id: "tx_existing", version: 2 },
      { ...transferDraft, status: "rejected", version: 5 },
    ]));

    render(<App />);
    await openPlanning();

    expect(screen.getByText("Giao dịch: tx_existing")).toBeInTheDocument();
    expect(screen.getByText("Đã từ chối")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Xác nhận Tiền nhà" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Từ chối Gửi tiết kiệm" })).not.toBeInTheDocument();
  });

  it("does not queue a decision or enable controls while offline", async () => {
    const fetchMock = installFetch(async (path) => defaultResponse(path, [expenseDraft]));
    render(<App />);
    await openPlanning();

    await act(async () => {
      setNavigatorOnline(false);
      window.dispatchEvent(new Event("offline"));
    });

    expect(screen.getByText(/Thay đổi chưa được xếp hàng chờ/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Xác nhận Tiền nhà" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Từ chối Tiền nhà" })).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: "Xác nhận Tiền nhà" }));
    expect(fetchMock.mock.calls.filter(([input]) => pathOf(input).includes("/transaction-drafts/draft_expense/"))).toHaveLength(0);
  });

  it("keeps confirmed draft and ledger state when an older general refresh finishes last", async () => {
    let draftGets = 0;
    let committed = false;
    let releaseStaleDrafts!: (response: Response) => void;
    const staleDrafts = new Promise<Response>((resolve) => { releaseStaleDrafts = resolve; });
    const confirmed = { ...expenseDraft, status: "confirmed" as const, confirmed_transaction_id: "tx_race", version: 2 };
    const confirmedTransaction = transaction("tx_race", confirmed.amount_vnd, confirmed.note);
    const fetchMock = installFetch(async (path, options) => {
      if (path === "/api/v1/transaction-drafts/draft_expense/confirm" && options?.method === "POST") {
        committed = true;
        return ok({ draft: confirmed, transaction: confirmedTransaction });
      }
      if (path === "/api/v1/transaction-drafts") {
        draftGets += 1;
        if (draftGets === 2) return staleDrafts;
        return ok({ drafts: committed ? [confirmed] : [expenseDraft] });
      }
      if (path === "/api/v1/transactions") return ok({ transactions: committed ? [confirmedTransaction] : [] });
      return defaultResponse(path, committed ? [confirmed] : [expenseDraft]);
    });

    render(<App />);
    await openPlanning();
    await act(async () => {
      setNavigatorOnline(false);
      window.dispatchEvent(new Event("offline"));
    });
    await screen.findByText("Offline");
    await act(async () => {
      setNavigatorOnline(true);
      window.dispatchEvent(new Event("online"));
    });
    await waitFor(() => expect(draftGets).toBe(2));

    await userEvent.click(screen.getByRole("button", { name: "Xác nhận Tiền nhà" }));
    expect(await screen.findByText("Giao dịch: tx_race")).toBeInTheDocument();
    await waitFor(() => expect(draftGets).toBe(3));

    await act(async () => {
      releaseStaleDrafts(ok({ drafts: [expenseDraft] }));
      await new Promise((resolve) => setTimeout(resolve, 0));
    });

    expect(screen.getByText("Giao dịch: tx_race")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Xác nhận Tiền nhà" })).not.toBeInTheDocument();
    await userEvent.click(screen.getByLabelText("Sổ giao dịch"));
    expect(await screen.findByText("Tiền nhà")).toBeInTheDocument();
    expect(fetchMock.mock.calls.filter(([input]) => pathOf(input) === "/api/v1/transaction-drafts")).toHaveLength(3);
  });
});

async function openPlanning() {
  await screen.findByText("Ví chính");
  await userEvent.click(screen.getByLabelText("Ngân sách"));
  await screen.findByRole("heading", { name: "Bản nháp giao dịch" });
}

function installFetch(handler: (path: string, options?: RequestInit) => Promise<Response>) {
  const fetchMock = vi.fn((input: RequestInfo | URL, options?: RequestInit) => handler(pathOf(input), options));
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

function setNavigatorOnline(value: boolean) {
  Object.defineProperty(window.navigator, "onLine", { configurable: true, value });
}

function pathOf(input: RequestInfo | URL) {
  const url = typeof input === "string" ? input : input instanceof URL ? input.toString() : input.url;
  return new URL(url, "http://localhost").pathname;
}

function defaultResponse(path: string, drafts: Draft[]) {
  const routes: Record<string, unknown> = {
    "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
    "/api/v1/wallets": { wallets: [
      { id: "wallet_main", name: "Ví chính", type: "cash", balance_vnd: 10_000_000, include_in_total: true, is_default_ai: true, version: 1 },
      { id: "wallet_savings", name: "Ví tiết kiệm", type: "savings", balance_vnd: 20_000_000, include_in_total: true, is_default_ai: false, version: 1 },
    ] },
    "/api/v1/categories": { categories: [{ id: "cat_home", kind: "expense", name: "Nhà cửa", is_system: true, version: 1 }] },
    "/api/v1/transactions": { transactions: [] },
    "/api/v1/budgets": { budgets: [] },
    "/api/v1/events": { events: [] },
    "/api/v1/obligations": { obligations: [] },
    "/api/v1/recurring-schedules": { schedules: [] },
    "/api/v1/transaction-drafts": { drafts },
    "/api/v1/notifications": { notifications: [] },
    "/api/v1/dashboard": { report: {
      net_worth_vnd: 30_000_000,
      summary: reportSummary(),
      wallets: [],
      recent_transactions: [],
    } },
    "/api/v1/reports/daily": { report: { summary: reportSummary(), daily: [] } },
    "/api/v1/reports/insider": { report: {
      spent_vnd: 0,
      average_daily_vnd: 0,
      prior_average_daily_vnd: 0,
      not_comparable: true,
      elapsed_days: 1,
      generated_at: "2026-09-10T00:00:00Z",
      timezone: "Asia/Ho_Chi_Minh",
      from: "2026-09-10",
      to: "2026-09-10",
      data_version: 1,
    } },
    "/api/v1/assets": { assets: [] },
    "/api/v1/portfolio/summary": { summary: { investment_market_value_vnd: 0, missing_price_count: 0, position_count: 0 } },
  };
  return ok(routes[path] ?? {});
}

function reportSummary() {
  return {
    income_vnd: 0,
    expense_vnd: 0,
    net_income_vnd: 0,
    generated_at: "2026-09-10T00:00:00Z",
    timezone: "Asia/Ho_Chi_Minh",
    from: "2026-09-01",
    to: "2026-09-10",
    data_version: 1,
  };
}

function transaction(id: string, amount: number, note: string) {
  return {
    id,
    type: "expense",
    source_wallet_id: "wallet_main",
    category_id: "cat_home",
    amount_vnd: amount,
    balance_after_vnd: 6_400_000,
    occurred_at: expenseDraft.occurred_at,
    note,
    with_person: "",
    event_ref: "",
    excluded_from_reports: false,
    version: 1,
  };
}

function ok(body: unknown) {
  return new Response(JSON.stringify({ status: "ok", correlation_id: "req_ok", ...(body as Record<string, unknown>) }), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
}

function error(code: string, message: string, correlationID: string, status: number) {
  return new Response(JSON.stringify({ status: "error", error: { code, message }, correlation_id: correlationID }), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
