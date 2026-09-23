import { FileText, LoaderCircle, Send, X } from "lucide-react";
import { useEffect, useMemo } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseFileUpload } from "../atoms/BaseFileUpload";
import { BaseTextArea } from "../atoms/FormField";
import { IconButton } from "../atoms/IconButton";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { MAX_AI_ENTRY_FILES } from "../../services/aiEntryLogic";

export function AssistantComposer({ text, files, disabled, uploadDisabled, loading, canSend, hint, mode = "entry", onTextChange, onFiles, onClearFiles, onSubmit }: {
  text: string; files: File[]; disabled: boolean; uploadDisabled: boolean; loading: boolean; canSend: boolean; hint: string;
  mode?: "entry" | "advisor";
  onTextChange: (value: string) => void; onFiles: (files: File[]) => void; onClearFiles: () => void; onSubmit: () => void;
}) {
  const previews = useMemo(() => files.map(file => {
    const isImage = file.type.startsWith("image/") || /\.(png|jpe?g|webp)$/i.test(file.name);
    if (!isImage || typeof Blob === "undefined" || !(file instanceof Blob) || typeof URL === "undefined" || typeof URL.createObjectURL !== "function") return { file, url: "", isImage };
    return { file, url: URL.createObjectURL(file), isImage };
  }), [files]);
  useEffect(() => () => {
    previews.forEach(preview => { if (preview.url) URL.revokeObjectURL(preview.url); });
  }, [previews]);
  return <form className="sticky bottom-0" onSubmit={event => { event.preventDefault(); onSubmit(); }}>
    <SurfaceCard tone="form" padding="md" elevation="raised" className="space-y-2">
      {loading ? <div data-processing-state role="status" aria-live="polite" className="flex items-center gap-2 rounded-lg border border-line bg-row px-3 py-2 text-sm text-secondary">
        <LoaderCircle aria-hidden size={16} className="animate-spin text-action" />
        <span>{mode === "advisor" ? "Đang phân tích..." : "Đang đọc chứng từ và phân tích giao dịch..."}</span>
      </div> : null}
      <BaseTextArea variant="composer" aria-label={mode === "advisor" ? "Câu hỏi tài chính" : "Mô tả giao dịch"} placeholder={mode === "advisor" ? "Hỏi về chi tiêu, ví, ngân sách..." : "Ví dụ: Ăn sáng 45.000đ và cà phê 30.000đ hôm qua"} rows={3} value={text} disabled={disabled} onChange={event => onTextChange(event.target.value)} />
      {files.length ? <div aria-label="Tệp đính kèm" className="flex items-end gap-2 overflow-x-auto pb-1">
        <div className="flex min-w-0 flex-1 gap-2">
          {previews.map(({ file, url, isImage }, index) => <div key={`${file.name}-${file.lastModified ?? index}`} data-file-thumbnail={file.name} className={`relative h-16 w-16 shrink-0 overflow-hidden rounded-lg border border-line bg-row ${loading ? "opacity-60" : ""}`}>
            {isImage && url ? <img src={url} alt={file.name} className="h-full w-full object-cover" /> : <div data-thumbnail-kind="pdf" className="grid h-full w-full place-items-center text-danger"><FileText aria-hidden size={22} /></div>}
            <span className="pointer-events-none absolute inset-x-1 bottom-1 truncate rounded bg-black/60 px-1 text-[9px] text-white">{file.name}</span>
            <span className="absolute right-0.5 top-0.5">
              <IconButton label={`Bỏ tệp ${file.name}`} variant="surface" shape="circle" disabled={disabled} onClick={() => onFiles(files.filter((_, fileIndex) => fileIndex !== index))}>
                <X aria-hidden size={13} />
              </IconButton>
            </span>
          </div>)}
        </div>
        <BaseButton size="sm" variant="ghost" disabled={disabled} onClick={onClearFiles}>Bỏ tất cả</BaseButton>
      </div> : null}
      <Text size="xs" tone="secondary">{hint}</Text>
      <div className="flex items-center justify-between gap-2">
      {mode === "entry" ? <BaseFileUpload variant="button" maxFiles={MAX_AI_ENTRY_FILES} label="Đính kèm ảnh hoặc PDF" disabled={disabled || uploadDisabled} onFiles={selected => {
        onFiles([...files, ...selected]);
      }} /> : <span />}
        <BaseButton type="submit" size="sm" loading={loading} loadingLabel={mode === "advisor" ? "Đang phân tích..." : "Đang xử lý..."} disabled={disabled || !canSend} aria-label={mode === "advisor" ? "Gửi câu hỏi tài chính" : "Gửi mô tả giao dịch"}><Send aria-hidden size={17} />Gửi</BaseButton>
      </div>
    </SurfaceCard>
  </form>;
}
