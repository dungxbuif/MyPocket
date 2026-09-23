import { dateKeyAt } from "./accountTime.ts";
import type { WalletType } from "./wallets.ts";

export function ledgerRangeForWallet(type: WalletType | undefined, range: { start: string; end: string }): { start: string; end: string } | null {
  return type === "goal" ? null : range;
}

export function filterLedgerTransactions<T extends { wallet_id: string; occurred_at: string }>(
  transactions: T[], walletID: string, range: { start: string; end: string } | null, timezone: string,
): T[] {
  return transactions.filter(transaction => {
    if (walletID && transaction.wallet_id !== walletID) return false;
    if (!range) return true;
    const day = dateKeyAt(transaction.occurred_at, timezone);
    return !!day && (!range.start || day >= range.start) && (!range.end || day <= range.end);
  }).sort((a, b) => Date.parse(b.occurred_at) - Date.parse(a.occurred_at));
}
