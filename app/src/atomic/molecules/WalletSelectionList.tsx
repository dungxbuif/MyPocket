import { Check, CreditCard, Globe, Link, PiggyBank, Plus, WalletCards } from "lucide-react";
import { BaseButton } from "../atoms/BaseButton";
import { Divider } from "../atoms/Divider";
import { IconBadge } from "../atoms/IconBadge";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { StatusMessage } from "../atoms/StatusMessage";
import { formatVND } from "../utils/format";
import { totalWalletBalance } from "../../services/walletLogic";
import type { Wallet } from "../../services/wallets";

export function WalletSelectionList({ wallets, selectedID, editing, onSelect, onAdd }: { wallets: Wallet[]; selectedID: string | null; editing: boolean; onSelect: (id: string | null) => void; onAdd: () => void }) {
  const groups = [{ label: "TÍNH VÀO TỔNG", items: wallets.filter(wallet => wallet.is_in_total) }, { label: "KHÔNG TÍNH VÀO TỔNG", items: wallets.filter(wallet => !wallet.is_in_total) }];
  return <div className="space-y-5">
    <SurfaceCard padding="sm"><BaseButton variant="row" size="row" disabled={editing} aria-pressed={selectedID === null} onClick={() => onSelect(null)}><IconBadge icon={Globe} size="lg" shape="circle" tone="categoryTeal" /><span className="min-w-0 flex-1"><Text as="span" size="lg" weight="bold" className="block">Tổng cộng</Text><Text as="span" numeric tone="secondary" className="block">{formatVND(totalWalletBalance(wallets))}</Text></span>{selectedID === null && !editing ? <Text as="span" tone="action"><Check size={20} aria-hidden="true" /></Text> : null}</BaseButton></SurfaceCard>
    {wallets.length === 0 ? <StatusMessage variant="plain">Chưa có ví.</StatusMessage> : null}
    {groups.filter(group => group.items.length > 0).map(group => <div key={group.label} className="space-y-2"><Text size="xs" weight="bold" tone="secondary" className="px-3">{group.label}</Text><SurfaceCard padding="sm">{group.items.map((wallet, index) => <div key={wallet.id}>{index > 0 ? <Divider className="ml-16" /> : null}<BaseButton variant="row" size="row" aria-pressed={!editing && selectedID === wallet.id} onClick={() => onSelect(wallet.id)}><IconBadge icon={wallet.type === "credit" ? CreditCard : wallet.type === "goal" ? PiggyBank : WalletCards} size="lg" shape="circle" tone={wallet.type === "credit" ? "categoryTeal" : wallet.type === "goal" ? "success" : "categoryOrange"} /><span className="min-w-0 flex-1"><Text as="span" size="base" weight="bold" className="block truncate">{wallet.name}</Text><Text as="span" numeric tone="secondary" className="block">{formatVND(wallet.current_balance)}</Text></span>{!editing && selectedID === wallet.id ? <Text as="span" tone="action"><Check size={20} aria-hidden="true" /></Text> : null}</BaseButton></div>)}</SurfaceCard></div>)}
    <SurfaceCard padding="sm"><BaseButton variant="row" size="row" onClick={onAdd}><IconBadge icon={Plus} size="sm" shape="circle" tone="success" /><Text as="span" size="base" weight="semibold" tone="action">Thêm ví</Text></BaseButton><Divider className="ml-12" /><BaseButton variant="row" size="row" disabled><IconBadge icon={Link} size="sm" shape="circle" tone="success" /><Text as="span" size="base" tone="action">Liên kết dịch vụ</Text></BaseButton></SurfaceCard>
    <Text size="xs" tone="secondary">Liên kết dịch vụ chưa được hỗ trợ.</Text>
  </div>;
}
