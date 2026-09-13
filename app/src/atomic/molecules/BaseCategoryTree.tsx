import { ChevronRight, FolderTree, LockKeyhole, type LucideIcon } from "lucide-react";
import type { ReactNode } from "react";
import { IconBadge } from "../atoms/IconBadge";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { CATEGORY_TREE_LAYOUTS } from "../atoms/tokens";
import type { ICON_BADGE_TONES } from "../atoms/tokens";

export type BaseCategoryTreeItem = { id: string; name: string; subtitle: string; isSystem: boolean; isEditable?: boolean; icon?: LucideIcon; tone?: keyof typeof ICON_BADGE_TONES; trailing?: ReactNode };

type BaseCategoryTreeProps = {
  root: BaseCategoryTreeItem;
  children: BaseCategoryTreeItem[];
  layout?: keyof typeof CATEGORY_TREE_LAYOUTS;
  onSelect?: (id: string) => void;
};

export function BaseCategoryTree({ root, children, layout = "nested", onSelect }: BaseCategoryTreeProps) {
  const styles = CATEGORY_TREE_LAYOUTS[layout];
  const hasChildren = children.length > 0;

  return <SurfaceCard padding={hasChildren ? "sm" : "none"} className={`relative overflow-hidden ${hasChildren ? "" : "px-3"}`}>
    {hasChildren ? <span aria-hidden className={styles.connector} /> : null}
    <TreeRow item={root} root onSelect={onSelect} branchClass={styles.branch} />
    {hasChildren ? <div className={styles.children}>{children.map((item) => <TreeRow key={item.id} item={item} onSelect={onSelect} branchClass={styles.branch} />)}</div> : null}
  </SurfaceCard>;
}

function TreeRow({ item, root = false, onSelect, branchClass }: {
  item: BaseCategoryTreeItem; root?: boolean; onSelect?: (id: string) => void; branchClass: string;
}) {
  const Icon = item.icon ?? FolderTree;
  const selectable = item.isEditable && onSelect;
  const className = `z-10 flex w-full items-center justify-between rounded-xl text-left transition ${root ? "relative gap-3 p-1.5" : `${branchClass} gap-2.5 pl-2 pr-1.5 py-1.5`} ${selectable ? "cursor-pointer hover:bg-slate-50" : ""}`;
  const content = <><span className="grid shrink-0 place-items-center"><IconBadge icon={Icon} size={root ? "md" : "sm"} shape="circle" tone={item.tone ?? (root ? "neutral" : "brand")} /></span>
    <div className="min-w-0 flex-1"><p className={`${root ? "text-sm" : "text-xs"} font-bold text-slate-900`}>{item.name}</p><p className={`${root ? "text-[11px]" : "text-[10px]"} font-medium text-slate-400`}>{item.subtitle}</p></div>
    {item.trailing ? <div className="shrink-0">{item.trailing}</div> : null}
    {item.isSystem ? <span className="flex shrink-0 items-center gap-1.5"><LockKeyhole size={13} className="text-slate-400" /><ChevronRight size={16} className="text-slate-300" /></span> : null}
  </>;
  return selectable ? <button type="button" onClick={() => onSelect(item.id)} className={className}>{content}</button> : <div className={className}>{content}</div>;
}
