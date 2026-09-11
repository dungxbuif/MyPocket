# R0 — focused receipt readback investigation

Status: **in_review for this diagnostic follow-up**; receipt/device acceptance remains **blocked**. Owner approved this investigation with “Tiếp” after the previous stop. No production application changes, storage migration or deployment.

Trace: [approved design](../phases/R0-RECEIPT-CONTROLS-DETAIL_DESIGN.md), [previous attempt record](R0-RECEIPT-CONTROLS-2026-09-09.md), [R0](../phases/R0-DETAIL_DESIGN.md), [TICKET-028](../tickets/TICKET-028-production-hardening-debug-audit-foundation.md), [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md).

## Finding and causal limits

`scripts/probe-receipt-storage.mjs` now checks bytes, not just successful writes. It uses synthetic three-byte data and a minimal loopback page with no MyPocket, API, service worker or private files.

| WebKit 26.5 condition | Observed result |
| --- | --- |
| Ephemeral context, online | File/Blob writes fail with `UnknownError`; ArrayBuffer writes and reads pass |
| Disposable persistent context, online | Native File, memory Blob and ArrayBuffer all write/read the exact bytes; native input reset true/false makes no difference |
| Either context, emulated offline | Reading the selected File already fails with `NotReadableError`, before a storage write |
| Same stored File, persistent context, no rewrite | Offline false → true → false yields byte-read pass → `NotReadableError` → byte-read pass |

The last differential isolates offline emulation as sufficient to produce this symptom in the tested environment. It contradicts the earlier suspicion that input reset alone destroys the receipt or that successful writes permanently lose the bytes. It does **not** prove that all physical Safari devices support this path. No serialization fix is justified by these results. The low-level cause inside the browser/automation implementation is not established.

## Browser proof and retained failure

The original `offline receipt bytes persist...` test remains active, uses `setOffline(true)`, and still fails on WebKit. No skip, expected-failure marker or weaker byte assertion was introduced.

The separately named `HTTP outage retains offline receipt bytes...` test reuses the real login, wallet creation, transaction form, IndexedDB receipt/outbox/transaction rows and byte/entity assertions. It aborts HTTP routes, confirms an intercepted health request and rejected fetch, then dispatches an offline signal with `navigator.onLine=false`. WebKit uses its own disposable persistent profile. Only this WebKit HTTP-outage scenario blocks service workers, because the first attempt's transport assertion proved routing did not block the health request with a worker enabled. Playwright documents this interception limitation and recommends blocking workers for route-based tests: [BrowserContext.route](https://playwright.dev/docs/api/class-browsercontext#browser-context-route).

This test is **application storage under simulated HTTP outage**, not PWA shell-reload, physical airplane-mode or reconnect-upload acceptance. The original PWA test and production worker are unchanged.

| Follow-up stage | Result |
| --- | --- |
| Readback probe, input-reset comparison | Persistent online byte reads pass with/without reset |
| Add offline comparison and same-record roundtrip | Reproduces source-read and stored-read failures only during offline emulation; restoring online restores same stored bytes |
| First expanded browser run (`r0-receipts-readback`) | 7 pass / 2 fail: original WebKit readback plus new transport precondition not met |
| Block workers only in new WebKit transport scenario (`r0-receipts-readback-final`) | 8 pass / 1 fail: only original WebKit emulation case |
| Add reviewer-requested interception assertion (`r0-receipts-readback-reviewed`) | 8 pass / 1 fail, same original emulation case |
| Full frontend unit suite | 92 pass / 10 files |

No production-code fix attempts occurred. One test-harness correction and one review-strengthening change occurred; do not reinterpret the repeated original failure as fixed.

## Reproduction commands

Run from `frontend/`, with isolated fixture PostgreSQL `mypocket-r0-20260909-pg` on loopback 64739 and database `mypocket_r0_e2e`, test API 18173 and web 4187. Never use the production database. Browser profiles close in `finally`.

```sh
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin node scripts/probe-receipt-storage.mjs
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npx playwright test --config=playwright.r0.config.ts --project=mobile --project=desktop --project=webkit-mobile e2e/receipt-controls.spec.ts --output=test-results/r0-receipts-readback-reviewed
rtk proxy env DEBUG_PRINT_LIMIT=300 PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm test -- --run
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npx playwright test --config=playwright.r0.config.ts --project=mobile --project=desktop --output=test-results/r0-receipts-readback-chrome
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm run build
```

The probe exits zero when diagnostic cases are recorded, including failing cases; it is not an all-green suite. The three-browser command exits one because the retained emulation test fails.

## Review, reconciliation and next gate

- Read-only independent review: no Critical/Important findings. Minor request to verify actual route interception implemented and verified.
- Requirements/API/ERD/security/runtime/architecture: no application behavior or contracts changed; no master-doc or ADR change needed. Storage representation remains Blob/File. No migration or memory-copy tradeoff adopted.
- Anti-pattern recorded here: do not migrate application storage solely to satisfy automation until byte readback and network-emulation boundaries are isolated; a write-only probe does not establish persistence integrity.
- Context/backlog/validation/changelog link this bounded proof. Full R0 and API completeness are not promoted.
- Physical Safari/PWA UAT still required on a non-production test account: select a synthetic image, enter airplane mode, save one expense, verify one pending transaction and readable queued image, restart/reopen the app offline, then reconnect to a configured test object store and verify one attachment without a duplicate transaction. Record device/OS/browser versions. This run does not claim to execute that checklist or permit deployment.

## Final regression and cleanup

- Full Chrome mobile/desktop regression: **32 passed**, exit 0 (`test-results/r0-receipts-readback-chrome`). This does not replace the targeted **8 pass / 1 fail** result.
- Normal `npm run build`, no fixture `VITE_*` override: typecheck and Vite build passed, exit 0. Assets `index-kt5iPbZM.js` / `index-DL_iVDHZ.css`; same application output as before this diagnostics-only follow-up.
- `rtk git diff --check`: passed.
- `rtk proxy docker stop mypocket-r0-20260909-pg`: stopped the exact container created for this follow-up. `rtk proxy docker ps -a --filter name=mypocket-r0-20260909-pg --format '{{.Names}} {{.Status}}'` returned no rows. Disposable tmpfs fixture data was removed and can be regenerated; no production data touched.
- No commit, deployment or release approval.
