# Agent and Receipt UI — Completion Evidence

Date: 2026-09-11

- Added a first-class `Trợ lý` destination for text analysis and review-first transaction proposals.
- Optional camera/library image selection uses the shared `FilePickerInput`; client validation rejects non-images and files over 15 MiB before upload.
- Networking, receipt upload, polling, and provider-independent state mapping live in `app/agent.ts`; rendering is split into focused base-composed components.
- Agent/provider text is rendered as React text, never injected HTML.
- Queued, submitting, processing, completed, failed, cancelled, and expired states have stable Vietnamese labels.
- Completed transaction proposals expose a deliberate `Mở bản nháp` action. Editing, confirmation, and rejection remain in the existing tested draft decision component.
- Offline mode keeps the composer text and disables submission rather than pretending agent/OCR is available.

Verification:

```text
npm test -- --run
PASS: 22 files, 175 tests

npm run build
PASS

npm run test:e2e -- agent.spec.ts agent-image.spec.ts
PASS: 6/6 across Chromium desktop, Chromium mobile, and mobile WebKit
```
