import { Text } from "../atoms/Text";
import { BaseButton } from "../atoms/BaseButton";
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
  const className = `z-10 flex w-full items-center justify-between ${root ? "relative gap-3" : `${branchClass} gap-2.5`}`;
  const content = <><span className="grid shrink-0 place-items-center"><IconBadge icon={Icon} size={root ? "md" : "sm"} shape="circle" tone={item.tone ?? (root ? "neutral" : "brand")} /></span>
    <div className="min-w-0 flex-1"><Text size={root ? "sm" : "xs"} weight="bold" tone="heading">{item.name}</Text><Text size={root ? "tiny" : "micro"} weight="medium" tone="muted">{item.subtitle}</Text></div>
    {item.trailing ? <div className="shrink-0">{item.trailing}</div> : null}
    {item.isSystem ? <span className="flex shrink-0 items-center gap-1.5"><LockKeyhole size={13} className="text-muted" /><ChevronRight size={16} className="text-muted" /></span> : null}
  </>;
  return selectable ? <BaseButton variant="row" size="row" type="button" onClick={() => onSelect(item.id)} className={className}>{content}</BaseButton> : <div className={className}>{content}</div>;
}
