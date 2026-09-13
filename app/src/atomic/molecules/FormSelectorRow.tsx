import { Text } from "../atoms/Text";
import { BaseButton } from "../atoms/BaseButton";
import type { LucideIcon } from "lucide-react";
import { ChevronRight } from "lucide-react";
import { IconBadge } from "../atoms/IconBadge";

export function FormSelectorRow({ label, value, icon: Icon }: { label: string; value: string; icon: LucideIcon }) {
  return (
    <BaseButton variant="row" size="row" type="button" className="flex w-full items-center gap-3">
      <IconBadge icon={Icon} shape="circle" tone="success" />
      <div className="min-w-0 flex-1">
        <Text size="xs" tone="secondary" className="">{label}</Text>
        <Text weight="semibold" className="truncate">{value}</Text>
      </div>
      <ChevronRight size={17} className="text-secondary" />
    </BaseButton>
  );
}
