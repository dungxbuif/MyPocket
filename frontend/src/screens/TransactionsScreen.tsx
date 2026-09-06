import * as React from "react";
import { Search } from "lucide-react";
import type { Transaction } from "../app/finance";
import { Card } from "../components/ui/card";
import { WalletFilterChip } from "../components/navigation/WalletFilterChip";

export interface TransactionsScreenProps {
  transactions: Transaction[];
  onEdit: (transaction: Transaction) => void;
  signedAmount: (transaction: Transaction) => string;
}

export function TransactionsScreen({
  transactions,
  onEdit,
  signedAmount,
}: TransactionsScreenProps) {
  const [query, setQuery] = React.useState("");

  const visible = transactions.filter((transaction) =>
    `${transaction.note} ${transaction.type}`.toLowerCase().includes(query.toLowerCase())
  );

  return (
    <section className="content-stack">
      <div className="sub-header flex items-center justify-between">
        <h1>Sổ giao dịch</h1>
        <div className="flex items-center gap-2">
          <WalletFilterChip walletName="Tổng cộng" />
          <button className="pill-button" type="button">Tháng 08/2026</button>
        </div>
      </div>

      <Card className="list-card">
        <label className="transaction-search">
          <Search size={18} />
          <input
            aria-label="Tìm giao dịch"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Tìm giao dịch"
          />
        </label>

        {visible.length === 0 ? (
          <p className="empty-state">Chưa có giao dịch</p>
        ) : (
          visible.map((transaction) => (
            <div
              key={transaction.id}
              role="button"
              tabIndex={0}
              onClick={() => onEdit(transaction)}
              className="transaction-row"
            >
              <div className="transaction-main">
                <span className="transaction-icon">
                  {transaction.type === "income" ? "💵" : "🛍️"}
                </span>
                <div>
                  <strong>{transaction.note || transaction.type}</strong>
                  <p>{new Date(transaction.occurred_at).toLocaleDateString("vi-VN")}</p>
                </div>
              </div>
              <span className={`transaction-amount ${transaction.type === "income" ? "income" : "expense"}`}>
                {signedAmount(transaction)}
              </span>
            </div>
          ))
        )}
      </Card>
    </section>
  );
}
