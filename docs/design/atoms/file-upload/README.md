# BaseFileUpload

Consumer: [AI entry](../../screens/assistant/README.md). Implementation: `app/src/atomic/atoms/BaseFileUpload.tsx`.

Native labeled file input, composed with FormField/BaseTextInput, owns browser file picking and keyboard access. Props: label, disabled, onFiles. Multiple selection accepts JPEG/PNG. The consumer validates at most three files, each at most 5 MiB, before reading or sending. Change clears the native input value so the same file can be selected again. Disabled forwards to the input. Uses the existing form-control tokens; no new visual palette. Selected-file summary and errors belong to the consumer. Proof: `node scripts/ai-entry.test.mjs`.
