import * as React from "react";
import { RefreshCw } from "lucide-react";
import { SubpageHeader } from "../../components/navigation/SubpageHeader";
import { GroupedCard } from "../../components/cards/GroupedCard";
import { FormRowItem } from "../../components/cards/FormRowItem";
import { SwitchRow } from "../../components/cards/SwitchRow";

export interface SettingsScreenProps {
  onBack: () => void;
  onExportCSV?: () => void;
}

export function SettingsScreen({ onBack, onExportCSV }: SettingsScreenProps) {
  const [dailyReminder, setDailyReminder] = React.useState(false);
  const [excludeFromReport, setExcludeFromReport] = React.useState(false);

  return (
    <div className="flex flex-col min-h-screen bg-[#f2f3f8] px-4 py-3 select-none pb-12">
      <SubpageHeader title="Cài đặt" onBack={onBack} />

      {/* Hiển Thị */}
      <GroupedCard title="HIỂN THỊ">
        <FormRowItem label="Kiểu hiển thị số tiền" value="1.083.096 đ" />
        <FormRowItem label="Định dạng thời gian" value="23/08/2026" />
        <FormRowItem label="Chọn ngôn ngữ" value="Tiếng Việt" />
        <FormRowItem label="Đơn vị tiền cho ví Tổng" value="VND" />
        <FormRowItem label="Chọn ngày đầu tuần" value="Thứ hai" />
        <FormRowItem label="Đặt ngày đầu tiên của tháng" value="5" />
        <FormRowItem label="Chọn tháng đầu tiên của năm" value="Tháng Một" />
        <FormRowItem label="Thiết lập khoảng thời gian..." value="Trong 3 tháng..." />
        <FormRowItem label="Chế độ hiển thị Tổng quan" value="Hiện theo tiền..." />
        <SwitchRow
          title="Không tính vào báo cáo"
          description="Các giao dịch mới tạo mặc định sẽ không tính vào báo cáo."
          checked={excludeFromReport}
          onChange={setExcludeFromReport}
        />
      </GroupedCard>

      {/* Hệ Thống */}
      <GroupedCard title="HỆ THỐNG">
        <SwitchRow
          title="Nhắc thêm giao dịch hàng ngày"
          description="Thông báo vào cuối ngày để ghi chép các khoản chi."
          checked={dailyReminder}
          onChange={setDailyReminder}
        />
        <FormRowItem label="Bảo mật" value="Mã khóa & Face ID" />
      </GroupedCard>

      {/* Cơ Sở Dữ Liệu */}
      <GroupedCard title="CƠ SỞ DỮ LIỆU">
        <FormRowItem label="Xuất file CSV" onClick={onExportCSV} />
        <div className="flex items-center justify-between py-3.5 px-2">
          <div className="flex flex-col">
            <span className="text-[15px] font-semibold text-[#111111]">Chạm để cập nhật tỷ giá</span>
            <span className="text-xs text-[#8e8e93] mt-0.5">Cập nhật lần cuối: 2026-06-24</span>
          </div>
          <button
            type="button"
            className="w-8 h-8 rounded-full bg-[#eef0f4] flex items-center justify-center text-[#29495a] active:scale-90 transition-all"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>
      </GroupedCard>
    </div>
  );
}
