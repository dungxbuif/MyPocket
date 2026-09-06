import * as React from "react";
import { Smartphone } from "lucide-react";
import { SubpageHeader } from "../../components/navigation/SubpageHeader";
import { GroupedCard } from "../../components/cards/GroupedCard";
import { DestructiveActionRow } from "../../components/cards/DestructiveActionRow";

export interface AccountSecurityScreenProps {
  onBack: () => void;
  onLogout: () => void;
  userEmail?: string;
  displayName?: string;
}

export function AccountSecurityScreen({
  onBack,
  onLogout,
  userEmail = "dungxbuif@gmail.com",
  displayName = "Dung Bui",
}: AccountSecurityScreenProps) {
  return (
    <div className="flex flex-col min-h-screen bg-[#f2f3f8] px-4 py-3 select-none">
      <SubpageHeader title="Quản Lý Tài Khoản" onBack={onBack} />

      {/* User Profile Badge Card */}
      <div className="rounded-[28px] bg-white p-5 shadow-xs flex flex-col items-center text-center my-3 relative">
        <div className="w-16 h-16 rounded-full bg-[#ff8800] text-white flex items-center justify-center font-bold text-2xl mb-2 shadow-xs">
          {displayName.slice(0, 1).toUpperCase()}
        </div>

        <span className="px-3 py-0.5 rounded-full bg-[#fff3e0] text-[#ff8800] text-[11px] font-bold tracking-wide uppercase mb-1">
          TÀI KHOẢN PREMIUM
        </span>

        <h3 className="font-bold text-lg text-[#111111]">{displayName}</h3>
        <span className="text-xs text-[#8e8e93] mt-0.5">{userEmail}</span>

        <button
          type="button"
          className="mt-4 h-10 px-5 rounded-full bg-[#eef0f4] hover:bg-[#e2e4e9] text-[#111111] text-xs font-semibold active:scale-95 transition-all"
        >
          Thay đổi mật khẩu
        </button>
      </div>

      {/* Active Devices */}
      <GroupedCard title="THIẾT BỊ 1/5">
        <div className="flex items-center justify-between py-2 px-1">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-[#eef0f4] flex items-center justify-center text-[#29495a]">
              <Smartphone className="w-5 h-5" />
            </div>
            <div className="flex flex-col">
              <span className="text-sm font-semibold text-[#111111]">iPhone</span>
              <span className="text-[11px] text-[#2dbd4f] font-medium">Thiết bị này</span>
            </div>
          </div>
        </div>
      </GroupedCard>

      {/* Logout Action */}
      <div className="my-3">
        <DestructiveActionRow label="Đăng xuất" onClick={onLogout} variant="pill" />
      </div>

      {/* Irreversible Operations */}
      <GroupedCard>
        <DestructiveActionRow label="Đặt lại tài khoản" onClick={() => undefined} />
        <DestructiveActionRow label="Xóa tài khoản" onClick={() => undefined} />
      </GroupedCard>
    </div>
  );
}
