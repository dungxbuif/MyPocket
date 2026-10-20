import * as React from "react";
import type { Transaction } from "../app/finance";
import { Card, CardHeader, CardTitle } from "../components/ui/card";
import { Button } from "../components/ui/button";
import { SearchInputField } from "../components/inputs/SearchInputField";
import { TransactionRow } from "../components/finance/TransactionRow";
import { EmptyStateView } from "../components/feedback/EmptyStateView";

export interface TransactionsScreenProps {
  transactions: Transaction[];
  onEdit: (transaction: Transaction) => void;
}

export function TransactionsScreen({ transactions, onEdit }: TransactionsScreenProps) {
  const [query, setQuery] = React.useState("");
  const normalizedQuery = query.trim().toLowerCase();
  const visibleTransactions = transactions.filter((transaction) =>
    `${transaction.note} ${transaction.type}`.toLowerCase().includes(normalizedQuery)
  );

  return (
    <section className="content-stack">
      <div className="sub-header">
        <h1>Sổ giao dịch</h1>
        <Button variant="outline" size="sm">Tháng hiện tại</Button>
      </div>
      <Card className="list-card">
        <CardHeader>
          <CardTitle>Tìm giao dịch</CardTitle>
        </CardHeader>
        <SearchInputField ariaLabel="Tìm giao dịch" value={query} onChange={setQuery} placeholder="Tìm giao dịch" />
        {visibleTransactions.length === 0 ? (
          <EmptyStateView
            title="Chưa có giao dịch"
            description={normalizedQuery ? "Không tìm thấy giao dịch phù hợp." : undefined}
          />
        ) : visibleTransactions.map((transaction) => (
          <TransactionRow
            key={transaction.id}
            id={transaction.id}
            categoryName={transaction.note || transaction.type}
            amount={transaction.type === "expense" ? -transaction.amount_vnd : transaction.amount_vnd}
            amountTone={transaction.type === "transfer" || transaction.type === "adjustment" ? "neutral" : "signed"}
            note={transaction.type === "transfer" ? "Chuyển khoản" : transaction.type === "adjustment" ? "Số dư sau điều chỉnh" : undefined}
            dateFormatted={new Date(transaction.occurred_at).toLocaleDateString("vi-VN")}
            onClick={() => onEdit(transaction)}
          />
        ))}
      </Card>
    </section>
  );
}
