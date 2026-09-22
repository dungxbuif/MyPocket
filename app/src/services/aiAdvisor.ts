import { apiRequest } from "./api";
import { getStoredToken } from "./auth";
export { mergeAdvisorMessages } from "./advisorHistory";

export type AdvisorPart = { type: string; text?: string; data?: unknown };
export type AdvisorMessage = { id: string; conversation_id: string; seq: number; role: "user" | "assistant"; parts: AdvisorPart[]; created_at: string };
export type AdvisorRun = { id: string; conversation_id: string; status: string; generation: number; created_at: string; finished_at?: string };
export type AdvisorSubmit = { run_id: string; conversation_id: string; text: string; status: string };
export type AdvisorCapabilities = { enabled: boolean; write: boolean; parts_version: number; tools: string[] };
export type AdvisorOverview = { view_kind: string; view: unknown; source: unknown; status: string };

const request = <T>(path: string, options: RequestInit = {}) => apiRequest<T>(`/api/v1/ai/advisor${path}`, { ...options }, getStoredToken());

export function fetchAdvisorCapabilities(): Promise<AdvisorCapabilities> {
  return request<AdvisorCapabilities>("/capabilities");
}

export function submitAdvisorMessage(input: { client_request_id: string; text: string }): Promise<AdvisorSubmit> {
  return request<AdvisorSubmit>("/messages", { method: "POST", body: JSON.stringify(input) });
}

export function fetchAdvisorOverview(month?: string): Promise<AdvisorOverview> {
  const suffix = month ? `?month=${encodeURIComponent(month)}` : "";
  return request<AdvisorOverview>(`/overview${suffix}`);
}

export function fetchAdvisorMessages(conversationId: string, beforeSeq?: number): Promise<AdvisorMessage[]> {
  const before = beforeSeq && beforeSeq > 0 ? `&before_seq=${encodeURIComponent(beforeSeq)}` : "";
  return request<AdvisorMessage[]>(`/conversation/messages?conversation_id=${encodeURIComponent(conversationId)}${before}`).then(messages => [...messages].sort((left, right) => left.seq - right.seq));
}

export function fetchAdvisorRun(id: string): Promise<AdvisorRun> {
  return request<AdvisorRun>(`/runs/${encodeURIComponent(id)}`);
}

export function cancelAdvisorRun(id: string): Promise<null> {
  return request<null>(`/runs/${encodeURIComponent(id)}/cancel`, { method: "POST" });
}
