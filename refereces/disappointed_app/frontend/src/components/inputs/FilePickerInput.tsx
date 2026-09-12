import { forwardRef, type InputHTMLAttributes } from 'react';

type FilePickerInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, 'type' | 'multiple' | 'value' | 'defaultValue' | 'onChange'> & {
  onFileSelected: (file: File) => void;
};

// Single-file selection: cancellation is not removal; allow choosing the same
// file again after the consumer explicitly removes or replaces its selection.
export const FilePickerInput = forwardRef<HTMLInputElement, FilePickerInputProps>(function FilePickerInput(
  { onFileSelected, disabled, ...props }, ref,
) {
  return <input {...props} ref={ref} type="file" disabled={disabled} onChange={event => {
    const file = event.currentTarget.files?.[0];
    event.currentTarget.value = '';
    if (file && !disabled) onFileSelected(file);
  }} />;
});
