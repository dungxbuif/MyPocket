import type { LucideIcon } from "lucide-react";

import { BaseButton } from "../atoms/BaseButton";
import { IconBadge } from "../atoms/IconBadge";
import { Text } from "../atoms/Text";
import { formatVND } from "../utils/format";
import type { BadgeTone } from "../../ui/variants";

export type TransactionItemModel = {
  title: string;
  metadata: string;
  amount: number;
  kind: "income" | "expense" | "transfer" | "debt";
  icon: LucideIcon;
  tone: BadgeTone;
};

export function TransactionItem({ item, onActivate }: { item: TransactionItemModel; onActivate?: () => void }) {
  return (
    <BaseButton variant="row" size="row" className="flex w-full items-center gap-3" onClick={onActivate} disabled={!onActivate}>
      <IconBadge icon={item.icon} tone={item.tone} />
      <div className="min-w-0 flex-1">
        <Text weight="semibold" className="truncate">{item.title}</Text>
        <Text size="xs" tone="secondary" className="truncate">{item.metadata}</Text>
      </div>
      <Text numeric weight="bold" tone={item.kind === "income" ? "action" : item.kind === "transfer" ? "secondary" : "danger"} className="tracking-tight">{formatVND(item.amount)}</Text>
    </BaseButton>
  );
}
