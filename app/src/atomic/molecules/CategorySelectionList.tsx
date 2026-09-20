import { useState } from "react";
import { Tags } from "lucide-react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseTextInput } from "../atoms/FormField";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { IconBadge } from "../atoms/IconBadge";
import { Text } from "../atoms/Text";
import { StatusMessage } from "../atoms/StatusMessage";
import { categoryPresentationFor } from "../atoms/categoryPresentation";
import type { Category } from "../../services/categories";

export function CategorySelectionList({ categories, onSelect }: {categories: Category[]; onSelect: (id: string) => void}) {
  const [query,setQuery] = useState("");
  const rows = categories.filter(category => category.name.toLocaleLowerCase("vi").includes(query.toLocaleLowerCase("vi").trim()));
  return <div className="space-y-3"><BaseTextInput type="search" aria-label="Tìm kiếm nhóm" placeholder="Tìm kiếm" value={query} onChange={event => setQuery(event.target.value)} />
    <BaseButton variant="ghost" onClick={() => onSelect("")}>Không chọn nhóm</BaseButton>
    {rows.map(category => { const display = categoryPresentationFor(category.system_key ?? category.icon_key); return <SurfaceCard key={category.id} padding="sm"><BaseButton variant="row" size="row" onClick={() => onSelect(category.id)}><IconBadge icon={display?.icon ?? Tags} tone={display?.tone ?? "neutral"} shape="circle" size="lg" /><Text as="span" size="base">{category.name}</Text></BaseButton></SurfaceCard>; })}
    {rows.length === 0 ? <StatusMessage variant="plain">Không có nhóm phù hợp.</StatusMessage> : null}
  </div>;
}
