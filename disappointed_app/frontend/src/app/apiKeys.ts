import { apiFetch } from "./apiClient";

export type APIKeySummary = {
  id: string;
  user_id: string;
  name: string;
  key_prefix: string;
  last_used_at?: string;
  revoked_at?: string;
  created_at: string;
};

export type CreatedAPIKey = APIKeySummary & {
  plaintext: string;
};

export async function loadAPIKeys(): Promise<APIKeySummary[]> {
  const response = await apiFetch<{ keys: APIKeySummary[] }>("/api/v1/api-keys");
  return response.keys ?? [];
}

export async function createAPIKey(name: string): Promise<CreatedAPIKey> {
  const response = await apiFetch<{ key: CreatedAPIKey }>("/api/v1/api-keys", {
    method: "POST",
    body: JSON.stringify({ name }),
    headers: { "Content-Type": "application/json" },
  });
  return response.key;
}

export async function revokeAPIKey(id: string): Promise<void> {
  await apiFetch(`/api/v1/api-keys/${id}/revoke`, { method: "POST" });
}
