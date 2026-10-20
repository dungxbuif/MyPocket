import { removeReceiptUpload, readPendingReceiptUploads } from "./db";
import type { OfflineReceiptUpload } from "./types";

export async function listPendingReceiptUploads() {
  return readPendingReceiptUploads();
}

export async function markReceiptUploadComplete(record: OfflineReceiptUpload) {
  await removeReceiptUpload(record.id);
}
