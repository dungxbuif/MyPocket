import { useEffect, useRef, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { StatusMessage } from "../atoms/StatusMessage";
import { AssistantComposer } from "../molecules/AssistantComposer";
import { AssistantResultCard } from "../molecules/AssistantResultCard";
import { BaseBottomSheet } from "../molecules/BaseBottomSheet";
import { EntryProposalRow } from "../molecules/EntryProposalRow";
import { saveEntryProposal, decideEntryProposal, type EntryDraft, fetchEntryCapabilities, fetchEntryRequest, processEntry, type EntryCapabilities, type EntryProcess } from "../../services/ai";
import { proposalIssues, sameDraft, validateEntryFiles } from "../../services/aiEntryLogic";
import { fetchCategories, type Category } from "../../services/categories";
import { fetchWallets, type Wallet } from "../../services/wallets";
import { useAccountTimezone } from "../../services/AccountTimezoneContext";

const COPY = {
  title: "Nhập giao dịch bằng AI", close: "Đóng", reload: "Tải lại kết quả", loading: "Đang tải...",
  unavailable: "AI chưa được cấu hình. Bạn vẫn có thể xem và xử lý đề xuất hoặc nhập giao dịch thủ công.",
  noOcr: "OCR chưa được cấu hình. Hiện chỉ gửi được nội dung chữ.", processing: "Yêu cầu đang xử lý. Tải lại kết quả để kiểm tra; không gửi lại nội dung.",
  noStorage: "Lưu trữ chứng từ chưa sẵn sàng. Bạn vẫn có thể nhập nội dung chữ.",
  empty: "Nhập mô tả hoặc đính kèm chứng từ để tạo danh sách giao dịch. Bạn sửa và duyệt từng dòng.",
  transient: "JPEG, PNG hoặc PDF; tối đa 5 MiB mỗi tệp.", failed: "Không thể thực hiện yêu cầu.",
  ambiguous: "Yêu cầu có thể đã được tiếp nhận. Tải lại kết quả để kiểm tra trước khi gửi nội dung khác. Không tự động gửi lại.",
  reloadConfirm: "Tải lại kết quả sẽ bỏ thay đổi chưa lưu trong các đề xuất. Tiếp tục?",
};

export function AiEntrySheet({ onClose, onSaved }: { onClose: () => void; onSaved: () => void }) {
  const timezone=useAccountTimezone();
  const [edits, setEdits] = useState<Record<string, EntryDraft>>({});
  const [process, setProcess] = useState<EntryProcess | null>(null);
  const [capabilities, setCapabilities] = useState<EntryCapabilities | null>(null);
  const [wallets, setWallets] = useState<Wallet[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);

  const [ready, setReady] = useState(false);
  const [uncertain, setUncertain] = useState(false);
  const [error, setError] = useState("");
  const [text, setText] = useState("");
  const [files, setFiles] = useState<File[]>([]);
  const requestActive = useRef(false);
  const requestID = useRef<string | null>(null);
  const mounted = useRef(false);

  useEffect(() => {
    mounted.current = true;
    return () => { mounted.current = false; };
  }, []);

  useEffect(() => {
    let active = true;
    setLoading(true); setReady(false); setError("");
    Promise.all([fetchEntryCapabilities(), fetchWallets(), fetchCategories()])
      .then(([nextCapabilities, nextWallets, nextCategories]) => {
        if (!active) return;
        setCapabilities(nextCapabilities); setWallets(nextWallets); setCategories(nextCategories); setReady(true);
      }).catch(reason => { if (active) setError(reason instanceof Error ? reason.message : COPY.failed); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, []);

  const refreshResult = async () => {
    const id = requestID.current;
    if (!id || requestActive.current || busy) return;
    requestActive.current = true; setBusy(true); setError("");
    try {
      const next = await fetchEntryRequest(id);
      if (!mounted.current) return;
      setProcess(next);
      if (!next.processing) {
        setUncertain(false); setText(""); setFiles([]);
      }
    } catch (reason) {
      if (!mounted.current) return;
      const status = (reason as { status?: number })?.status;
      if (status === 404) {
        requestID.current = null; setUncertain(false); setProcess(null);
        setError("Chưa tìm thấy yêu cầu đã tiếp nhận. Bạn có thể gửi lại nội dung này.");
      } else setError(reason instanceof Error ? reason.message : COPY.failed);
    } finally { requestActive.current = false; if (mounted.current) setBusy(false); }
  };

  const send = async () => {
    if (requestActive.current || busy || loading || !ready || uncertain || process?.processing || !capabilities?.ai_configured || (!text.trim() && !files.length)) return;
    if (files.length && !capabilities.files_configured) return;
    requestActive.current = true; setBusy(true); setError("");
    let submitted = false;
    try {
      const id = crypto.randomUUID();
      requestID.current = id;
      submitted = true;
      const next = await processEntry({
        request_id: id, text: text.trim(), timezone, files,
      });
      if (mounted.current) { setProcess(next); setUncertain(false); setText(""); setFiles([]); }
    } catch (reason) {
      if (mounted.current) {
        setError(reason instanceof Error ? reason.message : COPY.failed);
        const status = (reason as { status?: number })?.status;
        if (submitted && (!status || status >= 500)) setUncertain(true);
        else requestID.current = null;
      }
    } finally { requestActive.current = false; if (mounted.current) setBusy(false); }
  };

  const saveAll = async () => {
    if (requestActive.current || !process) return;
    const pending = process.proposals.filter(item => item.status === "pending");
    if (pending.some(item => proposalIssues(edits[item.id] ?? item.draft, wallets, categories).length)) {
      setError("Sửa các giao dịch còn thiếu thông tin trước khi lưu tất cả."); return;
    }
    requestActive.current = true; setBusy(true); setError("");
    try {
      for (const item of pending) {
        const draft = edits[item.id] ?? item.draft;
        const saved = sameDraft(draft, item.draft) ? item : await saveEntryProposal(item, draft);
        setProcess(current => current ? { ...current, proposals: current.proposals.map(p => p.id === saved.id ? saved : p) } : current);
        const approved = await decideEntryProposal(saved, "approve");
        setProcess(current => current ? { ...current, proposals: current.proposals.map(p => p.id === approved.id ? approved : p) } : current);
        onSaved();
      }
    } catch (reason) { setError(reason instanceof Error ? reason.message : COPY.failed); }
    finally { requestActive.current = false; setBusy(false); }
  };
  const blocked = busy || loading || !ready || !!process?.processing;
  return <BaseBottomSheet presentation="form" title={COPY.title} closeLabel={COPY.close} onClose={() => { if (!busy) onClose(); }} closingDisabled={busy}>
    <div className="space-y-4">
      {(uncertain || process?.processing) ? <BaseButton variant="secondary" disabled={loading || busy} onClick={() => {
        if (requestActive.current) return;
        if (!window.confirm(COPY.reloadConfirm)) return;
        void refreshResult();
      }}>{COPY.reload}</BaseButton> : null}
      {loading ? <StatusMessage>{COPY.loading}</StatusMessage> : null}
      {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
      {process?.error ? <StatusMessage tone="danger">{process.error}</StatusMessage> : null}
      {capabilities && !capabilities.ai_configured ? <StatusMessage>{COPY.unavailable}</StatusMessage> : null}
      {capabilities && !capabilities.ocr_configured ? <StatusMessage>{COPY.noOcr}</StatusMessage> : null}
      {capabilities && capabilities.ocr_configured && !capabilities.files_configured ? <StatusMessage>{COPY.noStorage}</StatusMessage> : null}
      {process?.processing ? <StatusMessage>{COPY.processing}</StatusMessage> : null}
      {uncertain ? <StatusMessage>{COPY.ambiguous}</StatusMessage> : null}
      {!loading && !(process?.proposals?.length) ? <StatusMessage variant="plain">{COPY.empty}</StatusMessage> : null}
      {(process?.proposals?.filter(proposal => proposal.status !== "rejected").length ?? 0) > 0 ? <AssistantResultCard count={process!.proposals.filter(proposal => proposal.status !== "rejected").length} action={<BaseButton size="sm" className="w-full" disabled={busy || loading || !process?.proposals.some(item => item.status === "pending")} onClick={() => void saveAll()}>Lưu tất cả</BaseButton>}>
        {(process?.proposals ?? []).filter(proposal => proposal.status !== "rejected").map(proposal => <EntryProposalRow key={`${proposal.id}:${proposal.version}:${proposal.status}`}
          onDraftChange={draft => setEdits(current => ({ ...current, [proposal.id]: draft }))} proposal={proposal} wallets={wallets} categories={categories} disabled={blocked} onBusy={value => { requestActive.current = value; setBusy(value); }} onSaved={onSaved}
          onUpdated={updated => setProcess(current => current ? { ...current, proposals: current.proposals.map(item => item.id === updated.id ? updated : item) } : current)} />)}
      </AssistantResultCard> : null}
      <AssistantComposer text={text} files={files} disabled={blocked || !capabilities?.ai_configured || uncertain} uploadDisabled={!capabilities?.files_configured} loading={busy} canSend={!!capabilities?.ai_configured && !uncertain && (!!text.trim() || !!files.length)} hint={COPY.transient}
        onTextChange={setText} onFiles={selected => {
          const issue = validateEntryFiles(selected);
          if (issue) { setError(issue); return; }
          setFiles(selected); setError("");
        }} onClearFiles={() => setFiles([])} onSubmit={() => void send()} />
      {uncertain && ready && !process?.processing ? <BaseButton variant="secondary" disabled={busy || loading} onClick={() => {
        setText(""); setFiles([]); setUncertain(false); setProcess(null); requestID.current = null;
      }}>Đã kiểm tra kết quả — soạn nội dung mới</BaseButton> : null}
    </div>
  </BaseBottomSheet>;
}
