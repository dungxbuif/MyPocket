import type { MockGoal } from "../data/mockFinance";
import { formatVND, ratioPercent } from "../utils/format";

export function GoalCard({ goal, masked }: { goal: MockGoal; masked: boolean }) {
  const Icon = goal.icon;
  const progress = ratioPercent(goal.saved, goal.target);
  return (
    <div className="rounded-xl bg-[#f5f3f3] p-3">
      <div className="flex items-center gap-3">
        <div className="grid h-10 w-10 place-items-center rounded-full bg-[#d9e6da] text-[#006e1c]">
          <Icon size={18} />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-2">
            <p className="truncate font-semibold">{goal.name}</p>
            <p className="text-sm font-bold text-[#006e1c]">{progress}%</p>
          </div>
          <p className="money text-xs text-[#3f4a3c]">
            {masked ? "••••••" : `${formatVND(goal.saved)} / ${formatVND(goal.target)}`}
          </p>
        </div>
      </div>
      <div className="mt-3 h-2 overflow-hidden rounded-full bg-[#e3e2e2]">
        <div className="h-full rounded-full bg-[#006e1c]" style={{ width: `${progress}%` }} />
      </div>
    </div>
  );
}
