import { ChevronDown } from "lucide-react";
import { useState } from "react";
import { IconBadge } from "../atoms/IconBadge";
import { IconButton } from "../atoms/IconButton";
import { categoryPresentationFor } from "../atoms/categoryPresentation";
import { BaseBottomSheet } from "./BaseBottomSheet";

const ICONS = ["tag", "expense_food", "expense_bills", "expense_shopping", "expense_family", "expense_transport", "expense_health", "expense_entertainment", "expense_savings", "expense_travel"] as const;
const ICON_LABELS: Record<(typeof ICONS)[number], string> = { tag: "Nhãn", expense_food: "Ăn uống", expense_bills: "Hóa đơn", expense_shopping: "Mua sắm", expense_family: "Gia đình", expense_transport: "Di chuyển", expense_health: "Sức khỏe", expense_entertainment: "Giải trí", expense_savings: "Tiết kiệm", expense_travel: "Du lịch" };
const PICKER_TEXT = { title: "Chọn biểu tượng", close: "Đóng chọn biểu tượng" } as const;

export function CategoryIconPicker({ value, onChange, disabled = false, compact = false }: { value: string; onChange: (key: string) => void; disabled?: boolean; compact?: boolean }) {
  const [expanded, setExpanded] = useState(false);
  const selected = categoryPresentationFor(value);
  const select = (key: string) => { onChange(key); setExpanded(false); };
  return <div className={compact ? "" : "mt-1.5"}><IconButton label={ICON_LABELS[value as keyof typeof ICON_LABELS] ?? ICON_LABELS.tag} variant={compact ? "bare" : "surface"} disabled={disabled} className={compact ? "h-11 w-11" : "h-11 w-full grid-cols-[32px_1fr_20px] justify-items-start px-2"} onClick={() => setExpanded(true)}><IconBadge icon={selected.icon} tone={selected.tone} size={compact ? "md" : "sm"} shape="circle" />{compact ? null : <><span className="text-sm font-semibold text-ink">{ICON_LABELS[value as keyof typeof ICON_LABELS] ?? ICON_LABELS.tag}</span><ChevronDown size={17} className="justify-self-end text-muted" /></>}</IconButton>{expanded ? <BaseBottomSheet title={PICKER_TEXT.title} closeLabel={PICKER_TEXT.close} onClose={() => setExpanded(false)}><div className="grid grid-cols-5 gap-3 pb-2">{ICONS.map((key) => { const presentation = categoryPresentationFor(key); return <IconButton key={key} label={ICON_LABELS[key]} variant="surface" selected={value === key} onClick={() => select(key)}><IconBadge icon={presentation.icon} tone={presentation.tone} size="sm" shape="circle" /></IconButton>; })}</div></BaseBottomSheet> : null}</div>;
}
