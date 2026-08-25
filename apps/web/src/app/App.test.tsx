import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";

describe("App shell", () => {
  afterEach(() => {
    mockNavigatorOnline(true);
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
});

function mockNavigatorOnline(value: boolean) {
  Object.defineProperty(window.navigator, "onLine", {
    configurable: true,
    value,
  });
}
