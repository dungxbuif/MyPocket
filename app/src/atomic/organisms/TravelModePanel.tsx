import { useEffect, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseTextInput, FormField } from "../atoms/FormField";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
import { Divider } from "../atoms/Divider";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { createTravelEvent, deleteTravelEvent, fetchTravelEvents, setTravelEventActive, updateTravelEvent, type TravelEvent, type TravelEventInput } from "../../services/travel";

export function TravelModePanel({ refreshKey = 0 }: { refreshKey?: number }) {
  const [events, setEvents] = useState<TravelEvent[]>([]);
  const [editor, setEditor] = useState<TravelEvent | "new" | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const load = () => {
    setLoading(true);
    fetchTravelEvents().then(value => { setEvents(value); setError(""); }).catch(() => setError("Không thể tải danh sách chuyến.")).finally(() => setLoading(false));
  };
  useEffect(() => { load(); }, [refreshKey]);

  const toggle = async (event: TravelEvent) => {
    try { await setTravelEventActive(event.id, !event.active); load(); } catch { setError("Không thể đổi trạng thái Travel Mode."); }
  };
  const remove = async (event: TravelEvent) => {
    if (!window.confirm(`Xóa chuyến “${event.name}”? Giao dịch vẫn giữ nguyên, chỉ gỡ liên kết chuyến.`)) return;
    try { await deleteTravelEvent(event.id); load(); } catch { setError("Không thể xóa chuyến."); }
  };

  if (loading) return <StatusMessage>Đang tải Travel Mode...</StatusMessage>;
  return <>
    <SurfaceCard padding="md">
      <div className="flex items-start justify-between gap-3">
        <div><Text size="xl" weight="bold">Travel Mode</Text><Text size="sm" tone="secondary" className="mt-1">Gắn giao dịch thu/chi vào một chuyến để xem bối cảnh. Chuyến không phải là ví và không đổi số dư.</Text></div>
        <BaseButton size="sm" onClick={() => setEditor("new")}>Thêm</BaseButton>
      </div>
      <div className="mt-3">{events.some(event => event.active) ? <StatusMessage>Đang bật: {events.find(event => event.active)?.name}. Giao dịch thường mới sẽ tự gắn chuyến này.</StatusMessage> : <StatusMessage variant="plain">Chưa bật chuyến nào. Giao dịch mới sẽ không tự gắn Travel Mode.</StatusMessage>}</div>
    </SurfaceCard>
    {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
    {events.length === 0 ? <StatusMessage variant="plain">Chưa có chuyến. Tạo chuyến đầu tiên để bắt đầu theo dõi.</StatusMessage> : <div className="space-y-2">{events.map(event => <SurfaceCard key={event.id} padding="md"><div className="flex items-start justify-between gap-3"><div className="min-w-0"><Text weight="bold" className="truncate">{event.name}</Text><Text size="sm" tone="secondary" className="mt-1">{[event.starts_on, event.ends_on].filter(Boolean).join(" → ") || "Không giới hạn ngày"}</Text>{event.context ? <Text size="sm" tone="secondary" className="mt-1">{event.context}</Text> : null}</div><Text size="xs" tone={event.active ? "action" : "secondary"}>{event.active ? "Đang bật" : "Tắt"}</Text></div><div className="mt-3 flex flex-wrap gap-2"><BaseButton variant={event.active ? "chip" : "secondary"} size="sm" onClick={() => void toggle(event)}>{event.active ? "Tắt" : "Bật"}</BaseButton><BaseButton variant="ghost" size="sm" onClick={() => setEditor(event)}>Sửa</BaseButton><BaseButton variant="ghost" size="sm" onClick={() => void remove(event)}>Xóa</BaseButton></div></SurfaceCard>)}</div>}
    {editor ? <TravelEventSheet event={editor === "new" ? undefined : editor} onClose={() => setEditor(null)} onSaved={() => { setEditor(null); load(); }} /> : null}
  </>;
}

function TravelEventSheet({ event, onClose, onSaved }: { event?: TravelEvent; onClose: () => void; onSaved: () => void }) {
  const [draft, setDraft] = useState<TravelEventInput>(() => event ? { name: event.name, context: event.context ?? "", starts_on: event.starts_on ?? "", ends_on: event.ends_on ?? "" } : { name: "", context: "", starts_on: "", ends_on: "", active: false });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const save = async () => {
    if (!draft.name?.trim()) { setError("Tên chuyến là bắt buộc."); return; }
    if (draft.starts_on && draft.ends_on && draft.ends_on < draft.starts_on) { setError("Ngày kết thúc không được trước ngày bắt đầu."); return; }
    setSaving(true); setError("");
    try { const input: TravelEventInput = { name: draft.name.trim(), context: draft.context?.trim() || null, starts_on: draft.starts_on || null, ends_on: draft.ends_on || null }; if (event) await updateTravelEvent(event.id, input); else await createTravelEvent(input); onSaved(); } catch (reason) { setError(reason instanceof Error ? reason.message : "Không thể lưu chuyến."); } finally { setSaving(false); }
  };
  return <BaseBottomSheet presentation="form" title={event ? "Sửa chuyến" : "Thêm chuyến"} closeLabel="Hủy" closingDisabled={saving} onClose={onClose} footer={<BaseButton className="w-full" size="lg" loading={saving} onClick={() => void save()}>Lưu chuyến</BaseButton>}><div className="space-y-3"><StatusMessage variant="plain">Chỉ giao dịch thu/chi tạo mới mới tự nhận chuyến đang bật. Giao dịch recurring, chuyển ví và điều chỉnh không tự gắn.</StatusMessage>{error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}<SurfaceCard tone="form" padding="md"><FormField label="Tên chuyến"><BaseTextInput value={draft.name} disabled={saving} onChange={event => setDraft(current => ({ ...current, name: event.target.value }))} /></FormField><Divider /><FormField label="Bối cảnh (tùy chọn)"><BaseTextInput variant="inline" value={draft.context ?? ""} disabled={saving} onChange={event => setDraft(current => ({ ...current, context: event.target.value }))} /></FormField><Divider /><div className="grid grid-cols-2 gap-3"><FormField label="Bắt đầu"><BaseTextInput type="date" value={draft.starts_on ?? ""} disabled={saving} onChange={event => setDraft(current => ({ ...current, starts_on: event.target.value }))} /></FormField><FormField label="Kết thúc"><BaseTextInput type="date" value={draft.ends_on ?? ""} disabled={saving} onChange={event => setDraft(current => ({ ...current, ends_on: event.target.value }))} /></FormField></div></SurfaceCard></div></BaseBottomSheet>;
}
