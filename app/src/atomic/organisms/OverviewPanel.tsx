import { Lightbulb } from "lucide-react";
import { MetricBox } from "../atoms/MetricBox";
import { SectionTitle } from "../atoms/SectionTitle";
import { BudgetProgressItem } from "../molecules/BudgetProgressItem";
import { TransactionItem } from "../molecules/TransactionItem";
import { WalletCard } from "../molecules/WalletCard";
import { budgets, insights, quickActions, reportBars, transactions, wallets } from "../data/mockFinance";
import { formatVND, ratioPercent } from "../utils/format";

export function OverviewPanel({ masked }: { masked: boolean }) {
  const totalBalance = wallets.reduce((sum, wallet) => sum + wallet.balance, 0);
  const totalBudget = budgets.reduce((sum, budget) => sum + budget.limit, 0);
  const totalSpent = budgets.reduce((sum, budget) => sum + budget.spent, 0);
  const safeToSpend = Math.max(0, Math.round((totalBudget - totalSpent) / 14));

  return (
    <>
      <section className="rounded-3xl bg-[#006e1c] p-5 text-white shadow-[0_16px_30px_rgb(0_110_28/0.20)]">
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-sm text-white/75">Tổng số dư</p>
            <p className="money mt-1 text-3xl font-bold">{masked ? "••••••••" : formatVND(totalBalance)}</p>
          </div>
          <span className="rounded-full bg-white/15 px-3 py-1 text-xs font-semibold">Tháng 09</span>
        </div>
        <div className="mt-5 grid grid-cols-2 gap-3">
          <div className="rounded-2xl bg-white/15 p-3">
            <p className="text-xs text-white/70">Có thể chi/ngày</p>
            <p className="money mt-1 text-sm font-bold">{masked ? "••••••" : formatVND(safeToSpend)}</p>
          </div>
          <div className="rounded-2xl bg-white/15 p-3">
            <p className="text-xs text-white/70">Đã dùng ngân sách</p>
            <p className="mt-1 text-sm font-bold">{ratioPercent(totalSpent, totalBudget)}%</p>
          </div>
        </div>
      </section>

      <section className="rounded-3xl border border-[#d9e6da] bg-white p-4">
        <div className="flex gap-3">
          <div className="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-[#d9e6da] text-[#006e1c]">
            <Lightbulb size={20} />
          </div>
          <div>
            <p className="font-semibold">Money Insight</p>
            <p className="mt-1 text-sm leading-5 text-[#3f4a3c]">{insights[0]}</p>
          </div>
        </div>
      </section>

      <section className="grid grid-cols-4 gap-2">
        {quickActions.map((action) => {
          const Icon = action.icon;
          return (
            <button key={action.label} className="rounded-3xl bg-white p-3 text-center shadow-sm">
              <span className="mx-auto grid h-10 w-10 place-items-center rounded-full bg-[#d9e6da] text-[#006e1c]">
                <Icon size={18} />
              </span>
              <span className="mt-2 block text-xs font-bold leading-4 text-[#3f4a3c]">{action.label}</span>
            </button>
          );
        })}
      </section>

      <section className="rounded-3xl bg-white p-4">
        <SectionTitle title="Ví của bạn" action="Quản lý" />
        <div className="mt-3 space-y-2">{wallets.map((wallet) => <WalletCard key={wallet.id} wallet={wallet} masked={masked} />)}</div>
      </section>

      <section className="rounded-3xl bg-white p-4">
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
      </section>

      <section className="rounded-3xl bg-white p-4">
        <SectionTitle title="Giao dịch gần đây" action="Xem tất cả" />
        <div className="mt-3 space-y-2">{transactions.slice(0, 4).map((transaction) => <TransactionItem key={transaction.id} transaction={transaction} />)}</div>
      </section>

      <section className="rounded-3xl bg-white p-4">
        <SectionTitle title="Ngân sách nổi bật" />
        <div className="mt-3 space-y-3">{budgets.slice(0, 2).map((budget) => <BudgetProgressItem key={budget.id} budget={budget} />)}</div>
      </section>
    </>
  );
}
