import type { EntryDraft } from "./ai";
import type { Category } from "./categories";
import type { Wallet } from "./wallets";
import { categoryAppliesToTransaction } from "./transactionLogic";

export function proposalIssues(draft: EntryDraft, wallets: Wallet[], categories: Category[]): string[] {
  const issues: string[] = [];
  if (draft.type !== "income" && draft.type !== "expense") issues.push(draft.type === "transfer" ? "Chuyển tiền chưa thể duyệt trong phiên bản này." : "Chọn khoản thu hoặc khoản chi, rồi lưu bản nháp trước khi duyệt.");
  if (!Number.isSafeInteger(draft.amount) || draft.amount <= 0) issues.push("Nhập số tiền nguyên lớn hơn 0.");
  const wallet = wallets.find(item => item.id === draft.wallet_id && (item.type === "basic" || item.type === "goal"));
  if (!wallet) issues.push("Chọn ví thường hoặc ví tiết kiệm.");
  if (!draft.occurred_at || !Number.isFinite(Date.parse(draft.occurred_at))) issues.push("Nhập ngày và giờ giao dịch.");
  if (draft.category_id) {
    const category = categories.find(item => item.id === draft.category_id);
    if (!category || (draft.type !== "income" && draft.type !== "expense") || !categoryAppliesToTransaction(category, draft.type, draft.wallet_id, wallet?.type)) issues.push("Chọn lại hoặc bỏ nhóm không phù hợp với loại giao dịch và ví.");
  }
  return issues;
}

export function sameDraft(a: EntryDraft, b: EntryDraft): boolean {
  return (Object.keys(a) as (keyof EntryDraft)[]).every(key => a[key] === b[key]);
}

export function localEntryDate(value: string): string {
  if (!value || !Number.isFinite(Date.parse(value))) return "";
  const date = new Date(value);
  return new Date(date.getTime() - date.getTimezoneOffset() * 60_000).toISOString().slice(0, 16);
}

export function validateImages(files: Pick<File, "type" | "size">[]): string {
	if (files.length > 3) return "Chọn tối đa 3 tệp.";
	if (files.some(file => !["image/jpeg", "image/png", "application/pdf"].includes(file.type))) return "Chỉ hỗ trợ JPEG, PNG hoặc PDF.";
	if (files.some(file => file.size > 5 * 1024 * 1024)) return "Mỗi tệp không được lớn hơn 5 MiB.";
  return "";
}

export async function encodeEntryImages(files: File[]) {
  const error = validateImages(files);
  if (error) throw new Error(error);
  return Promise.all(files.map(file => new Promise<{ name: string; mime_type: string; base64: string }>((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error("Không thể đọc ảnh. Hãy chọn lại."));
    reader.onload = () => resolve({ name: file.name, mime_type: file.type, base64: String(reader.result).split(",")[1] });
    reader.readAsDataURL(file);
  })));
}
