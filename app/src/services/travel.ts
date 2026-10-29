import { apiRequest } from "./api";
import { getStoredToken } from "./auth";

export type TravelEvent = {
  id: string;
  owner_id: string;
  name: string;
  context?: string | null;
  starts_on?: string | null;
  ends_on?: string | null;
  active: boolean;
  created_at: string;
  updated_at: string;
};

export type TravelEventInput = {
  name: string;
  context?: string | null;
  starts_on?: string | null;
  ends_on?: string | null;
  active?: boolean;
};

const PATH = "/api/v1/travel/events";

export function fetchTravelEvents(): Promise<TravelEvent[]> {
  return apiRequest<TravelEvent[]>(PATH, {}, getStoredToken());
}

export function createTravelEvent(input: TravelEventInput): Promise<TravelEvent> {
  return apiRequest<TravelEvent>(PATH, { method: "POST", body: JSON.stringify(input) }, getStoredToken());
}

export function updateTravelEvent(id: string, input: TravelEventInput): Promise<TravelEvent> {
  return apiRequest<TravelEvent>(`${PATH}/${encodeURIComponent(id)}`, { method: "PATCH", body: JSON.stringify(input) }, getStoredToken());
}

export function deleteTravelEvent(id: string): Promise<void> {
  return apiRequest<void>(`${PATH}/${encodeURIComponent(id)}`, { method: "DELETE" }, getStoredToken());
}

export function setTravelEventActive(id: string, active: boolean): Promise<TravelEvent> {
  return apiRequest<TravelEvent>(`${PATH}/${encodeURIComponent(id)}/${active ? "activate" : "deactivate"}`, { method: "POST" }, getStoredToken());
}

export function linkTransactionTravel(transactionID: string, eventID: string | null): Promise<unknown> {
  return apiRequest<unknown>(`/api/v1/transactions/${encodeURIComponent(transactionID)}/travel`, { method: "PATCH", body: JSON.stringify({ event_id: eventID }) }, getStoredToken());
}
