import { useRef, type ChangeEvent } from "react";
import { Paperclip } from "lucide-react";
import { BaseButton } from "./BaseButton";
import { BaseTextInput, FormField } from "./FormField";

export function BaseFileUpload({ label, disabled, onFiles, variant = "field" }: { label: string; disabled?: boolean; onFiles: (files: File[]) => void; variant?: "field" | "button" }) {
  const input = useRef<HTMLInputElement>(null);
  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    onFiles(Array.from(event.target.files ?? []));
    event.target.value = "";
  };
  if (variant === "button") return <>
    <BaseButton variant="chip" disabled={disabled} aria-label={label} onClick={() => input.current?.click()}><Paperclip aria-hidden size={16} />{label}</BaseButton>
    <input ref={input} className="sr-only" type="file" accept="image/jpeg,image/png,application/pdf" multiple disabled={disabled} aria-label={label} onChange={handleChange} />
  </>;
  return <FormField label={label}><BaseTextInput type="file" accept="image/jpeg,image/png,application/pdf" multiple disabled={disabled} onChange={handleChange} /></FormField>;
}
