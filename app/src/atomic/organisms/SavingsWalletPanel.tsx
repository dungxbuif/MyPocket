import { useState } from "react";
import { ArrowDownLeft, ArrowUpRight } from "lucide-react";
import { BaseButton } from "../atoms/BaseButton";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { StatusMessage } from "../atoms/StatusMessage";
import { SavingsSummary } from "../molecules/SavingsSummary";
import { TransactionItem } from "../molecules/TransactionItem";
import { QuickAddSheet } from "./QuickAddSheet";
import type { Wallet } from "../../services/wallets";
import { signedTransactionAmount, type Transaction } from "../../services/transactions";
import { formatVND } from "../utils/format";
import type { Category } from "../../services/categories";
import { categoryPresentationFor } from "../atoms/categoryPresentation";

export function SavingsWalletPanel({wallet, transactions, categories, onBack, onChanged}: {wallet: Wallet; transactions: Transaction[]; categories: Category[]; onBack: () => void; onChanged: () => void}) {
  const [editor, setEditor] = useState<Transaction | "new" | null>(null);
  const groups = new Map<string, Transaction[]>();
  [...transactions].filter(row => row.wallet_id === wallet.id).sort((a,b) => Date.parse(b.occurred_at)-Date.parse(a.occurred_at)).forEach(row => {
    const date = new Date(row.occurred_at);
    const key = `Tháng ${date.getMonth()+1} ${date.getFullYear()}`;
    groups.set(key,[...(groups.get(key) ?? []), row]);
  });
  return <section className="space-y-4"><header className="flex items-center justify-between gap-3"><BaseButton variant="chip" onClick={onBack}>Chọn ví</BaseButton><Text as="h1" size="lg" weight="bold" className="truncate">{wallet.name}</Text></header><SavingsSummary wallet={wallet} /><Text weight="bold">TẤT CẢ CÁC GIAO DỊCH</Text>
    {groups.size === 0 ? <StatusMessage variant="plain">Chưa có giao dịch.</StatusMessage> : null}
    {[...groups].map(([label,rows]) => <SurfaceCard key={label} padding="md"><div className="mb-3 flex items-center justify-between gap-2"><Text weight="bold">{label}</Text><Text numeric>{formatVND(rows.reduce((sum,row) => sum+signedTransactionAmount(row),0))}</Text></div>{rows.map(row => {
      const category = categories.find(item => item.id === row.category_id);
      const presentation = category ? categoryPresentationFor(category.system_key ?? category.icon_key) : {icon:row.type === "income" ? ArrowDownLeft : ArrowUpRight, tone:row.type === "income" ? "success" as const : "danger" as const};
      return <TransactionItem key={row.id} item={{title: category?.name ?? (row.type === "income" ? "Khoản thu" : "Khoản chi"), metadata: [new Date(row.occurred_at).toLocaleDateString("vi-VN"),row.note].filter(Boolean).join(" · "), amount:signedTransactionAmount(row),kind:row.type,icon:presentation.icon,tone:presentation.tone}} onActivate={() => setEditor(row)} />;
    })}</SurfaceCard>)}
    <BaseButton className="w-full" onClick={() => setEditor("new")}>Thêm giao dịch</BaseButton>
    {editor ? <QuickAddSheet initialWalletID={wallet.id} transaction={editor === "new" ? undefined : editor} onClose={() => setEditor(null)} onSaved={onChanged} /> : null}
  </section>;
}
