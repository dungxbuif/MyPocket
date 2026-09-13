import type { MockTransaction } from "../data/mockFinance";
import { formatVND } from "../utils/format";
import { IconBadge } from "../atoms/IconBadge";
import { TRANSACTION_BADGE_CLASSES } from "../atoms/tokens";

export function TransactionItem({ transaction }: { transaction: MockTransaction }) {
  const Icon = transaction.icon;
  const positive = transaction.amount > 0;
  const neutral = transaction.kind === "transfer";
  return (
    <button className="flex w-full cursor-pointer items-center gap-3 rounded-xl px-2 py-3 text-left transition hover:bg-[#f5f3f3]">
      <IconBadge icon={Icon} className={TRANSACTION_BADGE_CLASSES[transaction.kind]} />
      <div className="min-w-0 flex-1">
        <p className="truncate font-semibold">{transaction.title}</p>
        <p className="truncate text-xs text-[#3f4a3c]">
          {transaction.category} · {transaction.wallet}
          {transaction.note ? ` · ${transaction.note}` : ""}
        </p>
      </div>
      <p className={`money text-sm font-bold tracking-tight ${positive ? "text-[#006e1c]" : neutral ? "text-[#3f4a3c]" : "text-[#bb1614]"}`}>
        {formatVND(transaction.amount)}
      </p>
    </button>
  );
}
