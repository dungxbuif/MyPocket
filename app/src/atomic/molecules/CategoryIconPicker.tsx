import { IconBadge } from "../atoms/IconBadge";
import { IconButton } from "../atoms/IconButton";
import { categoryPresentationFor } from "../atoms/categoryPresentation";

const ICONS = ["tag", "expense_food", "expense_bills", "expense_shopping", "expense_family", "expense_transport", "expense_health", "expense_entertainment", "expense_savings", "expense_travel"] as const;
const ICON_LABELS: Record<(typeof ICONS)[number], string> = { tag: "Nhãn", expense_food: "Ăn uống", expense_bills: "Hóa đơn", expense_shopping: "Mua sắm", expense_family: "Gia đình", expense_transport: "Di chuyển", expense_health: "Sức khỏe", expense_entertainment: "Giải trí", expense_savings: "Tiết kiệm", expense_travel: "Du lịch" };

export function CategoryIconPicker({ value, onChange }: { value: string; onChange: (key: string) => void }) {
  return <div className="grid grid-cols-5 gap-2">{ICONS.map((key) => { const presentation = categoryPresentationFor(key); return <IconButton key={key} label={ICON_LABELS[key]} variant="surface" className={value === key ? "ring-2 ring-emerald-500 ring-offset-2" : ""} onClick={() => onChange(key)}><IconBadge icon={presentation.icon} tone={presentation.tone} size="sm" shape="circle" /></IconButton>; })}</div>;
}
