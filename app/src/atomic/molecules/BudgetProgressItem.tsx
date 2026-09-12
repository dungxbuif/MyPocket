import type { MockBudget } from "../data/mockFinance";
import { formatVND, ratioPercent } from "../utils/format";

export function BudgetProgressItem({ budget }: { budget: MockBudget }) {
  const Icon = budget.icon;
  const progress = ratioPercent(budget.spent, budget.limit);
  const over = budget.spent > budget.limit;
  return (
    <div className="rounded-2xl bg-[#f5f3f3] p-3">
      <div className="flex items-center gap-3">
        <div className={`grid h-10 w-10 place-items-center rounded-full ${over ? "bg-[#ffdad6] text-[#93000a]" : "bg-[#d9e6da] text-[#006e1c]"}`}>
          <Icon size={18} />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-2">
            <p className="font-semibold">{budget.name}</p>
            <p className={`text-sm font-bold ${over ? "text-[#bb1614]" : "text-[#1b1c1c]"}`}>{progress}%</p>
          </div>
          <p className="money text-xs text-[#3f4a3c]">
            {formatVND(budget.spent)} / {formatVND(budget.limit)}
          </p>
        </div>
      </div>
      <div className="mt-3 h-2 overflow-hidden rounded-full bg-[#e3e2e2]">
        <div className={`h-full rounded-full ${over ? "bg-[#bb1614]" : "bg-[#006e1c]"}`} style={{ width: `${progress}%` }} />
      </div>
    </div>
  );
}

