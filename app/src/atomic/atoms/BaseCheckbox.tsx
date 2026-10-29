import type { ChangeEventHandler, ReactNode } from "react";
import { BASE_COMPONENT_RADIUS } from "./tokens";

export function BaseCheckbox({ checked, disabled = false, label, onChange, children }: { checked: boolean; disabled?: boolean; label: string; onChange: ChangeEventHandler<HTMLInputElement>; children: ReactNode }) {
  return <label className={`flex min-h-11 items-center gap-3 ${BASE_COMPONENT_RADIUS} px-2 py-1 transition ${disabled ? "cursor-not-allowed opacity-70" : "cursor-pointer hover:bg-row"}`}><input aria-label={label} type="checkbox" checked={checked} disabled={disabled} onChange={onChange} className="h-4 w-4 rounded border-muted text-action focus:ring-accent" /><span className="min-w-0 flex-1 text-sm text-ink">{children}</span></label>;
}
