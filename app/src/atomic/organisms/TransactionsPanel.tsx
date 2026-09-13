import { Text } from "../atoms/Text";
import { Chip } from "../atoms/Chip";
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
        <Chip className="min-w-0 flex-1">
          <Search size={17} className="text-secondary" />
          <span className="truncate text-sm text-secondary">Tìm giao dịch, ví, danh mục</span>
        </Chip>
        <IconButton label="Bộ lọc">
          <ListFilter size={18} />
        </IconButton>
      </div>
      <div className="flex gap-2 overflow-x-auto pb-1">
        {["Tổng cộng", "Tháng 09", "Theo danh mục", "Không tính báo cáo"].map((filter) => (
          <Chip key={filter}>
            {filter}
          </Chip>
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
                <Text size="lg" weight="bold" className="">{group}</Text>
                <Text size="xs" tone="secondary" className="">{rows.length} giao dịch</Text>
              </div>
              <Text numeric weight="bold" tone={sum < 0 ? "danger" : "action"}>{formatVND(sum)}</Text>
            </div>
            <div className="space-y-2">{rows.map((transaction) => <TransactionItem key={transaction.id} transaction={transaction} />)}</div>
          </SurfaceCard>
        );
      })}
    </>
  );
}
