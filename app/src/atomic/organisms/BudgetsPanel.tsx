import { MetricBox } from "../atoms/MetricBox";
import { SectionTitle } from "../atoms/SectionTitle";
import { BudgetProgressItem } from "../molecules/BudgetProgressItem";
import { budgets } from "../data/mockFinance";
import { formatVND, ratioPercent } from "../utils/format";

export function BudgetsPanel({ masked }: { masked: boolean }) {
  const totalBudget = budgets.reduce((sum, budget) => sum + budget.limit, 0);
  const totalSpent = budgets.reduce((sum, budget) => sum + budget.spent, 0);
  const progress = ratioPercent(totalSpent, totalBudget);
  const safeToSpend = Math.max(0, Math.round((totalBudget - totalSpent) / 14));

  return (
    <>
      <section className="rounded-3xl bg-white p-5 text-center">
        <p className="text-sm font-semibold text-[#3f4a3c]">Số tiền bạn có thể chi mỗi ngày</p>
        <p className="money mt-1 text-3xl font-bold text-[#006e1c]">{masked ? "••••••" : formatVND(safeToSpend)}</p>
        <div className="relative mx-auto mt-5 h-28 w-56 overflow-hidden">
          <div className="absolute inset-x-0 top-0 h-56 rounded-full border-[18px] border-[#e3e2e2]" />
          <div
            className="absolute inset-x-0 top-0 h-56 rounded-full border-[18px] border-[#006e1c]"
            style={{ clipPath: `inset(0 ${100 - progress}% 50% 0)` }}
          />
          <div className="absolute inset-x-0 bottom-0 text-center">
            <p className="text-xs text-[#3f4a3c]">Đã dùng</p>
            <p className="text-xl font-bold">{progress}%</p>
          </div>
        </div>
        <div className="mt-5 grid grid-cols-3 gap-2">
          <MetricBox label="Ngân sách" value={masked ? "••••••" : formatVND(totalBudget)} />
          <MetricBox label="Đã chi" value={masked ? "••••••" : formatVND(totalSpent)} danger />
          <MetricBox label="Còn lại" value={masked ? "••••••" : formatVND(totalBudget - totalSpent)} />
        </div>
      </section>
      <section className="rounded-3xl bg-white p-4">
        <SectionTitle title="Theo danh mục" action="Thêm" />
        <div className="mt-3 space-y-3">{budgets.map((budget) => <BudgetProgressItem key={budget.id} budget={budget} />)}</div>
      </section>
    </>
  );
}

