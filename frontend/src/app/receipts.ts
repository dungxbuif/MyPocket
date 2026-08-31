import { apiFetch } from "./apiClient";

export type ReceiptObject = {
  id: string;
  object_key: string;
  content_type: string;
  size_bytes: number;
  original_filename: string;
  created_at: string;
};

export async function uploadReceipt(file: File): Promise<ReceiptObject> {
  const checksum = await sha256Hex(file);
  const prepared = await apiFetch<{ receipt: ReceiptObject; upload_url: string }>("/api/v1/receipts/uploads", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ filename: file.name, content_type: file.type, size_bytes: file.size, checksum_sha256: checksum }),
  });
  const response = await fetch(prepared.upload_url, { method: "PUT", headers: { "Content-Type": file.type }, body: file });
  if (!response.ok) throw new Error("receipt upload failed");
  return prepared.receipt;
}

export async function getReceiptURL(id: string) {
  const response = await apiFetch<{ download_url: string }>(`/api/v1/receipts/${encodeURIComponent(id)}`);
  return response.download_url;
}

async function sha256Hex(file: File) {
  const digest = await crypto.subtle.digest("SHA-256", await file.arrayBuffer());
  return Array.from(new Uint8Array(digest), (byte) => byte.toString(16).padStart(2, "0")).join("");
}
