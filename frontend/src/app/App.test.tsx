import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { enqueueMutation, saveFinanceMirror, saveOfflineConflict } from "../offline/db";
import { App } from "./App";
import { readOutbox } from "./outbox";
import * as offlineDB from "../offline/db";

describe("App shell", () => {
  async function openReceiptForm() {
    const fetcher = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "basic", balance_vnd: 988000, include_in_total: true, is_default_ai: true, version: 1 }] },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }] },
      "/api/v1/transactions": { transactions: [] },
    });
    render(<App />);
    await screen.findByText("Ví API");
    await userEvent.click(screen.getByRole("button", { name: "Thêm giao dịch" }));
    return fetcher;
  }

  it("keeps receipt selection visible without opening details and after cancelling the picker", async () => {
    await openReceiptForm();
    const input = screen.getByLabelText("Ảnh đính kèm") as HTMLInputElement;
    await userEvent.upload(input, new File(["receipt"], "bill.png", { type: "image/png" }));
    expect(screen.getByRole("status")).toHaveTextContent("bill.png");
    await userEvent.click(screen.getByRole("button", { name: "Thêm chi tiết" }));
    await userEvent.click(screen.getByRole("button", { name: "Ẩn chi tiết" }));
    fireEvent.change(input, { target: { files: [] } });
    expect(screen.getByRole("status")).toHaveTextContent("bill.png");
  });

  it("lets the user remove and reselect the same receipt", async () => {
    await openReceiptForm();
    const input = screen.getByLabelText("Ảnh đính kèm") as HTMLInputElement;
    const file = new File(["receipt"], "same.png", { type: "image/png" });
    await userEvent.upload(input, file);
    await userEvent.click(screen.getByRole("button", { name: "Bỏ ảnh" }));
    expect(screen.queryByText(/same.png/)).not.toBeInTheDocument();
    await userEvent.upload(input, file);
    expect(screen.getByRole("status")).toHaveTextContent("same.png");
  });

  it("links the selected receipt to exactly the queued offline transaction", async () => {
    await openReceiptForm();
    await userEvent.upload(screen.getByLabelText("Ảnh đính kèm"), new File(["receipt"], "offline.png", { type: "image/png" }));
    await userEvent.type(screen.getByLabelText("Số tiền"), "23000");
    mockNavigatorOnline(false);
    await userEvent.click(screen.getByRole("button", { name: "Lưu" }));
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Thêm Giao Dịch" })).not.toBeInTheDocument());
    const transactions = await offlineDB.readPendingMutations();
    const receipts = await offlineDB.readPendingReceiptUploads();
    expect(transactions).toHaveLength(1);
    expect(receipts).toHaveLength(1);
    expect(transactions[0]).toMatchObject({ entity_type: "transaction", operation: "create", payload: { amount_vnd: 23000 } });
    expect(receipts[0]).toMatchObject({ transaction_id: transactions[0].entity_id, filename: "offline.png", content_type: "image/png" });
  });

  it("does not silently discard an income or expense receipt when switching to debt", async () => {
    await openReceiptForm();
    await userEvent.upload(screen.getByLabelText("Ảnh đính kèm"), new File(["receipt"], "keep.png", { type: "image/png" }));
    await userEvent.type(screen.getByLabelText("Số tiền"), "23000");
    await userEvent.click(screen.getByRole("button", { name: "Vay/nợ" }));
    await userEvent.type(screen.getByLabelText("Đối tác"), "Lan");
    expect(screen.getByRole("button", { name: "Lưu" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Đính kèm ảnh" })).toBeDisabled();
    expect(screen.getByRole("status")).toHaveTextContent("keep.png");
    expect(screen.getByText(/Ảnh chỉ hỗ trợ/)).toBeVisible();
    await userEvent.click(screen.getByRole("button", { name: "Bỏ ảnh" }));
    expect(screen.getByRole("button", { name: "Lưu" })).toBeEnabled();
  });

  it("locks receipt controls and transaction type while a save is pending", async () => {
    const fetcher = await openReceiptForm();
    const base = fetcher.getMockImplementation()!;
    let finish!: () => void;
    fetcher.mockImplementation(async (url, options) => {
      if (String(url).endsWith('/transactions') && options?.method === 'POST') {
        await new Promise<void>(resolve => { finish = resolve; });
        return jsonResponse({ transaction: { id: "saved", ...JSON.parse(String(options.body)), balance_after_vnd: 965000, version: 1 } });
      }
      return base(url, options);
    });
    await userEvent.type(screen.getByLabelText("Số tiền"), "23000");
    await userEvent.click(screen.getByRole("button", { name: "Thêm chi tiết" }));
    await userEvent.click(screen.getByRole("button", { name: "Lưu" }));
    expect(screen.getByRole("button", { name: "Đính kèm ảnh" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Thêm Hình Ảnh" })).toBeDisabled();
    expect(screen.getByLabelText("Ảnh đính kèm")).toBeDisabled();
    expect(screen.getByRole("button", { name: "Vay/nợ" })).toBeDisabled();
    await act(async () => { finish(); });
  });

  it("locks selected receipt controls when storage-degraded mode goes offline", async () => {
    await openReceiptForm();
    await userEvent.upload(screen.getByLabelText("Ảnh đính kèm"), new File(["receipt"], "read-only.png", { type: "image/png" }));
    await userEvent.click(screen.getByRole("button", { name: "Thêm chi tiết" }));
    vi.stubGlobal("indexedDB", undefined);
    await act(async () => { mockNavigatorOnline(false); window.dispatchEvent(new Event("offline")); });
    await screen.findByText("Offline storage chưa sẵn sàng. Mở mạng lại để lưu giao dịch.");
    expect(screen.getByRole("button", { name: "Đính kèm ảnh" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Đổi ảnh" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Bỏ ảnh" })).toBeDisabled();
    expect(screen.getByLabelText("Ảnh đính kèm")).toBeDisabled();
    expect(screen.getByRole("status")).toHaveTextContent("read-only.png");
  });

  it("marks unsupported quick-add details as unavailable rather than inert actions", async () => {
    await openReceiptForm();
    await userEvent.click(screen.getByRole("button", { name: "Thêm chi tiết" }));
    for (const label of ["Với", "Đặt vị trí", "Chọn sự kiện", "Đặt nhắc nhở"]) {
      const button = screen.getByRole("button", { name: new RegExp(label) });
      expect(button).toBeDisabled();
      expect(button).toHaveTextContent("Chưa hỗ trợ trong form này");
    }
  });

  it("does not recreate a saved offline transaction when its receipt queue fails", async () => {
    mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "basic", balance_vnd: 988000, include_in_total: true, is_default_ai: true, version: 1 }] },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }] },
      "/api/v1/transactions": { transactions: [] },
    });
    render(<App />);
    await screen.findByText("Ví API");
    await userEvent.click(screen.getByRole("button", { name: "Thêm giao dịch" }));
    await userEvent.type(screen.getByLabelText("Số tiền"), "12000");
    await userEvent.click(screen.getByRole("button", { name: "Thêm chi tiết" }));
    await userEvent.upload(document.getElementById("receipt-image") as HTMLInputElement, new File(["receipt"], "bill.png", { type: "image/png" }));
    vi.spyOn(offlineDB, "queueReceiptUpload").mockRejectedValue(new DOMException("Quota exceeded", "QuotaExceededError"));
    mockNavigatorOnline(false);
    await userEvent.click(screen.getByRole("button", { name: "Lưu" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Giao dịch đã lưu");
    expect(screen.getByRole("button", { name: "Lưu" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Đính kèm ảnh" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Đổi ảnh" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Bỏ ảnh" })).toBeDisabled();
    expect(screen.getByLabelText("Ảnh đính kèm")).toBeDisabled();
    expect(await readOutbox()).toHaveLength(1);
    expect((await readOutbox())[0].input).toMatchObject({ type: "expense", amount_vnd: 12000 });
  });

  it.each(["create", "edit", "archive"])("preserves the transaction form and reports a rejected %s without automatic retry", async operation => {
    const transaction = { id: "tx_error", type: "expense", source_wallet_id: "wallet_live", category_id: "cat_food", amount_vnd: 12000, balance_after_vnd: 988000, occurred_at: "2026-09-09T00:00:00Z", note: "Giữ ghi chú", with_person: "", event_ref: "", excluded_from_reports: false, version: 1 };
    const fetcher = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "basic", balance_vnd: 988000, include_in_total: true, is_default_ai: true, version: 1 }] },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }] },
      "/api/v1/transactions": { transactions: [transaction] },
    });
    const base = fetcher.getMockImplementation()!;
    fetcher.mockImplementation(async (url, options) => options?.method && String(url).includes('/transactions')
      ? jsonResponse({ status: "error", error: { code: "VALIDATION_FAILED", message: "RAW_NOT_FOR_UI" }, correlation_id: "req_transaction_failure" }, 422)
      : base(url, options));
    render(<App />);
    await screen.findByText("Ví API");
    if (operation === "create") {
      await userEvent.click(screen.getByRole("button", { name: "Thêm giao dịch" }));
      await userEvent.type(screen.getByLabelText("Số tiền"), "12000");
      await userEvent.type(screen.getByLabelText("Ghi chú"), "Giữ ghi chú");
    } else {
      await userEvent.click(screen.getByRole("button", { name: "Sổ giao dịch" }));
      await userEvent.click(await screen.findByRole("button", { name: /Giữ ghi chú/ }));
    }
    await userEvent.click(screen.getByRole("button", { name: operation === "create" ? "Lưu" : operation === "edit" ? "Lưu thay đổi" : "Lưu trữ" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("req_transaction_failure");
    expect(screen.getByLabelText("Số tiền")).toHaveValue("12000");
    expect(screen.getByLabelText("Ghi chú")).toHaveValue("Giữ ghi chú");
    expect(screen.queryByText("RAW_NOT_FOR_UI")).not.toBeInTheDocument();
    expect(fetcher.mock.calls.filter(([url, options]) => options?.method && String(url).includes('/transactions'))).toHaveLength(1);
  });

  it("does not submit the previous owner's queued wallet after account switching", async () => {
    await saveFinanceMirror({ userID: "previous-owner", wallets: [] });
    await enqueueMutation({ entity_type: "wallet", entity_id: "private-wallet", operation: "create", base_version: 0, payload: { name: "Private previous owner", type: "basic" } });
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com" } },
      "/api/v1/wallets": { wallets: [] }, "/api/v1/categories": { categories: [] }, "/api/v1/transactions": { transactions: [] },
    });
    render(<App />);
    await waitFor(async () => expect(await readOutbox()).toEqual([]));
    expect(fetchMock.mock.calls.filter(([input]) => String(input).includes("/sync/mutations"))).toHaveLength(0);
  });
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

  it("renders the renewed product shell with a distinct operating desk header", async () => {
    mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [] },
      "/api/v1/categories": { categories: [] },
      "/api/v1/transactions": { transactions: [] },
    });
    render(<App />);

    expect(await screen.findByText("Money Command")).toBeInTheDocument();
    expect(screen.getByText("Today Desk")).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: "Điều hướng chính" })).toHaveClass("dock-nav");
  });

  it("shows a bottom-corner PWA install prompt when the browser offers installation", async () => {
    mockFetchRoutes({ "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } } });
    render(<App />);
    const installEvent = Object.assign(new Event("beforeinstallprompt", { cancelable: true }), {
      prompt: vi.fn().mockResolvedValue(undefined),
      userChoice: Promise.resolve({ outcome: "accepted" }),
    });

    window.dispatchEvent(installEvent);

    expect(await screen.findByRole("dialog", { name: "Cài MyPocket" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Cài ứng dụng" }));
    expect(installEvent.prompt).toHaveBeenCalledOnce();
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Cài MyPocket" })).not.toBeInTheDocument());
  });

  it.each(["dismissed", "rejected"])("keeps install help recoverable after native prompt is %s", async outcome => {
    mockFetchRoutes({ "/api/v1/me": { user: { id: "user_123", email: "a@example.com" } } });
    render(<App />);
    const installEvent = Object.assign(new Event("beforeinstallprompt", { cancelable: true }), {
      prompt: outcome === "rejected" ? vi.fn().mockRejectedValue(new Error("browser failure")) : vi.fn().mockResolvedValue(undefined),
      userChoice: Promise.resolve({ outcome: "dismissed" }),
    });
    window.dispatchEvent(installEvent);
    await userEvent.click(await screen.findByRole("button", { name: "Cài ứng dụng" }));
    if (outcome === "rejected") expect(await screen.findByRole("alert")).toHaveTextContent("Không mở được cài đặt");
    await userEvent.click(await screen.findByRole("button", { name: "Cài ứng dụng" }));
    expect(screen.getByText("Safari: Chia sẻ → Thêm vào Màn hình chính")).toBeInTheDocument();
    expect(installEvent.prompt).toHaveBeenCalledOnce();
    await userEvent.click(screen.getByRole("button", { name: "Để sau" }));
    expect(screen.queryByRole("dialog", { name: "Cài MyPocket" })).not.toBeInTheDocument();
  });

  it("shows iPhone installation instructions when no native install prompt is available", async () => {
    mockFetchRoutes({ "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } } });
    render(<App />);

    await userEvent.click(await screen.findByRole("button", { name: "Cài ứng dụng" }));

    expect(screen.getByText("Safari: Chia sẻ → Thêm vào Màn hình chính")).toBeInTheDocument();
  });

  it("allows dismissing installation help without losing account navigation", async () => {
    mockFetchRoutes({ "/api/v1/me": { user: { id: "user_123", email: "a@example.com" } } });
    render(<App />);
    await userEvent.click(await screen.findByRole("button", { name: "Để sau" }));
    expect(screen.queryByRole("dialog", { name: "Cài MyPocket" })).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Tài khoản" }));
    expect(screen.getByRole("button", { name: "Đăng xuất" })).toBeInTheDocument();
  });

  it("removes the install suggestion after appinstalled", async () => {
    mockFetchRoutes({ "/api/v1/me": { user: { id: "user_123", email: "a@example.com" } } });
    render(<App />);
    await screen.findByRole("dialog", { name: "Cài MyPocket" });
    window.dispatchEvent(new Event("appinstalled"));
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Cài MyPocket" })).not.toBeInTheDocument());
  });

  it("does not suggest installing when already running standalone", async () => {
    vi.stubGlobal("matchMedia", (query: string) => ({ matches: query === "(display-mode: standalone)", addEventListener: vi.fn(), removeEventListener: vi.fn() }));
    mockFetchRoutes({ "/api/v1/me": { user: { id: "user_123", email: "a@example.com" } } });
    render(<App />);
    await screen.findByRole("navigation");
    expect(screen.queryByRole("dialog", { name: "Cài MyPocket" })).not.toBeInTheDocument();
  });

  it("renders the overview expense chart from the daily API report", async () => {
    mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [] }, "/api/v1/categories": { categories: [] }, "/api/v1/transactions": { transactions: [] },
      "/api/v1/reports/daily": { report: { summary: { income_vnd: 0, expense_vnd: 42000, net_income_vnd: -42000, generated_at: "2026-09-08T00:00:00Z", timezone: "Asia/Ho_Chi_Minh", from: "2026-09-01", to: "2026-09-08", data_version: 1 }, daily: [{ date: "2026-09-07", income_vnd: 0, expense_vnd: 42000, net_income_vnd: -42000, cumulative_net_vnd: -42000 }] } },
    });

    render(<App />);

    expect((await screen.findAllByText("42.000 đ")).length).toBeGreaterThan(1);
    expect(screen.getByText("07/09")).toBeInTheDocument();
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
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "basic", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] },
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
          { id: "wallet_live", name: "Ví API", type: "basic", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 2 },
          { id: "wallet_savings", name: "Tiết kiệm API", type: "goal", balance_vnd: 234567, include_in_total: true, is_default_ai: false, version: 1 },
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

  it("creates goal wallets with behavior type and goal metadata", async () => {
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [] },
      "/api/v1/categories": { categories: [] },
      "/api/v1/transactions": { transactions: [] },
    });
    render(<App />);

    await screen.findByText("Tổng quan");
    await userEvent.click(await screen.findByRole("button", { name: "Xem tất cả" }));
    await userEvent.click(screen.getByRole("button", { name: "Thêm ví" }));
    await userEvent.type(screen.getByLabelText("Tên ví mới"), "Quỹ du lịch");
    const typeSelect = screen.getByLabelText("Loại ví") as HTMLSelectElement;
    expect([...typeSelect.options].map((option) => option.value)).toEqual(["basic", "goal", "credit"]);
    await userEvent.selectOptions(typeSelect, "goal");
    await userEvent.type(screen.getByLabelText("Mục tiêu số tiền"), "25000000");
    fireEvent.change(screen.getByLabelText("Ngày hoàn thành mục tiêu"), { target: { value: "2027-12-31" } });
    await userEvent.click(screen.getByRole("button", { name: "Tạo ví" }));

    await waitFor(() => {
      const createCall = fetchMock.mock.calls.find(([input, options]) => new URL(String(input), "http://localhost").pathname === "/api/v1/wallets" && options?.method === "POST");
      expect(createCall).toBeTruthy();
      expect(JSON.parse(String(createCall?.[1]?.body))).toMatchObject({ name: "Quỹ du lịch", type: "goal", goal_target_vnd: 25_000_000, goal_deadline_on: "2027-12-31" });
    });
  });

  it("shows loaded categories in the add transaction sheet", async () => {
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "basic", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] },
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
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "basic", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] },
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
      userID: "user_123",
      wallets: [{ id: "wallet_cached", name: "Ví cached", type: "basic", balance_vnd: 880000, include_in_total: true, is_default_ai: true, version: 3 }],
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
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví reload", type: "basic", balance_vnd: 990000, include_in_total: true, is_default_ai: true, version: 5 }] },
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
      if (path === "/api/v1/wallets") return new Response(JSON.stringify({ wallets: [{ id: "wallet_live", name: "Ví API", type: "basic", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] }), { status: 200 });
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
    await waitFor(() => expect(screen.getByLabelText("Thêm giao dịch")).toBeEnabled());
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
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "basic", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] },
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
      if (path === "/api/v1/wallets") return jsonResponse({ wallets: [{ id: "wallet_live", name: "Ví API", type: "basic", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] });
      if (path === "/api/v1/categories") return jsonResponse({ categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }] });
      if (path === "/api/v1/transactions/tx_1" && options?.method === "PATCH") {
        const body = JSON.parse(String(options.body));
        expect(body.note).toBe("Cà phê chiều");
        expect(body.base_version).toBe(1);
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
      userID: "user_123",
      wallets: [{ id: "wallet_live", name: "Ví API", type: "basic", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 2 }],
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
      "/api/v1/wallets": { wallets: [{ id: "wallet_live", name: "Ví API", type: "basic", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 2 }] },
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
    await userEvent.click(screen.getByRole("button", { name: "Tạo" }));
    await userEvent.type(await screen.findByLabelText("Tên ngân sách"), "Mua sắm");
    await userEvent.type(screen.getByLabelText("Số tiền ngân sách"), "1200000");
    await userEvent.click(screen.getAllByRole("button", { name: "Lưu" }).at(-1)!);

    await waitFor(() => {
      const createCall = fetchMock.mock.calls.find(([input, options]) => new URL(String(input), "http://localhost").pathname === "/api/v1/budgets" && options?.method === "POST");
      expect(createCall).toBeTruthy();
    });
  });

  it("assigns a selected budget when creating an expense transaction", async () => {
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_1", name: "Ví chính", type: "basic", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }] },
      "/api/v1/transactions": { transactions: [] },
      "/api/v1/budgets": {
        budgets: [{
          budget: { id: "budget_food", name: "Hũ ăn uống", period_type: "monthly", amount_vnd: 500000, category_ids: ["cat_food"], all_categories: false, version: 1 },
          period_start: "2026-09-01",
          period_end: "2026-09-30",
          spent_vnd: 0,
          remaining_vnd: 500000,
          percent: 0,
          alert_80: false,
          alert_100: false,
        }],
      },
      "/api/v1/events": { events: [] },
      "/api/v1/obligations": { obligations: [] },
      "/api/v1/recurring-schedules": { schedules: [] },
      "/api/v1/transaction-drafts": { drafts: [] },
    });

    render(<App />);
    await screen.findByText("Ví chính");
    await userEvent.click(screen.getByRole("button", { name: "Thêm giao dịch" }));
    await userEvent.type(await screen.findByLabelText("Số tiền"), "75000");
    await userEvent.selectOptions(screen.getByLabelText("Ngân sách giao dịch"), "budget_food");
    await userEvent.click(screen.getByRole("button", { name: "Lưu" }));

    await waitFor(() => {
      const createCall = fetchMock.mock.calls.find(([input, options]) => new URL(String(input), "http://localhost").pathname === "/api/v1/transactions" && options?.method === "POST");
      expect(createCall).toBeTruthy();
      expect(JSON.parse(String(createCall?.[1]?.body))).toMatchObject({ budget_id: "budget_food", type: "expense", amount_vnd: 75000 });
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
    await userEvent.click(screen.getAllByRole("button", { name: "Lưu" }).at(-1)!);

    await waitFor(() => {
      const createCall = fetchMock.mock.calls.find(([input, options]) => new URL(String(input), "http://localhost").pathname === "/api/v1/events" && options?.method === "POST");
      expect(createCall).toBeTruthy();
    });
  });

  it("creates a recurring schedule from the planning tab", async () => {
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_1", name: "Ví chính", type: "basic", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] },
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
    await userEvent.click(screen.getAllByRole("button", { name: "Lưu" }).at(-1)!);

    await waitFor(() => {
      const createCall = fetchMock.mock.calls.find(([input, options]) => new URL(String(input), "http://localhost").pathname === "/api/v1/recurring-schedules" && options?.method === "POST");
      expect(createCall).toBeTruthy();
    });
  });

  it("edits and pauses an existing recurring schedule from the planning tab", async () => {
    const fetchMock = mockFetchRoutes({
      "/api/v1/me": { user: { id: "user_123", email: "a@example.com", email_verified: true, display_name: "A", avatar_url: "" } },
      "/api/v1/wallets": { wallets: [{ id: "wallet_1", name: "Ví chính", type: "basic", balance_vnd: 1000000, include_in_total: true, is_default_ai: true, version: 1 }] },
      "/api/v1/categories": { categories: [{ id: "cat_food", kind: "expense", name: "Ăn uống", is_system: true, version: 1 }] },
      "/api/v1/transactions": { transactions: [] },
      "/api/v1/budgets": {
        budgets: [{
          budget: { id: "budget_food", name: "Hũ ăn uống", period_type: "monthly", amount_vnd: 500000, category_ids: ["cat_food"], all_categories: false, version: 1 },
          period_start: "2026-09-01",
          period_end: "2026-09-30",
          spent_vnd: 0,
          remaining_vnd: 500000,
          percent: 0,
          alert_80: false,
          alert_100: false,
        }],
      },
      "/api/v1/events": { events: [] },
      "/api/v1/obligations": { obligations: [] },
      "/api/v1/recurring-schedules": { schedules: [{ id: "schedule_1", name: "Tiền nhà", frequency: "monthly", timezone: "Asia/Ho_Chi_Minh", starts_at: "2026-09-01T02:00:00Z", next_occurs_at: "2026-10-01T02:00:00Z", type: "expense", source_wallet_id: "wallet_1", category_id: "cat_food", amount_vnd: 3500000, note: "Thuê nhà", posting_mode: "draft", budget_id: "budget_food", version: 2 }] },
      "/api/v1/transaction-drafts": { drafts: [] },
    });

    render(<App />);
    await userEvent.click(await screen.findByLabelText("Ngân sách"));
    await userEvent.click(await screen.findByRole("button", { name: /Tiền nhà/ }));
    await userEvent.clear(await screen.findByLabelText("Số tiền lịch lặp"));
    await userEvent.type(screen.getByLabelText("Số tiền lịch lặp"), "3600000");
    await userEvent.selectOptions(screen.getByLabelText("Cách ghi lịch lặp"), "auto_post");
    await userEvent.type(screen.getByLabelText("Ngày kết thúc lịch lặp"), "2026-12-31");
    await userEvent.click(screen.getByRole("button", { name: "Lưu thay đổi" }));

    await waitFor(() => {
      const updateCall = fetchMock.mock.calls.find(([input, options]) => new URL(String(input), "http://localhost").pathname === "/api/v1/recurring-schedules/schedule_1" && options?.method === "PATCH");
      expect(updateCall).toBeTruthy();
      expect(JSON.parse(String(updateCall?.[1]?.body))).toMatchObject({ base_version: 2, amount_vnd: 3600000, posting_mode: "auto_post", budget_id: "budget_food", ends_at: "2026-12-31T23:59:59+07:00" });
    });

    await userEvent.click(await screen.findByRole("button", { name: /Tiền nhà/ }));
    await userEvent.click(await screen.findByRole("button", { name: "Tạm dừng" }));

    await waitFor(() => {
      const pauseCall = fetchMock.mock.calls.find(([input, options]) => new URL(String(input), "http://localhost").pathname === "/api/v1/recurring-schedules/schedule_1/pause" && options?.method === "POST");
      expect(pauseCall).toBeTruthy();
      expect(JSON.parse(String(pauseCall?.[1]?.body))).toMatchObject({ base_version: 2 });
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
