# R0 — receipt controls and remaining WebKit readback blocker

Status: **blocked**, not complete. Selected UI controls pass; the WebKit offline readback acceptance does not. No deployment.

Follow-up after owner approval: [readback investigation](R0-RECEIPT-READBACK-2026-09-09.md) isolates the symptom to offline emulation in a minimal fixture; the same stored bytes recover without rewriting when emulation is disabled. A separate HTTP-outage browser scenario passes, while the original failing test remains active. This supersedes the unconfirmed file-lifetime hypothesis below, not the physical-device acceptance gate.

Trace: [design](../phases/R0-RECEIPT-CONTROLS-DETAIL_DESIGN.md), [R0 parent](../phases/R0-DETAIL_DESIGN.md), [action inventory](R0-MOUNTED-ACTIONS-2026-09-09.md), [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md).

## Implemented, with bounded proof

- One always-mounted `FilePickerInput` serves both image buttons, including the footer with collapsed details. Native input cancellation retains React's current selection; selection reset allows same-file reselection. Both controls use shared buttons.
- Filename remains visible outside the expandable panel, with explicit removal. Read-only/pending/completed states lock receipt actions; debt cannot silently discard an earlier chosen image.
- Four previously inert quick-add details use `UnavailableAction` with native disabled semantics and visible explanations. These are **not implemented location/contact/event/reminder features**.
- Optional shared `SheetFrame.footer` separates the scrolling body from header/toolbar. Existing consumers without a footer retain their original structure. The receipt screen's real WebKit removal click no longer hits the save bar.

## Environment and commands

Current dirty checkout, HEAD `d419eac4eba0ccdb7597dd1a319af7b568c3df7b`; no commits/reset/staging or unrelated changes. Test PostgreSQL: disposable tmpfs `mypocket-r0-20260909-pg`, loopback 64739, database `mypocket_r0_e2e`; API 18173, web 4187, fixture OAuth, no server reuse. No private user images: test PNG is synthetic.

Run from `frontend/`:

```sh
rtk proxy env DEBUG_PRINT_LIMIT=300 PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm test -- --run
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm run typecheck
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npx playwright test --config=playwright.r0.config.ts --project=mobile --project=desktop --project=webkit-mobile e2e/receipt-controls.spec.ts --output=test-results/r0-receipts-final
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin node scripts/probe-receipt-storage.mjs
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npx playwright test --config=playwright.r0.config.ts --project=mobile --project=desktop --output=test-results/r0-receipts-chrome-final
```

## Attempt and result record

| Stage | Observed evidence |
| --- | --- |
| Before implementation: targeted six new App tests | 6 failed, 32 skipped: input inaccessible without details, footer unnamed, unsupported buttons enabled |
| Initial implementation | 37 App tests passed, one test expectation failed: transaction ID lives in mutation `entity_id`, not compatibility wrapper's payload `input.id` |
| Corrected test to inspect real pending mutations | Full frontend **91 passed**, typecheck passed; actual offline receipt links to the queued entity |
| Reviewer lock-coverage additions | Added completed/details-picker locks and degraded-mode scenario. Two incorrect degraded fixture setups failed; corrected by making IndexedDB unavailable after the online form loads and before offline hydration |
| First three-browser receipt run | **4 passed, 2 WebKit failed**: toolbar intercepted Remove; offline receipt queue rejected Blob/File |
| Padding-only toolbar fix attempt | WebKit Remove still intercepted: padding was insufficient for a fixed child inside the scrollable filtered sheet |
| Shared optional footer/body layout | Real removal now passes on all three browser projects |
| Persistent-profile test setup | Initial run rejected nonexistent `browserType` fixture; corrected to explicit `webkit` launcher, without production changes |
| Latest full frontend suite | **92 passed / 10 files**, including 39 App tests |
| Latest three-browser receipt suite | **5 passed, 1 failed**: all picker flows + both Chrome offline byte checks pass; persistent WebKit fails stored byte readback with `NotReadableError` |

The repository loop guard was reached across these implementation/test-setup cycles. No further corrective changes are authorized by this run; stop for human/design review. Final regression/build checks and docs reconciliation do not erase the unresolved result.

## WebKit diagnosis and limits

Durable opt-in probe `scripts/probe-receipt-storage.mjs` uses a minimal loopback HTML fixture with no MyPocket code, no input reset, no auth or API. WebKit 26.5 results:

| Context | Native File write | Memory Blob write | ArrayBuffer write |
| --- | --- | --- | --- |
| Ephemeral | Fail: `UnknownError`, preparing Blob/File data | Same failure | Pass |
| Disposable persistent | Pass | Pass | Pass |

The script records **writes only**; it exits zero when cases are recorded, not when all cases pass. In the real app, the persistent-profile E2E confirms the form closes and one transaction is pending, but reading back `item.file.arrayBuffer()` throws `NotReadableError`. The byte-comparison assertion remains intact. Therefore neither successful `put()` nor using persistent mode is acceptance of usable stored images.

Current hypothesis: a WebKit File/Blob backing-data lifetime/readback issue; whether caused by native input reset, file-backed versus in-memory data, browser automation or app storage lifecycle is **not established**. Next review must compare readback, not just writes, across those boundaries before changing serialization. A storage format migration or material file-memory tradeoff needs explicit design review. Original WebKit PWA offline-reload failure from the prior slice remains separate and unchanged.

The storage E2E uses an empty-path disposable persistent profile **only for WebKit receipt storage**, closed in `finally`; picker/accessibility tests use normal contexts. This does not use a personal browser profile or demonstrate physical iOS/Android behavior. Actual S3 upload, receipt viewing/download, OCR, reconnect attachment upload and production provider/device acceptance remain outside this slice's proof.

## Review and docs checklist

- [x] Read-only reviewer: no critical/important code findings; additional lock assertions added. Follow-up review accepted the optional footer design **conditional on browser tests passing**; that condition is not satisfied for WebKit readback.
- [x] Current status/design/backlog/validation/changelog reconciled as blocked, not whole-R0 completion.
- [x] Requirements: restores access to existing receipt capability; unsupported controls remain explicitly unavailable. No new domain requirement accepted.
- [x] API/ERD/security/runtime: no changes. Existing upload, receipt metadata and offline storage representation preserved.
- [x] Architecture/SDD/ADR: shared visual base gains an optional layout slot, no new subsystem or durable architecture decision; no ADR needed for current changes. Possible future storage changes require separate review.
- [ ] Human UAT and WebKit stored-image readback accepted.

## Final regression, build and cleanup

- Full Chrome mobile + desktop regression after all code changes: **30 passed** (`test-results/r0-receipts-chrome-final`). This does not resolve or replace the separate three-browser result of **5 passed / 1 WebKit failed**.
- Normal `npm run build`, without fixture `VITE_*` overrides: typecheck and Vite build passed. Output: `dist/assets/index-kt5iPbZM.js` and `dist/assets/index-DL_iVDHZ.css`.
- Stopped the exact disposable container `mypocket-r0-20260909-pg`; subsequent filtered `docker ps -a` returned no container. Its tmpfs test data was removed and can be regenerated from fixtures. Production containers and data were not touched.
- No deployment, commit or release. Human/design review remains required before another corrective cycle for WebKit receipt readback.
