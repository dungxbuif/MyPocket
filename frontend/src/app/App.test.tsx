import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";

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
    });

    render(<App />);
    await waitFor(() => expect(screen.queryByText("Đang kiểm tra phiên đăng nhập...")).not.toBeInTheDocument());
    await userEvent.click(screen.getByLabelText("Thêm giao dịch"));

    expect(screen.getByText("Ăn uống API")).toBeInTheDocument();
  });
});

function mockNavigatorOnline(value: boolean) {
  Object.defineProperty(window.navigator, "onLine", {
    configurable: true,
    value,
  });
}

function mockFetchRoutes(routes: Record<string, unknown>) {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL) => {
      const url = typeof input === "string" ? input : input instanceof URL ? input.toString() : input.url;
      const path = new URL(url, "http://localhost").pathname;
      const body = routes[path];
      if (!body) {
        return new Response(JSON.stringify({ error: { code: "NOT_FOUND", message: "Not found" }, correlation_id: "req_test" }), {
          status: 404,
          headers: { "Content-Type": "application/json" },
        });
      }
      return new Response(JSON.stringify({ status: "ok", correlation_id: "req_test", ...body }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }),
  );
}
