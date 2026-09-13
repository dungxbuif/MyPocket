import type { ChangeEventHandler, ReactNode } from "react";
import { BASE_COMPONENT_RADIUS } from "./tokens";

export function BaseCheckbox({ checked, label, onChange, children }: { checked: boolean; label: string; onChange: ChangeEventHandler<HTMLInputElement>; children: ReactNode }) {
  return <label className={`flex min-h-11 cursor-pointer items-center gap-3 ${BASE_COMPONENT_RADIUS} px-2 py-1 transition hover:bg-slate-50`}><input aria-label={label} type="checkbox" checked={checked} onChange={onChange} className="h-4 w-4 rounded border-slate-300 text-emerald-600 focus:ring-emerald-500" /><span className="min-w-0 flex-1 text-sm text-slate-700">{children}</span></label>;
}
