import { act, configure, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";

const wallet = { id: "wallet-main", name: "Ví chính", type: "basic", balance_vnd: 500000, include_in_total: true, is_default_ai: true, version: 2 };
const parent = { id: "cat-food", kind: "expense", name: "Ăn uống", is_system: false, version: 3 };
const child = { id: "cat-cafe", parent_id: "cat-food", kind: "expense", name: "Cà phê", is_system: false, version: 4 };
const income = { id: "cat-salary", kind: "income", name: "Lương", is_system: false, version: 1 };

configure({ asyncUtilTimeout: 10_000 });

describe("mounted category hierarchy and wallet settings", { timeout: 15_000 }, () => {
  afterEach(() => {
    setOnline(true);
    localStorage.clear();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("submits the selected same-kind parent when creating a category", async () => {
    const fetcher = financeFetch();
    render(<App />);
    await openWalletManager();

    await userEvent.type(screen.getByLabelText("Tên nhóm mới"), "Ăn sáng");
    expect(screen.getByLabelText("Nhóm cha mới")).not.toHaveTextContent("Lương");
    await userEvent.selectOptions(screen.getByLabelText("Nhóm cha mới"), parent.id);
    await userEvent.click(screen.getByRole("button", { name: "Tạo nhóm" }));

    await waitFor(() => {
      const call = requestCall(fetcher, "/api/v1/categories", "POST");
      expect(call).toBeTruthy();
      expect(JSON.parse(String(call?.[1]?.body))).toEqual({ name: "Ăn sáng", kind: "expense", parent_id: "cat-food" });
    });
  });

  it("preserves the category draft and selected parent after a safe validation error", async () => {
    const fetcher = financeFetch({
      mutate: async (path, method) => path === "/api/v1/categories" && method === "POST"
        ? errorResponse("VALIDATION_FAILED", "RAW_PARENT_VALIDATION_DETAIL", "req_category_create", 400)
        : undefined,
    });
    render(<App />);
    await openWalletManager();

    await userEvent.type(screen.getByLabelText("Tên nhóm mới"), "Ăn sáng");
    await userEvent.selectOptions(screen.getByLabelText("Nhóm cha mới"), parent.id);
    await userEvent.click(screen.getByRole("button", { name: "Tạo nhóm" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Chưa lưu được thay đổi nhóm");
    expect(screen.getByLabelText("Tên nhóm mới")).toHaveValue("Ăn sáng");
    expect(screen.getByLabelText("Nhóm cha mới")).toHaveValue(parent.id);
    expect(screen.queryByText("RAW_PARENT_VALIDATION_DETAIL")).not.toBeInTheDocument();
    expect(requestCall(fetcher, "/api/v1/categories", "POST")).toBeTruthy();
  });

  it("submits name, parent and optimistic version when editing a custom category", async () => {
    const fetcher = financeFetch();
    render(<App />);
    await openWalletManager();

    expect(screen.getByLabelText("Nhóm cha Ăn uống")).not.toHaveTextContent("Cà phê");
    await userEvent.clear(screen.getByLabelText("Tên nhóm Cà phê"));
    await userEvent.type(screen.getByLabelText("Tên nhóm Cà phê"), "Cà phê sáng");
    await userEvent.selectOptions(screen.getByLabelText("Nhóm cha Cà phê"), "");
    await userEvent.click(screen.getByRole("button", { name: "Lưu nhóm Cà phê" }));

    await waitFor(() => {
      const call = requestCall(fetcher, "/api/v1/categories/cat-cafe", "PATCH");
      expect(call).toBeTruthy();
      expect(JSON.parse(String(call?.[1]?.body))).toEqual({ name: "Cà phê sáng", parent_id: null, base_version: 4 });
    });
  });

  it("loads wallet category settings and reloads canonical state after a toggle", async () => {
    let activeSetting = true;
    let finishToggle!: () => void;
    const fetcher = financeFetch({
      settings: () => {
        return [{ ...child, active: activeSetting }];
      },
      mutate: async (path, method) => {
        if (path !== "/api/v1/wallets/wallet-main/categories/cat-cafe" || method !== "PUT") return undefined;
        await new Promise<void>((resolve) => { finishToggle = resolve; });
        activeSetting = false;
        return ok({});
      },
    });
    render(<App />);
    await openWalletManager();

    const active = await findSettingToggle("Cà phê đang bật");
    expect(active).toHaveAttribute("aria-pressed", "true");
    await userEvent.click(active);
    expect(active).toBeDisabled();
    expect(active).toHaveTextContent("Đang lưu");
    await act(async () => { finishToggle(); });

    const inactive = await findSettingToggle("Cà phê đang tắt");
    expect(inactive).toHaveAttribute("aria-pressed", "false");
    const call = requestCall(fetcher, "/api/v1/wallets/wallet-main/categories/cat-cafe", "PUT");
    expect(JSON.parse(String(call?.[1]?.body))).toEqual({ active: false });
  }, 15_000);

  it("retains confirmed wallet settings and offers reload after a safe toggle error", async () => {
    const fetcher = financeFetch({
      mutate: async (path, method) => path === "/api/v1/wallets/wallet-main/categories/cat-cafe" && method === "PUT"
        ? errorResponse("VALIDATION_FAILED", "RAW_SETTING_DETAIL", "req_setting", 400)
        : undefined,
    });
    render(<App />);
    await openWalletManager();

    await userEvent.click(await findSettingToggle("Cà phê đang bật"));

    expect(await screen.findByRole("alert")).toHaveTextContent("Chưa cập nhật được cài đặt nhóm");
    expect(screen.getByRole("button", { name: "Cà phê đang bật" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.queryByText("RAW_SETTING_DETAIL")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Tải lại cài đặt nhóm" })).toBeEnabled();
    expect(requestCall(fetcher, "/api/v1/wallets/wallet-main/categories/cat-cafe", "PUT")).toBeTruthy();
  });

  it("recovers a failed wallet-settings read through an explicit reload", async () => {
    let failRead = true;
    financeFetch({
      mutate: async (path, method) => {
        if (path !== "/api/v1/wallets/wallet-main/category-settings" || method !== "GET" || !failRead) return undefined;
        failRead = false;
        return errorResponse("INTERNAL_RETRYABLE", "RAW_READ_DETAIL", "req_settings_read", 503);
      },
    });
    render(<App />);
    await openWalletManager();

    expect(await screen.findByRole("alert")).toHaveTextContent("Chưa tải được cài đặt nhóm theo ví");
    expect(screen.queryByText("RAW_READ_DETAIL")).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Tải lại cài đặt nhóm" }));

    expect(await findSettingToggle("Cà phê đang bật")).toBeEnabled();
  });

  it("loads settings for the wallet selected in the manager", async () => {
    const savings = { ...wallet, id: "wallet-savings", name: "Ví tiết kiệm", type: "goal", is_default_ai: false };
    const fetcher = financeFetch({
      wallets: [wallet, savings],
      mutate: async (path, method) => path === "/api/v1/wallets/wallet-savings/category-settings" && method === "GET"
        ? ok({ categories: [{ ...child, active: false }] })
        : undefined,
    });
    render(<App />);
    await openWalletManager();
    await waitFor(() => expect(requestCall(fetcher, "/api/v1/wallets/wallet-main/category-settings", "GET")).toBeTruthy(), { timeout: 3000 });
    await findSettingToggle("Cà phê đang bật");

    await userEvent.selectOptions(screen.getByLabelText("Ví cài đặt nhóm"), savings.id);

    expect(await findSettingToggle("Cà phê đang tắt")).toBeEnabled();
    expect(requestCall(fetcher, "/api/v1/wallets/wallet-savings/category-settings", "GET")).toBeTruthy();
  });

  it("disables wallet category mutations after the app goes offline", async () => {
    financeFetch();
    render(<App />);
    await openWalletManager();
    const toggle = await findSettingToggle("Cà phê đang bật");

    await act(async () => {
      setOnline(false);
      window.dispatchEvent(new Event("offline"));
    });

    expect(screen.getByText("Cần online để tải và thay đổi cài đặt nhóm theo ví.")).toBeVisible();
    expect(toggle).toBeDisabled();
  });
});

async function openWalletManager() {
  await userEvent.click(await screen.findByRole("button", { name: "Xem tất cả" }));
  expect(await screen.findByRole("dialog", { name: "Ví Của Tôi" })).toBeVisible();
}

function findSettingToggle(name: string) {
  return screen.findByRole("button", { name }, { timeout: 10_000 });
}

function requestCall(fetcher: ReturnType<typeof vi.fn>, path: string, method: string) {
  return fetcher.mock.calls.find(([input, options]) => new URL(String(input), "http://localhost").pathname === path && (options?.method ?? "GET") === method);
}

function financeFetch(options: {
  wallets?: Array<Record<string, unknown>>;
  settings?: () => Array<Record<string, unknown>>;
  mutate?: (path: string, method: string) => Promise<Response | undefined>;
} = {}) {
  const fetcher = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const path = new URL(String(input), "http://localhost").pathname;
    const method = init?.method ?? "GET";
    const mutationResponse = await options.mutate?.(path, method);
    if (mutationResponse) return mutationResponse;
    if (path === "/api/v1/me") return ok({ user: { id: "category-owner", email: "owner@example.com", email_verified: true, display_name: "Owner", avatar_url: "" } });
    if (path === "/api/v1/wallets") return ok({ wallets: options.wallets ?? [wallet] });
    if (path === "/api/v1/categories" && method === "GET") return ok({ categories: [parent, child, income] });
    if (path === "/api/v1/categories" && method === "POST") return ok({ category: { id: "cat-breakfast", ...JSON.parse(String(init?.body)), is_system: false, version: 1 } }, 201);
    if (path === "/api/v1/categories/cat-cafe" && method === "PATCH") return ok({ category: { ...child, ...JSON.parse(String(init?.body)), version: 5 } });
    if (path === "/api/v1/wallets/wallet-main/category-settings") return ok({ categories: options.settings?.() ?? [{ ...child, active: true }] });
    if (path === "/api/v1/wallets/wallet-main/categories/cat-cafe" && method === "PUT") return ok({});
    const empty: Record<string, unknown> = {
      "/api/v1/transactions": { transactions: [] },
      "/api/v1/budgets": { budgets: [] },
      "/api/v1/events": { events: [] },
      "/api/v1/obligations": { obligations: [] },
      "/api/v1/recurring-schedules": { schedules: [] },
      "/api/v1/transaction-drafts": { drafts: [] },
      "/api/v1/notifications": { notifications: [] },
      "/api/v1/assets": { assets: [] },
    };
    return path in empty ? ok(empty[path] as Record<string, unknown>) : errorResponse("INTERNAL_RETRYABLE", "Unavailable", "req_unavailable", 503);
  });
  vi.stubGlobal("fetch", fetcher);
  return fetcher;
}

function ok(body: Record<string, unknown>, status = 200) {
  return new Response(JSON.stringify({ status: "ok", correlation_id: "req_ok", ...body }), { status, headers: { "Content-Type": "application/json" } });
}

function errorResponse(code: string, message: string, correlationID: string, status: number) {
  return new Response(JSON.stringify({ error: { code, message }, correlation_id: correlationID }), { status, headers: { "Content-Type": "application/json" } });
}

function setOnline(value: boolean) {
  Object.defineProperty(window.navigator, "onLine", { configurable: true, value });
}
