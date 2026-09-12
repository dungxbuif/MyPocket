import type { MockTransaction } from "../data/mockFinance";
import { formatVND } from "../utils/format";

export function TransactionItem({ transaction }: { transaction: MockTransaction }) {
  const Icon = transaction.icon;
  const positive = transaction.amount > 0;
  const neutral = transaction.kind === "transfer";
  return (
    <button className="flex w-full items-center gap-3 rounded-2xl p-2 text-left transition hover:bg-[#f5f3f3]">
      <div className="grid h-10 w-10 place-items-center rounded-full bg-[#d9e6da] text-[#006e1c]">
        <Icon size={18} />
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate font-semibold">{transaction.title}</p>
        <p className="truncate text-xs text-[#3f4a3c]">
          {transaction.category} · {transaction.wallet}
          {transaction.note ? ` · ${transaction.note}` : ""}
        </p>
      </div>
      <p className={`money text-sm font-bold ${positive ? "text-[#006e1c]" : neutral ? "text-[#3f4a3c]" : "text-[#bb1614]"}`}>
        {formatVND(transaction.amount)}
      </p>
    </button>
  );
}

