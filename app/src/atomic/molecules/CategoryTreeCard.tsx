import { useState } from "react";
import { ChevronRight } from "lucide-react";
import type { MockCategoryNode } from "../data/mockFinance";
import { formatVND } from "../utils/format";
import { CATEGORY_TREE_CLASSES } from "../atoms/tokens";

export function CategoryTreeCard({ node, masked }: { node: MockCategoryNode; masked: boolean }) {
  const [expanded, setExpanded] = useState(true);
  const Icon = node.icon;
  return (
    <article className={CATEGORY_TREE_CLASSES.card}>
      <button type="button" className={CATEGORY_TREE_CLASSES.parent} onClick={() => setExpanded((value) => !value)} aria-expanded={expanded}>
        <div className="flex min-w-0 items-center gap-3">
          <div className={CATEGORY_TREE_CLASSES.parentIcon}>
          <Icon size={18} />
          </div>
          <div className="min-w-0 text-left">
            <p className="truncate text-sm font-bold text-[#1b1c1c]">{node.name}</p>
            <p className="text-[11px] font-medium text-[#6f7a6b]">Hoạt động trong {node.children?.length ?? 0} nhóm</p>
          </div>
        </div>
        <div className="flex items-center gap-2 pl-2">
          <p className="money text-sm font-bold text-[#1b1c1c]">{masked ? "••••••" : formatVND(node.amount)}</p>
          <ChevronRight size={17} className={`text-[#6f7a6b] transition ${expanded ? "rotate-90" : ""}`} />
        </div>
      </button>
      {expanded && node.children?.length ? (
        <div className={CATEGORY_TREE_CLASSES.children}>
          {node.children.map((child) => {
            const ChildIcon = child.icon;
            return (
              <div key={child.id} className={CATEGORY_TREE_CLASSES.child}>
                <span aria-hidden="true" className="absolute -left-4 top-1/2 h-3 w-3 -translate-y-[90%] rounded-bl-lg border-b-2 border-l-2 border-[#e3e2e2]" />
                <div className={CATEGORY_TREE_CLASSES.childIcon}>
                  <ChildIcon size={15} />
                </div>
                <p className="min-w-0 flex-1 truncate px-2 text-xs font-semibold text-[#1b1c1c]">{child.name}</p>
                <p className="money text-xs font-bold text-[#3f4a3c]">{masked ? "••••••" : formatVND(child.amount)}</p>
              </div>
            );
          })}
        </div>
      ) : null}
    </article>
  );
}
