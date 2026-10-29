# R0 functional stability — first slice

Status: in_review (selected fixes locally verified; human UAT and full R0 remain open).

Trace: [design](../phases/R0-DETAIL_DESIGN.md), [action inventory](R0-MOUNTED-ACTIONS-2026-09-09.md), [previous red evidence](MONEYLOVER-PARITY-2026-09-08.md), [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [release notes](../../releases/CHANGELOG.md).

## Changes and acceptance

- Retained current shell/theme/card appearance. Added reusable `PWAInstallPrompt` with existing `PillButton`, dismiss/standalone/appinstalled handling and error feedback. Installation remains user-triggered; no browser install is claimed to have happened in the tests. Dismiss lasts for the mounted app session, not across reloads.
- Prompt height is measured with ResizeObserver and reserved in page padding and scroll padding. **All three phone-frame rules**, including two mobile media blocks, use the same reservation. The original logout test passes without force-click, removing the prompt, or changing the original assertion.
- Reusable native `ActionCard` displays `budget.name` and keeps category context. Enter/Space open the correct budget; offline disables editing. Original live CRUD test now reaches edit and archive.
- Budget gauge uses actual totals, including zero/unspent/exhausted/overspent cases, not positive-only sample fallbacks. No sample-data notice or invented 12 remaining days. Null/unloaded budgets are distinguished from an empty loaded list; unavailable totals are not rendered as zero balances.
- No API/auth/schema/accounting/runtime changes. The accepted 80% warning is unchanged. Full category/report/asset/planning parity is not implied.

## Test environment

Repo HEAD `d419eac4eba0ccdb7597dd1a319af7b568c3df7b` plus pre-existing uncommitted work and this slice. No commits, reset or unrelated staging. Test API uses fixture OAuth, not real Google OAuth. Local API 18173, web 4187; no server reuse. Database container is isolated from running MyPocket containers and production DB.

Commands below are from repo root for Docker, frontend directory for npm/Playwright. Explicit PATH is required on this Mac:

```sh
rtk proxy docker run --rm -d --name mypocket-r0-20260909-pg -e POSTGRES_USER=mypocket -e POSTGRES_PASSWORD=mypocket -e POSTGRES_DB=mypocket_r0_e2e --tmpfs /var/lib/postgresql/data -p 127.0.0.1:64739:5432 postgres:16-alpine
rtk proxy env DEBUG_PRINT_LIMIT=500 PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm test -- --run
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npx playwright test --config=playwright.r0.config.ts --project=mobile --project=desktop --workers=1 --output=test-results/r0-final
```

`frontend/playwright.r0.config.ts` is an opt-in isolated test harness derived from the regular config, not deployment configuration. Credentials above are disposable fixture credentials, not production values. Do not run against production or reuse an existing server.

## Recorded results

| Run | Result | Evidence meaning |
| --- | --- | --- |
| Baseline `npx playwright test --config=playwright.r0.config.ts --project=mobile e2e/auth-shell.spec.ts e2e/planning-automation.spec.ts --workers=1` with the PATH prefix above | **2 fail, 2 pass** | Logout pointer intercepted by PWA/dock; budget row not found by persisted name. |
| Initial `npm test -- --run src/screens/BudgetsScreen.test.tsx` | **6 fail** | Missing name/native disabled semantics and sample-number fallbacks; test selector for empty zero was subsequently scoped to its visible span because a pre-existing hidden total has the same text. |
| `npm test -- --run src/app/App.test.tsx -t 'dismissing installation\|after appinstalled\|already running standalone'` | **3 fail** | No dismiss button; prompt remained after install and in standalone mode. |
| First implementation, original four mobile tests | **1 fail, 3 pass** | Budget CRUD fixed. Logout still obstructed: two later mobile CSS shorthands overrode the initial padding fix. No forced-click workaround was used. |
| After applying reservation in all phone-frame rules, complete original mobile suite | **9 pass** | Auth, CRUD, offline/conflicts, planning and cookie/API-key cross-user isolation. |
| Added expanded-help/dismiss browser tests; all three projects | **32 pass, 1 fail** | Chrome mobile 11/11, desktop 11/11, WebKit mobile 10/11. WebKit offline reload fails separately (below). |
| `npm test -- --run src/screens/BudgetsScreen.test.tsx -t 'unloaded budget'` before unavailable-state fix | **1 fail** | Null list was incorrectly represented as loaded empty budget data. |
| Final complete Vitest suite after review follow-up | **72 pass / 9 files** | Includes 7 budget regressions, 3 added App install-lifecycle regressions, accepted-install removal assertion, and 2 native prompt dismissal/rejection recovery tests. |
| Final Chrome mobile + desktop suite after unavailable-state fix | **22 pass** | Both expanded and dismissed install prompt flows included. Build/typecheck also run by fresh web-server setup. |

All npm/Playwright commands used the explicit PATH prefix above; rows with shortened commands omit that unchanged prefix only. Color-environment warnings (`NO_COLOR` versus `FORCE_COLOR`) were emitted by Playwright; they are not test failures.

## WebKit residual failure — release gate

`e2e/pwa-shell.spec.ts:17` fails at `page.reload()` after `context.setOffline(true)` with `WebKit encountered an internal error`. Before reload, the test verifies registered/ready/controlling service worker, cached `/`, and cached identity. This failed in the 33-test run and again, unchanged, via:

```sh
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npx playwright test --config=playwright.r0.config.ts --project=webkit-mobile e2e/pwa-shell.spec.ts --workers=1 --output=test-results/r0-webkit-recheck
```

Result: **1 fail**. Cause is **not established**; do not label this an engine-only issue or assert it reproduces on physical Safari. Service-worker code was read but not changed in this slice. Track an independent reproduction/instrumentation task before claiming iPhone offline release readiness. No retries/skips added to hide this failure.

## Visual/manual and provider limitations

Inspected generated mobile screenshot `frontend/test-results/r0/install-prompt-expanded-in-3debb-st-account-action-reachable-mobile/account-install-help.png`: logout is above the expanded prompt and the dock remains reachable. Screenshots/trace ZIPs live under ignored test-results directories and may be replaced by future test runs; this document retains the durable command/result record.

Physical-device installation, production OAuth, Redis/S3/Web Push, API completeness and human UAT were not revalidated by this frontend slice. Backend package suites were not re-run: no backend code changed in this slice; live E2E used the real fixture API/PostgreSQL. This is not a replacement for the earlier full backend audit.

## Docs review

- [x] Design/root cause/alternatives/impact recorded; current actions inventoried with remaining work.
- [x] Context, backlog, validation and changelog reconciled without upgrading full R0 or whole-app completion.
- [x] Requirements unchanged: restores existing budget naming/data correctness and accessible PWA interaction; no new business rule.
- [x] API, ERD, architecture, SDD, ADR unchanged: no contract/schema/security/runtime boundary changed.
- [x] Final normal production build: `rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm run build` passed tsc/Vite; output `index-CcU2W969.js`, `index-CkpeUcnY.css`. Rebuilt without fixture VITE_API_BASE_URL, restoring the ordinary production build after E2E.
- [x] Cleanup: `rtk proxy docker stop mypocket-r0-20260909-pg` succeeded; subsequent `docker ps --filter name=mypocket-r0-20260909-pg` returned no container. Only disposable tmpfs test data was destroyed; it is not recoverable and can be regenerated by rerunning the tests. Production data unchanged.
- [x] Read-only peer review: no critical/important findings in this slice. Minor request for native-install accepted/dismissed/rejected coverage was addressed, followed by the 72-test passing run.
- [x] `rtk git diff --check` passed. Read-only local Markdown-link validation checked all three new R0 documents: zero broken links.

No deployment performed. Production containers/database untouched.
