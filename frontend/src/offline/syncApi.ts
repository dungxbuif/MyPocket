import { apiFetch } from "../app/apiClient";
import type { OfflineMutation } from "./types";

export type SyncMutationResult = {
  mutation_id: string;
  entity_type: OfflineMutation["entity_type"];
  entity_id: string;
  operation: OfflineMutation["operation"];
  state: "applied" | "replayed" | "rejected" | "conflict";
  version?: number;
  payload?: Record<string, unknown>;
  reason?: string;
  conflict?: Record<string, unknown>;
};

export async function submitSyncMutations(mutations: OfflineMutation[]) {
  const response = await apiFetch<{ results: SyncMutationResult[] }>("/api/v1/sync/mutations", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ mutations }),
  });
  return response.results;
}
