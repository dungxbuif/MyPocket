import { SurfaceCard } from "../atoms/SurfaceCard";
import { IconBadge } from "../atoms/IconBadge";
import { Progress } from "../atoms/Progress";
import { Text } from "../atoms/Text";
import type { MockGoal } from "../data/mockFinance";
import { formatVND, ratioPercent } from "../utils/format";

export function GoalCard({ goal, masked }: { goal: MockGoal; masked: boolean }) {
  const Icon = goal.icon;
  const progress = ratioPercent(goal.saved, goal.target);
  return (
    <SurfaceCard padding="sm" tone="muted" elevation="flat" className="">
      <div className="flex items-center gap-3">
        <IconBadge icon={Icon} shape="circle" tone="success" />
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-2">
            <Text weight="semibold" className="truncate">{goal.name}</Text>
            <Text size="sm" weight="bold" tone="action" className="">{progress}%</Text>
          </div>
          <Text numeric size="xs" tone="secondary" className="">
            {masked ? "••••••" : `${formatVND(goal.saved)} / ${formatVND(goal.target)}`}
          </Text>
        </div>
      </div>
      <Progress value={progress} label={goal.name} className="mt-3" />
    </SurfaceCard>
  );
}
