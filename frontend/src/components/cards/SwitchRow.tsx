import * as React from "react";
import { cn } from "../../lib/utils";
import { Switch } from "../ui/switch";

export interface SwitchRowProps {
  title: string;
  description?: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
  className?: string;
}

export function SwitchRow({ title, description, checked, onChange, disabled, className }: SwitchRowProps) {
  return (
    <div className={cn("w-full flex items-center justify-between py-3.5 px-2", className)}>
      <div className="flex flex-col pr-4">
        <span className="text-[15px] font-semibold text-[#111111]">{title}</span>
        {description && <span className="text-xs text-[#8e8e93] leading-snug mt-0.5">{description}</span>}
      </div>
      <Switch checked={checked} onCheckedChange={onChange} disabled={disabled} />
    </div>
  );
}
