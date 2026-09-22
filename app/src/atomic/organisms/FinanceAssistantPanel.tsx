import { useEffect, useMemo, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { Heading } from "../atoms/Heading";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { AssistantComposer } from "../molecules/AssistantComposer";
import { FinanceAssistantPart } from "../molecules/FinanceAssistantPart";
import { fetchAdvisorCapabilities, fetchAdvisorMessages, fetchAdvisorOverview, submitAdvisorMessage, type AdvisorMessage, type AdvisorOverview } from "../../services/aiAdvisor";

const CONVERSATION_KEY = "mypocket.advisor.conversation";

export function FinanceAssistantPanel({ ownerID, masked = false }: { ownerID: string; masked?: boolean }) {
  const [messages, setMessages] = useState<AdvisorMessage[]>([]);
  const [overview, setOverview] = useState<AdvisorOverview | null>(null);
  const [text, setText] = useState("");
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [enabled, setEnabled] = useState(false);
  const [error, setError] = useState("");
  const storageKey = `${CONVERSATION_KEY}:${ownerID}`;
  const conversationID = typeof window === "undefined" ? "" : window.localStorage.getItem(storageKey) ?? "";

  useEffect(() => {
    let cancelled = false;
    void Promise.all([fetchAdvisorCapabilities(), conversationID ? fetchAdvisorMessages(conversationID) : Promise.resolve([]), fetchAdvisorOverview()])
      .then(([capabilities, history, summary]) => {
        if (cancelled) return;
        setEnabled(capabilities.enabled);
        setMessages(history);
        setOverview(summary);
        setError("");
      })
      .catch(caught => { if (!cancelled) setError(caught instanceof Error ? caught.message : "Không thể tải trợ lý tài chính"); })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [conversationID]);

  const suggestions = useMemo(() => ["Tháng này tôi tiêu thế nào?", "Tìm các giao dịch trên 200.000đ", "So sánh chi tiêu với tháng trước"], []);

  const submit = async (): Promise<void> => {
    const value = text.trim();
    if (!value || sending || !enabled) return;
    setSending(true);
    setError("");
    try {
      const result = await submitAdvisorMessage({ client_request_id: crypto.randomUUID(), text: value });
      if (typeof window !== "undefined") window.localStorage.setItem(storageKey, result.conversation_id);
      setText("");
      const history = await fetchAdvisorMessages(result.conversation_id);
      setMessages(history);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : "Không thể gửi câu hỏi");
    } finally {
      setSending(false);
    }
  };

  if (loading) return <div className="space-y-3 p-4"><Text tone="secondary">Đang tải Finance Assistant...</Text></div>;

  return <div className="flex min-h-[calc(100vh-8rem)] flex-col gap-3 p-4 pb-24">
    <div className="flex items-start justify-between gap-3">
      <div><Heading as="h1" size="screen">Finance Assistant</Heading><Text tone="secondary" size="xs">Hỏi đáp trên dữ liệu tài chính thật · chỉ đọc</Text></div>
      {messages.length ? <BaseButton variant="ghost" size="sm" onClick={() => conversationID ? void fetchAdvisorMessages(conversationID).then(setMessages).catch(caught => setError(caught instanceof Error ? caught.message : "Không thể tải lại lịch sử")) : undefined}>Tải lại lịch sử</BaseButton> : null}
    </div>
    {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
    {overview ? <FinanceAssistantPart type={overview.view_kind} data={{ view: overview.view, source: overview.source }} masked={masked} /> : null}
    {!enabled ? <SurfaceCard tone="muted" padding="md"><Text>Trợ lý chưa được bật hoặc chưa cấu hình nhà cung cấp AI.</Text></SurfaceCard> : null}
    {!messages.length ? <SurfaceCard padding="md" className="space-y-3"><Text weight="semibold">Bạn có thể hỏi</Text><div className="flex flex-wrap gap-2">{suggestions.map(suggestion => <BaseButton key={suggestion} variant="secondary" size="sm" onClick={() => setText(suggestion)}>{suggestion}</BaseButton>)}</div></SurfaceCard> : null}
    <div className="flex flex-1 flex-col gap-2" aria-live="polite">{messages.map(message => <SurfaceCard key={message.id} padding="sm" tone={message.role === "user" ? "muted" : "default"} className={message.role === "user" ? "ml-8" : "mr-8"}><div className="space-y-2">{message.parts.map((part, index) => part.type === "text" ? <Text key={`${message.id}-text-${index}`} weight={message.role === "assistant" ? "normal" : "semibold"}>{masked ? "Nội dung ẩn khi bật chế độ che số dư." : part.text}</Text> : <FinanceAssistantPart key={`${message.id}-${part.type}-${index}`} type={part.type} data={part.data} masked={masked} />)}</div></SurfaceCard>)}</div>
    <AssistantComposer mode="advisor" text={text} files={[]} disabled={!enabled} uploadDisabled loading={sending} canSend={!!text.trim()} hint="Dữ liệu chỉ được đọc; không có thao tác ngân hàng hoặc thanh toán." onTextChange={setText} onFiles={() => undefined} onClearFiles={() => undefined} onSubmit={() => void submit()} />
  </div>;
}
