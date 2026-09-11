import * as React from "react";
import { Plus, Eye, Search } from "lucide-react";
import { SubpageHeader } from "../../components/navigation/SubpageHeader";
import { WalletFilterChip } from "../../components/navigation/WalletFilterChip";
import { SegmentedControl } from "../../components/ui/segmented-control";
import { CategoryTreeRow } from "../../components/finance/CategoryTreeRow";
import { TooltipPopover } from "../../components/ui/tooltip";

export interface CategoryManagementScreenProps {
  onBack: () => void;
  onSelectCategory?: (category: { id: string; name: string }) => void;
}

export function CategoryManagementScreen({ onBack, onSelectCategory }: CategoryManagementScreenProps) {
  const [tab, setTab] = React.useState<"expense" | "income" | "debt">("expense");
  const [tooltipOpen, setTooltipOpen] = React.useState(false);
  const [showInactive, setShowInactive] = React.useState(false);

  return (
    <div className="flex flex-col min-h-screen bg-[#f2f3f8] px-4 py-3 select-none">
      {/* Header */}
      <SubpageHeader
        title="Nhóm"
        onBack={onBack}
        centerExtra={<WalletFilterChip walletName="Tổng cộng" />}
        rightAction={{
          icon: <Search className="w-5 h-5 text-[#29495a]" />,
          onClick: () => undefined,
        }}
      />

      {/* Top Action Pill */}
      <div className="flex justify-center my-2">
        <button
          type="button"
          className="h-9 px-4 rounded-full bg-[#2dbd4f] text-white text-xs font-bold flex items-center gap-1.5 shadow-xs hover:bg-[#25a443] active:scale-95 transition-all"
        >
          <Plus className="w-3.5 h-3.5 stroke-[3]" />
          <span>Nhóm mới</span>
        </button>
      </div>

      {/* 3-Segment Switcher */}
      <SegmentedControl
        options={[
          { key: "expense", label: "Khoản chi" },
          { key: "income", label: "Khoản thu" },
          { key: "debt", label: "Vay/Nợ" },
        ]}
        selectedKey={tab}
        onChange={(k) => setTab(k as "expense" | "income" | "debt")}
        className="my-2"
      />

      {/* Tooltip for locked system categories */}
      {tooltipOpen && (
        <div className="relative flex justify-center">
          <TooltipPopover
            isOpen={tooltipOpen}
            text="Đây là danh mục của hệ thống nên bạn không thể chỉnh sửa hoặc xóa bỏ."
            onClose={() => setTooltipOpen(false)}
            className="top-0"
          />
        </div>
      )}

      {/* Categories List */}
      <div className="rounded-[24px] bg-white p-2 shadow-xs my-2 divide-y divide-[#e8e8ec]/80">
        {tab === "expense" ? (
          <>
            <CategoryTreeRow
              id="cat_food"
              name="Ăn uống"
              onClick={() => onSelectCategory?.({ id: "cat_food", name: "Ăn uống" })}
            />
            <CategoryTreeRow
              id="cat_snack"
              name="Ăn vặt"
              isChild
              onClick={() => onSelectCategory?.({ id: "cat_snack", name: "Ăn vặt" })}
            />
            <CategoryTreeRow
              id="cat_coffee"
              name="Cà phê"
              isChild
              onClick={() => onSelectCategory?.({ id: "cat_coffee", name: "Cà phê" })}
            />
            <CategoryTreeRow
              id="cat_meal"
              name="Cơm Bữa"
              isChild
              onClick={() => onSelectCategory?.({ id: "cat_meal", name: "Cơm Bữa" })}
            />

            <CategoryTreeRow
              id="cat_bills"
              name="Hoá đơn & Tiện ích"
              onClick={() => onSelectCategory?.({ id: "cat_bills", name: "Hoá đơn & Tiện ích" })}
            />
            <CategoryTreeRow
              id="cat_shopping"
              name="Mua sắm"
              onClick={() => onSelectCategory?.({ id: "cat_shopping", name: "Mua sắm" })}
            />
            <CategoryTreeRow
              id="cat_family"
              name="Gia đình"
              onClick={() => onSelectCategory?.({ id: "cat_family", name: "Gia đình" })}
            />
            <CategoryTreeRow
              id="cat_transit"
              name="Di chuyển"
              onClick={() => onSelectCategory?.({ id: "cat_transit", name: "Di chuyển" })}
            />
            <CategoryTreeRow
              id="cat_health"
              name="Sức khỏe"
              onClick={() => onSelectCategory?.({ id: "cat_health", name: "Sức khỏe" })}
            />

            {/* Locked System Categories */}
            <CategoryTreeRow
              id="cat_locked_other"
              name="Các chi phí khác"
              isSystemLocked
              onLockedClick={() => setTooltipOpen(true)}
            />
            <CategoryTreeRow
              id="cat_locked_transfer"
              name="Tiền chuyển đi"
              isSystemLocked
              onLockedClick={() => setTooltipOpen(true)}
            />
          </>
        ) : tab === "income" ? (
          <>
            <CategoryTreeRow
              id="cat_salary"
              name="Lương"
              onClick={() => onSelectCategory?.({ id: "cat_salary", name: "Lương" })}
            />
            <CategoryTreeRow
              id="cat_bonus"
              name="Thưởng"
              onClick={() => onSelectCategory?.({ id: "cat_bonus", name: "Thưởng" })}
            />
            <CategoryTreeRow
              id="cat_locked_other_inc"
              name="Thu nhập khác"
              isSystemLocked
              onLockedClick={() => setTooltipOpen(true)}
            />
          </>
        ) : (
          <>
            <CategoryTreeRow
              id="cat_lend"
              name="Cho vay"
              isSystemLocked
              onLockedClick={() => setTooltipOpen(true)}
            />
            <CategoryTreeRow
              id="cat_repay"
              name="Trả nợ"
              isSystemLocked
              onLockedClick={() => setTooltipOpen(true)}
            />
            <CategoryTreeRow
              id="cat_borrow"
              name="Đi vay"
              isSystemLocked
              onLockedClick={() => setTooltipOpen(true)}
            />
          </>
        )}
      </div>

      {/* Footer Action */}
      <div className="flex justify-center my-3">
        <button
          type="button"
          onClick={() => setShowInactive(!showInactive)}
          className="h-9 px-4 rounded-full bg-white text-[#8e8e93] text-xs font-semibold flex items-center gap-1.5 shadow-xs hover:bg-[#f8f9fa] active:scale-95 transition-all"
        >
          <Eye className="w-3.5 h-3.5" />
          <span>Hiển thị nhóm không hoạt động</span>
        </button>
      </div>
    </div>
  );
}
