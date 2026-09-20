export type TransactionDirection = "income" | "expense";

export function categoryAppliesToTransaction(category: { kind: string; wallet_ids: string[]; system_key?: string | null }, type: TransactionDirection, walletID: string, walletType = "basic"): boolean {
  const savingsAllowed = type === "income"
    ? category.system_key === "income_transfer_in" || category.system_key === "income_interest"
    : category.system_key === "expense_transfer_out";
  return category.kind === type && (walletType !== "goal" || savingsAllowed) && (category.wallet_ids.length === 0 || category.wallet_ids.includes(walletID));
}

export function signedTransactionAmount(transaction: { type: TransactionDirection; amount: number }): number {
  return transaction.type === "income" ? transaction.amount : -transaction.amount;
}
