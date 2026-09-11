import { apiFetch } from "./apiClient";
import { uploadFile } from "./receipts";

export type AgentRunStatus = "queued" | "submitting" | "processing" | "completed" | "failed" | "cancelled" | "expired";
export type AgentToolRun = { id: string; status: AgentRunStatus; attempts: number; error_code?: string; result?: { text?: string; fields?: Record<string, unknown> } };
export type AgentRunView = {
  id: string;
  kind: "transaction_draft" | "analysis";
  status: AgentRunStatus;
  request_text: string;
  response_text?: string;
  error_code?: string;
  attempts: number;
  draft_ids: string[];
  tool_runs?: AgentToolRun[];
};

export async function submitAgentMessage(input: { kind: AgentRunView["kind"]; message: string; image?: File }, idempotencyKey = newAgentKey()) {
  const receipt = input.image ? await uploadFile(input.image) : undefined;
  const response = await apiFetch<{ run: AgentRunView }>("/api/v1/agent/messages", {
    method: "POST",
    headers: { "Content-Type": "application/json", "Idempotency-Key": idempotencyKey },
    body: JSON.stringify({ kind: input.kind, message: input.message, receipt_id: receipt?.id }),
  });
  return response.run;
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
