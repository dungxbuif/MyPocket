import * as React from "react";
import { cn } from "../../lib/utils";
import { Search, X } from "lucide-react";

export interface SearchInputFieldProps {
  value: string;
  onChange: (val: string) => void;
  onClear?: () => void;
  placeholder?: string;
  autoFocus?: boolean;
  className?: string;
}

export function SearchInputField({
  value,
  onChange,
  onClear,
  placeholder = "Tìm kiếm giao dịch, nhóm...",
  autoFocus,
  className,
}: SearchInputFieldProps) {
  return (
    <div className={cn("relative flex items-center w-full", className)}>
      <Search className="absolute left-3.5 w-4 h-4 text-[#8e8e93] pointer-events-none" />
      <input
        type="text"
        value={value}
        autoFocus={autoFocus}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        className="w-full h-10 pl-10 pr-9 rounded-full bg-[#eef0f4] text-sm text-[#111111] placeholder:text-[#8e8e93] focus:outline-none focus:ring-2 focus:ring-[#2dbd4f] transition-all"
      />
      {value ? (
        <button
          type="button"
          onClick={() => (onClear ? onClear() : onChange(""))}
          className="absolute right-3 w-5 h-5 rounded-full bg-[#8e8e93]/30 text-[#111111] flex items-center justify-center hover:bg-[#8e8e93]/50 transition-colors"
        >
          <X className="w-3 h-3" />
        </button>
      ) : null}
    </div>
  );
}
