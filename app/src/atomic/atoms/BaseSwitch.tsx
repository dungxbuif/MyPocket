import type { InputHTMLAttributes } from "react";
import { Text } from "./Text";

export function BaseSwitch({ label, ...props }: Omit<InputHTMLAttributes<HTMLInputElement>, "type" | "role" | "className"> & { label: string }) {
  return <label className="flex min-h-11 cursor-pointer items-center justify-between gap-4">
    <Text as="span" weight="medium">{label}</Text>
    <span className="relative flex h-11 w-12 shrink-0 items-center">
      <input {...props} type="checkbox" role="switch" aria-label={label} className="peer sr-only" />
      <span aria-hidden="true" className="h-7 w-12 rounded-full bg-line transition peer-checked:bg-accent peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-action peer-disabled:opacity-50" />
      <span aria-hidden="true" className="pointer-events-none absolute left-0.5 h-6 w-6 rounded-full bg-card shadow-control transition-transform peer-checked:translate-x-5" />
    </span>
  </label>;
}
