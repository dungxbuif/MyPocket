import { act, fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ReportsPanel, type ReportsPanelProps } from "./ReportsPanel";

const summary = { income_vnd: 900000, expense_vnd: 120000, net_income_vnd: 780000, daily_average_vnd: 4000, from: "2026-08-01", to: "2026-08-31", timezone: "Asia/Ho_Chi_Minh", generated_at: "2026-09-09T03:00:00Z", data_version: 1, not_comparable: false, income_change_percent: 50, expense_change_percent: 20 };
const report = { summary, categories: [{ category_name: "Ăn uống", amount_vnd: 120000, share_percent: 100 }], daily: [{ date: "2026-08-02", income_vnd: 500000, expense_vnd: 25000, net_income_vnd: 475000, cumulative_net_vnd: 650000 }], prior: { ...summary, from: "2026-07-01", to: "2026-07-31", income_vnd: 600000, expense_vnd: 100000, net_income_vnd: 500000 } };
const props: ReportsPanelProps = { userID: "owner", online: true, privacyMasked: false, wallets: [{ id: "cash", name: "Tiền mặt", type: "cash", balance_vnd: 1000000, include_in_total: true, is_default_ai: false, version: 1 }], onClose: vi.fn() };
const ok = (value: unknown = report) => new Response(JSON.stringify({ report: value }), { status: 200 });

function filters(kind = "cash-flow") {
  fireEvent.change(screen.getByLabelText("Loại báo cáo"), { target: { value: kind } });
  fireEvent.change(screen.getByLabelText("Từ ngày"), { target: { value: "2026-08-01" } });
  fireEvent.change(screen.getByLabelText("Đến ngày"), { target: { value: "2026-08-31" } });
  fireEvent.change(screen.getByLabelText("Ví báo cáo"), { target: { value: "cash" } });
}

afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks(); vi.unstubAllGlobals(); });

describe("ReportsPanel", () => {
  it("defaults dates to the current Ho Chi Minh month at a UTC month boundary", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-08-31T18:00:00Z"));
    render(<ReportsPanel {...props} />);
    expect(screen.getByLabelText("Từ ngày")).toHaveValue("2026-09-01");
    expect(screen.getByLabelText("Đến ngày")).toHaveValue("2026-09-30");
  });

  it.each([
    ["cash-flow", "Tổng hợp báo cáo", "780.000 đ"],
    ["categories", "Chi theo danh mục", "100%"],
    ["daily", "Báo cáo theo ngày", "475.000 đ"],
    ["comparison", "So sánh kỳ", "600.000 đ"],
    ["cumulative", "Lũy kế theo ngày", "650.000 đ"],
  ])("loads server %s values for selected dates and wallet", async (kind, tableName, amount) => {
    const fetcher = vi.fn(async (_url: RequestInfo | URL) => ok());
    vi.stubGlobal("fetch", fetcher);
    render(<ReportsPanel {...props} />);
    expect(screen.queryByText("0 đ")).not.toBeInTheDocument();
    filters(kind);
    await userEvent.click(screen.getByRole("button", { name: "Xem báo cáo" }));
    expect(within(await screen.findByRole("table", { name: tableName })).getByText(amount)).toBeVisible();
    const url = new URL(String(fetcher.mock.calls[0][0]), "http://localhost");
    expect(url.pathname).toBe(`/api/v1/reports/${kind}`);
    expect(Object.fromEntries(url.searchParams)).toEqual({ from: "2026-08-01", to: "2026-08-31", wallet_id: "cash" });
  });

  it("retains filters after safe failure and retries without showing zero totals", async () => {
    let failed = true;
    vi.stubGlobal("fetch", vi.fn(async () => failed ? new Response(JSON.stringify({ error: { code: "INTERNAL_FAILURE", message: "SECRET_TRACE" }, correlation_id: "req_report" }), { status: 500 }) : ok()));
    render(<ReportsPanel {...props} />);
    filters();
    await userEvent.click(screen.getByRole("button", { name: "Xem báo cáo" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("req_report");
    expect(screen.queryByText("SECRET_TRACE")).not.toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.queryByText("0 đ")).not.toBeInTheDocument();
    expect(screen.getByLabelText("Ví báo cáo")).toHaveValue("cash");
    failed = false;
    await userEvent.click(screen.getByRole("button", { name: "Thử lại báo cáo" }));
    expect(await screen.findByRole("table", { name: "Tổng hợp báo cáo" })).toBeVisible();
  });

  it("ignores an obsolete response after changing report filters", async () => {
    let resolveOld!: (value: Response) => void;
    const fetcher = vi.fn().mockImplementationOnce(() => new Promise<Response>(resolve => { resolveOld = resolve; })).mockImplementationOnce(async () => ok());
    vi.stubGlobal("fetch", fetcher);
    render(<ReportsPanel {...props} />);
    await userEvent.click(screen.getByRole("button", { name: "Xem báo cáo" }));
    filters("categories");
    await userEvent.click(screen.getByRole("button", { name: "Xem báo cáo" }));
    await screen.findByRole("table", { name: "Chi theo danh mục" });
    await act(async () => { resolveOld(ok({ ...report, summary: { ...summary, net_income_vnd: 999999 } })); });
    expect(screen.queryByText("999.999 đ")).not.toBeInTheDocument();
    expect(screen.getByRole("table", { name: "Chi theo danh mục" })).toBeVisible();
  });

  it("masks financial values and percentages already loaded in the DOM", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ok()));
    const view = render(<ReportsPanel {...props} />);
    filters("categories");
    await userEvent.click(screen.getByRole("button", { name: "Xem báo cáo" }));
    await screen.findByRole("table", { name: "Chi theo danh mục" });
    view.rerender(<ReportsPanel {...props} privacyMasked />);
    expect(screen.queryByText("120.000 đ")).not.toBeInTheDocument();
    expect(screen.queryByText("780.000 đ")).not.toBeInTheDocument();
    expect(screen.queryByText("100%")).not.toBeInTheDocument();
    expect(screen.getAllByText("••••••").length).toBeGreaterThan(0);
  });

  it("marks previously loaded data as stale offline and disables refresh", async () => {
    const fetcher = vi.fn(async () => ok());
    vi.stubGlobal("fetch", fetcher);
    const view = render(<ReportsPanel {...props} />);
    await userEvent.click(screen.getByRole("button", { name: "Xem báo cáo" }));
    await screen.findByRole("table", { name: "Tổng hợp báo cáo" });
    view.rerender(<ReportsPanel {...props} online={false} />);
    expect(screen.getByRole("status")).toHaveTextContent("Ngoại tuyến");
    expect(screen.getByRole("status")).toHaveTextContent("chưa đồng bộ");
    expect(screen.getByRole("button", { name: "Xem báo cáo" })).toBeDisabled();
    expect(screen.getByText(/Cập nhật/)).toHaveTextContent("9/9/2026");
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it("rejects reversed date ranges before requesting data", async () => {
    const fetcher = vi.fn();
    vi.stubGlobal("fetch", fetcher);
    render(<ReportsPanel {...props} />);
    filters();
    fireEvent.change(screen.getByLabelText("Từ ngày"), { target: { value: "2026-09-01" } });
    await userEvent.click(screen.getByRole("button", { name: "Xem báo cáo" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Ngày bắt đầu");
    expect(fetcher).not.toHaveBeenCalled();
  });

  it("clears another user's report and ignores their pending response", async () => {
    let finish!: (value: Response) => void;
    vi.stubGlobal("fetch", vi.fn(() => new Promise<Response>(resolve => { finish = resolve; })));
    const view = render(<ReportsPanel {...props} />);
    await userEvent.click(screen.getByRole("button", { name: "Xem báo cáo" }));
    view.rerender(<ReportsPanel {...props} userID="another-owner" wallets={[]} />);
    await act(async () => { finish(ok()); });
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.queryByText("780.000 đ")).not.toBeInTheDocument();
  });

  it("does not invent a percentage for a zero prior baseline", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => ok({ ...report, summary: { ...summary, income_change_percent: undefined }, prior: { ...report.prior, income_vnd: 0 } })));
    render(<ReportsPanel {...props} />);
    filters("comparison");
    await userEvent.click(screen.getByRole("button", { name: "Xem báo cáo" }));
    const comparison = await screen.findByRole("table", { name: "So sánh kỳ" });
    expect(within(comparison).getByRole("row", { name: /^Thu / })).toHaveTextContent("Không thể so sánh");
    expect(within(comparison).getByRole("row", { name: /^Chi / })).toHaveTextContent("20%");
  });

  it("closes the report panel through its close control", async () => {
    const onClose = vi.fn();
    render(<ReportsPanel {...props} onClose={onClose} />);
    await userEvent.click(screen.getByRole("button", { name: "Đóng báo cáo" }));
    expect(onClose).toHaveBeenCalledOnce();
  });
});
