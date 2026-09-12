import type { AgentRunStatus as Status } from "../../app/agent";
import { Badge } from "../ui/badge";

const labels: Record<Status, string> = { queued: "Đang chờ", submitting: "Đang gửi ảnh", processing: "Đang xử lý", completed: "Hoàn tất", failed: "Không thành công", cancelled: "Đã hủy", expired: "Đã hết hạn" };

export function AgentRunStatus({ status }: { status: Status }) {
  return <div role="status" aria-live="polite" className="agent-run-status"><Badge>{labels[status]}</Badge></div>;
}
