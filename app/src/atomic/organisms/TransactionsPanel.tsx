import { ListFilter, Search } from "lucide-react";
import { IconButton } from "../atoms/IconButton";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { TransactionItem } from "../molecules/TransactionItem";
import { transactions } from "../data/mockFinance";
import { formatVND } from "../utils/format";

export function TransactionsPanel() {
  const groups = ["Hôm nay", "Hôm qua", "Thứ 5"];
  return (
    <>
      <div className="flex gap-2">
        <div className="flex min-w-0 flex-1 items-center gap-2 rounded-full bg-white px-4 py-3">
          <Search size={17} className="text-[#3f4a3c]" />
          <span className="truncate text-sm text-[#3f4a3c]">Tìm giao dịch, ví, danh mục</span>
        </div>
        <IconButton label="Bộ lọc">
          <ListFilter size={18} />
        </IconButton>
      </div>
      <div className="flex gap-2 overflow-x-auto pb-1">
        {["Tổng cộng", "Tháng 09", "Theo danh mục", "Không tính báo cáo"].map((filter) => (
          <span key={filter} className="shrink-0 rounded-full bg-white px-4 py-2 text-sm font-semibold text-[#3f4a3c]">
            {filter}
          </span>
        ))}
      </div>
      {groups.map((group) => {
        const rows = transactions.filter((transaction) => transaction.occurredLabel === group);
        if (!rows.length) return null;
        const sum = rows.reduce((value, row) => value + row.amount, 0);
        return (
          <SurfaceCard key={group} padding="md">
            <div className="mb-3 flex items-center justify-between">
              <div>
                <p className="text-lg font-bold">{group}</p>
                <p className="text-xs text-[#3f4a3c]">{rows.length} giao dịch</p>
              </div>
              <p className={`money text-sm font-bold ${sum < 0 ? "text-[#bb1614]" : "text-[#006e1c]"}`}>{formatVND(sum)}</p>
            </div>
            <div className="space-y-2">{rows.map((transaction) => <TransactionItem key={transaction.id} transaction={transaction} />)}</div>
          </SurfaceCard>
        );
      })}
    </>
  );
}
