import * as React from "react";
import { Plus, CalendarSync } from "lucide-react";
import { SubpageHeader } from "../../components/navigation/SubpageHeader";
import { WalletFilterChip } from "../../components/navigation/WalletFilterChip";
import { EmptyStateView } from "../../components/feedback/EmptyStateView";

export interface RecurringTransactionsScreenProps {
  onBack: () => void;
  onCreateSchedule?: () => void;
}

export function RecurringTransactionsScreen({
  onBack,
  onCreateSchedule,
}: RecurringTransactionsScreenProps) {
  return (
    <div className="flex flex-col min-h-screen bg-[#f2f3f8] px-4 py-3 select-none">
      <SubpageHeader
        title="Giao dịch định kì"
        onBack={onBack}
        centerExtra={<WalletFilterChip walletName="Tổng cộng" />}
        rightAction={{
          icon: <Plus className="w-5 h-5 text-[#29495a]" />,
          onClick: () => onCreateSchedule?.(),
        }}
      />

      <p className="text-xs text-[#8e8e93] text-center px-4 my-2 leading-relaxed">
        Tạo ra định kỳ các giao dịch sẽ được tự động thêm trong tương lai.
      </p>

      <div className="flex-1 flex items-center justify-center">
        <EmptyStateView
          illustration={<CalendarSync className="w-16 h-16 text-[#8e8e93]/60" />}
          title="Không có giao dịch định kỳ nào"
          description="Thiết lập các khoản thanh toán tiền nhà, tiền mạng hoặc lương cố định hàng tháng."
          actionButton={{
            label: "Tạo giao dịch định kì",
            onClick: () => onCreateSchedule?.(),
          }}
        />
      </div>
    </div>
  );
}
