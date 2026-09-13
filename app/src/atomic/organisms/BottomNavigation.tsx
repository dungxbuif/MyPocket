import { Home, PieChart, Plus, ReceiptText, Settings } from "lucide-react";
import type { PrototypeTab } from "../pages/FinancePrototypePage";

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
    <nav className="fixed inset-x-0 bottom-0 z-30 mx-auto max-w-[430px] border-t border-[#e3e2e2] bg-white/95 px-3 pb-[calc(env(safe-area-inset-bottom)+8px)] pt-2 backdrop-blur">
      <button
        type="button"
        onClick={onAdd}
        className="absolute left-1/2 top-1 grid h-14 w-14 -translate-x-1/2 -translate-y-1/3 place-items-center rounded-full bg-[#006e1c] text-white shadow-[0_12px_24px_rgb(0_110_28/0.25)]"
        aria-label="Thêm giao dịch"
      >
        <Plus size={26} />
      </button>
      <div className="grid grid-cols-5">
        {items.map((item, index) => {
          const Icon = item.icon;
          const active = tab === item.key;
          return (
            <button
              key={item.key}
              type="button"
              onClick={() => onTabChange(item.key)}
              className={`flex min-h-[52px] flex-col items-center justify-center gap-1 text-xs font-bold ${
                index === 1 ? "pr-5" : index === 2 ? "col-start-4 pl-5" : index === 3 ? "col-start-5" : ""
              } ${active ? "text-[#006e1c]" : "text-[#6f7a6b]"}`}
            >
              <Icon size={20} />
              {item.label}
            </button>
          );
        })}
      </div>
    </nav>
  );
}
