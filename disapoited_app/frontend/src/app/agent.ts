import { apiFetch } from "./apiClient";
import { uploadFile } from "./receipts";

export type AgentRunStatus = "queued" | "submitting" | "processing" | "completed" | "failed" | "cancelled" | "expired";
export type AgentToolRun = { id: string; status: AgentRunStatus; attempts: number; error_code?: string; result?: { text?: string; fields?: Record<string, unknown> } };
export type AgentRunView = {
  id: string;
  kind: "transaction_draft" | "analysis" | "intake" | "advisor";
  status: AgentRunStatus;
  request_text: string;
  response_text?: string;
  error_code?: string;
  attempts: number;
  draft_ids: string[];
  tool_runs?: AgentToolRun[];
};
export type AgentSessionView = {
  id: string;
  kind: "intake" | "advisor";
  title: string;
  context_summary?: string;
  status: "active" | "archived";
  created_at: string;
  updated_at: string;
};
export type AgentActionCard = { type: "drafts_created"; draft_ids: string[] };
export type AgentMessageView = {
  id: string;
  session_id: string;
  run_id?: string;
  role: "user" | "assistant" | "system";
  text: string;
  action?: AgentActionCard;
  created_at: string;
};
export type AgentSessionRunView = { run: AgentRunView; session: AgentSessionView; message: AgentMessageView };
export type AgentSessionHistory = { session: AgentSessionView; messages: AgentMessageView[] };

export async function submitAgentMessage(input: { kind: AgentRunView["kind"]; message: string; image?: File }, idempotencyKey = newAgentKey()) {
  if (input.kind === "intake" || input.kind === "transaction_draft") {
    return submitIntakeMessage({ message: input.message, image: input.image }, idempotencyKey);
  }
  return submitAdvisorMessage({ message: input.message }, idempotencyKey);
}

export async function submitIntakeMessage(input: { message: string; image?: File }, idempotencyKey = newAgentKey()) {
  const receipt = input.image ? await uploadFile(input.image) : undefined;
  const response = await apiFetch<AgentSessionRunView>("/api/v1/agent/intakes", {
    method: "POST",
    headers: { "Content-Type": "application/json", "Idempotency-Key": idempotencyKey },
    body: JSON.stringify({ message: input.message, receipt_id: receipt?.id }),
  });
  return response;
}

export async function submitAdvisorMessage(input: { message: string }, idempotencyKey = newAgentKey()) {
  const response = await apiFetch<AgentSessionRunView>("/api/v1/agent/advisor/messages", {
    method: "POST",
    headers: { "Content-Type": "application/json", "Idempotency-Key": idempotencyKey },
    body: JSON.stringify({ message: input.message }),
  });
  return response;
}

export async function loadIntakeSession(id: string) {
  const response = await apiFetch<AgentSessionHistory>(`/api/v1/agent/intakes/${encodeURIComponent(id)}`);
  return response;
}

export async function loadAgentRun(id: string) {
  const response = await apiFetch<{ run: AgentRunView }>(`/api/v1/agent/runs/${encodeURIComponent(id)}`);
  return response.run;
}

export function effectiveAgentStatus(run: AgentRunView): AgentRunStatus {
  const activeTool = run.tool_runs?.find((tool) => !["completed", "failed", "cancelled", "expired"].includes(tool.status));
  if (activeTool) return activeTool.status;
  const failedTool = run.tool_runs?.find((tool) => ["failed", "cancelled", "expired"].includes(tool.status));
  return run.status === "queued" && failedTool ? failedTool.status : run.status;
}

function newAgentKey() {
  return `agent-${globalThis.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`}`;
}
