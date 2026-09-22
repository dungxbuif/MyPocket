# BaseFileUpload

Consumer: [AI entry](../../screens/assistant/README.md). Implementation: `app/src/atomic/atoms/BaseFileUpload.tsx`.

Native labeled file input, composed with FormField/BaseTextInput, owns browser file picking, keyboard access, and drag/drop. Props: label, disabled, onFiles, variant, and maxFiles. Multiple selection accepts JPEG/PNG/PDF. The consumer validates at most 20 files, each at most 5 MiB, before reading or sending; duplicate name/type/size selections are rejected. Change clears the native input value so the same file can be selected again. Disabled forwards to the input and drop handlers. Uses the existing form-control tokens; no new visual palette. Selected-file summary and errors belong to the consumer. Proof: `node scripts/ai-attachments.test.mjs` and `node scripts/ai-entry.test.mjs`.
