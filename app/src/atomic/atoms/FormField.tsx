import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes } from "react";
import { ChevronDown } from "lucide-react";
import { Heading } from "./Heading";
import { FORM_CONTROL_CLASS } from "./tokens";

type FieldProps = { label: string; children: ReactNode };

export function FormField({ label, children }: FieldProps) {
  return <label className="block"><Heading as="span" size="field">{label}</Heading>{children}</label>;
}

export function BaseTextInput(props: InputHTMLAttributes<HTMLInputElement>) {
  return <input {...props} className={`${FORM_CONTROL_CLASS} ${props.className ?? ""}`} />;
}

export function BaseSelect({ className = "", ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return <span className="relative block"><select {...props} className={`${FORM_CONTROL_CLASS} appearance-none pr-10 ${className}`} /><ChevronDown aria-hidden size={18} className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-slate-400" /></span>;
}
