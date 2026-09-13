import { Text } from "../atoms/Text";
import { BaseButton } from "../atoms/BaseButton";
import type { MockTransaction } from "../data/mockFinance";
import { formatVND } from "../utils/format";
import { IconBadge } from "../atoms/IconBadge";
import { TRANSACTION_BADGE_TONES } from "../../ui/domainVariants";

export function TransactionItem({ transaction }: { transaction: MockTransaction }) {
  const Icon = transaction.icon;
  const positive = transaction.amount > 0;
  const neutral = transaction.kind === "transfer";
  return (
    <BaseButton variant="row" size="row" className="flex w-full items-center gap-3">
      <IconBadge icon={Icon} tone={TRANSACTION_BADGE_TONES[transaction.kind]} />
      <div className="min-w-0 flex-1">
        <Text weight="semibold" className="truncate">{transaction.title}</Text>
        <Text size="xs" tone="secondary" className="truncate">
          {transaction.category} · {transaction.wallet}
          {transaction.note ? ` · ${transaction.note}` : ""}
        </Text>
      </div>
      <Text numeric weight="bold" tone={positive ? "action" : neutral ? "secondary" : "danger"} className="tracking-tight">
        {formatVND(transaction.amount)}
      </Text>
    </BaseButton>
  );
}
