import { describe, expect, it } from "vitest";

import { buildTransactionInput, calendarDateInHoChiMinh } from "./transactionInput";

describe("buildTransactionInput", () => {
  it("anchors a selected calendar date to noon in Asia/Ho_Chi_Minh", () => {
    const input = buildTransactionInput({
      type: "expense",
      amount: "50000",
      sourceWalletID: "wallet_cash",
      categoryID: "category_food",
      note: "Bữa trưa",
      excludedFromReports: false,
      occurredOn: "2026-09-10",
    });

    expect(input.occurred_at).toBe("2026-09-10T05:00:00.000Z");
  });
});

describe("calendarDateInHoChiMinh", () => {
  it("uses the Vietnam calendar day instead of UTC or device timezone", () => {
    expect(calendarDateInHoChiMinh(new Date("2026-09-09T18:30:00.000Z"))).toBe("2026-09-10");
  });
});
