import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { Transaction } from "../app/finance";
import { TransactionsScreen } from "./TransactionsScreen";

const transfer: Transaction = {
  id: "transfer-1",
  type: "transfer",
  source_wallet_id: "cash",
  destination_wallet_id: "bank",
  amount_vnd: 350000,
  balance_after_vnd: 650000,
  occurred_at: "2026-09-09T03:00:00Z",
  note: "Gửi tiền tiết kiệm",
  with_person: "",
  event_ref: "",
  excluded_from_reports: true,
  version: 1,
};

describe("TransactionsScreen accounting presentation", () => {
  it("shows the transfer amount neutrally and opens that same single transaction for editing", async () => {
    const onEdit = vi.fn();
    render(<TransactionsScreen transactions={[transfer]} onEdit={onEdit} />);

    const rows = screen.getAllByRole("button", { name: /Gửi tiền tiết kiệm/ });
    expect(rows).toHaveLength(1);
    const amount = within(rows[0]).getByText("350.000 đ");
    expect(amount).toBeVisible();
    expect(amount).toHaveClass("text-[#111111]");
    expect(amount).not.toHaveClass("text-[#32a9df]", "text-[#ff5a66]");
    expect(within(rows[0]).getByText("Chuyển khoản")).toBeVisible();
    await userEvent.click(rows[0]);
    expect(onEdit).toHaveBeenCalledExactlyOnceWith(transfer);
  });

  it("shows the adjustment target balance neutrally and identifies it as a balance", () => {
    const adjustment: Transaction = {
      ...transfer,
      id: "adjustment-1",
      type: "adjustment",
      destination_wallet_id: undefined,
      amount_vnd: 800000,
      balance_after_vnd: 800000,
      note: "Kiểm kê tiền mặt",
    };
    render(<TransactionsScreen transactions={[adjustment]} onEdit={vi.fn()} />);

    const row = screen.getByRole("button", { name: /Kiểm kê tiền mặt/ });
    const amount = within(row).getByText("800.000 đ");
    expect(amount).toBeVisible();
    expect(amount).toHaveClass("text-[#111111]");
    expect(amount).not.toHaveClass("text-[#32a9df]", "text-[#ff5a66]");
    expect(within(row).getByText("Số dư sau điều chỉnh")).toBeVisible();
  });

  it("keeps income positive and expenses negative", () => {
    render(<TransactionsScreen transactions={[
      { ...transfer, id: "income-1", type: "income", destination_wallet_id: undefined, amount_vnd: 900000, note: "Lương" },
      { ...transfer, id: "expense-1", type: "expense", destination_wallet_id: undefined, amount_vnd: 125000, note: "Ăn uống" },
    ]} onEdit={vi.fn()} />);

    expect(within(screen.getByRole("button", { name: /Lương/ })).getByText("+900.000 đ")).toHaveClass("text-[#32a9df]");
    expect(within(screen.getByRole("button", { name: /Ăn uống/ })).getByText("-125.000 đ")).toHaveClass("text-[#ff5a66]");
  });
});
