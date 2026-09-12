import { Bell, Eye, EyeOff } from "lucide-react";
import { IconButton } from "../atoms/IconButton";

export function AppHeader({ masked, onToggleMask }: { masked: boolean; onToggleMask: () => void }) {
  return (
    <header className="sticky top-0 z-20 bg-[#fbf9f9]/90 px-4 pb-3 pt-[calc(env(safe-area-inset-top)+16px)] backdrop-blur">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#3f4a3c]">MyPocket</p>
          <h1 className="text-xl font-bold text-[#1b1c1c]">Tài chính hôm nay</h1>
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

