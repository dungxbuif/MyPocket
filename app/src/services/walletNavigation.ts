export type TransactionsSearch = {} | { wallet_id: string };

export function transactionsSearch(walletID: string | null | undefined): TransactionsSearch {
  const value = walletID?.trim() ?? "";
  return value ? { wallet_id: value } : {};
}

export function walletIDFromTransactionsSearch(searchString: string): string {
  const query = searchString.startsWith("?") ? searchString.slice(1) : searchString;
  return new URLSearchParams(query).get("wallet_id")?.trim() ?? "";
}
