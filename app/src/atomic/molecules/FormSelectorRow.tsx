import type { LucideIcon } from "lucide-react";
import { ChevronRight } from "lucide-react";

export function FormSelectorRow({ label, value, icon: Icon }: { label: string; value: string; icon: LucideIcon }) {
  return (
    <button type="button" className="flex w-full cursor-pointer items-center gap-3 rounded-xl p-3 text-left transition hover:bg-[#f5f3f3]">
      <div className="grid h-9 w-9 place-items-center rounded-full bg-[#d9e6da] text-[#006e1c]">
        <Icon size={17} />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-xs text-[#3f4a3c]">{label}</p>
        <p className="truncate font-semibold">{value}</p>
      </div>
      <ChevronRight size={17} className="text-[#6f7a6b]" />
    </button>
  );
}
