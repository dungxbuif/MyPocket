import { useEffect, useMemo, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { Divider } from "../atoms/Divider";
import { BaseSelect, FormField } from "../atoms/FormField";
import { StatusMessage } from "../atoms/StatusMessage";
import { Text } from "../atoms/Text";
import { updateAccountTimezone, type UserProfile } from "../../services/auth";
import { ProfileHeroCard } from "../molecules/ProfileHeroCard";
import { Wallet, Layers3, MessageSquareText } from "lucide-react";
import { AccountMenuRow } from "../molecules/AccountMenuRow";
import { SurfaceCard } from "../atoms/SurfaceCard";

export function AccountPanel({
  user,
  onLogout,
  onTimezoneChange,
}: {
  user: UserProfile | null;
  onLogout: () => void;
  onTimezoneChange: (user: UserProfile) => void;
}) {
  const [timezone, setTimezone] = useState(user?.timezone ?? "Asia/Ho_Chi_Minh");
  const [savingTimezone, setSavingTimezone] = useState(false);
  const [timezoneMessage, setTimezoneMessage] = useState("");
  const [timezoneError, setTimezoneError] = useState("");
  const timezones = useMemo(() => {
    const intl = Intl as typeof Intl & { supportedValuesOf?: (key: "timeZone") => string[] };
    const supported = intl.supportedValuesOf?.("timeZone") ?? ["Asia/Ho_Chi_Minh", "America/Los_Angeles", "Europe/London", "UTC"];
    return [...new Set(["UTC", ...(user?.timezone ? [user.timezone] : []), ...supported])].sort((a, b) => a.localeCompare(b));
  }, [user?.timezone]);

  useEffect(() => {
    setTimezone(user?.timezone ?? "Asia/Ho_Chi_Minh");
  }, [user?.timezone]);

  const saveTimezone = async () => {
    if (!user || savingTimezone || timezone === user.timezone) return;
    setSavingTimezone(true);
    setTimezoneMessage("");
    setTimezoneError("");
    try {
      const updated = await updateAccountTimezone(timezone);
      onTimezoneChange(updated);
      setTimezoneMessage("Đã cập nhật múi giờ tài khoản.");
    } catch (error) {
      setTimezoneError(error instanceof Error ? error.message : "Không thể cập nhật múi giờ.");
    } finally {
      setSavingTimezone(false);
    }
  };

  return (
    <>
      <ProfileHeroCard user={user} />
      <BaseButton variant="primary" size="md" type="button" onClick={onLogout} className="w-full">Đăng xuất</BaseButton>
      {/* Tạm ẩn các màn ví, nhóm, mục tiêu, định kỳ, nợ và OCR cho đến khi backend tương ứng hoàn tất. */}
      <SurfaceCard radius="md" className="overflow-hidden">
        <AccountMenuRow to="/account/wallets" icon={Wallet} label="Ví của tôi" />
        <Divider className="mx-4" />
        <AccountMenuRow to="/account/groups" icon={Layers3} label="Nhóm" />
        <Divider className="mx-4" />
        <AccountMenuRow to="/account/feedback" icon={MessageSquareText} label="Phản hồi" />
      </SurfaceCard>
      <SurfaceCard padding="md" className="space-y-3">
        <Text weight="semibold">Múi giờ</Text>
        <Text size="sm" tone="secondary">Ngày và tháng trong báo cáo dùng múi giờ này. Đổi múi giờ không thay đổi thời điểm giao dịch đã lưu.</Text>
        <FormField label="Múi giờ tài khoản">
          <BaseSelect value={timezone} disabled={savingTimezone} onChange={event => { setTimezone(event.target.value); setTimezoneMessage(""); setTimezoneError(""); }}>
            {timezones.map(value => <option key={value} value={value}>{value}</option>)}
          </BaseSelect>
        </FormField>
        {timezoneError ? <StatusMessage tone="danger">{timezoneError}</StatusMessage> : null}
        {timezoneMessage ? <StatusMessage>{timezoneMessage}</StatusMessage> : null}
        <BaseButton className="w-full" disabled={savingTimezone || timezone === user?.timezone} loading={savingTimezone} onClick={() => void saveTimezone()}>Lưu múi giờ</BaseButton>
      </SurfaceCard>
      {/* Cài đặt dữ liệu chưa có backend contract nên tạm không render. */}
      {/* <section className="rounded-3xl bg-card p-2">
        {items.map((item) => {
          const Icon = item.icon;
          return (
            <button key={item.title} className="flex w-full items-center gap-3 rounded-2xl p-3 text-left hover:bg-row">
              <div className="grid h-10 w-10 place-items-center rounded-full bg-success-soft text-action">
                <Icon size={18} />
              </div>
              <div className="min-w-0 flex-1">
                <p className="font-semibold">{item.title}</p>
                <p className="truncate text-sm text-secondary">{item.detail}</p>
              </div>
              <ChevronRight size={18} className="text-secondary" />
            </button>
          );
        })}
      </section> */}
    </>
  );
}
