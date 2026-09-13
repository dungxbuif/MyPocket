import { UserProfile } from "../../services/auth";
import { ProfileHeroCard } from "../molecules/ProfileHeroCard";
import { Wallet, Layers3 } from "lucide-react";
import { AccountMenuRow } from "../molecules/AccountMenuRow";
import { SurfaceCard } from "../atoms/SurfaceCard";

export function AccountPanel({
  user,
  onLogout,
}: {
  user: UserProfile | null;
  onLogout: () => void;
}) {
  return (
    <>
      <ProfileHeroCard user={user} />
      <button type="button" onClick={onLogout} className="w-full cursor-pointer rounded-full bg-[#006e1c] py-3 text-sm font-semibold text-white transition hover:bg-[#005313]">Đăng xuất</button>
      {/* Tạm ẩn các màn ví, nhóm, mục tiêu, định kỳ, nợ và OCR cho đến khi backend tương ứng hoàn tất. */}
      <SurfaceCard radius="md" className="overflow-hidden">
        <AccountMenuRow to="/account/wallets" icon={Wallet} label="Ví của tôi" />
        <div className="mx-4 h-px bg-[#e3e2e2]" />
        <AccountMenuRow to="/account/groups" icon={Layers3} label="Nhóm" />
      </SurfaceCard>
      {/* Cài đặt dữ liệu chưa có backend contract nên tạm không render. */}
      {/* <section className="rounded-3xl bg-white p-2">
        {items.map((item) => {
          const Icon = item.icon;
          return (
            <button key={item.title} className="flex w-full items-center gap-3 rounded-2xl p-3 text-left hover:bg-[#f5f3f3]">
              <div className="grid h-10 w-10 place-items-center rounded-full bg-[#d9e6da] text-[#006e1c]">
                <Icon size={18} />
              </div>
              <div className="min-w-0 flex-1">
                <p className="font-semibold">{item.title}</p>
                <p className="truncate text-sm text-[#3f4a3c]">{item.detail}</p>
              </div>
              <ChevronRight size={18} className="text-[#6f7a6b]" />
            </button>
          );
        })}
      </section> */}
    </>
  );
}
