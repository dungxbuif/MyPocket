import { Text } from "../atoms/Text";
import { BaseButton } from "../atoms/BaseButton";
import type { LucideIcon } from "lucide-react";
import { ChevronRight } from "lucide-react";
import { IconBadge } from "../atoms/IconBadge";
import type { BadgeTone } from "../atoms/tokens";
export function FormSelectorRow({ label, value, icon, onClick, disabled, placeholder = false, tone = "success" }: { label: string; value: string; icon: LucideIcon; onClick?: () => void; disabled?: boolean; placeholder?: boolean; tone?: BadgeTone }) {
  return <BaseButton variant="row" size="row" aria-label={label} disabled={disabled} onClick={onClick} className="w-full gap-3">
    <IconBadge icon={icon} shape="circle" tone={tone} />
    <Text as="span" size="base" tone={placeholder ? "muted" : "ink"} className="min-w-0 flex-1 truncate">{value}</Text>
    <ChevronRight size={18} aria-hidden />
  </BaseButton>;
}
