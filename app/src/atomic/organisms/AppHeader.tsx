import { useEffect, useState } from "react";
import { Bell, Eye, EyeOff } from "lucide-react";

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
    <header className="sticky top-0 z-20 bg-canvas/90 px-4 pb-3 pt-[calc(env(safe-area-inset-top)+16px)] backdrop-blur">
      <div className="flex items-center justify-between gap-3">
        <div><Text size="xs" weight="semibold" tone="secondary">Tổng số dư</Text><Text numeric size="lg" weight="bold" tone="ink">{masked ? "••••••••" : totalBalance === null ? "—" : formatVND(totalBalance)}</Text></div>
        <div className="flex items-center gap-2"><IconButton label="Ẩn hiện số dư" onClick={onToggleMask}>{masked ? <EyeOff size={18} /> : <Eye size={18} />}</IconButton><IconButton label="Thông báo"><Bell size={18} /></IconButton></div>
      </div>
    </header>
  );
}
