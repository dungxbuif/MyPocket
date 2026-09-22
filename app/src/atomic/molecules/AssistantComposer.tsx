import { Send } from "lucide-react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseFileUpload } from "../atoms/BaseFileUpload";
import { BaseTextArea } from "../atoms/FormField";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";

export function AssistantComposer({ text, files, disabled, uploadDisabled, loading, canSend, hint, onTextChange, onFiles, onClearFiles, onSubmit }: {
  text: string; files: File[]; disabled: boolean; uploadDisabled: boolean; loading: boolean; canSend: boolean; hint: string;
  onTextChange: (value: string) => void; onFiles: (files: File[]) => void; onClearFiles: () => void; onSubmit: () => void;
}) {
  return <form className="sticky bottom-0" onSubmit={event => { event.preventDefault(); onSubmit(); }}>
    <SurfaceCard tone="form" padding="md" elevation="raised" className="space-y-2">
      <BaseTextArea variant="composer" aria-label="Mô tả giao dịch" placeholder="Ví dụ: Ăn sáng 45.000đ và cà phê 30.000đ hôm qua" rows={3} value={text} disabled={disabled} onChange={event => onTextChange(event.target.value)} />
      {files.length ? <div className="flex flex-wrap items-center gap-2"><Text size="xs" tone="secondary" className="min-w-0 flex-1 break-all">{files.map(file => file.name).join(", ")}</Text><BaseButton size="sm" variant="ghost" disabled={disabled} onClick={onClearFiles}>Bỏ tệp</BaseButton></div> : null}
      <Text size="xs" tone="secondary">{hint}</Text>
      <div className="flex items-center justify-between gap-2">
        <BaseFileUpload variant="button" label="Đính kèm ảnh hoặc PDF" disabled={disabled || uploadDisabled} onFiles={onFiles} />
        <BaseButton type="submit" size="sm" loading={loading} loadingLabel="Đang xử lý..." disabled={disabled || !canSend} aria-label="Gửi mô tả giao dịch"><Send aria-hidden size={17} />Gửi</BaseButton>
      </div>
    </SurfaceCard>
  </form>;
}
