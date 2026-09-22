import { Progress } from "../atoms/Progress";
import { Text } from "../atoms/Text";
import { formatVND } from "../utils/format";
import type { Wallet } from "../../services/wallets";
import { todayDateKey } from "../../services/accountTime";
import { useAccountTimezone } from "../../services/AccountTimezoneContext";

export function savingsProgress(balance: number, target: number) {
  return { remaining: Math.max(target - balance, 0), percentage: target > 0 ? Math.max(0, Math.min(100, balance / target * 100)) : 0, reached: target > 0 && balance >= target };
}

export function SavingsSummary({ wallet }: { wallet: Wallet }) {
  const timezone=useAccountTimezone();
  const goal = savingsProgress(wallet.current_balance, wallet.target_amount ?? 0);
  const date = wallet.target_date?.slice(0,10);
  const today = todayDateKey(timezone);
  const todayUTC = Date.parse(`${today}T00:00:00Z`);
  const days = date ? Math.round((Date.parse(`${date}T00:00:00Z`) - todayUTC) / 86400000) : null;
  return <div className="space-y-5 py-4">
    <div className="text-center"><Text tone="secondary">Số dư</Text><Text size="3xl" weight="bold" numeric>{formatVND(wallet.current_balance)}</Text></div>
    <div className="text-center"><Text tone="secondary">{goal.reached ? "ĐÃ ĐẠT MỤC TIÊU" : "CẦN THÊM"}</Text><Text size="4xl" numeric weight="bold">{formatVND(goal.remaining)}</Text>{days !== null ? <Text size="lg">{days < 0 ? `Đã qua hạn ${-days} ngày` : days === 0 ? "Đến hạn hôm nay" : `Còn ${days} ngày`}</Text> : <Text tone="secondary">Chưa đặt ngày kết thúc</Text>}</div>
    <Progress value={goal.percentage} label="Tiến độ tiết kiệm" />
    <Text size="sm" tone="secondary" className="text-center">Mục tiêu {formatVND(wallet.target_amount ?? 0)}</Text>
  </div>;
}
