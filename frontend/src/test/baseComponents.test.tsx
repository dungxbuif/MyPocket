import * as React from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { AmountInputHero } from "../components/inputs/AmountInputHero";
import { DateNavigationRow } from "../components/inputs/DateNavigationRow";
import { SegmentedControl } from "../components/ui/segmented-control";
import { Switch } from "../components/ui/switch";
import { GaugeArcSummary } from "../components/charts/GaugeArcSummary";
import { ProgressBarWithMarker } from "../components/charts/ProgressBarWithMarker";
import { ComparisonBarChart } from "../components/charts/ComparisonBarChart";
import { WalletFilterChip } from "../components/navigation/WalletFilterChip";
import { ExpenseDetailSheet } from "../screens/sheets/ExpenseDetailSheet";
import { WalletPickerSheet } from "../screens/sheets/WalletPickerSheet";

describe("Base UI Components & Charts", () => {
  it("renders AmountInputHero with VND currency badge and numeric input", async () => {
    function ControlledHero({ onValChange }: { onValChange: (v: number) => void }) {
      const [val, setVal] = React.useState(0);
      return (
        <AmountInputHero
          value={val}
          onChange={(v) => {
            setVal(v);
            onValChange(v);
          }}
        />
      );
    }

    const handleChange = vi.fn();
    render(<ControlledHero onValChange={handleChange} />);

    expect(screen.getByText("VND")).toBeInTheDocument();
    const input = screen.getByRole("textbox") as HTMLInputElement;

    await userEvent.type(input, "500000");
    expect(handleChange).toHaveBeenLastCalledWith(500000);
    expect(input.value).toBe("500.000");
  });

  it("renders DateNavigationRow and steps day forward and backward", async () => {
    const handleDateChange = vi.fn();
    render(<DateNavigationRow currentDate="2026-08-23" onDateChange={handleDateChange} />);

    expect(screen.getByText("Ngày")).toBeInTheDocument();
    expect(screen.getByText(/Chủ Nhật, 23\/08\/2026/)).toBeInTheDocument();

    const buttons = screen.getAllByRole("button");
    // Button 0 is previous day (<), Button 1 is next day (>)
    await userEvent.click(buttons[0]);
    expect(handleDateChange).toHaveBeenCalledWith("2026-08-22");

    await userEvent.click(buttons[1]);
    expect(handleDateChange).toHaveBeenCalledWith("2026-08-24");
  });

  it("handles SegmentedControl tab switching", async () => {
    const handleChange = vi.fn();
    render(
      <SegmentedControl
        options={[
          { key: "expense", label: "Khoản chi" },
          { key: "income", label: "Khoản thu" },
          { key: "debt", label: "Vay/nợ" },
        ]}
        selectedKey="expense"
        onChange={handleChange}
      />
    );

    const incomeBtn = screen.getByRole("tab", { name: "Khoản thu" });
    expect(incomeBtn).toBeInTheDocument();
    expect(incomeBtn).toHaveAttribute("aria-selected", "false");

    await userEvent.click(incomeBtn);
    expect(handleChange).toHaveBeenCalledWith("income");
  });

  it("toggles Switch component", async () => {
    const handleChange = vi.fn();
    render(<Switch checked={false} onCheckedChange={handleChange} />);

    const switchBtn = screen.getByRole("switch");
    expect(switchBtn).toHaveAttribute("aria-checked", "false");

    await userEvent.click(switchBtn);
    expect(handleChange).toHaveBeenCalledWith(true);
  });

  it("renders GaugeArcSummary with spendable amount and 3-stat breakdown", () => {
    render(
      <GaugeArcSummary
        availableAmount={45075000}
        totalBudget={65000000}
        totalSpent={19920000}
        daysRemaining={12}
      />
    );

    expect(screen.getByText("Số tiền bạn có thể chi")).toBeInTheDocument();
    expect(screen.getByText("45.075.000 đ")).toBeInTheDocument();
    expect(screen.getByText("65 M đ")).toBeInTheDocument();
    expect(screen.getByText(/19(\.|,)92 M đ/)).toBeInTheDocument();
    expect(screen.getByText("12 ngày")).toBeInTheDocument();
  });

  it("renders ProgressBarWithMarker with 'Hôm nay' indicator", () => {
    render(
      <ProgressBarWithMarker
        spentPercent={35}
        dayProgressPercent={50}
        spentAmountFormatted="2.190.000 đ"
        remainingAmountFormatted="Còn lại 5.810.000 đ"
      />
    );

    expect(screen.getByText("2.190.000 đ")).toBeInTheDocument();
    expect(screen.getByText("Còn lại 5.810.000 đ")).toBeInTheDocument();
    expect(screen.getByText("Hôm nay")).toBeInTheDocument();
  });

  it("renders ComparisonBarChart with comparative bars", () => {
    render(
      <ComparisonBarChart
        data={[
          { label: "Tuần trước", amount: 150000, amountFormatted: "150.000 đ" },
          { label: "Tuần này", amount: 250000, amountFormatted: "250.000 đ", isCurrent: true },
        ]}
      />
    );

    expect(screen.getByText("Tuần trước")).toBeInTheDocument();
    expect(screen.getByText("Tuần này")).toBeInTheDocument();
    expect(screen.getByText("150.000 đ")).toBeInTheDocument();
    expect(screen.getByText("250.000 đ")).toBeInTheDocument();
  });

  it("renders WalletFilterChip with wallet title", () => {
    render(<WalletFilterChip walletName="Techcombank" />);
    expect(screen.getByText("Techcombank")).toBeInTheDocument();
  });

  it("renders ExpenseDetailSheet with donut breakdown and close button", async () => {
    const handleClose = vi.fn();
    render(
      <ExpenseDetailSheet
        isOpen={true}
        onClose={handleClose}
        categoryName="Ăn uống"
        totalSpent={250000}
      />
    );

    expect(screen.getByRole("dialog", { name: "Chi tiết khoản chi" })).toBeInTheDocument();
    expect(screen.getAllByText("250.000 đ")[0]).toBeInTheDocument();

    const closeBtn = screen.getByRole("button", { name: "Đóng" });
    await userEvent.click(closeBtn);
    expect(handleClose).toHaveBeenCalled();
  });

  it("renders WalletPickerSheet and handles wallet selection", async () => {
    const handleSelect = vi.fn();
    const handleClose = vi.fn();
    render(
      <WalletPickerSheet
        isOpen={true}
        onClose={handleClose}
        wallets={[{ id: "w_1", name: "Ví Tiền Mặt", balance_vnd: 500000, type: "cash", include_in_total: true, is_default_ai: false, version: 1 }]}
        onSelectWallet={handleSelect}
      />
    );

    expect(screen.getByRole("dialog", { name: "Chọn Ví" })).toBeInTheDocument();
    expect(screen.getByText("Tổng cộng")).toBeInTheDocument();
    expect(screen.getByText("Ví Tiền Mặt")).toBeInTheDocument();

    await userEvent.click(screen.getByText("Ví Tiền Mặt"));
    expect(handleSelect).toHaveBeenCalledWith("w_1");
  });
});
