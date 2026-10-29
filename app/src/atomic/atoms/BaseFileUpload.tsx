import { useRef, useState, type ChangeEvent, type DragEvent } from "react";
import { Paperclip } from "lucide-react";
import { BaseButton } from "./BaseButton";
import { BaseTextInput, FormField } from "./FormField";
import { Text } from "./Text";

export function BaseFileUpload({ label, disabled, onFiles, variant = "field", maxFiles = 20 }: { label: string; disabled?: boolean; onFiles: (files: File[]) => void; variant?: "field" | "button"; maxFiles?: number }) {
  const fileInput = useRef<HTMLInputElement>(null);
  const [dragging, setDragging] = useState(false);
  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    onFiles(Array.from(event.target.files ?? []).slice(0, maxFiles));
    event.target.value = "";
  };
  const handleDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setDragging(false);
    if (!disabled) onFiles(Array.from(event.dataTransfer.files).slice(0, maxFiles));
  };
  const inputElement = <input ref={fileInput} className="sr-only" type="file" accept="image/jpeg,image/png,application/pdf" multiple disabled={disabled} aria-label={label} onChange={handleChange} />;
  if (variant === "button") return <div data-dropzone="true" className={dragging ? "rounded-control ring-2 ring-accent" : ""} onDragOver={event => { event.preventDefault(); if (!disabled) setDragging(true); }} onDragLeave={() => setDragging(false)} onDrop={handleDrop}>
    <BaseButton variant="chip" disabled={disabled} aria-label={label} onClick={() => fileInput.current?.click()}><Paperclip aria-hidden size={16} />{label}</BaseButton>
    <Text size="micro" tone="secondary" className="sr-only">Kéo thả tệp vào đây, tối đa {maxFiles} tệp.</Text>
    {inputElement}
  </div>;
  return <FormField label={label}><div data-dropzone="true" className={dragging ? "rounded-control ring-2 ring-accent" : ""} onDragOver={event => { event.preventDefault(); if (!disabled) setDragging(true); }} onDragLeave={() => setDragging(false)} onDrop={handleDrop}><BaseTextInput type="file" accept="image/jpeg,image/png,application/pdf" multiple disabled={disabled} onChange={handleChange} /><Text size="micro" tone="secondary">Kéo thả tệp vào đây, tối đa {maxFiles} tệp.</Text></div></FormField>;
}
