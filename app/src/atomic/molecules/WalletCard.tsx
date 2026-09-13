import type { MockWallet } from "../data/mockFinance";
import { formatVND } from "../utils/format";
import { IconBadge } from "../atoms/IconBadge";
import { WALLET_BADGE_CLASSES } from "../atoms/tokens";

export function WalletCard({ wallet, masked }: { wallet: MockWallet; masked: boolean }) {
  const Icon = wallet.icon;
  const negative = wallet.balance < 0;
  return (
    <button className="flex w-full cursor-pointer items-center gap-3 rounded-xl px-2 py-3.5 text-left transition hover:bg-[#f5f3f3]">
      <IconBadge icon={Icon} className={WALLET_BADGE_CLASSES[wallet.kind]} />
      <div className="min-w-0 flex-1">
        <p className="truncate font-semibold">{wallet.name}</p>
        <p className="truncate text-xs text-[#3f4a3c]">{wallet.subtitle}</p>
      </div>
      <p className={`money text-sm font-bold tracking-tight ${negative ? "text-[#bb1614]" : "text-[#1b1c1c]"}`}>
        {masked ? "••••••" : formatVND(wallet.balance)}
      </p>
    </button>
  );
}
