import { apiRequest } from "./api";
import { getStoredToken } from "./auth";

export type EntryDraft = { type: string; amount: number; wallet_id: string; category_id: string | null; jar_id?: string | null; occurred_at: string; note: string; included_in_reports: boolean };
export type EntryProposal = { id: string; process_id: string; version: number; status: "pending" | "approved" | "rejected"; draft: EntryDraft; questions: string[]; transaction_id?: string };
export type EntryProcess = { id: string; proposals: EntryProposal[]; processing: boolean; error?: string };
export type EntryCapabilities = { ai_configured: boolean; ocr_configured: boolean; files_configured: boolean };
export type EntryProcessInput = { request_id: string; text: string; timezone: string; files: File[] };
const ENTRY_PATH = "/api/v1/ai/entry";
const request = <T>(path: string, method = "GET", body?: unknown) => apiRequest<T>(`${ENTRY_PATH}${path}`, { method, ...(body === undefined ? {} : { body: body instanceof FormData ? body : JSON.stringify(body) }) }, getStoredToken());

export const fetchEntryCapabilities = () => request<EntryCapabilities>("/capabilities");
export const fetchEntryRequest = (requestID: string) => request<EntryProcess>(`/requests/${encodeURIComponent(requestID)}`);
export const processEntry = (input: EntryProcessInput) => {
  const form = new FormData();
  form.append("request_id", input.request_id);
  form.append("text", input.text);
  form.append("timezone", input.timezone);
  for (const file of input.files) form.append("files", file, file.name);
  return request<EntryProcess>("/process", "POST", form);
};
export const saveEntryProposal = (proposal: EntryProposal, draft: EntryDraft) => request<EntryProposal>(`/proposals/${encodeURIComponent(proposal.id)}`, "PATCH", { version: proposal.version, draft });
export const decideEntryProposal = (proposal: EntryProposal, decision: "approve" | "reject") => request<EntryProposal>(`/proposals/${encodeURIComponent(proposal.id)}/${decision}`, "POST", { version: proposal.version });
