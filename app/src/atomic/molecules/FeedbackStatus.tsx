import { Text } from "../atoms/Text";
import type { FeedbackStatus as Status } from "../../services/feedback";

const labels: Record<Status, string> = {
  open: "Đã gửi",
  triaged: "Đã phân loại",
  in_progress: "Đang xử lý",
  fixed: "Đã xử lý",
  rejected: "Không thực hiện",
};

export function FeedbackStatus({ status, changelogVersion }: { status: Status; changelogVersion?: string }) {
  return <div className="space-y-1"><Text size="sm" weight="semibold">{labels[status]}</Text>{status === "fixed" && changelogVersion ? <Text size="xs" tone="secondary">Fixed in {changelogVersion}</Text> : null}</div>;
}
