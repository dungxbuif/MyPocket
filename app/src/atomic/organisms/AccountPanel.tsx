import { CalendarDays, Camera, ChevronRight, CreditCard, Settings, Target, Wallet } from "lucide-react";
import { SectionTitle } from "../atoms/SectionTitle";
import { goals } from "../data/mockFinance";
import { GoalCard } from "../molecules/GoalCard";
import { UserProfile } from "../../services/auth";

const items = [
  { icon: Wallet, title: "Quản lý ví", detail: "Tiền mặt, ngân hàng, thẻ tín dụng" },
  { icon: CalendarDays, title: "Giao dịch định kỳ", detail: "Lương, thuê nhà, hóa đơn" },
  { icon: CreditCard, title: "Nợ và cho vay", detail: "Nhắc hạn, trả một phần, lịch sử" },
  { icon: Target, title: "Mục tiêu & quỹ", detail: "Du lịch Đà Lạt, laptop mới" },
  { icon: Camera, title: "OCR hóa đơn", detail: "Mock workflow, chưa nối service" },
  { icon: Settings, title: "Cài đặt dữ liệu", detail: "Export, privacy, passcode" },
];

export function AccountPanel({
  user,
  onLogout,
}: {
  user: UserProfile | null;
  onLogout: () => void;
}) {
  const displayName = user?.name?.trim() || "Người dùng";
  const displayEmail = user?.email?.trim() || "Chưa có email";

  return (
    <>
      <section className="rounded-3xl bg-white p-4">
        <div className="flex items-center gap-3">
          <div className="grid h-14 w-14 place-items-center rounded-full bg-[#006e1c] text-xl font-bold text-white">D</div>
          <div>
            <p className="text-lg font-bold">{displayName}</p>
            <p className="text-sm text-[#3f4a3c]">{displayEmail}</p>
          </div>
        </div>
        <div className="mt-4">
          <button
            type="button"
            onClick={onLogout}
            className="w-full rounded-full bg-[#006e1c] py-2 text-sm font-semibold text-white"
          >
            Đăng xuất
          </button>
        </div>
      </section>
      <section className="rounded-3xl bg-white p-4">
        <SectionTitle title="Mục tiêu cá nhân" action="Thêm" />
        <div className="mt-3 space-y-3">{goals.map((goal) => <GoalCard key={goal.id} goal={goal} masked={false} />)}</div>
      </section>
      <section className="rounded-3xl bg-white p-2">
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
      </section>
    </>
  );
}
