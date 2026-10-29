import * as React from "react";
import { cn } from "../../lib/utils";
import { Search, X } from "lucide-react";
import { ActionButton } from '../../app/components';

export interface SearchInputFieldProps {
  value: string;
  onChange: (val: string) => void;
  onClear?: () => void;
  placeholder?: string;
  ariaLabel?: string;
  autoFocus?: boolean;
  className?: string;
}

export function SearchInputField({
  value,
  onChange,
  onClear,
  placeholder = "Tìm kiếm giao dịch, nhóm...",
  ariaLabel = "Tìm kiếm",
  autoFocus,
  className,
}: SearchInputFieldProps) {
  return (
    <div className={cn("relative flex items-center w-full", className)}>
      <Search className="absolute left-3.5 w-4 h-4 text-neutral-500 pointer-events-none" />
      <input
        type="text"
        aria-label={ariaLabel}
        value={value}
        autoFocus={autoFocus}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        className="w-full h-10 pl-10 pr-9 rounded-full bg-neutral-100 text-sm text-black placeholder:text-neutral-500 focus:outline-none focus:ring-2 focus:ring-black transition-all"
      />
      {value ? (
        <ActionButton
          aria-label="Xóa từ khóa tìm kiếm"
          onClick={() => (onClear ? onClear() : onChange(""))}
          className="absolute right-3 w-5 h-5 rounded-full bg-neutral-300 text-black flex items-center justify-center hover:bg-neutral-400 transition-colors"
        >
          <X className="w-3 h-3" />
        </ActionButton>
      ) : null}
    </div>
  );
}
