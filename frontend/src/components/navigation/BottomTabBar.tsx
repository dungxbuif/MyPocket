import * as React from "react";
import { cn } from "../../lib/utils";
import { Home, Wallet, PieChart, User, Plus } from "lucide-react";

export type TabKey = "overview" | "transactions" | "budgets" | "account";

export interface BottomTabBarProps {
  activeTab: TabKey;
  onTabSelect: (tab: TabKey) => void;
  onAddClick: () => void;
  className?: string;
}

export function BottomTabBar({ activeTab, onTabSelect, onAddClick, className }: BottomTabBarProps) {
  return (
    <nav
      role="navigation"
      aria-label="Điều hướng chính"
      className={cn(
        "fixed bottom-[max(16px,env(safe-area-inset-bottom))] left-4 right-4 max-w-[408px] mx-auto h-16 rounded-full bg-white/95 backdrop-blur-md shadow-[0_4px_24px_rgba(0,0,0,0.08)] border border-white/60 flex items-center justify-around px-2 z-40",
        className
      )}
    >
      {/* Tab 1: Tổng quan */}
      <button
        type="button"
        aria-label="Tổng quan"
        onClick={() => onTabSelect("overview")}
        className={cn(
          "flex flex-col items-center justify-center gap-0.5 px-3 py-1.5 rounded-full transition-all outline-none",
          activeTab === "overview" ? "bg-[#eef0f4] text-[#111111]" : "text-[#8e8e93] hover:text-[#111111]"
        )}
      >
        <Home className="w-5 h-5 stroke-[2.2]" />
        <span className={cn("text-[11px]", activeTab === "overview" ? "font-bold text-[#111111]" : "font-medium")}>
          Tổng quan
        </span>
      </button>

      {/* Tab 2: Sổ giao dịch */}
      <button
        type="button"
        aria-label="Sổ giao dịch"
        onClick={() => onTabSelect("transactions")}
        className={cn(
          "flex flex-col items-center justify-center gap-0.5 px-3 py-1.5 rounded-full transition-all outline-none",
          activeTab === "transactions" ? "bg-[#eef0f4] text-[#111111]" : "text-[#8e8e93] hover:text-[#111111]"
        )}
      >
        <Wallet className="w-5 h-5 stroke-[2.2]" />
        <span className={cn("text-[11px]", activeTab === "transactions" ? "font-bold text-[#111111]" : "font-medium")}>
          Sổ giao dịch
        </span>
      </button>

      {/* Center Action: Thêm giao dịch */}
      <button
        type="button"
        aria-label="Thêm giao dịch"
        onClick={onAddClick}
        className="w-[52px] h-[52px] rounded-full bg-[#2dbd4f] text-white flex items-center justify-center -mt-4 shadow-[0_6px_16px_rgba(45,189,79,0.4)] active:scale-95 transition-all hover:bg-[#25a443] outline-none"
      >
        <Plus className="w-7 h-7 stroke-[2.5]" />
      </button>

      {/* Tab 3: Ngân sách */}
      <button
        type="button"
        aria-label="Ngân sách"
        onClick={() => onTabSelect("budgets")}
        className={cn(
          "flex flex-col items-center justify-center gap-0.5 px-3 py-1.5 rounded-full transition-all outline-none",
          activeTab === "budgets" ? "bg-[#eef0f4] text-[#111111]" : "text-[#8e8e93] hover:text-[#111111]"
        )}
      >
        <PieChart className="w-5 h-5 stroke-[2.2]" />
        <span className={cn("text-[11px]", activeTab === "budgets" ? "font-bold text-[#111111]" : "font-medium")}>
          Ngân sách
        </span>
      </button>

      {/* Tab 4: Tài khoản */}
      <button
        type="button"
        aria-label="Tài khoản"
        onClick={() => onTabSelect("account")}
        className={cn(
          "flex flex-col items-center justify-center gap-0.5 px-3 py-1.5 rounded-full transition-all outline-none",
          activeTab === "account" ? "bg-[#eef0f4] text-[#111111]" : "text-[#8e8e93] hover:text-[#111111]"
        )}
      >
        <User className="w-5 h-5 stroke-[2.2]" />
        <span className={cn("text-[11px]", activeTab === "account" ? "font-bold text-[#111111]" : "font-medium")}>
          Tài khoản
        </span>
      </button>
    </nav>
  );
}
