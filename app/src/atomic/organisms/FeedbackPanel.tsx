import { useEffect, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseSelect, BaseTextArea, BaseTextInput, FormField } from "../atoms/FormField";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { createFeedback, fetchChangelog, fetchFeedback, type Changelog, type Feedback, type FeedbackType } from "../../services/feedback";
import { FeedbackStatus } from "../molecules/FeedbackStatus";

export function FeedbackPanel() {
  const [items, setItems] = useState<Feedback[]>([]);
  const [changelog, setChangelog] = useState<Changelog[]>([]);
  const [type, setType] = useState<FeedbackType>("bug");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const load = async () => {
    setLoading(true); setError("");
    try { const [feedback, releases] = await Promise.all([fetchFeedback(), fetchChangelog()]); setItems(feedback); setChangelog(releases); }
    catch (reason) { setError(reason instanceof Error ? reason.message : "Không thể tải phản hồi."); }
    finally { setLoading(false); }
  };

  useEffect(() => { void load(); }, []);

  const submit = async () => {
    if (saving || !title.trim() || !description.trim()) return;
    setSaving(true); setMessage(""); setError("");
    try { await createFeedback({ type, title, description }); setTitle(""); setDescription(""); setMessage("Đã gửi phản hồi. Cảm ơn bạn đã giúp cải thiện MyPocket."); await load(); }
    catch (reason) { setError(reason instanceof Error ? reason.message : "Không thể gửi phản hồi."); }
    finally { setSaving(false); }
  };

  const versionByID = new Map(changelog.map(entry => [entry.id, entry.version]));
  return <div className="space-y-4">
    <Text as="h1" size="2xl" weight="bold">Phản hồi</Text>
    <SurfaceCard padding="md" className="space-y-3">
      <Text weight="semibold">Báo lỗi hoặc góp ý</Text>
      <FormField label="Loại"><BaseSelect value={type} onChange={event => setType(event.target.value as FeedbackType)}><option value="bug">Báo lỗi</option><option value="feature">Tính năng</option><option value="improvement">Cải thiện</option></BaseSelect></FormField>
      <FormField label="Tiêu đề"><BaseTextInput value={title} maxLength={200} onChange={event => setTitle(event.target.value)} placeholder="Ví dụ: Số dư chưa cập nhật" /></FormField>
      <FormField label="Mô tả"><BaseTextArea rows={5} maxLength={10000} value={description} onChange={event => setDescription(event.target.value)} placeholder="Mô tả điều đã xảy ra hoặc điều bạn muốn cải thiện" /></FormField>
      {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
      {message ? <StatusMessage>{message}</StatusMessage> : null}
      <BaseButton type="button" className="w-full" loading={saving} disabled={saving || !title.trim() || !description.trim()} onClick={() => void submit()}>Gửi phản hồi</BaseButton>
    </SurfaceCard>
    <Text weight="semibold">Phản hồi của bạn</Text>
    {loading ? <StatusMessage>Đang tải...</StatusMessage> : null}
    {!loading && !items.length ? <StatusMessage>Chưa có phản hồi nào.</StatusMessage> : null}
    {items.map(item => <SurfaceCard key={item.id} padding="md" className="space-y-2"><div className="flex items-start justify-between gap-3"><Text weight="semibold">{item.title}</Text><FeedbackStatus status={item.status} changelogVersion={item.changelog_id ? versionByID.get(item.changelog_id) : undefined} /></div><Text size="sm" tone="secondary">{item.description}</Text><Text size="xs" tone="secondary">{new Date(item.created_at).toLocaleDateString("vi-VN")}</Text></SurfaceCard>)}
  </div>;
}
