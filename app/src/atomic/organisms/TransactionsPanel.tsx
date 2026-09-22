import { MoreHorizontal } from "lucide-react";
import { BaseModal } from "../atoms/BaseModal";
import { Divider } from "../atoms/Divider";
import { DateField } from "../molecules/DateField";
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
import { dateKeyAt, todayDateKey } from "../../services/accountTime";
import { useAccountTimezone } from "../../services/AccountTimezoneContext";

const COPY = { title: "Giao dịch", loading: "Đang tải giao dịch...", empty: "Chưa có giao dịch. Dùng nút + để ghi khoản thu hoặc chi đầu tiên.", error: "Không thể tải giao dịch.", retry: "Thử lại" } as const;

type TransactionGroup = { key: string; label: string; rows: Transaction[] };

function groupLabel(key: string, timezone:string): string {
  const today = todayDateKey(timezone);
  const yesterdayDate = new Date(`${today}T12:00:00Z`);
  yesterdayDate.setUTCDate(yesterdayDate.getUTCDate() - 1);
  const yesterday = yesterdayDate.toISOString().slice(0,10);
  if (key === today) return "Hôm nay";
  if (key === yesterday) return "Hôm qua";
  const [year, month, day] = key.split("-").map(Number);
  return new Intl.DateTimeFormat("vi-VN", { weekday: "long", day: "2-digit", month: "2-digit", year: "numeric", timeZone: "UTC" }).format(new Date(Date.UTC(year, month - 1, day, 12)));
}

export function TransactionsPanel({ refreshKey = 0, onChanged }: { refreshKey?: number; onChanged: () => void }) {
  const timezone=useAccountTimezone();
  const [menu, setMenu] = useState(false), [period, setPeriod] = useState(false), [byCategory, setByCategory] = useState(false);
  const [from, setFrom] = useState(""), [to, setTo] = useState("");
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
  const visible = useMemo(() => transactions.filter(item => { const day = dateKeyAt(item.occurred_at,timezone); return (!from || day >= from) && (!to || day <= to); }).sort((a, b) => Date.parse(b.occurred_at) - Date.parse(a.occurred_at)), [transactions, from, to,timezone]);
  const groups = useMemo<TransactionGroup[]>(() => {
    const byDate = new Map<string, Transaction[]>();
    visible.forEach((transaction) => {
      const key = byCategory ? transaction.category_id ?? transaction.type : dateKeyAt(transaction.occurred_at,timezone);
      byDate.set(key, [...(byDate.get(key) ?? []), transaction]);
    });
    return [...byDate.entries()].map(([key, rows]) => ({ key, label: byCategory ? categoryByID.get(key)?.name ?? (key === "income" ? "Khoản thu" : "Khoản chi") : groupLabel(key,timezone), rows }));
  }, [visible, byCategory, categoryByID,timezone]);

  return (
    <section className="space-y-3">
      <div className="flex items-center justify-between"><Text as="h1" size="xl" weight="bold">{COPY.title}</Text><BaseButton variant="chip" size="sm" aria-label="Tùy chọn giao dịch" onClick={() => setMenu(true)}><MoreHorizontal size={20} /></BaseButton></div>
      {from || to ? <BaseButton variant="ghost" onClick={() => setPeriod(true)}>{from || "Từ đầu"} — {to || "Hiện tại"}</BaseButton> : null}
      {!loading && !error && transactions.length > 0 ? <SurfaceCard tone="form" padding="md"><div className="flex justify-between"><Text>Tiền vào</Text><Text tone="action" numeric>{formatVND(visible.filter(item => item.type === "income").reduce((sum, item) => sum + item.amount, 0))}</Text></div><div className="mt-3 flex justify-between"><Text>Tiền ra</Text><Text tone="danger" numeric>{formatVND(visible.filter(item => item.type === "expense").reduce((sum, item) => sum + item.amount, 0))}</Text></div></SurfaceCard> : null}
      {loading ? <StatusMessage>{COPY.loading}</StatusMessage> : null}
      {error ? <StatusMessage tone="danger">{COPY.error}</StatusMessage> : null}
      {!loading && !error && visible.length === 0 ? <StatusMessage variant="plain">{COPY.empty}</StatusMessage> : null}
      {groups.map((group) => {
        const total = group.rows.reduce((sum, transaction) => sum + signedTransactionAmount(transaction), 0);
        return <SurfaceCard key={group.key} tone="form" padding="md"><div className="mb-3 flex items-center justify-between gap-2"><div className="flex items-center gap-3">{!byCategory ? <Text size="4xl" weight="bold">{Number(group.key.slice(8))}</Text> : null}<div><Text size="sm" tone="secondary">{byCategory ? group.label : new Intl.DateTimeFormat("vi-VN", { weekday: "long", timeZone:"UTC" }).format(new Date(`${group.key}T12:00:00Z`))}</Text><Text size="xs" tone="secondary">{byCategory ? `${group.rows.length} giao dịch` : new Intl.DateTimeFormat("vi-VN", { month: "long", year: "numeric", timeZone:"UTC" }).format(new Date(`${group.key}T12:00:00Z`))}</Text></div></div><Text numeric weight="bold">{formatVND(total)}</Text></div><Divider /><div className="mt-3 space-y-2">{group.rows.map((transaction) => {
          const category = transaction.category_id ? categoryByID.get(transaction.category_id) : undefined;
          const presentation = category ? categoryPresentationFor(category.system_key ?? category.icon_key) : { icon: Tags, tone: "categorySlate" as const };
          return <TransactionItem key={transaction.id} item={{ title: category?.name ?? (transaction.type === "income" ? "Khoản thu" : "Khoản chi"), metadata: [walletNames.get(transaction.wallet_id) ?? "Ví đã xóa", transaction.note].filter(Boolean).join(" · "), amount: signedTransactionAmount(transaction), kind: transaction.type, icon: presentation.icon, tone: presentation.tone }} onActivate={() => setEditing(transaction)} />;
        })}</div></SurfaceCard>;
      })}
      {menu ? <BaseModal label="Tùy chọn giao dịch" onClose={() => setMenu(false)}><div className="space-y-2"><BaseButton className="w-full" variant="chip" onClick={() => { setMenu(false); setPeriod(true); }}>Khoảng thời gian</BaseButton><BaseButton className="w-full" variant="chip" onClick={() => { setByCategory(!byCategory); setMenu(false); }}>{byCategory ? "Xem theo ngày" : "Xem theo nhóm"}</BaseButton><BaseButton className="w-full" variant="chip" onClick={() => { setLocalRefresh(v => v + 1); setMenu(false); }}>Tải lại giao dịch</BaseButton><BaseButton className="w-full" variant="chip" disabled>Xóa nhiều giao dịch</BaseButton><BaseButton className="w-full" variant="chip" disabled>Chuyển tiền đến ví khác</BaseButton><BaseButton className="w-full" variant="chip" disabled>Điều chỉnh số dư</BaseButton><BaseButton className="w-full" variant="chip" disabled>Đồng bộ ví</BaseButton><Text size="xs" tone="secondary">Các thao tác bị vô hiệu hóa chưa được hỗ trợ.</Text><BaseButton variant="ghost" onClick={() => setMenu(false)}>Đóng</BaseButton></div></BaseModal> : null}
      {period ? <BaseModal label="Khoảng thời gian" onClose={() => setPeriod(false)}><DateField label="Từ ngày" value={from} stepper={false} onChange={setFrom} /><DateField label="Đến ngày" value={to} stepper={false} onChange={setTo} />{from && to && to < from ? <StatusMessage tone="danger">Ngày kết thúc phải sau ngày bắt đầu.</StatusMessage> : null}<div className="flex gap-2"><BaseButton variant="ghost" onClick={() => { setFrom(""); setTo(""); setPeriod(false); }}>Tất cả thời gian</BaseButton><BaseButton disabled={!!(from && to && to < from)} onClick={() => setPeriod(false)}>Xong</BaseButton></div></BaseModal> : null}
      {editing ? <QuickAddSheet transaction={editing} onClose={() => setEditing(null)} onSaved={() => { setLocalRefresh((value) => value + 1); onChanged(); }} /> : null}
    </section>
  );
}
