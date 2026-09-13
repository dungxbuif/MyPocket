import { Text } from "../atoms/Text";
import { BudgetGauge } from "../atoms/Progress";
import { MetricBox } from "../atoms/MetricBox";
import { SectionTitle } from "../atoms/SectionTitle";
import { BudgetProgressItem } from "../molecules/BudgetProgressItem";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { budgets } from "../data/mockFinance";
import { formatVND, ratioPercent } from "../utils/format";

export function BudgetsPanel({ masked }: { masked: boolean }) {
  const totalBudget = budgets.reduce((sum, budget) => sum + budget.limit, 0);
  const totalSpent = budgets.reduce((sum, budget) => sum + budget.spent, 0);
  const progress = ratioPercent(totalSpent, totalBudget);
  const safeToSpend = Math.max(0, Math.round((totalBudget - totalSpent) / 14));

  return (
    <>
      <SurfaceCard padding="lg" className="text-center">
        <Text size="sm" weight="semibold" tone="secondary" className="">Số tiền bạn có thể chi mỗi ngày</Text>
        <Text numeric size="3xl" weight="bold" tone="action" className="mt-1">{masked ? "••••••" : formatVND(safeToSpend)}</Text>
        <BudgetGauge value={progress} label="Đã dùng" />
        <div className="mt-5 grid grid-cols-3 gap-2">
          <MetricBox label="Ngân sách" value={masked ? "••••••" : formatVND(totalBudget)} />
          <MetricBox label="Đã chi" value={masked ? "••••••" : formatVND(totalSpent)} danger />
          <MetricBox label="Còn lại" value={masked ? "••••••" : formatVND(totalBudget - totalSpent)} />
        </div>
      </SurfaceCard>
      <SurfaceCard padding="md">
        <SectionTitle title="Theo danh mục" action="Thêm" />
        <div className="mt-3 space-y-3">{budgets.map((budget) => <BudgetProgressItem key={budget.id} budget={budget} />)}</div>
      </SurfaceCard>
    </>
  );
}
