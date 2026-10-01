export type TransactionDirection = "income" | "expense";
export type AdjustmentDirection = "increase" | "decrease";

export function categoryAppliesToTransaction(category: { kind: string; wallet_ids: string[]; system_key?: string | null }, type: TransactionDirection, walletID: string, walletType = "basic"): boolean {
  const savingsAllowed = type === "income"
    ? category.system_key === "income_transfer_in" || category.system_key === "income_interest"
    : category.system_key === "expense_transfer_out";
  return category.kind === type && (walletType !== "goal" || savingsAllowed) && (category.wallet_ids.length === 0 || category.wallet_ids.includes(walletID));
}

export function signedTransactionAmount(transaction: { type: TransactionDirection | "adjustment"; adjustment_direction?: AdjustmentDirection | null; amount: number }): number {
  if (transaction.type === "income") return transaction.amount;
  if (transaction.type === "adjustment") return transaction.adjustment_direction === "increase" ? transaction.amount : -transaction.amount;
  return -transaction.amount;
}
