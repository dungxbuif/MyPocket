import type { ReactNode } from "react";
import { Heading } from "../atoms/Heading";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";

export function AssistantResultCard({ count, action, children }: { count: number; action: ReactNode; children: ReactNode }) {
  return <SurfaceCard aria-label="Kết quả AI" padding="md" tone="muted" className="space-y-3">
    <div className="space-y-2">
      <div><Heading as="h3" size="field">AI đã nhận diện {count} giao dịch</Heading><Text size="xs" tone="secondary">Kiểm tra thông tin rồi chọn giao dịch cần lưu.</Text></div>
      {action}
    </div>
    <div className="space-y-3">{children}</div>
  </SurfaceCard>;
}
