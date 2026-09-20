import { useEffect, useMemo, useState } from "react";
import { RefreshCw } from "lucide-react";

import { BaseButton } from "../atoms/BaseButton";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { TransactionItem } from "../molecules/TransactionItem";
import { formatVND } from "../utils/format";
import { QuickAddSheet } from "./QuickAddSheet";
import { categoryPresentationFor } from "../atoms/categoryPresentation";
import { Tags } from "lucide-react";
import { fetchCategories, type Category } from "../../services/categories";
import { fetchTransactions, signedTransactionAmount, type Transaction } from "../../services/transactions";
import { fetchWallets, type Wallet } from "../../services/wallets";

const COPY = { title: "Giao dịch", loading: "Đang tải giao dịch...", empty: "Chưa có giao dịch. Dùng nút + để ghi khoản thu hoặc chi đầu tiên.", error: "Không thể tải giao dịch.", retry: "Thử lại" } as const;

type TransactionGroup = { key: string; label: string; rows: Transaction[] };

function localDateKey(value: string): string {
  const date = new Date(value);
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

function groupLabel(key: string): string {
  const today = localDateKey(new Date().toISOString());
  const yesterdayDate = new Date();
  yesterdayDate.setDate(yesterdayDate.getDate() - 1);
  const yesterday = localDateKey(yesterdayDate.toISOString());
  if (key === today) return "Hôm nay";
  if (key === yesterday) return "Hôm qua";
  const [year, month, day] = key.split("-").map(Number);
  return new Intl.DateTimeFormat("vi-VN", { weekday: "long", day: "2-digit", month: "2-digit", year: "numeric" }).format(new Date(year, month - 1, day));
}

export function TransactionsPanel({ refreshKey = 0, onChanged }: { refreshKey?: number; onChanged: () => void }) {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [editing, setEditing] = useState<Transaction | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [localRefresh, setLocalRefresh] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError("");
    Promise.all([fetchTransactions(), fetchWallets(), fetchCategories()])
      .then(([nextTransactions, nextWallets, nextCategories]) => {
        if (cancelled) return;
        setTransactions(nextTransactions);
        setWallets(nextWallets);
        setCategories(nextCategories);
      })
      .catch(() => { if (!cancelled) setError(COPY.error); })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [refreshKey, localRefresh]);

  const walletNames = useMemo(() => new Map(wallets.map((wallet) => [wallet.id, wallet.name])), [wallets]);
  const categoryByID = useMemo(() => new Map(categories.map((category) => [category.id, category])), [categories]);
  const groups = useMemo<TransactionGroup[]>(() => {
    const byDate = new Map<string, Transaction[]>();
    transactions.forEach((transaction) => {
      const key = localDateKey(transaction.occurred_at);
      byDate.set(key, [...(byDate.get(key) ?? []), transaction]);
    });
    return [...byDate.entries()].map(([key, rows]) => ({ key, label: groupLabel(key), rows }));
  }, [transactions]);

  return (
    <section className="space-y-3">
      <div className="flex items-center justify-between"><Text as="h1" size="xl" weight="bold">{COPY.title}</Text><BaseButton variant="ghost" size="sm" onClick={() => setLocalRefresh((value) => value + 1)}><RefreshCw size={16} />{COPY.retry}</BaseButton></div>
      {loading ? <StatusMessage>{COPY.loading}</StatusMessage> : null}
      {error ? <StatusMessage tone="danger">{COPY.error}</StatusMessage> : null}
      {!loading && !error && transactions.length === 0 ? <StatusMessage variant="plain">{COPY.empty}</StatusMessage> : null}
      {groups.map((group) => {
        const total = group.rows.reduce((sum, transaction) => sum + signedTransactionAmount(transaction), 0);
        return <SurfaceCard key={group.key} padding="md"><div className="mb-3 flex items-center justify-between"><div><Text size="lg" weight="bold">{group.label}</Text><Text size="xs" tone="secondary">{group.rows.length} giao dịch</Text></div><Text numeric weight="bold" tone={total < 0 ? "danger" : "action"}>{formatVND(total)}</Text></div><div className="space-y-2">{group.rows.map((transaction) => {
          const category = transaction.category_id ? categoryByID.get(transaction.category_id) : undefined;
          const presentation = category ? categoryPresentationFor(category.system_key ?? category.icon_key) : { icon: Tags, tone: "categorySlate" as const };
          return <TransactionItem key={transaction.id} item={{ title: category?.name ?? (transaction.type === "income" ? "Khoản thu" : "Khoản chi"), metadata: [walletNames.get(transaction.wallet_id) ?? "Ví đã xóa", transaction.note].filter(Boolean).join(" · "), amount: signedTransactionAmount(transaction), kind: transaction.type, icon: presentation.icon, tone: presentation.tone }} onActivate={() => setEditing(transaction)} />;
        })}</div></SurfaceCard>;
      })}
      {editing ? <QuickAddSheet transaction={editing} onClose={() => setEditing(null)} onSaved={() => { setLocalRefresh((value) => value + 1); onChanged(); }} /> : null}
    </section>
  );
}
