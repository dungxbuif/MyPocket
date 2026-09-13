import { Home, PieChart, Plus, ReceiptText, Settings } from "lucide-react";
import type { PrototypeTab } from "../pages/FinancePrototypePage";
import { BaseFab, BaseNavigationItem } from "../atoms/BaseNavigation";

const items = [
  { key: "overview", label: "Tổng quan", icon: Home },
  { key: "transactions", label: "Giao dịch", icon: ReceiptText },
  { key: "budgets", label: "Ngân sách", icon: PieChart },
  { key: "account", label: "Tài khoản", icon: Settings },
] as const;

export function BottomNavigation({
  tab,
  onTabChange,
  onAdd,
}: {
  tab: PrototypeTab;
  onTabChange: (tab: PrototypeTab) => void;
  onAdd: () => void;
}) {
  return (
    <nav className="fixed inset-x-0 bottom-0 z-30 mx-auto max-w-[430px] border-t border-line bg-card/95 px-3 pb-[calc(env(safe-area-inset-bottom)+8px)] pt-2 backdrop-blur">
      <BaseFab
        type="button"
        onClick={onAdd}
        className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2"
        aria-label="Thêm giao dịch"
      >
        <Plus size={26} />
      </BaseFab>
      <div className="grid grid-cols-5">
        {items.map((item, index) => {
          const Icon = item.icon;
          const active = tab === item.key;
          return (
            <BaseNavigationItem
              active={active}
              key={item.key}
              type="button"
              onClick={() => onTabChange(item.key)}
              className={`${
                index === 1 ? "pr-5" : index === 2 ? "col-start-4 pl-5" : index === 3 ? "col-start-5" : ""
              }`}
            >
              <Icon size={20} />
              {item.label}
            </BaseNavigationItem>
          );
        })}
      </div>
    </nav>
  );
}
