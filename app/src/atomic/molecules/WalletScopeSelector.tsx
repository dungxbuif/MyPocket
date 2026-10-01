import { ChevronDown, CreditCard, Globe, PiggyBank, WalletCards } from "lucide-react";
import { useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { IconBadge } from "../atoms/IconBadge";
import { Text } from "../atoms/Text";
import { BaseBottomSheet } from "./BaseBottomSheet";
import { WalletSelectionList } from "./WalletSelectionList";
import type { Wallet } from "../../services/wallets";

function walletIcon(wallet?: Wallet) {
  if (!wallet) return Globe;
  if (wallet.type === "credit") return CreditCard;
  if (wallet.type === "goal") return PiggyBank;
  return WalletCards;
}

export function WalletScopeSelector({ wallets, selectedID, onSelect, onAdd, onEdit }: {
  wallets: Wallet[];
  selectedID: string | null;
  onSelect: (id: string | null) => void;
  onAdd: () => void;
  onEdit?: () => void;
}) {
  const [open, setOpen] = useState(false);
  const selected = wallets.find(wallet => wallet.id === selectedID);
  const choose = (id: string | null) => {
    setOpen(false);
    onSelect(id);
  };
  return <>
    <BaseButton
      variant="chip"
      size="sm"
      className="mx-auto flex w-fit min-w-[12rem] justify-center gap-2"
      aria-label="Chọn ví giao dịch"
      aria-haspopup="dialog"
      aria-expanded={open}
      onClick={() => setOpen(true)}
    >
      <IconBadge icon={walletIcon(selected)} size="sm" shape="circle" tone={selected?.type === "goal" ? "success" : selected?.type === "credit" ? "categoryTeal" : "categorySlate"} />
      <Text as="span" size="sm" weight="semibold" className="max-w-[10rem] truncate">{selected?.name ?? "Tổng cộng"}</Text>
      <ChevronDown size={17} aria-hidden="true" />
    </BaseButton>
    {open ? <BaseBottomSheet
      presentation="form"
      title="Chọn Ví"
      closeLabel="Đóng"
      onClose={() => setOpen(false)}
      headerAction={onEdit ? <BaseButton variant="chip" size="sm" onClick={onEdit}>Sửa</BaseButton> : undefined}
    >
      <WalletSelectionList
        wallets={wallets}
        selectedID={selectedID}
        editing={false}
        onSelect={choose}
        onAdd={() => { setOpen(false); onAdd(); }}
      />
    </BaseBottomSheet> : null}
  </>;
}
