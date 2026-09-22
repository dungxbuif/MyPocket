import { useState } from "react";
import html2canvas from "html2canvas";
import { MessageSquarePlus, X } from "lucide-react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseModal } from "../atoms/BaseModal";
import { BaseSelect, BaseTextArea, BaseTextInput, FormField } from "../atoms/FormField";
import { IconButton } from "../atoms/IconButton";
import { StatusMessage } from "../atoms/StatusMessage";
import { Text } from "../atoms/Text";
import { createFeedback, type FeedbackType } from "../../services/feedback";

async function captureCurrentScreen(): Promise<Blob | null> {
  try {
    const canvas = await html2canvas(document.documentElement, {
      backgroundColor: getComputedStyle(document.documentElement).getPropertyValue("--color-canvas").trim() || undefined,
      scale: Math.min(window.devicePixelRatio || 1, 2),
      ignoreElements: element => element.getAttribute("data-feedback-overlay") === "true",
    });
    return await new Promise(resolve => canvas.toBlob(resolve, "image/png"));
  } catch {
    return null;
  }
}

export function FeedbackFloatingBubble() {
  const [open, setOpen] = useState(false);
  const [type, setType] = useState<FeedbackType>("bug");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [screenshot, setScreenshot] = useState<Blob | null>(null);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  const openFeedback = () => {
    setMessage("");
    setError("");
    setOpen(true);
    void captureCurrentScreen().then(setScreenshot);
  };

  const closeFeedback = () => {
    if (saving) return;
    setOpen(false);
    setScreenshot(null);
  };

  const submit = async () => {
    if (saving || !title.trim() || !description.trim()) return;
    setSaving(true);
    setError("");
    setMessage("");
    try {
      await createFeedback({ type, title, description, screenshot });
      setTitle("");
      setDescription("");
      setScreenshot(null);
      setMessage("Đã gửi phản hồi kèm hiện trạng màn hình.");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Không thể gửi phản hồi.");
    } finally {
      setSaving(false);
    }
  };

  return <>
    <div className="fixed bottom-24 right-4 z-40" data-feedback-overlay="true">
      <IconButton label="Gửi phản hồi" variant="brand" onClick={openFeedback}>
        <MessageSquarePlus size={20} aria-hidden />
      </IconButton>
    </div>
    {open ? <BaseModal label="Gửi phản hồi" onClose={closeFeedback}>
      <div data-feedback-overlay="true" className="space-y-4">
        <div className="flex items-start justify-between gap-3">
          <div><Text as="h2" size="xl" weight="bold">Gửi phản hồi</Text><Text size="sm" tone="secondary">Chúng tôi đã âm thầm đính kèm ảnh hiện trạng màn hình.</Text></div>
          <IconButton label="Đóng" variant="bare" onClick={closeFeedback}><X size={20} aria-hidden /></IconButton>
        </div>
        <FormField label="Loại"><BaseSelect value={type} onChange={event => setType(event.target.value as FeedbackType)}><option value="bug">Báo lỗi</option><option value="feature">Tính năng</option><option value="improvement">Cải thiện</option></BaseSelect></FormField>
        <FormField label="Tiêu đề"><BaseTextInput value={title} maxLength={200} onChange={event => setTitle(event.target.value)} placeholder="Ví dụ: Số dư chưa cập nhật" /></FormField>
        <FormField label="Mô tả"><BaseTextArea rows={5} maxLength={10000} value={description} onChange={event => setDescription(event.target.value)} placeholder="Mô tả điều đã xảy ra hoặc điều bạn muốn cải thiện" /></FormField>
        {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
        {message ? <StatusMessage>{message}</StatusMessage> : null}
        <BaseButton type="button" className="w-full" loading={saving} disabled={saving || !title.trim() || !description.trim()} onClick={() => void submit()}>Gửi phản hồi</BaseButton>
      </div>
    </BaseModal> : null}
  </>;
}
