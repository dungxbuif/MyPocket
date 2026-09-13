import { Bell, Eye, EyeOff } from "lucide-react";
import { IconButton } from "../atoms/IconButton";
import { wallets } from "../data/mockFinance";
import { formatVND } from "../utils/format";

export function AppHeader({ masked, onToggleMask }: { masked: boolean; onToggleMask: () => void }) {
  const totalBalance = wallets.reduce((sum, wallet) => sum + wallet.balance, 0);

  return (
    <header className="sticky top-0 z-20 bg-[#fbf9f9]/90 px-4 pb-3 pt-[calc(env(safe-area-inset-top)+16px)] backdrop-blur">
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-xs font-semibold text-[#3f4a3c]">Tổng số dư</p>
          <p className="money text-lg font-bold text-[#1b1c1c]">{masked ? "••••••••" : formatVND(totalBalance)}</p>
        </div>
        <div className="flex items-center gap-2">
          <IconButton label="Ẩn hiện số dư" onClick={onToggleMask}>
            {masked ? <EyeOff size={18} /> : <Eye size={18} />}
          </IconButton>
          <IconButton label="Thông báo">
            <Bell size={18} />
          </IconButton>
        </div>
      </div>
    </header>
  );
}
