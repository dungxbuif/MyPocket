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
  conflict?: {
    entity_type: OfflineMutation["entity_type"];
    entity_id: string;
    operation: OfflineMutation["operation"];
    base_version: number;
    server_version: number;
    local_payload: Record<string, unknown>;
    server_payload: Record<string, unknown>;
  };
};

export async function submitSyncMutations(mutations: OfflineMutation[]) {
  const response = await apiFetch<{ results: SyncMutationResult[] }>("/api/v1/sync/mutations", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ mutations }),
  });
  return response.results;
}

export async function fetchAuthoritativeSnapshot() {
  const response = await apiFetch<{
    snapshot: {
      wallets: unknown[];
      categories: unknown[];
      transactions: unknown[];
      next_cursor: number;
    };
  }>("/api/v1/sync/resync", { method: "POST" });
  return response.snapshot;
}
