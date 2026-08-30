import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { saveFinanceMirror } from "../offline/db";
import { App } from "./App";
import { readOutbox } from "./outbox";

describe("App shell", () => {
  afterEach(() => {
    mockNavigatorOnline(true);
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("renders mobile navigation destinations without finance data", () => {
    render(<App />);

    expect(screen.getByRole("navigation")).toBeInTheDocument();
    expect(screen.getByLabelText("Tổng quan")).toBeInTheDocument();
    expect(screen.getByLabelText("Sổ giao dịch")).toBeInTheDocument();
    expect(screen.getByLabelText("Thêm giao dịch")).toBeInTheDocument();
    expect(screen.getByLabelText("Ngân sách")).toBeInTheDocument();
    expect(screen.getByLabelText("Tài khoản")).toBeInTheDocument();
  });

  it("shows offline state when the browser is offline", () => {
    mockNavigatorOnline(false);

    render(<App />);

    expect(screen.getByText("Offline")).toBeInTheDocument();
  });

  it("opens the transaction add sheet from the raised add action", async () => {
    render(<App />);

    await userEvent.click(screen.getByLabelText("Thêm giao dịch"));

    expect(screen.getByRole("dialog", { name: "Thêm Giao Dịch" })).toBeInTheDocument();
    expect(screen.getByText("Chọn nhóm")).toBeInTheDocument();
    expect(screen.getByText("Thêm Hình Ảnh")).toBeInTheDocument();
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
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống API", system_key: "expense_food", is_system: true }] },
      "/api/v1/transactions": { transactions: [] },
    });

    render(<App />);

    expect(await screen.findByText("Ví API")).toBeInTheDocument();
    expect(screen.getByText("Tiết kiệm API")).toBeInTheDocument();
    expect(screen.getAllByText("1.234.567 đ")).toHaveLength(2);
    expect(screen.getByText("1.000.000 đ")).toBeInTheDocument();
    expect(screen.queryByText("Techcombank")).not.toBeInTheDocument();
  });

  it("shows loaded categories in the add transaction sheet", async () => {
    mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [] },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống API", system_key: "expense_food", is_system: true }] },
      "/api/v1/transactions": { transactions: [] },
    });

    render(<App />);
    await waitFor(() => expect(screen.queryByText("Đang kiểm tra phiên đăng nhập...")).not.toBeInTheDocument());
    await userEvent.click(screen.getByLabelText("Thêm giao dịch"));

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
      categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống cached", is_system: false }],
      transactions: [{ id: "tx_cached", type: "expense", source_wallet_id: "wallet_cached", category_id: "cat_food", amount_vnd: 12000, balance_after_vnd: 868000, occurred_at: "2026-08-31T00:00:00Z", note: "Cached lunch", with_person: "", event_ref: "", excluded_from_reports: false, version: 2 }],
    });
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
    });

    render(<App />);

    expect(await screen.findByText("Ví cached")).toBeInTheDocument();
    await userEvent.click(screen.getByLabelText("Sổ giao dịch"));
    expect(await screen.findByText("Cached lunch")).toBeInTheDocument();
    await userEvent.click(screen.getByLabelText("Thêm giao dịch"));
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
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn reload", is_system: false }] },
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
    await userEvent.click(screen.getByLabelText("Thêm giao dịch"));
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

    await userEvent.click(screen.getByLabelText("Thêm giao dịch"));

    expect(await screen.findByText("Chỉ đọc offline")).toBeInTheDocument();
    expect(screen.getByText("Offline storage chưa sẵn sàng. Mở mạng lại để lưu giao dịch.")).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Lưu" })[0]).toBeDisabled();
  });

  it("creates an expense from the add sheet", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL, options?: RequestInit) => {
      const url = typeof input === "string" ? input : input instanceof URL ? input.toString() : input.url;
      const path = new URL(url, "http://localhost").pathname;
      if (path === "/api/v1/me") return new Response(JSON.stringify({ user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } }), { status: 200 });
      if (path === "/api/v1/wallets") return new Response(JSON.stringify({ wallets: [{ id: "wallet_live", name: "Ví API", type: "cash", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] }), { status: 200 });
      if (path === "/api/v1/categories") return new Response(JSON.stringify({ categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true }] }), { status: 200 });
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
    await userEvent.click(screen.getByLabelText("Thêm giao dịch"));
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
      "/api/v1/categories": { categories: [{ id: "cat_custom", kind: "expense", name: "Cafe", is_system: false }] },
      "/api/v1/transactions": { transactions: [] },
    });

    render(<App />);
    await userEvent.click(await screen.findByRole("button", { name: "Xem tất cả" }));
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
      if (path === "/api/v1/categories") return jsonResponse({ categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true }] });
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
