import { apiFetch } from "./apiClient";

export type LifecycleJob = { id: string; kind: "import"|"export"|"reset"|"delete"; status: "queued"|"running"|"awaiting_confirmation"|"completed"|"failed"; result?: { valid_rows?: unknown[]; errors?: {row:number;message:string}[]; confirmable?: boolean; imported_rows?: number }; version: number; error_code?: string };
export type DestructivePreview = { preview_token: string; affected_counts: Record<string,number>; expires_at: string };
const idempotency = () => crypto.randomUUID();

export async function createExport(){return (await apiFetch<{job:LifecycleJob}>("/api/v1/exports",{method:"POST",headers:{"Content-Type":"application/json","Idempotency-Key":idempotency()},body:JSON.stringify({datasets:["wallets","categories","transactions"]})})).job}
export async function uploadImport(file:File){return (await apiFetch<{job:LifecycleJob}>("/api/v1/imports",{method:"POST",headers:{"Content-Type":"text/csv","Idempotency-Key":idempotency()},body:file})).job}
export async function getLifecycleJob(job:LifecycleJob){const root=job.kind==="import"?"imports":job.kind==="export"?"exports":"account/jobs";return (await apiFetch<{job:LifecycleJob}>(`/api/v1/${root}/${job.id}`)).job}
export async function confirmImport(job:LifecycleJob){return (await apiFetch<{job:LifecycleJob}>(`/api/v1/imports/${job.id}/confirm`,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({version:job.version})})).job}
export async function getExportDownload(job:LifecycleJob){return (await apiFetch<{download_url:string}>(`/api/v1/exports/${job.id}/download`)).download_url}
export async function previewDestructive(kind:"reset"|"delete"){return (await apiFetch<{preview:DestructivePreview}>(`/api/v1/account/${kind}`,{method:"POST",headers:{"Content-Type":"application/json"},body:"{}"})).preview}
export async function confirmDestructive(kind:"reset"|"delete",confirmation:string,preview_token:string){return (await apiFetch<{job:LifecycleJob}>(`/api/v1/account/${kind}`,{method:"POST",headers:{"Content-Type":"application/json","Idempotency-Key":idempotency()},body:JSON.stringify({confirmation,preview_token})})).job}
