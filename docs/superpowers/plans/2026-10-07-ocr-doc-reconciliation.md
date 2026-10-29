# OCR contract documentation reconciliation

## Intent

Reconcile the OCR architecture page with the implemented provider adapter so
future agents do not infer that OCR is still design-only.

## Contract

- AI Entry accepts JPEG/PNG/PDF attachments through the MyPocket backend.
- Backend stores originals privately, invokes OCR Platform with wait-mode plus
  bounded polling fallback, persists normalized OCR evidence and passes only
  text to the model.
- OCR remains an assistant suggestion source; proposal review is required
  before a transaction is written.
- Provider credentials remain server-side and live image/provider UAT is a
  separate deployment gate.

## Verification

Update only the stale linked-decision wording, then run docs link/design checks
and `git diff --check`; no runtime behavior or provider call changes are part
of this reconciliation.

## Result

The OCR page now distinguishes implemented local adapter behavior from the
remaining live provider/image/retention gates, and its machine-readable test
field points to the backend contract checks instead of `not-run-docs-only`.
