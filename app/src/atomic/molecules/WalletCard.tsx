import { CreditCard, PiggyBank, WalletCards } from "lucide-react";

import { BaseButton } from "../atoms/BaseButton";
import { IconBadge } from "../atoms/IconBadge";
import { Text } from "../atoms/Text";
import { formatVND } from "../utils/format";
import type { Wallet } from "../../services/wallets";

const PRESENTATION = {
  basic: { icon: WalletCards, tone: "warning" as const, label: "Ví thường" },
  goal: { icon: PiggyBank, tone: "success" as const, label: "Ví tiết kiệm" },
  credit: { icon: CreditCard, tone: "categoryTeal" as const, label: "Ví tín dụng" },
};

export function WalletCard({ wallet, masked, onActivate }: { wallet: Wallet; masked: boolean; onActivate?: () => void }) {
  const presentation = PRESENTATION[wallet.type];
  return (
    <BaseButton variant="row" size="row" className="flex w-full items-center gap-3" onClick={onActivate}>
      <IconBadge icon={presentation.icon} tone={presentation.tone} />
      <div className="min-w-0 flex-1">
        <Text weight="semibold" className="truncate">{wallet.name}</Text>
        <Text size="xs" tone="secondary" className="truncate">{wallet.description || presentation.label}</Text>
      </div>
      <Text numeric weight="bold" tone={wallet.current_balance < 0 ? "danger" : "ink"} className="tracking-tight">{masked ? "••••••" : formatVND(wallet.current_balance)}</Text>
    </BaseButton>
  );
}
