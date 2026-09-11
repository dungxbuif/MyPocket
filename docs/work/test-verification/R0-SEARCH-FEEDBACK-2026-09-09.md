# R0 — search feedback verification

Status: in_review for this bounded slice; full R0 remains in_progress. Owner explicitly approved the search feedback design. No deployment.

Trace: [design](../phases/R0-SEARCH-FEEDBACK-DETAIL_DESIGN.md), [R0](../phases/R0-DETAIL_DESIGN.md), [TICKET-015](../tickets/TICKET-015-pwa-navigation-search-wallet-views.md), [backlog](../BACKLOG.md), [inventory](R0-MOUNTED-ACTIONS-2026-09-09.md), [validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md).

## Changed behavior

- Mounted search uses a dedicated `SearchPanel` with request-scoped loading/success/error state. Empty appears only after the current successful response contains no results.
- Safe shared `OperationError` retains the typed query, exposes the sanitized correlation ID and permits explicit retry. Raw server detail is not rendered.
- Effect cleanup ignores late responses after query/connectivity/owner changes or panel unmount. App retains the query across panel toggles, keys the panel by owner, and resets query/open state when unauthenticated.
- Existing `Card` and `SearchInputField` are reused. Shared clear action uses `ActionButton` and an accessible name; shell/theme/results layout retained. This does not claim pixel-identical input styling: the global panel now uses the existing shared search-input presentation.
- Offline search does not issue a request or falsely claim no results; it explicitly requires connectivity. No new offline index/cache behavior was added. The analytics adapter's existing TypeError-only user-scoped cache fallback is unchanged; cached-result freshness labeling remains a separate limitation, not newly verified here.

## Test evidence and attempt log

| Stage | Result |
| --- | --- |
| Before implementation: new real-App search tests | 7 failed as expected: missing error/loading, stale success/failure overwrite, results return after clear/close/offline |
| Request-scoped panel and shared primitives | Same 7 tests passed |
| Full frontend first run | 99 passed / 11 files |
| First browser startup, then diagnostic rerun | No browser cases ran: build's typecheck rejected three Testing Library `exact` options mistakenly copied from Playwright syntax (TS2769) |
| Correct test selector options; review coverage addition | Typecheck passed; full frontend **100 passed / 11 files**, including 8 real-App search tests |
| Targeted search E2E | **6 passed** across Chrome mobile, Chrome desktop and WebKit mobile |

The added eighth test verifies that a successful late response after logout remains hidden. Query clearing on logout is also code-reviewed; a same-mounted-App reauthentication/owner switch is not separately exercised by that test. Clear-query coverage now uses the actual named clear button.

Browser tests create real fixture wallets and query the real API. Only the first search failure is injected; retry returns the real wallet. A second test delays a real search response until a newer no-match request completes. Unit tests deterministically flush deferred completions and verify stale success and failure paths. Route-based browser fault injection blocks service workers **only in this spec**, matching prior fault-test practice; this is not PWA offline-reload acceptance. Existing PWA and receipt failure cases are untouched. Screenshots: `frontend/test-results/r0-search/**/search-error.png` (local generated artifacts).

## Commands and environment

All frontend commands use Node 24 from the explicit PATH below. Tests ran against disposable tmpfs PostgreSQL `mypocket-r0-20260909-pg`, loopback 64739, database `mypocket_r0_e2e`, API 18173, web 4187 and fixture OAuth. No production DB or real user records/images.

```sh
rtk proxy env DEBUG_PRINT_LIMIT=300 PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm test -- --run src/app/App.search.test.tsx
rtk proxy env DEBUG_PRINT_LIMIT=300 PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm test -- --run
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm run typecheck
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npx playwright test --config=playwright.r0.config.ts --project=mobile --project=desktop --project=webkit-mobile e2e/search-feedback.spec.ts --output=test-results/r0-search
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npx playwright test --config=playwright.r0.config.ts --project=mobile --project=desktop --output=test-results/r0-search-chrome
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm run build
```

The diagnostic startup rerun used the same targeted browser command with `DEBUG=pw:webserver`; it confirmed the frontend typecheck failure, not a PostgreSQL outage. No runtime config workaround was made.

## Docs and review checklist

- [x] Independent read-only reviewer: no Critical/Important findings. Minor authentication-lifecycle coverage request addressed with pending-search/logout test; scope limit above retained.
- [x] Requirements/API/ERD/security/runtime: no API, schema, auth-policy, backend, deployment or dependency changes; existing minimum two-character search remains.
- [x] Architecture/ADR: component ownership is local to the existing UI boundary, no durable subsystem decision or migration; no ADR needed.
- [x] Current design/context/backlog/inventory/validation/changelog reconciled; no whole-R0 or search parity completion claim.
- [x] Horizontal inspection: other swallowed errors remain in notification mark-read, wallet-detail/Insider loading, logout and unmounted asset archive code; tracked in the inventory, not bundled into this change.
- [ ] Human physical-device UAT/release acceptance. Suggested check: search, induce a test API failure, retry, quickly change/clear query and close/reopen panel. Never use production data for fault injection.

## Final checks and cleanup

- Full Chrome mobile/desktop regression: **36 passed**, exit 0 (`test-results/r0-search-chrome`). Targeted search **6 passed** includes four Chrome cases also present in that regression plus two WebKit cases; do not sum them as distinct scenarios.
- Normal `npm run build` without fixture `VITE_*` overrides: typecheck and Vite passed, exit 0. Output: `index-Dz4PSmjw.js` / `index-DL_iVDHZ.css`.
- Visually inspected the WebKit search-error screenshot: query, clear control, wrapped safe message, correlation ID and retry button are readable without overlap. Real retry/clear clicks pass on all three projects.
- `rtk git diff --check` passed; local trace-link check on this document and the design found zero broken links.
- `rtk proxy docker stop mypocket-r0-20260909-pg` stopped the exact disposable container created for this slice. Subsequent `rtk proxy docker ps -a --filter name=mypocket-r0-20260909-pg --format '{{.Names}} {{.Status}}'` returned no rows. Its tmpfs fixture data was removed and can be regenerated; production untouched.
- No commits, deployment or release approval. Physical Safari/PWA/receipt-provider gates from prior slices remain unresolved and were not rerun or weakened here.
