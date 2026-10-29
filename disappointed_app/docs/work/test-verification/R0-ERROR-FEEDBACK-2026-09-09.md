# R0 — transaction and API-key error feedback

Status: in_review; this bounded slice is locally verified, full R0 and device UAT remain open. No deployment.

Trace: [design](../phases/R0-ERROR-FEEDBACK-DETAIL_DESIGN.md), [parent](../phases/R0-DETAIL_DESIGN.md), [action inventory](R0-MOUNTED-ACTIONS-2026-09-09.md), [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md).

## Implemented behavior

- Existing add/edit/archive transaction sheets show action-specific errors through shared `OperationError`. Raw exceptions are not shown; API correlation IDs are shown only when they match the bounded safe character set. Input stays available and no financial write is automatically retried. Uncertain responses instruct the user to check transaction history before resubmission.
- If an offline transaction has been saved but the separate receipt queue fails, the sheet reports partial success, retains the file, offers Close and disables another save. The regression proves exactly one queued transaction. It does not prove remote receipt upload/OCR readiness.
- Account API-key list loading/failure differs from an empty list, with a manual read-only reload. Failed creation retains its requested name; confirmed creation keeps its one-time plaintext if the subsequent list refresh fails. Confirmed revoke stays disabled and removes matching displayed plaintext despite refresh failure.
- Shared request generations discard old list responses and old busy/error updates across disconnect/reconnect. Mutations are disabled while a list load is active. Copy reports success/failure/unavailable; stale copy completion cannot report success for a replacement/revoked key.
- Touched action buttons use the native `ActionButton` base while keeping existing CSS classes. Existing UI theme/layout, backend contracts, bearer/cookie authorization, database schema and accounting logic are unchanged.

## Test record

All shell commands use `rtk`; npm/Playwright commands run from `frontend/` with this explicit Mac PATH:

```sh
rtk proxy env DEBUG_PRINT_LIMIT=300 PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm test -- --run
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npx playwright test --config=playwright.r0.config.ts --project=mobile --project=desktop
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npx playwright test --config=playwright.r0.config.ts --project=webkit-mobile e2e/operation-feedback.spec.ts --output=test-results/r0-webkit-feedback
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin node scripts/probe-offline-webkit.mjs
```

| Check | Observed result |
| --- | --- |
| Initial targeted API-key + rejected transaction tests | 10 failed, 28 skipped, 3 unhandled write errors before implementation |
| Initial partial receipt-queue failure test | Failed with unhandled quota error before catch/saved-state lock |
| First post-implementation run | 38 passed, 1 failed: test incorrectly expected `entity_type` on compatibility `readOutbox()` result; corrected to its actual `input` contract |
| Reconnect stale-response regression / delayed copy-after-revoke regression | Each failed before its generation guard, then passed after correction |
| Initial browser setup | Blocked by three new test-only TypeScript errors (`exact` is not a Testing Library role option); corrected without changing product behavior |
| Initial full Chrome suite | 24 passed, 2 failed: new test selected nonexistent exact label `Thu`; actual UI label is `Khoản thu`, corrected using DOM evidence |
| Intermediate full unit suite | 84 passed, 1 failed: old expense-create test clicked Add before asynchronous wallet load enabled it; now awaits enabled state, not a timing sleep |
| Final full Vitest suite | **85 passed / 10 files**, including 9 Account lifecycle/race tests and 32 App tests |
| Final Chrome mobile + desktop | **26 passed**; real fixture API/PostgreSQL, existing 22 journeys plus 4 new project-specific error/recovery journeys |
| WebKit new online fault-injection journeys, worker blocked only in this file | **2 passed** after proving failures are injected; not offline or installed-PWA acceptance |

New E2E tests inject a rejected transaction request, assert preserved input and one attempted write, then allow a deliberate user resubmission and verify one persisted entry after reload. API-key tests use real create/revoke with injected subsequent GET failures. Trace, screenshot and video recording are disabled for the new fixture-key journeys. Existing cookie/bearer user-isolation tests also pass; this does not establish whole-API completeness.

WebKit fault-injection setup initially failed both new cases: no alert appeared. Instrumented rerun established both interception counters were **0**; the intended failure responses had never reached the app. The new `operation-feedback.spec.ts` alone now sets `serviceWorkers: 'block'`, following [Playwright's network-interception guidance](https://playwright.dev/docs/network#missing-network-events-and-service-workers), and asserts that each failure was actually injected. These tests prove online error feedback without a worker, not PWA behavior. Existing PWA/offline tests and the isolated worker diagnostic retain service workers; their original offline-reload failure is not suppressed.

## WebKit diagnosis (not a physical Safari acceptance)

`scripts/probe-offline-webkit.mjs` launches disposable contexts with a loopback HTML fixture. It compares a minimal cache-only service worker with the unchanged MyPocket worker; neither case loads application UI, authentication or finance APIs. Every case confirms a controlling worker and cached root before interruption.

| Engine | Worker | `setOffline(true)` + automation reload | `setOffline(true)` + page reload | Server actually closed, no emulation |
| --- | --- | --- | --- | --- |
| Chromium 151.0.7922.34 | Minimal cache-only | Pass | Pass | Pass |
| Chromium 151.0.7922.34 | Current app worker | Pass | Pass | Pass |
| WebKit 26.5 | Minimal cache-only | Internal-error failure | Load timeout | Pass |
| WebKit 26.5 | Current app worker | Internal-error failure | Load timeout | Pass |

This isolates the observed trigger to the WebKit offline-emulation/reload path, independent of MyPocket application logic. It does **not** prove physical airplane-mode Safari works, nor identify the underlying browser/automation defect. The original WebKit offline-reload test remains an unresolved release gate, not skipped, rewritten or declared passing. The diagnostic exits zero when all cases have been recorded, even when individual cases fail; read its JSON results rather than treating exit status as suite success. No service-worker changes or dependency upgrades were made.

## Review, boundaries and handoff

- Read-only peer review found one important list-request race and one minor stale-copy feedback race. Both were reproduced with deferred promises, fixed and re-reviewed with no remaining findings in the bounded changes.
- Test environment: unchanged HEAD `d419eac4eba0ccdb7597dd1a319af7b568c3df7b` plus dirty worktree; fixture API 18173, web 4187, PostgreSQL tmpfs container `mypocket-r0-20260909-pg` on 64739/database `mypocket_r0_e2e`. No existing server reused and no production database accessed.
- No backend package rerun or new endpoint: live tests exercise existing handlers, not full backend/provider coverage. API/ERD/ADR/SDD need no contract edits. Financial retry/idempotency semantics remain unchanged.
- Remaining R0: receipt footer wiring/unsupported controls; other planning/search/audit/notification error paths. Then finance/report journeys per parent design. Production OAuth, actual iOS/Android install, S3/Web Push and third-party contract completeness still need their own validation.

## Docs review checklist

- [x] Root cause, approved scope and design refinements recorded; reviewer findings closed with regression evidence.
- [x] Context, backlog, validation matrix, mounted-action inventory and unreleased changelog updated; no whole-phase promotion.
- [x] Requirements/master API/schema/architecture/ADR assessed: no new domain, endpoint, auth, runtime or schema contract; no ADR needed.
- [x] Human UAT remains required; physical devices and providers are explicitly excluded from automated acceptance.
- [x] Local Markdown links in the new design, new verification and updated action inventory checked: 3 files, zero missing targets. `rtk git diff --check` passed.

- [x] Final normal build: `rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm run build` passed TypeScript and Vite. Outputs: `index-CetNAiXY.js`, `index-BeEcDUyx.css`. Rebuilt without fixture `VITE_API_BASE_URL`, restoring ordinary production artifacts after E2E; this is not a deployment.
- [x] Cleanup: `rtk proxy docker stop mypocket-r0-20260909-pg` succeeded; subsequent filtered `docker ps -a` returned no container. Its disposable tmpfs test data is removed and unrecoverable, but regenerable from the fixture tests. Production containers/data unchanged.
