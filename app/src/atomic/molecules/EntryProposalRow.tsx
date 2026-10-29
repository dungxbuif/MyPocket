import { useEffect, useRef, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseSelect, FormField } from "../atoms/FormField";
import { Heading } from "../atoms/Heading";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { TransactionFields } from "./TransactionFields";
import { decideEntryProposal, saveEntryProposal, type EntryDraft, type EntryProposal } from "../../services/ai";
import { localEntryDate, proposalIssues, sameDraft } from "../../services/aiEntryLogic";
import type { Category } from "../../services/categories";
import type { Wallet } from "../../services/wallets";
import { instantFromLocalDateTime } from "../../services/accountTime";
import { useAccountTimezone } from "../../services/AccountTimezoneContext";
import { fetchJarMonth, type JarMonthSummary } from "../../services/jars";
import { jarAssignmentNeedsSelection, monthKeyForInstant } from "../../services/monthJarLogic";

export function EntryProposalRow({ proposal, wallets, categories, disabled = false, onUpdated, onSaved, onBusy, onDraftChange }: {
  proposal: EntryProposal; wallets: Wallet[]; categories: Category[]; disabled?: boolean;
  onUpdated: (proposal: EntryProposal) => void; onSaved: () => void; onBusy?: (busy: boolean) => void; onDraftChange?: (draft: EntryDraft) => void;
}) {
  const timezone=useAccountTimezone();
  const [draft, setDraft] = useState(proposal.draft);
  const [jarSummary, setJarSummary] = useState<JarMonthSummary | null>(null);
  const [busy, setBusy] = useState(false), [error, setError] = useState("");
  const running = useRef(false), saved = useRef(proposal);
  const transfer = proposal.draft.type === "transfer";
  const supported = draft.type === "income" || draft.type === "expense";
  const draftMonth = monthKeyForInstant(draft.occurred_at, timezone);
  useEffect(() => {
    if (!supported || draft.type !== "expense") {
      setJarSummary(null);
      return;
    }
    let cancelled = false;
    fetchJarMonth(draftMonth).then(value => { if (!cancelled) setJarSummary(value); }).catch(() => { if (!cancelled) setJarSummary(null); });
    return () => { cancelled = true; };
  }, [supported, draft.type, draftMonth]);
  const jarOptions = (jarSummary?.items ?? []).filter(item => item.active || item.jar_id === draft.jar_id).map(item => ({ jar_id: item.jar_id, name: item.name, active: item.active }));
  const stableJarName = jarSummary?.jars.find(item => item.jar_id === draft.jar_id)?.name;
  if (draft.jar_id && !jarOptions.some(item => item.jar_id === draft.jar_id)) jarOptions.push({ jar_id: draft.jar_id, name: stableJarName ?? "Hũ không có cấu hình tháng", active: false });
  const issues = [
    ...proposalIssues(draft, wallets, categories),
    ...(draft.jar_id && draft.type !== "expense" ? ["Chỉ khoản chi thông thường mới được gắn hũ."] : []),
    ...(jarAssignmentNeedsSelection(draft.jar_id, jarOptions, proposal.draft.occurred_at, draft.occurred_at) ? ["Hũ đã gỡ khỏi tháng này. Bỏ chọn hoặc chọn hũ đang hoạt động trước khi lưu."] : []),
  ];
  const locked = busy || disabled;
  const change = (next: EntryDraft) => { setDraft(next); onDraftChange?.(next); };
  const perform = async (action: "approve" | "reject") => {
    if (running.current || disabled || proposal.status !== "pending") return;
    if (action === "approve" && (issues.length || !supported || transfer)) return;
    running.current = true; setBusy(true); onBusy?.(true); setError("");
    try {
      if (action === "approve" && !sameDraft(draft, saved.current.draft)) saved.current = await saveEntryProposal(saved.current, draft);
      const updated = await decideEntryProposal(saved.current, action);
      onUpdated(updated);
      if (updated.status === "approved") onSaved();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Không thể lưu giao dịch. Dữ liệu đang sửa được giữ lại.");
    } finally { running.current = false; setBusy(false); onBusy?.(false); }
  };
  if (proposal.status !== "pending") return <SurfaceCard padding="md" className="space-y-2">
    <Heading as="h3" size="field">{proposal.status === "approved" ? "Đã lưu" : "Đã xóa"}</Heading>
    <Text>{proposal.draft.amount.toLocaleString("vi-VN")} VND · {wallets.find(item => item.id === proposal.draft.wallet_id)?.name}</Text>
    <Text>{proposal.draft.note}</Text>
  </SurfaceCard>;
  return <section className="space-y-3" aria-label="Giao dịch đề xuất">
    {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
    {!supported ? <FormField label="Loại giao dịch"><BaseSelect disabled={locked || transfer} value={draft.type} onChange={e => change({ ...draft, type: e.target.value })}><option value={draft.type}>{transfer ? "Chuyển tiền — chưa hỗ trợ" : "Chọn loại giao dịch"}</option><option value="expense">Khoản chi</option><option value="income">Khoản thu</option></BaseSelect></FormField> : null}
    <TransactionFields state={{ type: draft.type === "income" ? "income" : "expense", amount: draft.amount ? String(draft.amount) : "", walletID: draft.wallet_id, categoryID: draft.category_id ?? "", jarID: draft.jar_id ?? "", occurredAt: localEntryDate(draft.occurred_at,timezone), note: draft.note, includedInReports: draft.included_in_reports }}
      disabled={locked || !supported || transfer} wallets={wallets} categories={categories} jars={jarOptions}
      onChange={next => { let occurredAt=""; try { if(next.occurredAt) occurredAt=instantFromLocalDateTime(next.occurredAt,timezone); } catch { /* invalid wall times remain unapprovable */ } change({ type: next.type, amount: Number(next.amount), wallet_id: next.walletID, category_id: next.categoryID || null, jar_id: next.jarID || null, occurred_at: occurredAt, note: next.note, included_in_reports: next.includedInReports }); }} />
    {issues.map(issue => <StatusMessage variant="plain" key={issue}>{issue}</StatusMessage>)}
    <div className="flex gap-2"><BaseButton className="flex-1" disabled={locked || issues.length > 0 || !supported || transfer} onClick={() => void perform("approve")}>Lưu giao dịch</BaseButton><BaseButton variant="ghost" disabled={locked} onClick={() => void perform("reject")}>Xóa item</BaseButton></div>
  </section>;
}
