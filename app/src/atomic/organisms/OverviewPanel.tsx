import { SectionTitle } from "../atoms/SectionTitle";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { TransactionItem } from "../molecules/TransactionItem";
import { WalletCard } from "../molecules/WalletCard";
import { transactions, wallets } from "../data/mockFinance";

type OverviewPanelProps = {
  masked: boolean;
};

export function OverviewPanel({ masked }: OverviewPanelProps) {
  return (
    <>
      <SurfaceCard className="p-5">
        <SectionTitle title="Ví của tôi" action="Quản lý" />
        <div className="mt-3 divide-y divide-slate-50">{wallets.slice(0, 3).map((wallet) => <WalletCard key={wallet.id} wallet={wallet} masked={masked} />)}</div>
      </SurfaceCard>

      {/* Insight, shortcuts and quick reports remain source-only until their APIs exist. */}

      <SurfaceCard className="p-5">
        <SectionTitle title="Giao dịch gần đây" action="Xem tất cả" />
        <div className="mt-3 space-y-2">{transactions.slice(0, 4).map((transaction) => <TransactionItem key={transaction.id} transaction={transaction} />)}</div>
      </SurfaceCard>

      {/* Featured budgets remain source-only until the budget API exists. */}
    </>
  );
}
