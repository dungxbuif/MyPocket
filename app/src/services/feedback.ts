import { apiRequest } from "./api";
import { getStoredToken } from "./auth";

export type FeedbackType = "bug" | "feature" | "improvement";
export type FeedbackStatus = "open" | "triaged" | "in_progress" | "fixed" | "rejected";

export type Feedback = {
  id: string;
  type: FeedbackType;
  title: string;
  description: string;
  status: FeedbackStatus;
  fixed_at?: string;
  changelog_id?: string;
  created_at: string;
  updated_at: string;
};

export type Changelog = {
  id: string;
  version: string;
  title: string;
  description: string;
  published_at: string;
};

const request = <T>(path: string, options: RequestInit = {}) => apiRequest<T>(`/api/v1${path}`, { ...options }, getStoredToken());

export function fetchFeedback(): Promise<Feedback[]> {
  return request<Feedback[]>("/feedback");
}

export function createFeedback(input: { type: FeedbackType; title: string; description: string }): Promise<Feedback> {
  return request<Feedback>("/feedback", { method: "POST", body: JSON.stringify({ type: input.type, title: input.title.trim(), description: input.description.trim() }) });
}

export function fetchChangelog(): Promise<Changelog[]> {
  return request<Changelog[]>("/changelog?limit=20");
}
