import { beforeEach, expect, test, vi } from "vitest";

const { apiFetch, uploadFile } = vi.hoisted(() => ({ apiFetch: vi.fn(), uploadFile: vi.fn() }));

vi.mock("./apiClient", () => ({ apiFetch }));
vi.mock("./receipts", () => ({ uploadFile }));

import { effectiveAgentStatus, loadIntakeSession, submitAdvisorMessage, submitIntakeMessage, type AgentMessageView, type AgentRunStatus, type AgentRunView, type AgentSessionView } from "./agent";

for (const status of ["queued","submitting","processing","completed","failed","cancelled","expired"] satisfies AgentRunStatus[]) {
  test(`exposes ${status} image-tool state`,()=>{
    const run:AgentRunView={id:"r",kind:"analysis",status:"queued",request_text:"x",attempts:0,draft_ids:[],tool_runs:[{id:"t",status,attempts:1}]};
    expect(effectiveAgentStatus(run)).toBe(status==="completed"?"queued":status);
  });
}

beforeEach(() => {
  apiFetch.mockReset();
  uploadFile.mockReset();
});

test("posts transaction intake messages to the intake endpoint", async () => {
  const run: AgentRunView = { id: "run-intake", kind: "intake", status: "queued", request_text: "Ăn trưa 80k", attempts: 0, draft_ids: [] };
  const session: AgentSessionView = { id: "session-1", kind: "intake", title: "Nhập giao dịch", status: "active", created_at: "2026-09-12T00:00:00Z", updated_at: "2026-09-12T00:00:00Z" };
  const message: AgentMessageView = { id: "message-1", session_id: "session-1", run_id: "run-intake", role: "user", text: "Ăn trưa 80k", created_at: "2026-09-12T00:00:00Z" };
  apiFetch.mockResolvedValue({ run, session, message });

  await expect(submitIntakeMessage({ message: "Ăn trưa 80k" }, "intake-key")).resolves.toEqual({ run, session, message });

  expect(apiFetch).toHaveBeenCalledWith("/api/v1/agent/intakes", expect.objectContaining({
    method: "POST",
    headers: expect.objectContaining({ "Idempotency-Key": "intake-key" }),
    body: JSON.stringify({ message: "Ăn trưa 80k", receipt_id: undefined }),
  }));
});

test("uploads intake images before submitting the receipt reference", async () => {
  const file = new File(["fake"], "receipt.jpg", { type: "image/jpeg" });
  const run: AgentRunView = { id: "run-image", kind: "intake", status: "queued", request_text: "Ảnh hóa đơn", attempts: 0, draft_ids: [] };
  const session: AgentSessionView = { id: "session-1", kind: "intake", title: "Nhập giao dịch", status: "active", created_at: "2026-09-12T00:00:00Z", updated_at: "2026-09-12T00:00:00Z" };
  const message: AgentMessageView = { id: "message-1", session_id: "session-1", run_id: "run-image", role: "user", text: "Ảnh hóa đơn", created_at: "2026-09-12T00:00:00Z" };
  uploadFile.mockResolvedValue({ id: "receipt-1" });
  apiFetch.mockResolvedValue({ run, session, message });

  await submitIntakeMessage({ message: "Ảnh hóa đơn", image: file }, "image-key");

  expect(uploadFile).toHaveBeenCalledWith(file);
  expect(apiFetch).toHaveBeenCalledWith("/api/v1/agent/intakes", expect.objectContaining({
    body: JSON.stringify({ message: "Ảnh hóa đơn", receipt_id: "receipt-1" }),
  }));
});

test("posts advisor messages to the read-only advisor endpoint", async () => {
  const run: AgentRunView = { id: "run-advisor", kind: "advisor", status: "queued", request_text: "Tháng này sao?", attempts: 0, draft_ids: [] };
  const session: AgentSessionView = { id: "session-advisor", kind: "advisor", title: "Tư vấn tài chính", status: "active", created_at: "2026-09-12T00:00:00Z", updated_at: "2026-09-12T00:00:00Z" };
  const message: AgentMessageView = { id: "message-1", session_id: "session-advisor", run_id: "run-advisor", role: "user", text: "Tháng này sao?", created_at: "2026-09-12T00:00:00Z" };
  apiFetch.mockResolvedValue({ run, session, message });

  await expect(submitAdvisorMessage({ message: "Tháng này sao?" }, "advisor-key")).resolves.toEqual({ run, session, message });

  expect(apiFetch).toHaveBeenCalledWith("/api/v1/agent/advisor/messages", expect.objectContaining({
    method: "POST",
    headers: expect.objectContaining({ "Idempotency-Key": "advisor-key" }),
    body: JSON.stringify({ message: "Tháng này sao?" }),
  }));
});

test("loads intake chat history with assistant draft action cards", async () => {
  const session: AgentSessionView = { id: "session-1", kind: "intake", title: "Nhập giao dịch", status: "active", created_at: "2026-09-12T00:00:00Z", updated_at: "2026-09-12T00:00:00Z" };
  const messages: AgentMessageView[] = [
    { id: "message-1", session_id: "session-1", role: "user", text: "Ăn trưa 80k", created_at: "2026-09-12T00:00:00Z" },
    { id: "message-2", session_id: "session-1", role: "assistant", text: "Đã tạo nháp", action: { type: "drafts_created", draft_ids: ["draft-1"] }, created_at: "2026-09-12T00:00:01Z" },
  ];
  apiFetch.mockResolvedValue({ session, messages });

  await expect(loadIntakeSession("session-1")).resolves.toEqual({ session, messages });

  expect(apiFetch).toHaveBeenCalledWith("/api/v1/agent/intakes/session-1");
});
