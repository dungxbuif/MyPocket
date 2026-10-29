export function totalWalletBalance(wallets: Array<{ current_balance: number; is_in_total: boolean }>): number {
  return wallets.reduce((total, wallet) => total + (wallet.is_in_total ? wallet.current_balance : 0), 0);
}
