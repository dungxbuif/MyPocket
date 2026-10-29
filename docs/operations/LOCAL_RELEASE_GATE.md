# Local release gate

The web release gate is intentionally runnable without production credentials
or a live AI provider:

```sh
cd app
npm run check:design
npm run typecheck
npm run test:release-gate
npm run build
```

`test:release-gate` checks the PWA manifest, service-worker registration and
install prompt contract, then runs the AI entry unit/browser fixture. The AI
fixture verifies the one-shot review flow, proposal edits/approval/rejection,
twenty-file attachments, shared controls and mobile-sized interaction state.
It does not claim provider quality or OCR availability; those require the
separate live-provider checklist and configured staging credentials.

The install prompt stores the deferred browser event only after
`beforeinstallprompt`, calls `prompt()` only from the user's click, and keeps
API routes out of the service-worker cache. A browser that does not expose the
install event simply renders no prompt.
