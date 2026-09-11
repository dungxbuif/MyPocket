import type { TransactionInput } from "./finance";

export function calendarDateInHoChiMinh(value: Date = new Date()): string {
  const parts = new Intl.DateTimeFormat("en", {
    timeZone: "Asia/Ho_Chi_Minh",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(value);
  const part = (type: Intl.DateTimeFormatPartTypes) => parts.find((item) => item.type === type)?.value ?? "";
  return `${part("year")}-${part("month")}-${part("day")}`;
}

export function buildTransactionInput(input: {
  type: "expense" | "income" | "transfer"; amount: string; sourceWalletID: string;
  destinationWalletID?: string;
  categoryID?: string; note: string; excludedFromReports: boolean;
  occurredOn: string; receiptObjectID?: string;
}): TransactionInput {
  return {
    type: input.type, source_wallet_id: input.sourceWalletID,
    destination_wallet_id: input.type === "transfer" ? input.destinationWalletID : undefined,
    category_id: input.type === "transfer" ? undefined : input.categoryID,
    receipt_object_id: input.receiptObjectID, amount_vnd: Number(input.amount),
    target_balance_vnd: null, occurred_at: new Date(`${input.occurredOn}T12:00:00+07:00`).toISOString(),
    note: input.note, excluded_from_reports: input.excludedFromReports,
  };
}
