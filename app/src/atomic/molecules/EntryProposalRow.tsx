import { useRef, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseCheckbox } from "../atoms/BaseCheckbox";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { Heading } from "../atoms/Heading";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { decideEntryProposal, saveEntryProposal, type EntryDraft, type EntryProposal } from "../../services/ai";
import { localEntryDate, proposalIssues, sameDraft } from "../../services/aiEntryLogic";
import type { Category } from "../../services/categories";
import { categoryAppliesToTransaction } from "../../services/transactionLogic";
import type { Wallet } from "../../services/wallets";

const COPY = {
  pending: "Đề xuất chưa ghi sổ", approved: "Đã duyệt", rejected: "Đã từ chối", type: "Loại giao dịch",
  expense: "Khoản chi", income: "Khoản thu", amount: "Số tiền (VND)", wallet: "Ví", category: "Nhóm",
  chooseWallet: "Chọn ví", noCategory: "Không chọn nhóm", date: "Ngày và giờ", note: "Ghi chú", reports: "Tính vào báo cáo",
  save: "Lưu bản nháp", approve: "Duyệt giao dịch", reject: "Từ chối", dirty: "Có thay đổi chưa lưu. Lưu bản nháp trước khi duyệt.",
  confirmReject: "Từ chối đề xuất này? Không tạo giao dịch và không thể sửa lại đề xuất đã từ chối.",
  unknown: "Chưa xác định — chọn loại", transfer: "Chuyển tiền — chưa hỗ trợ",
  transferBlocked: "Chuyển tiền chưa thể duyệt cho đến khi có giao dịch chuyển tiền ghi sổ hai ví.",
  failed: "Thao tác thất bại. Dữ liệu đang sửa được giữ lại. Tải lại phiên để kiểm tra trạng thái trước khi thử lại.",
};

export function EntryProposalRow({ proposal, wallets, categories, disabled = false, onUpdated, onSaved, onBusy }: {
  proposal: EntryProposal; wallets: Wallet[]; categories: Category[]; disabled?: boolean;
  onUpdated: (proposal: EntryProposal) => void; onSaved: () => void; onBusy?: (busy: boolean) => void;
}) {
  const [draft, setDraft] = useState(proposal.draft);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const running = useRef(false);
  const dirty = !sameDraft(draft, proposal.draft);
  const transfer = proposal.draft.type === "transfer";
  const supported = draft.type === "income" || draft.type === "expense";
  const issues = proposalIssues(draft, wallets, categories);
  const locked = busy || disabled;
  const wallet = wallets.find(item => item.id === draft.wallet_id);
  const applicable = categories.filter(category => (draft.type === "income" || draft.type === "expense") && categoryAppliesToTransaction(category, draft.type, draft.wallet_id, wallet?.type));
  const change = (patch: Partial<EntryDraft>) => setDraft(current => ({ ...current, ...patch }));
  const perform = async (action: "save" | "approve" | "reject") => {
    if (running.current || disabled || proposal.status !== "pending") return;
    if (action === "approve" && (dirty || issues.length || !supported || transfer)) return;
    if (action === "reject" && !window.confirm(COPY.confirmReject)) return;
    running.current = true; setBusy(true); onBusy?.(true); setError("");
    try {
      const updated = action === "save" ? await saveEntryProposal(proposal, draft) : await decideEntryProposal(proposal, action);
      onUpdated(updated);
      if (updated.status === "approved") onSaved();
    } catch (reason) {
      setError(`${reason instanceof Error ? reason.message : COPY.failed} ${COPY.failed}`);
    } finally { running.current = false; setBusy(false); onBusy?.(false); }
  };

  if (proposal.status !== "pending") return <SurfaceCard padding="md" className="space-y-2">
    <Heading as="h3" size="field">{COPY[proposal.status]}</Heading>
    <Text>{proposal.draft.type} · {proposal.draft.amount.toLocaleString("vi-VN")} VND</Text>
    <Text>{wallets.find(item => item.id === proposal.draft.wallet_id)?.name ?? proposal.draft.wallet_id}</Text>
    <Text>{categories.find(item => item.id === proposal.draft.category_id)?.name ?? COPY.noCategory}</Text>
    <Text>{localEntryDate(proposal.draft.occurred_at).replace("T", " ")}</Text>
    <Text>{proposal.draft.note}</Text>
    <Text tone="secondary">{COPY.reports}: {proposal.draft.included_in_reports ? "Có" : "Không"}</Text>
  </SurfaceCard>;

  return <SurfaceCard padding="md" className="space-y-3">
    <Heading as="h3" size="field">{COPY.pending}</Heading>
    {(proposal.questions ?? []).map((question, index) => <StatusMessage key={index}>{question}</StatusMessage>)}
    {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
    {transfer ? <StatusMessage>{COPY.transferBlocked}</StatusMessage> : null}
    <FormField label={COPY.type}><BaseSelect value={draft.type} disabled={locked || transfer} onChange={event => change({ type: event.target.value })}>
      {!supported ? <option value={draft.type}>{transfer ? COPY.transfer : COPY.unknown}</option> : null}
      <option value="expense">{COPY.expense}</option><option value="income">{COPY.income}</option>
    </BaseSelect></FormField>
    <FormField label={COPY.amount}><BaseTextInput inputMode="numeric" type="number" min="1" step="1" value={draft.amount || ""} disabled={locked} onChange={event => change({ amount: Number(event.target.value) })} /></FormField>
    <FormField label={COPY.wallet}><BaseSelect value={draft.wallet_id} disabled={locked} onChange={event => change({ wallet_id: event.target.value })}>
      <option value="">{COPY.chooseWallet}</option>
      {draft.wallet_id && !wallets.some(item => item.id === draft.wallet_id && item.type !== "credit") ? <option value={draft.wallet_id}>Ví không phù hợp</option> : null}
      {wallets.filter(item => item.type === "basic" || item.type === "goal").map(item => <option key={item.id} value={item.id}>{item.name}</option>)}
    </BaseSelect></FormField>
    <FormField label={COPY.category}><BaseSelect value={draft.category_id ?? ""} disabled={locked} onChange={event => change({ category_id: event.target.value || null })}>
      <option value="">{COPY.noCategory}</option>
      {draft.category_id && !applicable.some(item => item.id === draft.category_id) ? <option value={draft.category_id}>Nhóm không phù hợp</option> : null}
      {applicable.map(item => <option key={item.id} value={item.id}>{item.name}</option>)}
    </BaseSelect></FormField>
    <FormField label={COPY.date}><BaseTextInput type="datetime-local" disabled={locked} value={localEntryDate(draft.occurred_at)} onChange={event => change({ occurred_at: event.target.value && Number.isFinite(Date.parse(event.target.value)) ? new Date(event.target.value).toISOString() : "" })} /></FormField>
    <FormField label={COPY.note}><BaseTextInput disabled={locked} value={draft.note} onChange={event => change({ note: event.target.value })} /></FormField>
    <BaseCheckbox label={COPY.reports} disabled={locked} checked={draft.included_in_reports} onChange={event => change({ included_in_reports: event.target.checked })}>{COPY.reports}</BaseCheckbox>
    {issues.map(issue => <StatusMessage key={issue}>{issue}</StatusMessage>)}
    {dirty ? <StatusMessage>{COPY.dirty}</StatusMessage> : null}
    <div className="flex flex-wrap gap-2">
      <BaseButton variant="secondary" disabled={locked || !dirty} onClick={() => void perform("save")}>{COPY.save}</BaseButton>
      <BaseButton disabled={locked || dirty || issues.length > 0 || !supported || transfer} onClick={() => void perform("approve")}>{COPY.approve}</BaseButton>
      <BaseButton variant="danger" disabled={locked} onClick={() => void perform("reject")}>{COPY.reject}</BaseButton>
    </div>
  </SurfaceCard>;
}
