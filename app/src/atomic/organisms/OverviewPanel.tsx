import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { Tags } from "lucide-react";

import { SectionTitle } from "../atoms/SectionTitle";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { TransactionItem } from "../molecules/TransactionItem";
import { WalletCard } from "../molecules/WalletCard";
import { categoryPresentationFor } from "../atoms/categoryPresentation";
import { fetchCategories, type Category } from "../../services/categories";
import { fetchTransactions, signedTransactionAmount, type Transaction } from "../../services/transactions";
import { fetchWallets, type Wallet } from "../../services/wallets";

export function OverviewPanel({ masked, refreshKey = 0 }: { masked: boolean; refreshKey?: number }) {
  const navigate = useNavigate();
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    Promise.all([fetchWallets(), fetchTransactions(), fetchCategories()])
      .then(([nextWallets, nextTransactions, nextCategories]) => {
        if (cancelled) return;
        setWallets(nextWallets);
        setTransactions(nextTransactions);
        setCategories(nextCategories);
        setError("");
      })
      .catch(() => { if (!cancelled) setError("Không thể tải dữ liệu tổng quan."); })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [refreshKey]);

  const walletNames = useMemo(() => new Map(wallets.map((wallet) => [wallet.id, wallet.name])), [wallets]);
  const categoryByID = useMemo(() => new Map(categories.map((category) => [category.id, category])), [categories]);

  if (loading) return <StatusMessage>Đang tải tổng quan...</StatusMessage>;
  if (error) return <StatusMessage tone="danger">{error}</StatusMessage>;

  return (
    <>
      <SurfaceCard padding="md">
        <SectionTitle title="Ví của tôi" action="Quản lý" onAction={() => void navigate({ to: "/account/wallets" })} />
        {wallets.length === 0 ? <StatusMessage variant="plain">Chưa có ví.</StatusMessage> : <div className="mt-3 divide-y divide-row">{wallets.slice(0, 3).map((wallet) => <WalletCard key={wallet.id} wallet={wallet} masked={masked} onActivate={() => void navigate({ to: "/account/wallets" })} />)}</div>}
      </SurfaceCard>

      <SurfaceCard padding="md">
        <SectionTitle title="Giao dịch gần đây" action="Xem tất cả" onAction={() => void navigate({ to: "/transactions" })} />
        {transactions.length === 0 ? <StatusMessage variant="plain">Chưa có giao dịch.</StatusMessage> : <div className="mt-3 space-y-2">{transactions.slice(0, 4).map((transaction) => {
          const category = transaction.category_id ? categoryByID.get(transaction.category_id) : undefined;
          const presentation = category ? categoryPresentationFor(category.system_key ?? category.icon_key) : { icon: Tags, tone: "categorySlate" as const };
          return <TransactionItem key={transaction.id} item={{ title: category?.name ?? (transaction.type === "income" ? "Khoản thu" : "Khoản chi"), metadata: [walletNames.get(transaction.wallet_id) ?? "Ví đã xóa", transaction.note].filter(Boolean).join(" · "), amount: signedTransactionAmount(transaction), kind: transaction.type, icon: presentation.icon, tone: presentation.tone }} onActivate={() => void navigate({ to: "/transactions" })} />;
        })}</div>}
      </SurfaceCard>
    </>
  );
}
