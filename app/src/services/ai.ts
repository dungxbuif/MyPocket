import { apiRequest } from "./api";
import { getStoredToken } from "./auth";

export type EntryDraft = { type: string; amount: number; wallet_id: string; category_id: string | null; occurred_at: string; note: string; included_in_reports: boolean };
export type EntryProposal = { id: string; session_id: string; version: number; status: "pending" | "approved" | "rejected"; draft: EntryDraft; questions: string[]; transaction_id?: string };
export type EntrySession = { id: string; messages: { id: string; role: string; content: string; created_at: string }[]; proposals: EntryProposal[]; processing: boolean; error?: string };
export type EntryCapabilities = { ai_configured: boolean; ocr_configured: boolean };
export type EntryMessageInput = { request_id: string; text: string; timezone: string; images?: { name: string; mime_type: string; base64: string }[] };
const ENTRY_PATH = "/api/v1/ai/entry";
const request = <T>(path: string, method = "GET", body?: unknown) => apiRequest<T>(`${ENTRY_PATH}${path}`, { method, ...(body === undefined ? {} : { body: JSON.stringify(body) }) }, getStoredToken());

export const fetchEntryCapabilities = () => request<EntryCapabilities>("/capabilities");
export const fetchLatestEntrySession = () => request<EntrySession | null>("/sessions/latest");
export const createEntrySession = () => request<EntrySession>("/sessions", "POST", {});
export const fetchEntrySession = (id: string) => request<EntrySession>(`/sessions/${encodeURIComponent(id)}`);
export const sendEntryMessage = (id: string, input: EntryMessageInput) => request<EntrySession>(`/sessions/${encodeURIComponent(id)}/messages`, "POST", input);
export const saveEntryProposal = (proposal: EntryProposal, draft: EntryDraft) => request<EntryProposal>(`/proposals/${encodeURIComponent(proposal.id)}`, "PATCH", { version: proposal.version, draft });
export const decideEntryProposal = (proposal: EntryProposal, decision: "approve" | "reject") => request<EntryProposal>(`/proposals/${encodeURIComponent(proposal.id)}/${decision}`, "POST", { version: proposal.version });
