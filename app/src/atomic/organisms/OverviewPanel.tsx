import { Lightbulb } from "lucide-react";
import { MetricBox } from "../atoms/MetricBox";
import { SectionTitle } from "../atoms/SectionTitle";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { IconBadge } from "../atoms/IconBadge";
import { BudgetProgressItem } from "../molecules/BudgetProgressItem";
import { TransactionItem } from "../molecules/TransactionItem";
import { WalletCard } from "../molecules/WalletCard";
import { budgets, insights, quickActions, reportBars, transactions, wallets } from "../data/mockFinance";
import { formatVND } from "../utils/format";

type OverviewPanelProps = {
  masked: boolean;
};

export function OverviewPanel({ masked }: OverviewPanelProps) {
  const totalBudget = budgets.reduce((sum, budget) => sum + budget.limit, 0);
  const totalSpent = budgets.reduce((sum, budget) => sum + budget.spent, 0);

  return (
    <>
      <SurfaceCard className="p-5">
        <SectionTitle title="Ví của tôi" action="Quản lý" />
        <div className="mt-3 divide-y divide-slate-50">{wallets.slice(0, 3).map((wallet) => <WalletCard key={wallet.id} wallet={wallet} masked={masked} />)}</div>
      </SurfaceCard>

      <SurfaceCard className="p-5">
        <div className="flex gap-3">
          <IconBadge icon={Lightbulb} shape="circle" className="bg-[#d9e6da] text-[#006e1c]" />
          <div>
            <p className="font-semibold">Money Insight</p>
            <p className="mt-1 text-sm leading-5 text-[#3f4a3c]">{insights[0]}</p>
          </div>
        </div>
      </SurfaceCard>

      <section className="grid grid-cols-4 gap-2">
        {quickActions.map((action) => {
          const Icon = action.icon;
          return (
            <button key={action.label} className="cursor-pointer rounded-3xl border border-slate-100 bg-white p-3 text-center shadow-[0_4px_20px_rgb(0_0_0/0.03)] transition hover:bg-[#f5f3f3]">
              <IconBadge icon={Icon} shape="circle" size="md" className="mx-auto bg-[#d9e6da] text-[#006e1c]" />
              <span className="mt-2 block text-xs font-bold leading-4 text-[#3f4a3c]">{action.label}</span>
            </button>
          );
        })}
      </section>

      <SurfaceCard className="p-5">
        <SectionTitle title="Báo cáo nhanh" action="Chi tiết" />
        <div className="mt-4 flex h-24 items-end gap-2">
          {reportBars.map((bar, index) => (
            <div key={index} className="flex flex-1 items-end">
              <div className="w-full rounded-t-full bg-[#d9e6da]" style={{ height: `${bar}%` }}>
                <div className="h-1/2 rounded-t-full bg-[#006e1c]" />
              </div>
            </div>
          ))}
        </div>
        <div className="mt-4 grid grid-cols-2 gap-3">
          <MetricBox label="Chi tháng này" value={masked ? "••••••" : formatVND(-7830000)} danger />
          <MetricBox label="So với tháng trước" value="↓ 8.4%" />
        </div>
      </SurfaceCard>

      <SurfaceCard className="p-5">
        <SectionTitle title="Giao dịch gần đây" action="Xem tất cả" />
        <div className="mt-3 space-y-2">{transactions.slice(0, 4).map((transaction) => <TransactionItem key={transaction.id} transaction={transaction} />)}</div>
      </SurfaceCard>

      <SurfaceCard className="p-5">
        <SectionTitle title="Ngân sách nổi bật" />
        <div className="mt-3 space-y-3">{budgets.slice(0, 2).map((budget) => <BudgetProgressItem key={budget.id} budget={budget} />)}</div>
      </SurfaceCard>
    </>
  );
}
