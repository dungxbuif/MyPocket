import { Text } from "../atoms/Text";
import { BaseButton } from "../atoms/BaseButton";
import type { MockWallet } from "../data/mockFinance";
import { formatVND } from "../utils/format";
import { IconBadge } from "../atoms/IconBadge";
import { WALLET_BADGE_TONES } from "../../ui/domainVariants";

export function WalletCard({ wallet, masked }: { wallet: MockWallet; masked: boolean }) {
  const Icon = wallet.icon;
  const negative = wallet.balance < 0;
  return (
    <BaseButton variant="row" size="row" className="flex w-full items-center gap-3">
      <IconBadge icon={Icon} tone={WALLET_BADGE_TONES[wallet.kind]} />
      <div className="min-w-0 flex-1">
        <Text weight="semibold" className="truncate">{wallet.name}</Text>
        <Text size="xs" tone="secondary" className="truncate">{wallet.subtitle}</Text>
      </div>
      <Text numeric weight="bold" tone={negative ? "danger" : "ink"} className="tracking-tight">
        {masked ? "••••••" : formatVND(wallet.balance)}
      </Text>
    </BaseButton>
  );
}
