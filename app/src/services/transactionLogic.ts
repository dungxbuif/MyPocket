export type TransactionDirection = "income" | "expense";

export function categoryAppliesToTransaction(category: { kind: string; wallet_ids: string[] }, type: TransactionDirection, walletID: string): boolean {
  return category.kind === type && (category.wallet_ids.length === 0 || category.wallet_ids.includes(walletID));
}

export function signedTransactionAmount(transaction: { type: TransactionDirection; amount: number }): number {
  return transaction.type === "income" ? transaction.amount : -transaction.amount;
}
