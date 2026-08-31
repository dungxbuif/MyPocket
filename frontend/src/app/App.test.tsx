import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { enqueueMutation, saveFinanceMirror, saveOfflineConflict } from "../offline/db";
import { App } from "./App";
import { readOutbox } from "./outbox";

describe("App shell", () => {
  afterEach(() => {
    mockNavigatorOnline(true);
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("renders mobile navigation destinations without finance data", async () => {
    mockFetchRoutes({ "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } } });
    render(<App />);

    expect(await screen.findByRole("navigation")).toBeInTheDocument();
    expect(screen.getByLabelText("Tổng quan")).toBeInTheDocument();
    expect(screen.getByLabelText("Sổ giao dịch")).toBeInTheDocument();
    expect(screen.getByLabelText("Thêm giao dịch")).toBeInTheDocument();
    expect(screen.getByLabelText("Ngân sách")).toBeInTheDocument();
    expect(screen.getByLabelText("Tài khoản")).toBeInTheDocument();
  });

  it("shows offline state when the browser is offline", async () => {
    mockNavigatorOnline(false);
    localStorage.setItem("mypocket.current-user.v1", JSON.stringify({ id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" }));

    render(<App />);

    expect(await screen.findByText("Offline")).toBeInTheDocument();
  });

  it("opens the transaction add sheet from the raised add action", async () => {
    mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "cash", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống API", system_key: "expense_food", is_system: true, version: 1 }] },
      "/api/v1/transactions": { transactions: [] },
    });
    render(<App />);

    expect(await screen.findByText("Ví API")).toBeInTheDocument();
    await userEvent.click(await screen.findByLabelText("Thêm giao dịch"));

    expect(screen.getByRole("dialog", { name: "Thêm Giao Dịch" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Khoản chi" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Khoản thu" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Vay/nợ" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Chuyển" })).not.toBeInTheDocument();
    expect(screen.getByText("Chọn nhóm")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Thêm chi tiết" }));
    expect(screen.getByText("Thêm Hình Ảnh")).toBeInTheDocument();
  });

  it("blocks quick add until at least one wallet exists", async () => {
    mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [] },
      "/api/v1/categories": { categories: [] },
      "/api/v1/transactions": { transactions: [] },
    });
    render(<App />);

    expect(await screen.findByLabelText("Thêm giao dịch")).toBeDisabled();
    expect(screen.queryByRole("dialog", { name: "Thêm Giao Dịch" })).not.toBeInTheDocument();
  });

  it("renders authenticated wallet data from the finance API", async () => {
    mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": {
        wallets: [
          { id: "wallet_live", name: "Ví API", type: "cash", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 2 },
          { id: "wallet_savings", name: "Tiết kiệm API", type: "savings", balance_vnd: 234567, include_in_total: true, is_default_ai: false, version: 1 },
        ],
      },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống API", system_key: "expense_food", is_system: true, version: 1 }] },
      "/api/v1/transactions": { transactions: [] },
    });

    render(<App />);

    expect(await screen.findByText("Ví API")).toBeInTheDocument();
    expect(screen.getByText("Tiết kiệm API")).toBeInTheDocument();
    expect(screen.getAllByText("1.234.567 đ")).toHaveLength(1);
    expect(screen.queryByText("Tổng hiển thị")).not.toBeInTheDocument();
    expect(screen.queryByText("Thu nhập ròng")).not.toBeInTheDocument();
    expect(screen.getByText("1.000.000 đ")).toBeInTheDocument();
    expect(screen.queryByText("Techcombank")).not.toBeInTheDocument();
  });

  it("shows loaded categories in the add transaction sheet", async () => {
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "cash", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống API", system_key: "expense_food", is_system: true, version: 1 }] },
      "/api/v1/transactions": { transactions: [] },
    });

    render(<App />);
    await waitFor(() => expect(screen.queryByText("Đang kiểm tra phiên đăng nhập...")).not.toBeInTheDocument());
    await waitFor(() => expect(fetchMock.mock.calls.some(([input]) => new URL(String(input), "http://localhost").pathname === "/api/v1/categories")).toBe(true));
    expect(await screen.findByText("Ví API")).toBeInTheDocument();
    await userEvent.click(await screen.findByLabelText("Thêm giao dịch"));

    expect(screen.getByText("Ăn uống API")).toBeInTheDocument();
  });

  it("renders API transactions and can search them", async () => {
    mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "cash", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] },
      "/api/v1/categories": { categories: [] },
      "/api/v1/transactions": { transactions: [{ id: "tx_1", type: "expense", source_wallet_id: "wallet_live", amount_vnd: 125000, balance_after_vnd: 875000, occurred_at: "2026-08-30T00:00:00Z", note: "Cà phê sáng", with_person: "", event_ref: "", excluded_from_reports: false, version: 1 }] },
    });
    render(<App />);
    await userEvent.click(await screen.findByLabelText("Sổ giao dịch"));
    expect(await screen.findByText("Cà phê sáng")).toBeInTheDocument();
    expect(screen.getByText("-125.000 đ")).toBeInTheDocument();
    await userEvent.type(screen.getByLabelText("Tìm giao dịch"), "không tồn tại");
    expect(screen.getByText("Chưa có giao dịch")).toBeInTheDocument();
  });

  it("hydrates cached finance data and queues a transaction while offline", async () => {
    mockNavigatorOnline(false);
    await saveFinanceMirror({
      wallets: [{ id: "wallet_cached", name: "Ví cached", type: "cash", balance_vnd: 880000, include_in_total: true, is_default_ai: true, version: 3 }],
      categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống cached", is_system: false, version: 1 }],
      transactions: [{ id: "tx_cached", type: "expense", source_wallet_id: "wallet_cached", category_id: "cat_food", amount_vnd: 12000, balance_after_vnd: 868000, occurred_at: "2026-08-31T00:00:00Z", note: "Cached lunch", with_person: "", event_ref: "", excluded_from_reports: false, version: 2 }],
    });
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
    });

    render(<App />);

    expect(await screen.findByText("Ví cached")).toBeInTheDocument();
    await userEvent.click(screen.getByLabelText("Sổ giao dịch"));
    expect(await screen.findByText("Cached lunch")).toBeInTheDocument();
    await userEvent.click(await screen.findByLabelText("Thêm giao dịch"));
    await userEvent.type(await screen.findByLabelText("Số tiền"), "33000");
    await userEvent.type(screen.getByLabelText("Ghi chú"), "Offline dinner");
    await userEvent.click(screen.getAllByRole("button", { name: "Lưu" })[0]);

    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Thêm Giao Dịch" })).not.toBeInTheDocument());
    expect(await screen.findByText("1 chờ đồng bộ")).toBeInTheDocument();
    expect((await readOutbox())[0].input.note).toBe("Offline dinner");
    expect(fetchMock.mock.calls.map(([input]) => new URL(String(input), "http://localhost").pathname)).toEqual(["/api/v1/me"]);
  });

  it("reloads offline with cached auth, cached finance data, and pending mutation visibility", async () => {
    const onlineFetch = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví reload", type: "cash", balance_vnd: 990000, include_in_total: true, is_default_ai: true, version: 5 }] },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn reload", is_system: false, version: 1 }] },
      "/api/v1/transactions": { transactions: [{ id: "tx_reload", type: "expense", source_wallet_id: "wallet_live", category_id: "cat_food", amount_vnd: 11000, balance_after_vnd: 979000, occurred_at: "2026-08-31T00:00:00Z", note: "Online cached", with_person: "", event_ref: "", excluded_from_reports: false, version: 1 }] },
    });
    const firstRender = render(<App />);
    expect(await screen.findByText("Ví reload")).toBeInTheDocument();
    expect(onlineFetch).toHaveBeenCalled();
    firstRender.unmount();

    mockNavigatorOnline(false);
    const offlineFetch = vi.fn(async () => { throw new TypeError("offline"); });
    vi.stubGlobal("fetch", offlineFetch);
    render(<App />);

    expect(await screen.findByText("Ví reload")).toBeInTheDocument();
    await userEvent.click(screen.getByLabelText("Sổ giao dịch"));
    expect(await screen.findByText("Online cached")).toBeInTheDocument();
    await userEvent.click(await screen.findByLabelText("Thêm giao dịch"));
    await userEvent.type(await screen.findByLabelText("Số tiền"), "44000");
    await userEvent.type(screen.getByLabelText("Ghi chú"), "Reload pending");
    await userEvent.click(screen.getAllByRole("button", { name: "Lưu" })[0]);

    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Thêm Giao Dịch" })).not.toBeInTheDocument());
    expect(await screen.findByText("1 chờ đồng bộ")).toBeInTheDocument();
    expect(await screen.findByText("Reload pending")).toBeInTheDocument();
    expect(offlineFetch).toHaveBeenCalledTimes(1);
  });

  it("blocks offline writes when IndexedDB is unavailable", async () => {
    mockNavigatorOnline(false);
    vi.stubGlobal("indexedDB", undefined);
    mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
    });

    render(<App />);

    expect(await screen.findByText("Chỉ đọc offline")).toBeInTheDocument();
    expect(screen.getByLabelText("Thêm giao dịch")).toBeDisabled();
  });

  it("creates an expense from the add sheet", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, options?: RequestInit) => {
      const url = typeof input === "string" ? input : input instanceof URL ? input.toString() : input.url;
      const path = new URL(url, "http://localhost").pathname;
      if (path === "/api/v1/me") return new Response(JSON.stringify({ user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } }), { status: 200 });
      if (path === "/api/v1/wallets") return new Response(JSON.stringify({ wallets: [{ id: "wallet_live", name: "Ví API", type: "cash", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] }), { status: 200 });
      if (path === "/api/v1/categories") return new Response(JSON.stringify({ categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }] }), { status: 200 });
      if (path === "/api/v1/transactions" && options?.method === "POST") {
        expect(new Headers(options.headers).get("Idempotency-Key")).toBeTruthy();
        const body = JSON.parse(String(options.body));
        expect(body.amount_vnd).toBe(50000);
        return new Response(JSON.stringify({ transaction: { id: "tx_new", ...body, balance_after_vnd: 950000, version: 1 } }), { status: 201 });
      }
      return new Response(JSON.stringify({ transactions: [] }), { status: 200 });
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<App />);
    await userEvent.click(await screen.findByLabelText("Thêm giao dịch"));
    await userEvent.type(await screen.findByLabelText("Số tiền"), "50000");
    await userEvent.type(screen.getByLabelText("Ghi chú"), "Ăn sáng");
    await userEvent.click(screen.getAllByRole("button", { name: "Lưu" })[0]);
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Thêm Giao Dịch" })).not.toBeInTheDocument());
    expect(fetchMock).toHaveBeenCalled();
  });

  it("updates wallet/category settings from the manager sheet", async () => {
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "cash", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] },
      "/api/v1/categories": { categories: [{ id: "cat_custom", kind: "expense", name: "Cafe", is_system: false, version: 1 }] },
      "/api/v1/transactions": { transactions: [] },
    });

    render(<App />);
    await userEvent.click(await screen.findByRole("button", { name: "Xem tất cả" }));
    await userEvent.click(await screen.findByRole("button", { name: "Sửa" }));
    await userEvent.clear(await screen.findByLabelText("Tên ví Ví API"));
    await userEvent.type(screen.getByLabelText("Tên ví Ví API"), "Ví chính");
    await userEvent.click(screen.getAllByRole("button", { name: "Lưu" })[0]);

    await waitFor(() => {
      const updateCall = fetchMock.mock.calls.find(([input, options]) => new URL(String(input), "http://localhost").pathname === "/api/v1/wallets/wallet_live" && options?.method === "PATCH");
      expect(updateCall).toBeTruthy();
    });
  });

  it("edits and archives a transaction from the transaction list", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, options?: RequestInit) => {
      const url = typeof input === "string" ? input : input instanceof URL ? input.toString() : input.url;
      const path = new URL(url, "http://localhost").pathname;
      if (path === "/api/v1/me") return jsonResponse({ user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } });
      if (path === "/api/v1/wallets") return jsonResponse({ wallets: [{ id: "wallet_live", name: "Ví API", type: "cash", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] });
      if (path === "/api/v1/categories") return jsonResponse({ categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }] });
      if (path === "/api/v1/transactions/tx_1" && options?.method === "PATCH") {
        const body = JSON.parse(String(options.body));
        expect(body.note).toBe("Cà phê chiều");
        return jsonResponse({ transaction: { id: "tx_1", ...body, balance_after_vnd: 850000, version: 2 } });
      }
      if (path === "/api/v1/transactions/tx_1/archive" && options?.method === "POST") return jsonResponse({});
      return jsonResponse({ transactions: [{ id: "tx_1", type: "expense", source_wallet_id: "wallet_live", category_id: "cat_food", amount_vnd: 125000, balance_after_vnd: 875000, occurred_at: "2026-08-30T00:00:00Z", note: "Cà phê sáng", with_person: "", event_ref: "", excluded_from_reports: false, version: 1 }] });
    });
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    await userEvent.click(await screen.findByLabelText("Sổ giao dịch"));
    await userEvent.click(await screen.findByText("Cà phê sáng"));
    await userEvent.clear(screen.getByLabelText("Ghi chú"));
    await userEvent.type(screen.getByLabelText("Ghi chú"), "Cà phê chiều");
    await userEvent.click(screen.getByRole("button", { name: "Lưu thay đổi" }));
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Sửa Giao Dịch" })).not.toBeInTheDocument());

    await userEvent.click(await screen.findByText("Cà phê sáng"));
    await userEvent.click(screen.getByRole("button", { name: "Lưu trữ" }));
    await waitFor(() => expect(fetchMock.mock.calls.some(([input, options]) => new URL(String(input), "http://localhost").pathname === "/api/v1/transactions/tx_1/archive" && options?.method === "POST")).toBe(true));
  });

  it("shows conflict inbox and lets the user keep the server version", async () => {
    await saveFinanceMirror({
      wallets: [{ id: "wallet_live", name: "Ví API", type: "cash", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 2 }],
      categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }],
      transactions: [{ id: "tx_1", type: "expense", source_wallet_id: "wallet_live", category_id: "cat_food", amount_vnd: 12000, balance_after_vnd: 988000, occurred_at: "2026-08-31T00:00:00Z", note: "Server note", with_person: "", event_ref: "", excluded_from_reports: false, version: 2 }],
    });
    const mutation = await enqueueMutation({ entity_type: "transaction", entity_id: "tx_1", operation: "update", base_version: 1, payload: { note: "Offline note", amount_vnd: 13000 } });
    await saveOfflineConflict({
      conflict_id: mutation.mutation_id,
      mutation_id: mutation.mutation_id,
      entity_type: "transaction",
      entity_id: "tx_1",
      operation: "update",
      base_version: 1,
      server_version: 2,
      local_payload: { note: "Offline note", amount_vnd: 13000 },
      server_payload: { id: "tx_1", type: "expense", source_wallet_id: "wallet_live", category_id: "cat_food", amount_vnd: 12000, balance_after_vnd: 988000, occurred_at: "2026-08-31T00:00:00Z", note: "Server note", with_person: "", event_ref: "", excluded_from_reports: false, version: 2 },
      status: "open",
      created_at: "2026-08-31T00:00:00Z",
    });
    mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "cash", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 2 }] },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }] },
      "/api/v1/transactions": { transactions: [{ id: "tx_1", type: "expense", source_wallet_id: "wallet_live", category_id: "cat_food", amount_vnd: 12000, balance_after_vnd: 988000, occurred_at: "2026-08-31T00:00:00Z", note: "Server note", with_person: "", event_ref: "", excluded_from_reports: false, version: 2 }] },
    });

    render(<App />);

    expect(await screen.findByLabelText("Xung đột đồng bộ")).toBeInTheDocument();
    expect(screen.getByText("Offline: Offline note")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Giữ server" }));
    await waitFor(() => expect(screen.queryByLabelText("Xung đột đồng bộ")).not.toBeInTheDocument());
  });

  it("shows budget progress and creates a monthly budget", async () => {
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [] },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }] },
      "/api/v1/transactions": { transactions: [] },
      "/api/v1/budgets": {
        budgets: [{
          budget: { id: "budget_1", name: "Ăn uống", period_type: "monthly", amount_vnd: 500000, category_ids: ["cat_food"], all_categories: false, version: 1 },
          period_start: "2026-08-01",
          period_end: "2026-08-31",
          spent_vnd: 410000,
          remaining_vnd: 90000,
          percent: 82,
          alert_80: true,
          alert_100: false,
        }],
      },
      "/api/v1/events": { events: [] },
      "/api/v1/obligations": { obligations: [] },
      "/api/v1/recurring-schedules": { schedules: [] },
      "/api/v1/transaction-drafts": { drafts: [] },
    });

    render(<App />);
    await userEvent.click(await screen.findByLabelText("Ngân sách"));
    expect((await screen.findAllByText("Ăn uống")).length).toBeGreaterThan(0);
    expect(screen.getByText("Đã chạm 80%")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Tạo", exact: true }));
    await userEvent.type(await screen.findByLabelText("Tên ngân sách"), "Mua sắm");
    await userEvent.type(screen.getByLabelText("Số tiền ngân sách"), "1200000");
    await userEvent.click(screen.getAllByRole("button", { name: "Lưu" }).at(-1)!);

    await waitFor(() => {
      const createCall = fetchMock.mock.calls.find(([input, options]) => new URL(String(input), "http://localhost").pathname === "/api/v1/budgets" && options?.method === "POST");
      expect(createCall).toBeTruthy();
    });
  });

  it("shows event and debt planning flows", async () => {
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [] },
      "/api/v1/categories": { categories: [] },
      "/api/v1/transactions": { transactions: [] },
      "/api/v1/budgets": { budgets: [] },
      "/api/v1/events": { events: [{ id: "event_1", name: "Đà Lạt", starts_on: "2026-08-31", ends_on: "2026-09-02", note: "Trip", total_vnd: 125000, transaction_count: 1, version: 1 }] },
      "/api/v1/obligations": { obligations: [{ id: "debt_1", direction: "borrowed", principal_vnd: 1000000, counterparty: "Anh Minh", due_on: "2026-09-30", note: "Vay sửa nhà", repaid_vnd: 600000, remaining_vnd: 400000, version: 1 }] },
      "/api/v1/recurring-schedules": { schedules: [{ id: "schedule_1", name: "Tiền nhà", frequency: "monthly", timezone: "Asia/Ho_Chi_Minh", starts_at: "2026-08-01T02:00:00Z", next_occurs_at: "2026-09-01T02:00:00Z", type: "expense", source_wallet_id: "wallet_1", category_id: "cat_food", amount_vnd: 3500000, note: "Thuê nhà", version: 1 }] },
      "/api/v1/transaction-drafts": { drafts: [{ id: "draft_1", schedule_id: "schedule_1", occurrence_key: "recurring:schedule_1:2026-08-01T02:00:00Z", type: "expense", source_wallet_id: "wallet_1", category_id: "cat_food", amount_vnd: 3500000, occurred_at: "2026-08-01T02:00:00Z", note: "Thuê nhà", status: "pending", version: 1 }] },
    });

    render(<App />);
    await userEvent.click(await screen.findByLabelText("Ngân sách"));
    expect(await screen.findByText("Đà Lạt")).toBeInTheDocument();
    expect(screen.getByText("Đã dùng 125.000 đ · 1 giao dịch")).toBeInTheDocument();
    expect(screen.getByText("Anh Minh")).toBeInTheDocument();
    expect(screen.getByText("Còn 400.000 đ")).toBeInTheDocument();
    expect(screen.getByText("Tiền nhà")).toBeInTheDocument();
    expect(screen.getByText("3.500.000 đ · Chờ duyệt")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Tạo sự kiện" }));
    await userEvent.type(await screen.findByLabelText("Tên sự kiện"), "Du lịch Huế");
    await userEvent.click(screen.getAllByRole("button", { name: "Lưu", exact: true }).at(-1)!);

    await waitFor(() => {
      const createCall = fetchMock.mock.calls.find(([input, options]) => new URL(String(input), "http://localhost").pathname === "/api/v1/events" && options?.method === "POST");
      expect(createCall).toBeTruthy();
    });
  });

  it("creates a recurring schedule from the planning tab", async () => {
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_1", name: "Ví chính", type: "cash", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }] },
      "/api/v1/transactions": { transactions: [] },
      "/api/v1/budgets": { budgets: [] },
      "/api/v1/events": { events: [] },
      "/api/v1/obligations": { obligations: [] },
      "/api/v1/recurring-schedules": { schedules: [] },
      "/api/v1/transaction-drafts": { drafts: [] },
    });

    render(<App />);
    await userEvent.click(await screen.findByLabelText("Ngân sách"));
    await userEvent.click(screen.getByRole("button", { name: "Tạo lịch" }));
    await userEvent.type(await screen.findByLabelText("Tên lịch lặp"), "Tiền nhà");
    await userEvent.type(screen.getByLabelText("Số tiền lịch lặp"), "3500000");
    await userEvent.click(screen.getAllByRole("button", { name: "Lưu", exact: true }).at(-1)!);

    await waitFor(() => {
      const createCall = fetchMock.mock.calls.find(([input, options]) => new URL(String(input), "http://localhost").pathname === "/api/v1/recurring-schedules" && options?.method === "POST");
      expect(createCall).toBeTruthy();
    });
  });

  it("shows docs link in the Account tab", async () => {
    mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [] },
      "/api/v1/categories": { categories: [] },
      "/api/v1/transactions": { transactions: [] },
      "/api/v1/assets": { assets: [] },
      "/api/v1/portfolio/summary": { summary: { investment_market_value_vnd: 0, missing_price_count: 0, position_count: 0 } },
      "/api/v1/api-keys": { keys: [] },
    });
    render(<App />);
    await userEvent.click(await screen.findByLabelText("Tài khoản"));
    expect(screen.getByText("Tài liệu")).toBeInTheDocument();
    expect(screen.getByText("Sơ đồ CSDL và tài liệu API")).toBeInTheDocument();
    const link = screen.getByRole("link", { name: /tài liệu/i });
    expect(link).toHaveAttribute("href", "/docs/");
    expect(link).toHaveAttribute("target", "_blank");
  });
});

function mockNavigatorOnline(value: boolean) {
  Object.defineProperty(window.navigator, "onLine", {
    configurable: true,
    value,
  });
}

function mockFetchRoutes(routes: Record<string, unknown>) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL, options?: RequestInit) => {
    const url = typeof input === "string" ? input : input instanceof URL ? input.toString() : input.url;
    const path = new URL(url, "http://localhost").pathname;
    if (routes[path]) {
      return jsonResponse(routes[path]);
    }
    if (path === "/api/v1/budgets") {
      return jsonResponse({ budgets: [] });
    }
    if (path === "/api/v1/events") {
      return jsonResponse({ events: [] });
    }
    if (path === "/api/v1/obligations") {
      return jsonResponse({ obligations: [] });
    }
    if (path === "/api/v1/recurring-schedules") {
      return jsonResponse({ schedules: [] });
    }
    if (path === "/api/v1/transaction-drafts") {
      return jsonResponse({ drafts: [] });
    }
    if (options?.method && options.method !== "GET") {
      return jsonResponse({});
    }
    return new Response(JSON.stringify({ error: { code: "NOT_FOUND", message: "Not found" }, correlation_id: "req_test" }), {
      status: 404,
      headers: { "Content-Type": "application/json" },
    });
  });
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify({ status: "ok", correlation_id: "req_test", ...(body as Record<string, unknown>) }), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
