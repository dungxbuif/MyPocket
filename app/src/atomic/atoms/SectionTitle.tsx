import { BaseButton } from "./BaseButton";
import { Heading } from "./Heading";

export function SectionTitle({ title, action, onAction }: { title: string; action?: string; onAction?: () => void }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <Heading as="h2" size="section" className="text-secondary">{title}</Heading>
      {action ? <BaseButton variant="ghost" size="sm" onClick={onAction} className="shrink-0 px-2">{action}</BaseButton> : null}
    </div>
  );
}
