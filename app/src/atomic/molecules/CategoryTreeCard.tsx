import type { MockCategoryNode } from "../data/mockFinance";
import { formatVND } from "../utils/format";

export function CategoryTreeCard({ node, masked }: { node: MockCategoryNode; masked: boolean }) {
  const Icon = node.icon;
  return (
    <div className="rounded-2xl bg-[#f5f3f3] p-3">
      <div className="flex items-center gap-3">
        <div className="grid h-10 w-10 place-items-center rounded-full bg-[#d9e6da] text-[#006e1c]">
          <Icon size={18} />
        </div>
        <div className="min-w-0 flex-1">
          <p className="font-semibold">{node.name}</p>
          <p className="text-xs text-[#3f4a3c]">{node.children?.length ?? 0} nhóm con</p>
        </div>
        <p className="money text-sm font-bold text-[#1b1c1c]">{masked ? "••••••" : formatVND(node.amount)}</p>
      </div>
      {node.children?.length ? (
        <div className="relative ml-5 mt-2 space-y-2 border-l border-[#becab9] pl-5">
          {node.children.map((child) => {
            const ChildIcon = child.icon;
            return (
              <div key={child.id} className="relative flex items-center gap-3 rounded-xl bg-white p-2">
                <span className="absolute -left-5 top-1/2 h-px w-4 bg-[#becab9]" />
                <div className="grid h-8 w-8 place-items-center rounded-full bg-[#efeded] text-[#3f4a3c]">
                  <ChildIcon size={15} />
                </div>
                <p className="min-w-0 flex-1 truncate text-sm font-semibold">{child.name}</p>
                <p className="money text-xs font-bold text-[#3f4a3c]">{masked ? "••••••" : formatVND(child.amount)}</p>
              </div>
            );
          })}
        </div>
      ) : null}
    </div>
  );
}
