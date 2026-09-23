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
import { accountMonthKey, dateKeyAt } from "../../services/accountTime";
import { useAccountTimezone } from "../../services/AccountTimezoneContext";

export function WalletDetailPanel({wallet, transactions, categories, onBack, onChanged, embedded = false}: {wallet: Wallet; transactions: Transaction[]; categories: Category[]; onBack: () => void; onChanged: () => void; embedded?: boolean}) {
  const timezone=useAccountTimezone();
  const [editor, setEditor] = useState<Transaction | "new" | null>(null);
  const groups = new Map<string, Transaction[]>();
  [...transactions].filter(row => row.wallet_id === wallet.id).sort((a,b) => Date.parse(b.occurred_at)-Date.parse(a.occurred_at)).forEach(row => {
    const key = accountMonthKey(row.occurred_at,timezone);
    groups.set(key,[...(groups.get(key) ?? []), row]);
  });
  const income = walletTransactions(transactions, wallet.id).filter(row => row.type === "income").reduce((sum, row) => sum + row.amount, 0);
  const expense = walletTransactions(transactions, wallet.id).filter(row => row.type === "expense").reduce((sum, row) => sum + row.amount, 0);
  return <section className="space-y-4">{!embedded ? <header className="flex items-center justify-between gap-3"><BaseButton variant="chip" onClick={onBack}>Chọn ví</BaseButton><Text as="h1" size="lg" weight="bold" className="truncate">{wallet.name}</Text></header> : null}{wallet.type === "goal" ? <SavingsSummary wallet={wallet} /> : <SurfaceCard tone="form" padding="md"><div className="flex items-center justify-between"><Text tone="secondary">Số dư hiện tại</Text><Text numeric weight="bold" tone={wallet.current_balance < 0 ? "danger" : "ink"}>{formatVND(wallet.current_balance)}</Text></div><div className="mt-3 grid grid-cols-2 gap-3"><div><Text size="xs" tone="secondary">Tiền vào</Text><Text numeric tone="action">{formatVND(income)}</Text></div><div><Text size="xs" tone="secondary">Tiền ra</Text><Text numeric tone="danger">{formatVND(expense)}</Text></div></div></SurfaceCard>}<Text weight="bold">TẤT CẢ CÁC GIAO DỊCH</Text>
    {groups.size === 0 ? <StatusMessage variant="plain">Chưa có giao dịch.</StatusMessage> : null}
    {[...groups].map(([label,rows]) => <SurfaceCard key={label} padding="md"><div className="mb-3 flex items-center justify-between gap-2"><Text weight="bold">Tháng {label.slice(5,7)} {label.slice(0,4)}</Text><Text numeric>{formatVND(rows.reduce((sum,row) => sum+signedTransactionAmount(row),0))}</Text></div>{rows.map(row => {
      const category = categories.find(item => item.id === row.category_id);
      const presentation = category ? categoryPresentationFor(category.system_key ?? category.icon_key) : {icon:row.type === "income" ? ArrowDownLeft : ArrowUpRight, tone:row.type === "income" ? "success" as const : "danger" as const};
      const dateKey=dateKeyAt(row.occurred_at,timezone);
      const [year,month,day]=dateKey.split("-").map(Number);
      const dateLabel=new Intl.DateTimeFormat("vi-VN",{dateStyle:"medium",timeZone:"UTC"}).format(new Date(Date.UTC(year,month-1,day,12)));
      return <TransactionItem key={row.id} item={{title: category?.name ?? (row.type === "income" ? "Khoản thu" : "Khoản chi"), metadata: [dateLabel,row.note].filter(Boolean).join(" · "), amount:signedTransactionAmount(row),kind:row.type,icon:presentation.icon,tone:presentation.tone}} onActivate={wallet.type === "credit" ? undefined : () => setEditor(row)} />;
    })}</SurfaceCard>)}
    {wallet.type !== "credit" ? <BaseButton className="w-full" onClick={() => setEditor("new")}>Thêm giao dịch</BaseButton> : <Text size="xs" tone="secondary">Ví tín dụng hiện chỉ hiển thị giao dịch.</Text>}
    {editor ? <QuickAddSheet initialWalletID={wallet.id} transaction={editor === "new" ? undefined : editor} onClose={() => setEditor(null)} onSaved={onChanged} /> : null}
  </section>;
}

function walletTransactions(transactions: Transaction[], walletID: string) {
  return transactions.filter(row => row.wallet_id === walletID);
}
