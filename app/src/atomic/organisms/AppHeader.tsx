import { useEffect, useState } from "react";
import { Bell, Eye, EyeOff, Search } from "lucide-react";

import { IconButton } from "../atoms/IconButton";
import { Text } from "../atoms/Text";
import { formatVND } from "../utils/format";
import { fetchWallets } from "../../services/wallets";
import { totalWalletBalance } from "../../services/walletLogic";

export function AppHeader({ masked, onToggleMask, refreshKey = 0 }: { masked: boolean; onToggleMask: () => void; refreshKey?: number }) {
  const [totalBalance, setTotalBalance] = useState<number | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetchWallets().then((wallets) => { if (!cancelled) setTotalBalance(totalWalletBalance(wallets)); }).catch(() => { if (!cancelled) setTotalBalance(null); });
    return () => { cancelled = true; };
  }, [refreshKey]);

  return (
    <header className="sticky top-0 z-20 bg-canvas/95 px-5 pb-5 pt-[calc(env(safe-area-inset-top)+18px)] backdrop-blur">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-1">
            <Text numeric size="3xl" weight="extrabold" tone="heading" className="tracking-tight">{masked ? "••••••••" : totalBalance === null ? "—" : formatVND(totalBalance)}</Text>
            <IconButton variant="bare" label="Ẩn hiện số dư" onClick={onToggleMask}>{masked ? <EyeOff size={20} /> : <Eye size={20} />}</IconButton>
          </div>
          <Text size="base" weight="medium" tone="secondary">Tổng số dư</Text>
        </div>
        <div className="flex shrink-0 items-center gap-1 pt-2"><IconButton variant="bare" label="Tìm kiếm"><Search size={21} /></IconButton><IconButton variant="bare" label="Thông báo"><Bell size={21} /></IconButton></div>
      </div>
    </header>
  );
}
