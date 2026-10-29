import { Text } from "../atoms/Text";
import type { LucideIcon } from "lucide-react";
import { formatVND, ratioPercent } from "../utils/format";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { IconBadge } from "../atoms/IconBadge";
import { Progress } from "../atoms/Progress";

export function BudgetProgressItem({ budget, masked = false }: { budget: {name:string; spent:number; limit:number; icon:LucideIcon}; masked?:boolean }) {
  const Icon = budget.icon;
  const progress = ratioPercent(budget.spent, budget.limit);
  const over = budget.spent > budget.limit;
  return (
    <SurfaceCard padding="sm" tone="muted" elevation="flat">
      <div className="flex items-center gap-3">
        <IconBadge icon={Icon} shape="circle" tone={over ? "danger" : "success"} />
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-2">
            <Text weight="semibold" className="">{budget.name}</Text>
            <Text weight="bold" tone={over ? "danger" : "ink"}>{progress}%</Text>
          </div>
          <Text numeric size="xs" tone="secondary">
            {masked ? "•••••• / ••••••" : `${formatVND(budget.spent)} / ${formatVND(budget.limit)}`}
          </Text>
        </div>
      </div>
      <Progress value={progress} label={budget.name} danger={over} className="mt-3" />
    </SurfaceCard>
  );
}
