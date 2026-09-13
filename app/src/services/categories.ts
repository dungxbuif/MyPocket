import { apiRequest } from "./api";
import { getStoredToken } from "./auth";

export type Category = { id: string; parent_id?: string | null; kind: string; name: string; is_system: boolean };
export async function fetchCategories(): Promise<Category[]> {
  return apiRequest<Category[]>("/api/v1/categories", {}, getStoredToken());
}
