import { apiFetch } from "./apiClient";

export type AuditEvent = {
  id: string;
  occurred_at: string;
  correlation_id: string;
  action: string;
  entity_type: string;
  entity_id?: string;
  outcome: string;
  severity: string;
  source: string;
  error_code?: string;
  request_method?: string;
  request_path?: string;
  metadata?: Record<string, unknown>;
};

export async function checkAuditAccess() {
  return apiFetch<{ allowed: boolean; correlation_id: string }>("/api/v1/audit/access");
}

export async function loadAuditEvents(correlationID = "") {
  const query = correlationID.trim() ? `?correlation_id=${encodeURIComponent(correlationID.trim())}&limit=100` : "?limit=100";
  const response = await apiFetch<{ events: AuditEvent[]; correlation_id: string }>(`/api/v1/audit/events${query}`);
  return response.events ?? [];
}
