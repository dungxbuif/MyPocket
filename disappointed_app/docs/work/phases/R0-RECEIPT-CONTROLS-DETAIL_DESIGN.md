# R0 — existing receipt picker and truthful form controls

Status: blocked on WebKit receipt readback; selected controls implemented, not accepted as a complete receipt flow. Approval: owner said “Tiếp” after the preceding handoff explicitly proposed connecting the receipt footer and handling inert controls. This is a bounded continuation of [R0](R0-DETAIL_DESIGN.md), not a redesign or promotion of deferred capabilities. [Verification and stop record](../test-verification/R0-RECEIPT-CONTROLS-2026-09-09.md).

Trace: [backlog](../BACKLOG.md), [action inventory](../test-verification/R0-MOUNTED-ACTIONS-2026-09-09.md), [previous error-feedback evidence](../test-verification/R0-ERROR-FEEDBACK-2026-09-09.md), [TICKET-028 media foundation](../tickets/TICKET-028-production-hardening-debug-audit-foundation.md), [API](../../architecture/API.md), [validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md).

## Context and root cause

Inspected current context/backlog/standards, R0 parent, TICKET-028/media design, mounted `App.tsx` AddTransactionSheet, shared actions, `receipts.ts` and offline receipt adapter. The footer image button has no click handler or accessible name. The only input is mounted inside the expandable details panel. Four `SheetRow` controls have no callbacks despite button/chevron affordances. The debt path returns before processing receipts, so exposing attachments without a debt guard would silently discard selected files.

## Implementation boundary

- Keep current shell/theme, form classes, upload and offline queue contracts. Work in the existing dirty checkout as the parent design directs; preserve unrelated edits.
- Add a reusable native file-input base. Mount one receipt input for the entire form, independent of expanded details. Both existing image actions use the same ref through shared action buttons. Preserve `receipt-image` ID compatibility and current JPEG/PNG/WebP accept filter. Do not introduce browser-specific camera assumptions.
- Show the selected filename outside the collapsible details; an explicit remove action clears it. Cancelling selection retains the previous file. Reset the native input after a choice so the same file can be chosen again. Lock picker/removal while saving, read-only, or after a confirmed save.
- Debt has no attachment API here: disable picker in debt mode, explain that only income/expense supports it, and prevent debt save while an earlier selected image remains. Let the user switch back or explicitly remove it; never silently discard it.
- Replace four inert details controls with one shared disabled-action base and visible “Chưa hỗ trợ trong form này” state. Do not claim the underlying event domain is unavailable: this is only the quick-add form linkage. No new location/contact/reminder/event features in this slice.

## Alternatives and impact

Rejected: duplicate file inputs/state for each button; opening a new prototype attachment screen; automatic image removal when switching type; pretending unimplemented actions are functional; changing backend/provider/auth/runtime contracts.

Touched: mounted AddTransactionSheet, small shared input/feedback components, tests and docs. No dependencies, API/schema/security/deployment or financial accounting changes. Existing online upload may still fail when S3 is unconfigured; preserve input through the previously added error boundary. Remote storage/OCR and physical camera/library permissions remain external UAT.

## Verification and reconciliation

TDD: footer selection without expanding details; selection retained on collapse/cancel; explicit remove and same-file reselection; read-only/pending/debt guard; unsupported controls disabled with explanations; offline receipt linked to the actual queued transaction. Live browser filechooser proof on Chrome mobile/desktop and WebKit with isolated fixture DB; no real user files or production mutations. Run complete frontend suite, typecheck/build, existing Chrome E2E and read-only review. Do not alter the known WebKit offline-emulation test.

Update this design, action inventory, verification record, context/backlog/validation/changelog. Requirements restored within current receipt capability; API/ERD/architecture/SDD/ADR unchanged with reasons recorded. Full R0 and human device/provider acceptance remain open.

## Implementation refinement and stop gate

### Approved follow-up: isolate offline emulation (2026-09-09)

Owner said “Tiếp” after the explicit request for a focused storage-design review. This resumes bounded investigation, not authorization for a storage migration or production release. Receipt acceptance remains blocked until the residual device gate is resolved; investigation is in progress.

Outcome: focused investigation is now **in_review**; see [readback verification](../test-verification/R0-RECEIPT-READBACK-2026-09-09.md). No app/storage changes. HTTP-outage proof passes with service workers blocked only for that separate WebKit scenario; original emulation failure remains active. Latest receipt suite 8 pass/1 fail, full frontend 92 pass and Chrome regression 32 pass.

New minimal-fixture evidence: persistent WebKit reads File/Blob/ArrayBuffer correctly with input reset both enabled and disabled. With `context.setOffline(true)`, even reading the original selected File fails before any write. Reading the **same already-stored File**, with no rewrite, gives pass → `NotReadableError` → pass when toggling offline false → true → false. Thus the reproduced failure does not require MyPocket, input reset or corrupted stored data. This identifies an offline-emulation boundary in this environment, not physical Safari support.

Plan: retain the original failing network-emulation E2E unchanged in intent. Add a separately named HTTP-outage test using real route aborts plus a synthetic browser offline signal, preserving exact image-byte/entity assertions and verifying HTTP really fails. Reuse a disposable persistent WebKit profile for its known ephemeral File/Blob write limitation. No app, storage representation, schema, API, runtime, dependency or shared-component changes. Reject ArrayBuffer migration without evidence that application serialization is at fault. Verify browser results, unit/build and trace docs; record simulated-network and physical-device proof separately. No durable architecture decision/ADR is introduced.

Real WebKit clicks exposed the fixed save bar overlapping the new remove action. Padding-only did not fix this. `SheetFrame` now has an optional footer, a scrollable body and a nonshrinking toolbar; only opted-in consumers receive the new layout. AddTransactionSheet consumes this base; theme/buttons unchanged. Real picker/remove/reselect passes on Chrome mobile/desktop and WebKit.

The offline storage path needs a separate review: a minimal WebKit 26.5 fixture rejects File/Blob writes in ephemeral contexts, while a disposable persistent context accepts writes. However, the real app's persistent-profile E2E then fails reading the saved File bytes with `NotReadableError`. A write-only probe is not proof of readable persistence. No storage serialization/schema or upload implementation was changed. Do not declare this engine-only, solved by persistent profiles, or physical-Safari ready.

Stopped under the repository fix/test loop guard after repeated implementation/test-setup refinements (see attempt record). Required human/design review: approve a focused follow-up on receipt-file lifetime and IndexedDB readback before changing file preparation or the existing offline storage representation. Keep the failing readback assertion and do not deploy this slice until the receipt gate is resolved or an explicit release decision is made.
