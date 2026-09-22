import { useMemo, useState } from "react";
import { Tags } from "lucide-react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseTextInput } from "../atoms/FormField";
import { StatusMessage } from "../atoms/StatusMessage";
import { categoryPresentationFor } from "../atoms/categoryPresentation";
import { BaseCategoryTree, type BaseCategoryTreeItem } from "./BaseCategoryTree";
import type { Category } from "../../services/categories";

export function CategoryTreeSelector({ categories, selectableIDs, selectedID, onSelect, clearLabel = "Không chọn nhóm" }: {
  categories: Category[]; selectableIDs?: string[]; selectedID?: string; onSelect: (id: string) => void; clearLabel?: string;
}) {
  const [query, setQuery] = useState("");
  const groups = useMemo(() => {
    const selectable = new Set(selectableIDs ?? categories.map(category => category.id));
    const normalized = query.toLocaleLowerCase("vi").trim();
    const byID = new Map(categories.map(category => [category.id, category]));
    const roots = categories.filter(category => !category.parent_id || !byID.has(category.parent_id));
    const result: { root: BaseCategoryTreeItem; children: BaseCategoryTreeItem[] }[] = [];

    for (const category of roots) {
      const children = categories.filter(item => item.parent_id === category.id && selectable.has(item.id));
      const rootMatches = !normalized || category.name.toLocaleLowerCase("vi").includes(normalized);
      const matchingChildren = normalized && !rootMatches
        ? children.filter(item => item.name.toLocaleLowerCase("vi").includes(normalized))
        : children;
      if (!selectable.has(category.id) && matchingChildren.length === 0) continue;
      const presentation = categoryPresentationFor(category.system_key ?? category.icon_key);
      result.push({
        root: {
          id: category.id, name: category.name,
          subtitle: selectable.has(category.id) ? category.is_system ? "Nhóm mặc định" : "Nhóm cá nhân" : "Không áp dụng cho lựa chọn này",
          isSystem: category.is_system, isSelectable: selectable.has(category.id), ...presentation,
        },
        children: matchingChildren.map(child => {
          const childPresentation = categoryPresentationFor(child.system_key ?? child.icon_key);
          return { id: child.id, name: child.name, subtitle: child.is_system ? "Nhóm mặc định" : "Nhóm cá nhân", isSystem: child.is_system, isSelectable: true, ...childPresentation };
        }),
      });
    }
    return result;
  }, [categories, query, selectableIDs]);

  return <div className="space-y-3">
    <BaseTextInput type="search" aria-label="Tìm kiếm nhóm" placeholder="Tìm kiếm nhóm" value={query} onChange={event => setQuery(event.target.value)} />
    <BaseButton variant="ghost" className="w-full justify-start" aria-pressed={!selectedID} onClick={() => onSelect("")}>{clearLabel}</BaseButton>
    {groups.map(group => <BaseCategoryTree key={group.root.id} mode="selection" selectedID={selectedID} root={group.root} children={group.children} onSelect={onSelect} />)}
    {groups.length === 0 ? <StatusMessage variant="plain">Không có nhóm phù hợp.</StatusMessage> : null}
  </div>;
}
