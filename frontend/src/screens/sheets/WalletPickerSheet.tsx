import * as React from "react";
import { Plus, Link2 } from "lucide-react";
import type { WalletSummary } from "../../app/finance";
import { Sheet } from "../../components/ui/sheet";
import { ModalHeader } from "../../components/navigation/ModalHeader";
import { GroupedCard } from "../../components/cards/GroupedCard";
import { RadioCheckItem } from "../../components/cards/RadioCheckItem";

export interface WalletPickerSheetProps {
  isOpen: boolean;
  onClose: () => void;
  wallets: WalletSummary[];
  selectedWalletId?: string;
  onSelectWallet: (walletId: string) => void;
  onAddWallet?: () => void;
}

export function WalletPickerSheet({
  isOpen,
  onClose,
  wallets,
  selectedWalletId = "all",
  onSelectWallet,
  onAddWallet,
}: WalletPickerSheetProps) {
  return (
    <Sheet
      isOpen={isOpen}
      onClose={onClose}
      title="Chọn Ví"
      header={
        <ModalHeader
          title="Chọn Ví"
          dismissLabel="Đóng"
          onDismiss={onClose}
          rightAction={{ label: "Sửa", onClick: () => undefined }}
        />
      }
    >
      <div className="flex flex-col gap-4 py-2">
        {/* All Wallets Option */}
        <div className="rounded-[24px] bg-white p-2 shadow-xs">
          <RadioCheckItem
            label="Tổng cộng"
            selected={selectedWalletId === "all"}
            onSelect={() => {
              onSelectWallet("all");
              onClose();
            }}
          />
        </div>

        {/* Individual Wallets */}
        <GroupedCard title="TÍNH VÀO TỔNG">
          {wallets.map((w) => (
            <RadioCheckItem
              key={w.id}
              label={w.name}
              sublabel={`${new Intl.NumberFormat("vi-VN").format(w.balance_vnd)} đ`}
              selected={selectedWalletId === w.id}
              onSelect={() => {
                onSelectWallet(w.id);
                onClose();
              }}
            />
          ))}
        </GroupedCard>

        {/* Bottom Actions */}
        <div className="flex flex-col gap-2 mt-2">
          <button
            type="button"
            onClick={onAddWallet}
            className="w-full h-12 rounded-full bg-white text-[#2dbd4f] font-semibold text-sm flex items-center justify-center gap-2 shadow-xs hover:bg-[#f8f9fa] active:scale-98 transition-all"
          >
            <Plus className="w-4 h-4" />
            <span>Thêm ví</span>
          </button>
          <button
            type="button"
            className="w-full h-12 rounded-full bg-white text-[#111111] font-semibold text-sm flex items-center justify-center gap-2 shadow-xs hover:bg-[#f8f9fa] active:scale-98 transition-all"
          >
            <Link2 className="w-4 h-4" />
            <span>Liên kết dịch vụ</span>
          </button>
        </div>
      </div>
    </Sheet>
  );
}
