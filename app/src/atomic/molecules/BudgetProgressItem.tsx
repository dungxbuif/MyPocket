import type { MockBudget } from "../data/mockFinance";
import { formatVND, ratioPercent } from "../utils/format";
import { BUDGET_PROGRESS_CLASSES } from "../atoms/tokens";

export function BudgetProgressItem({ budget }: { budget: MockBudget }) {
  const Icon = budget.icon;
  const progress = ratioPercent(budget.spent, budget.limit);
  const over = budget.spent > budget.limit;
  const colors = over ? BUDGET_PROGRESS_CLASSES.over : BUDGET_PROGRESS_CLASSES.normal;
  return (
    <div className={`${BUDGET_PROGRESS_CLASSES.container} p-3`}>
      <div className="flex items-center gap-3">
        <div className={`grid h-10 w-10 place-items-center rounded-full ${colors.badge}`}>
          <Icon size={18} />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-2">
            <p className="font-semibold">{budget.name}</p>
            <p className={`text-sm font-bold ${colors.text}`}>{progress}%</p>
          </div>
          <p className={`money text-xs ${BUDGET_PROGRESS_CLASSES.detail}`}>
            {formatVND(budget.spent)} / {formatVND(budget.limit)}
          </p>
        </div>
      </div>
      <div className={`mt-3 h-2 overflow-hidden rounded-full ${BUDGET_PROGRESS_CLASSES.track}`}>
        <div className={`h-full rounded-full ${colors.fill}`} style={{ width: `${progress}%` }} />
      </div>
    </div>
  );
}
