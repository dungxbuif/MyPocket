import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes } from "react";
import { ChevronDown } from "lucide-react";
import { Heading } from "./Heading";
import { FORM_CONTROL_VARIANTS } from "./tokens";

type FieldProps = { label: string; children: ReactNode };

export function FormField({ label, children }: FieldProps) {
  return <label className="block"><Heading as="span" size="field">{label}</Heading>{children}</label>;
}

type ControlVariant = keyof typeof FORM_CONTROL_VARIANTS;
export function BaseTextInput({ variant = "default", className = "", ...props }: InputHTMLAttributes<HTMLInputElement> & { variant?: ControlVariant }) {
  return <input {...props} className={`${FORM_CONTROL_VARIANTS[variant]} ${className}`} />;
}

export function BaseSelect({ className = "", variant = "default", ...props }: SelectHTMLAttributes<HTMLSelectElement> & { variant?: ControlVariant }) {
  return <span className="relative block"><select {...props} className={`${FORM_CONTROL_VARIANTS[variant]} appearance-none pr-10 ${className}`} /><ChevronDown aria-hidden size={18} className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-muted" /></span>;
}
