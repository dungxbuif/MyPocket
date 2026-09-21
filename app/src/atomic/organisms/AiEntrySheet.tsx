import { useEffect, useRef, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseFileUpload } from "../atoms/BaseFileUpload";
import { BaseTextInput, FormField } from "../atoms/FormField";
import { Heading } from "../atoms/Heading";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
import { EntryProposalRow } from "../molecules/EntryProposalRow";
import { createEntrySession, fetchEntryCapabilities, fetchLatestEntrySession, sendEntryMessage, type EntryCapabilities, type EntrySession } from "../../services/ai";
import { encodeEntryImages, validateImages } from "../../services/aiEntryLogic";
import { fetchCategories, type Category } from "../../services/categories";
import { fetchWallets, type Wallet } from "../../services/wallets";

const COPY = {
  title: "Nhập giao dịch bằng AI", close: "Đóng", reload: "Tải lại", loading: "Đang tải...",
  unavailable: "AI chưa được cấu hình. Bạn vẫn có thể xem và xử lý đề xuất đã lưu hoặc nhập giao dịch thủ công.",
  noOcr: "OCR chưa được cấu hình. Hiện chỉ gửi được nội dung chữ.", processing: "Phiên đang xử lý. Tải lại phiên để kiểm tra kết quả; không gửi lại nội dung.",
  empty: "Nhập mô tả hoặc đính kèm chứng từ để tạo danh sách giao dịch. Bạn sửa và duyệt từng dòng.",
  proposals: "Danh sách giao dịch", text: "Nhập giao dịch", images: "Đính kèm tệp (tối đa 3)",
  transient: "JPEG, PNG hoặc PDF; tối đa 5 MiB mỗi tệp.",
  clear: "Bỏ ảnh đã chọn", send: "Gửi nội dung", sending: "Đang xử lý...", failed: "Không thể thực hiện yêu cầu.",
  ambiguous: "Yêu cầu có thể đã được tiếp nhận. Tải lại phiên để kiểm tra trước khi gửi nội dung khác. Không tự động gửi lại.",
  reloadConfirm: "Tải lại phiên sẽ bỏ thay đổi chưa lưu trong các đề xuất. Tiếp tục?",
};

export function AiEntrySheet({ onClose, onSaved }: { onClose: () => void; onSaved: () => void }) {
  const [session, setSession] = useState<EntrySession | null>(null);
  const [capabilities, setCapabilities] = useState<EntryCapabilities | null>(null);
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [creating, setCreating] = useState(false);
  const [ready, setReady] = useState(false);
  const [uncertain, setUncertain] = useState(false);
  const [error, setError] = useState("");
  const [text, setText] = useState("");
  const [files, setFiles] = useState<File[]>([]);
  const [revision, setRevision] = useState(0);
  const requestActive = useRef(false);
  const mounted = useRef(false);

  useEffect(() => {
    mounted.current = true;
    return () => { mounted.current = false; };
  }, []);

  useEffect(() => {
    let active = true;
    setLoading(true); setReady(false); setError("");
    Promise.all([fetchEntryCapabilities(), fetchLatestEntrySession(), fetchWallets(), fetchCategories()])
      .then(([nextCapabilities, nextSession, nextWallets, nextCategories]) => {
        if (!active) return;
        setCapabilities(nextCapabilities); setSession(nextSession); setWallets(nextWallets); setCategories(nextCategories); setReady(true);
      }).catch(reason => { if (active) setError(reason instanceof Error ? reason.message : COPY.failed); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [revision]);

  const send = async () => {
    if (requestActive.current || busy || loading || !ready || uncertain || session?.processing || !capabilities?.ai_configured || (!text.trim() && !files.length)) return;
    if (files.length && !capabilities.ocr_configured) return;
    requestActive.current = true; setBusy(true); setError("");
    let submitted = false;
    try {
      const images = await encodeEntryImages(files);
      const current = session ?? await createEntrySession();
      if (!mounted.current) return;
      setSession(current);
      submitted = true;
      const next = await sendEntryMessage(current.id, {
        request_id: crypto.randomUUID(), text: text.trim(), timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
        ...(images.length ? { images } : {}),
      });
      if (mounted.current) { setSession(next); setText(""); setFiles([]); }
    } catch (reason) {
      if (mounted.current) { setError(reason instanceof Error ? reason.message : COPY.failed); if (submitted) setUncertain(true); }
    } finally { requestActive.current = false; if (mounted.current) setBusy(false); }
  };

  const newConversation = async () => {
    if (requestActive.current || busy || loading || !ready || session?.processing) return;
    if (!window.confirm(COPY.newConversationConfirm)) return;
    requestActive.current = true; setBusy(true); setCreating(true); setError("");
    try {
      const next = await createEntrySession();
      if (!mounted.current) return;
      setSession(next); setText(""); setFiles([]); setUncertain(false);
    } catch (reason) {
      if (mounted.current) setError(reason instanceof Error ? reason.message : COPY.failed);
    } finally {
      requestActive.current = false;
      if (mounted.current) { setBusy(false); setCreating(false); }
    }
  };

  const blocked = busy || loading || !ready || !!session?.processing;
  return <BaseBottomSheet presentation="form" title={COPY.title} closeLabel={COPY.close} onClose={() => { if (!busy) onClose(); }} closingDisabled={busy}>
    <div className="space-y-4">
      <BaseButton variant="secondary" disabled={loading || busy} onClick={() => {
        if (requestActive.current) return;
        if (!window.confirm(COPY.reloadConfirm)) return;
        setRevision(value => value + 1);
      }}>{COPY.reload}</BaseButton>
      {loading ? <StatusMessage>{COPY.loading}</StatusMessage> : null}
      {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
      {session?.error ? <StatusMessage tone="danger">{session.error}</StatusMessage> : null}
      {capabilities && !capabilities.ai_configured ? <StatusMessage>{COPY.unavailable}</StatusMessage> : null}
      {capabilities && !capabilities.ocr_configured ? <StatusMessage>{COPY.noOcr}</StatusMessage> : null}
      {session?.processing ? <StatusMessage>{COPY.processing}</StatusMessage> : null}
      {uncertain ? <StatusMessage>{COPY.ambiguous}</StatusMessage> : null}
      {!loading && !session?.messages?.length ? <StatusMessage variant="plain">{COPY.empty}</StatusMessage> : null}
      {(session?.proposals?.length ?? 0) > 0 ? <Heading as="h3" size="field">{COPY.proposals}</Heading> : null}
      {(session?.proposals ?? []).map(proposal => <EntryProposalRow key={`${revision}:${proposal.id}:${proposal.version}:${proposal.status}`}
        proposal={proposal} wallets={wallets} categories={categories} disabled={blocked} onBusy={value => { requestActive.current = value; setBusy(value); }} onSaved={onSaved}
        onUpdated={updated => setSession(current => current ? { ...current, proposals: current.proposals.map(item => item.id === updated.id ? updated : item) } : current)} />)}
      <form className="sticky bottom-0 space-y-3 border-t border-line bg-card pt-3" onSubmit={event => { event.preventDefault(); void send(); }}>
        <FormField label={COPY.text}><BaseTextInput value={text} disabled={blocked || !capabilities?.ai_configured || uncertain} onChange={event => setText(event.target.value)} /></FormField>
        <BaseFileUpload label={COPY.images} disabled={blocked || !capabilities?.ai_configured || !capabilities?.ocr_configured || uncertain} onFiles={selected => {
          const issue = validateImages(selected);
          if (issue) { setError(issue); return; }
          setFiles(selected); setError("");
        }} />
        <Text size="sm" tone="secondary">{COPY.transient}</Text>
        {files.length ? <div className="space-y-2"><Text className="break-words">{files.map(file => file.name).join(", ")}</Text><BaseButton variant="secondary" disabled={busy} onClick={() => setFiles([])}>{COPY.clear}</BaseButton></div> : null}
        <BaseButton type="submit" loading={busy} loadingLabel={COPY.sending} disabled={blocked || uncertain || !capabilities?.ai_configured || (!text.trim() && !files.length)}>{COPY.send}</BaseButton>
      </form>
      {uncertain && ready && !session?.processing ? <BaseButton variant="secondary" disabled={busy || loading} onClick={() => {
        setText(""); setFiles([]); setUncertain(false);
      }}>Đã kiểm tra phiên — soạn nội dung mới</BaseButton> : null}
    </div>
  </BaseBottomSheet>;
}
