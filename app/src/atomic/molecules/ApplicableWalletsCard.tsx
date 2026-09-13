import { CreditCard, PiggyBank, WalletCards } from "lucide-react";
import { BaseCheckbox } from "../atoms/BaseCheckbox";
import { IconBadge } from "../atoms/IconBadge";
import { SurfaceCard } from "../atoms/SurfaceCard";
import type { Wallet } from "../../services/wallets";

const COPY = {
  title: "Ví áp dụng",
} as const;

const WALLET_ICONS = {
  basic: WalletCards,
  goal: PiggyBank,
  credit: CreditCard,
} as const;

export function ApplicableWalletsCard({ wallets, selectedWalletIDs, onToggle }: { wallets: Wallet[]; selectedWalletIDs: string[]; onToggle: (walletID: string) => void }) {
  return <SurfaceCard padding="md"><p className="pb-3 pt-1 text-[13.5px] font-medium leading-snug text-slate-600">{COPY.title}</p><div className="divide-y divide-slate-100">{wallets.map((wallet) => <BaseCheckbox key={wallet.id} label={wallet.name} checked={selectedWalletIDs.includes(wallet.id)} onChange={() => onToggle(wallet.id)}><WalletName wallet={wallet} /></BaseCheckbox>)}</div></SurfaceCard>;
}

function WalletName({ wallet }: { wallet: Wallet }) {
  const Icon = WALLET_ICONS[wallet.type];
  return <span className="flex items-center gap-3"><IconBadge icon={Icon} size="sm" shape="circle" /><span className="min-w-0 flex-1 truncate text-base font-semibold tracking-tight text-slate-800">{wallet.name}</span></span>;
}
