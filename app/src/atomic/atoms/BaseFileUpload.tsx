import { BaseTextInput, FormField } from "./FormField";

export function BaseFileUpload({ label, disabled, onFiles }: { label: string; disabled?: boolean; onFiles: (files: File[]) => void }) {
  return <FormField label={label}><BaseTextInput type="file" accept="image/jpeg,image/png" multiple disabled={disabled} onChange={event => {
    onFiles(Array.from(event.target.files ?? []));
    event.target.value = "";
  }} /></FormField>;
}
