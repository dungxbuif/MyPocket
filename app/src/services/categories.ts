import { apiRequest } from "./api";
import { getStoredToken } from "./auth";

export type Category = { id: string; owner_id?: string | null; parent_id?: string | null; kind: string; name: string; system_key?: string | null; icon_key: string; is_system: boolean; wallet_ids: string[] };

export type CategoryInput = { name: string; kind: string; parent_id?: string | null; wallet_ids: string[]; icon_key: string };

const CATEGORY_API_PATH = "/api/v1/categories";
const categoryPath = (id: string) => `${CATEGORY_API_PATH}/${id}`;

export async function fetchCategories(): Promise<Category[]> {
  return apiRequest<Category[]>(CATEGORY_API_PATH, {}, getStoredToken());
}

export async function createCategory(input: CategoryInput): Promise<Category> {
  return apiRequest<Category>(CATEGORY_API_PATH, { method: "POST", body: JSON.stringify(input) }, getStoredToken());
}

export async function updateCategory(id: string, input: CategoryInput): Promise<Category> {
  return apiRequest<Category>(categoryPath(id), { method: "PATCH", body: JSON.stringify(input) }, getStoredToken());
}

export async function deleteCategory(id: string): Promise<void> {
  await apiRequest<null>(categoryPath(id), { method: "DELETE" }, getStoredToken());
}
