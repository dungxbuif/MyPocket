import { useEffect, useMemo, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
import { AmountField } from "../molecules/AmountField";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { Divider } from "../atoms/Divider";
import { fetchWallets, type Wallet } from "../../services/wallets";
import { fetchCategories, type Category } from "../../services/categories";
import { createRecurring, deleteRecurring, fetchRecurring, runDueRecurring, updateRecurring, type RecurringFrequency, type RecurringInput, type RecurringSchedule } from "../../services/recurring";
import { instantFromLocalDateTime, localDateTimeAt } from "../../services/accountTime";
import { useAccountTimezone } from "../../services/AccountTimezoneContext";
import { formatVND } from "../utils/format";

type Draft = { name: string; walletID: string; categoryID: string; type: "income" | "expense"; amount: string; note: string; frequency: RecurringFrequency; interval: string; nextRunAt: string; endsAt: string; active: boolean };
const frequencies: Array<{ value: RecurringFrequency; label: string }> = [{ value: "daily", label: "Hàng ngày" }, { value: "weekly", label: "Hàng tuần" }, { value: "monthly", label: "Hàng tháng" }, { value: "yearly", label: "Hàng năm" }];

export function RecurringPanel({ refreshKey = 0, onChanged }: { refreshKey?: number; onChanged: () => void }) {
  const timezone = useAccountTimezone();
  const [rows, setRows] = useState<RecurringSchedule[]>([]);
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [editor, setEditor] = useState<RecurringSchedule | "new" | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [runMessage, setRunMessage] = useState("");
  const [running, setRunning] = useState(false);

  const load = () => { setLoading(true); Promise.all([fetchRecurring(), fetchWallets(), fetchCategories()]).then(([nextRows, nextWallets, nextCategories]) => { setRows(nextRows); setWallets(nextWallets.filter(wallet => wallet.type !== "credit")); setCategories(nextCategories); setError(""); }).catch(() => setError("Không thể tải lịch định kỳ.")).finally(() => setLoading(false)); };
  useEffect(() => { load(); }, [refreshKey]);
  const walletNames = useMemo(() => new Map(wallets.map(wallet => [wallet.id, wallet.name])), [wallets]);
  const runDue = async () => { if (running) return; try { setRunning(true); const result = await runDueRecurring(); setRunMessage(`Đã tạo ${result.created} giao dịch đến hạn.`); load(); onChanged(); } catch { setRunMessage("Không thể chạy lịch đến hạn."); } finally { setRunning(false); } };

  if (loading) return <StatusMessage>Đang tải lịch định kỳ...</StatusMessage>;
  return <>
    <SurfaceCard padding="md"><div className="flex items-center justify-between gap-3"><div><Text size="xl" weight="bold">Giao dịch định kỳ</Text><Text size="sm" tone="secondary" className="mt-1">Lịch bật sẽ sinh giao dịch thường, mỗi kỳ chỉ một lần.</Text></div><BaseButton size="sm" onClick={() => setEditor("new")}>Thêm</BaseButton></div><div className="mt-3 flex gap-2"><BaseButton variant="chip" size="sm" loading={running} onClick={() => void runDue()}>Chạy đến hạn</BaseButton>{runMessage ? <Text size="xs" tone="secondary" className="self-center">{runMessage}</Text> : null}</div></SurfaceCard>
    {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
    {rows.length === 0 ? <StatusMessage variant="plain">Chưa có lịch định kỳ.</StatusMessage> : <div className="space-y-2">{rows.map(row => <SurfaceCard key={row.id} padding="md"><div className="flex items-start justify-between gap-3"><div className="min-w-0"><Text weight="bold" className="truncate">{row.name}</Text><Text size="xs" tone="secondary">{row.type === "expense" ? "Khoản chi" : "Khoản thu"} · {frequencyLabel(row.frequency, row.interval)} · {walletNames.get(row.wallet_id) ?? "Ví đã xóa"}</Text><Text size="sm" numeric className="mt-2">{formatVND(row.amount)} · lần tới {new Intl.DateTimeFormat("vi-VN", { dateStyle: "medium", timeStyle: "short", timeZone: timezone }).format(new Date(row.next_run_at))}</Text></div><Text size="xs" tone={row.active ? "action" : "secondary"}>{row.active ? "Đang bật" : "Tạm dừng"}</Text></div><div className="mt-3 flex gap-2"><BaseButton variant="chip" size="sm" onClick={() => setEditor(row)}>Sửa</BaseButton><BaseButton variant="ghost" size="sm" onClick={() => void remove(row)}>Xóa</BaseButton></div></SurfaceCard>)}</div>}
    {editor ? <RecurringSheet schedule={editor === "new" ? undefined : editor} wallets={wallets} categories={categories} timezone={timezone} onClose={() => setEditor(null)} onSaved={() => { setEditor(null); load(); onChanged(); }} /> : null}
  </>;
  async function remove(row: RecurringSchedule) { if (!window.confirm(`Xóa lịch “${row.name}”? Các giao dịch đã tạo vẫn giữ nguyên.`)) return; try { await deleteRecurring(row.id); load(); onChanged(); } catch { setError("Không thể xóa lịch định kỳ."); } }
}

function RecurringSheet({ schedule, wallets, categories, timezone, onClose, onSaved }: { schedule?: RecurringSchedule; wallets: Wallet[]; categories: Category[]; timezone: string; onClose: () => void; onSaved: () => void }) {
  const [draft, setDraft] = useState<Draft>(() => schedule ? { name: schedule.name, walletID: schedule.wallet_id, categoryID: schedule.category_id ?? "", type: schedule.type, amount: String(schedule.amount), note: schedule.note ?? "", frequency: schedule.frequency, interval: String(schedule.interval), nextRunAt: localDateTimeAt(new Date(schedule.next_run_at), timezone), endsAt: schedule.ends_at ? localDateTimeAt(new Date(schedule.ends_at), timezone) : "", active: schedule.active } : { name: "", walletID: wallets[0]?.id ?? "", categoryID: "", type: "expense", amount: "", note: "", frequency: "monthly", interval: "1", nextRunAt: localDateTimeAt(new Date(), timezone), endsAt: "", active: true });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const visibleCategories = categories.filter(category => category.kind === draft.type);
  const save = async () => {
    const amount = Number(draft.amount); const interval = Number(draft.interval);
    if (!draft.name.trim() || !draft.walletID || !Number.isSafeInteger(amount) || amount <= 0 || !Number.isSafeInteger(interval) || interval <= 0) { setError("Nhập tên, ví, số tiền và khoảng lặp hợp lệ."); return; }
    try {
      const nextRunAt = instantFromLocalDateTime(draft.nextRunAt, timezone);
      const endsAt = draft.endsAt ? instantFromLocalDateTime(draft.endsAt, timezone) : null;
      if (endsAt && endsAt < nextRunAt) { setError("Ngày kết thúc phải sau lần chạy kế tiếp."); return; }
      const input: RecurringInput = { name: draft.name.trim(), wallet_id: draft.walletID, category_id: draft.categoryID || null, type: draft.type, amount, note: draft.note.trim() || null, frequency: draft.frequency, interval, next_run_at: nextRunAt, ends_at: endsAt, active: draft.active };
      setSaving(true); setError(""); if (schedule) await updateRecurring(schedule.id, input); else await createRecurring(input); onSaved();
    } catch { setError("Không thể lưu lịch định kỳ."); } finally { setSaving(false); }
  };
  return <BaseBottomSheet presentation="form" title={schedule ? "Sửa lịch định kỳ" : "Thêm lịch định kỳ"} closeLabel="Hủy" closingDisabled={saving} onClose={onClose} footer={<BaseButton className="w-full" size="lg" loading={saving} onClick={() => void save()}>Lưu lịch</BaseButton>}><div className="space-y-3"><StatusMessage variant="plain">Lịch chạy theo múi giờ {timezone}. Lần chạy tạo giao dịch độc lập để bạn sửa/xóa sau.</StatusMessage>{error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}<SurfaceCard tone="form" padding="md"><FormField label="Tên lịch"><BaseTextInput aria-label="Tên lịch" value={draft.name} disabled={saving} onChange={event => setDraft(current => ({ ...current, name: event.target.value }))} /></FormField><Divider /><FormField label="Loại"><BaseSelect value={draft.type} disabled={saving} onChange={event => setDraft(current => ({ ...current, type: event.target.value as Draft["type"], categoryID: "" }))}><option value="expense">Khoản chi</option><option value="income">Khoản thu</option></BaseSelect></FormField><Divider /><FormField label="Ví"><BaseSelect value={draft.walletID} disabled={saving} onChange={event => setDraft(current => ({ ...current, walletID: event.target.value }))}>{wallets.map(wallet => <option key={wallet.id} value={wallet.id}>{wallet.name}</option>)}</BaseSelect></FormField><Divider /><AmountField value={draft.amount} disabled={saving} onChange={value => setDraft(current => ({ ...current, amount: value }))} /><Divider /><FormField label="Nhóm (tùy chọn)"><BaseSelect value={draft.categoryID} disabled={saving} onChange={event => setDraft(current => ({ ...current, categoryID: event.target.value }))}><option value="">Không chọn nhóm</option>{visibleCategories.map(category => <option key={category.id} value={category.id}>{category.name}</option>)}</BaseSelect></FormField><Divider /><FormField label="Tần suất"><BaseSelect value={draft.frequency} disabled={saving} onChange={event => setDraft(current => ({ ...current, frequency: event.target.value as RecurringFrequency }))}>{frequencies.map(option => <option key={option.value} value={option.value}>{option.label}</option>)}</BaseSelect></FormField><div className="grid grid-cols-2 gap-3"><FormField label="Mỗi"><BaseTextInput type="number" min="1" inputMode="numeric" value={draft.interval} disabled={saving} onChange={event => setDraft(current => ({ ...current, interval: event.target.value.replace(/\D/g, "") }))} /></FormField><FormField label="Trạng thái"><BaseSelect value={draft.active ? "active" : "paused"} disabled={saving} onChange={event => setDraft(current => ({ ...current, active: event.target.value === "active" }))}><option value="active">Đang bật</option><option value="paused">Tạm dừng</option></BaseSelect></FormField></div><Divider /><FormField label="Lần chạy kế tiếp"><BaseTextInput type="datetime-local" value={draft.nextRunAt} disabled={saving} onChange={event => setDraft(current => ({ ...current, nextRunAt: event.target.value }))} /></FormField><FormField label="Kết thúc (tùy chọn)"><BaseTextInput type="datetime-local" value={draft.endsAt} disabled={saving} onChange={event => setDraft(current => ({ ...current, endsAt: event.target.value }))} /></FormField><Divider /><FormField label="Ghi chú"><BaseTextInput variant="inline" value={draft.note} disabled={saving} onChange={event => setDraft(current => ({ ...current, note: event.target.value }))} /></FormField></SurfaceCard></div></BaseBottomSheet>;
}

function frequencyLabel(frequency: RecurringFrequency, interval: number): string { const label = frequencies.find(item => item.value === frequency)?.label ?? frequency; return interval === 1 ? label : `Mỗi ${interval} ${label.toLowerCase().replace("hàng ", "")}`; }
