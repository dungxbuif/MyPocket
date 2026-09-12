import type { MockWallet } from "../data/mockFinance";
import { formatVND } from "../utils/format";

export function WalletCard({ wallet, masked }: { wallet: MockWallet; masked: boolean }) {
  const Icon = wallet.icon;
  const negative = wallet.balance < 0;
  return (
    <button className="flex w-full items-center gap-3 rounded-2xl bg-[#f5f3f3] p-3 text-left transition hover:bg-[#efeded]">
      <div className="grid h-10 w-10 place-items-center rounded-full bg-[#d9e6da] text-[#006e1c]">
        <Icon size={19} />
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate font-semibold">{wallet.name}</p>
        <p className="truncate text-xs text-[#3f4a3c]">{wallet.subtitle}</p>
      </div>
      <p className={`money text-sm font-bold ${negative ? "text-[#bb1614]" : "text-[#1b1c1c]"}`}>
        {masked ? "••••••" : formatVND(wallet.balance)}
      </p>
    </button>
  );
}

