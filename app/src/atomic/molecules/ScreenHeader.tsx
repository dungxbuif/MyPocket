import type { ReactNode } from "react";
import { Text } from "../atoms/Text";

export function ScreenHeader({ title, subtitle, action }: { title: string; subtitle?: string; action?: ReactNode }) {
  return <div className="flex items-center justify-between gap-3"><div className="min-w-0"><Text as="h1" size="xl" weight="bold">{title}</Text>{subtitle ? <Text size="xs" tone="secondary">{subtitle}</Text> : null}</div>{action}</div>;
}
