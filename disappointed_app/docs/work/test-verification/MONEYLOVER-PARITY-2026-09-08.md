---
artifact_type: test_verification
id: MONEYLOVER-PARITY-VERIFY-2026-09-08
status: done
owner: shared
updated: 2026-09-08
trace:
  audit: ../../research/moneylover/PARITY-AUDIT-2026-09-08.md
  source_catalog: ../../research/moneylover/SOURCE-CATALOG-2026-09-08.md
  matrix: ../../research/moneylover/FEATURE-PARITY-2026-09-08.md
  backlog: ../BACKLOG.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Fresh audit evidence — 2026-09-08

Status `done` means the audit/evidence record is complete. Product verification is **not passed**: two mobile checks remain failed, providers/UAT and many parity capabilities remain unverified. No application fix/deploy was performed.

## Environment and isolation

- Worktree: `/Users/dungxbuif/workspace/MyPocket`, HEAD `d419eac4eba0ccdb7597dd1a319af7b568c3df7b` plus pre-existing dirty changes.
- PostgreSQL: disposable container `mypocket-parity-audit-20260908-pg`, image `postgres:16-alpine`, tmpfs data, loopback-only randomly published port `64738` in this run. No production DB credentials or volumes used.
- Databases: `mypocket_verify_parity_20260908` (serialized repository tests) and `mypocket_verify_parity_e2e_20260908` (browser fixture server).
- Node commands required explicit PATH because `rtk npm test -- --run` initially failed with `env: node: No such file or directory`. This is recorded as an environment invocation failure, not an application failure.
- First PostgreSQL readiness check occurred before startup and returned no response; the next check passed. An unquoted URL attempt was rejected by zsh before Go ran; the quoted command below executed successfully.

## Commands/results

All shell commands use `rtk`. Run Go commands from `backend`, frontend commands from `frontend`, portal build from `frontend/docs`.

| Check | Result | What this proves / does not prove |
| --- | --- | --- |
| Public Zendesk categories/sections/articles JSON | PASS: 4/24/96, next_page=null | Coverage of public locale index, not videos/private pages or visual parity. |
| Go without DB: `rtk proxy env -u MYPOCKET_TEST_DATABASE_URL go test -json ./... -count=1` | PASS: 102 test pass events, 39 skips | Unit/HTTP coverage; initial run intentionally lacked repository environment. |
| Go with disposable PostgreSQL: command below | PASS: 139 test pass events, 2 skips | Includes real SQL accounting, idempotency, ownership, migration, planning, sync and portfolio tests. |
| `rtk proxy env -u MYPOCKET_TEST_DATABASE_URL go vet ./...` | PASS | Static Go diagnostics. |
| `rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm test -- --run` | PASS: 8 files, 60 tests | Components/offline helpers; many APIs mocked. Not live provider/UI parity. |
| Same PATH, `npm run build` | PASS | TypeScript + Vite; original production-default build emitted index-BGbJLwH0.js and index-BR8N5UaX.css. |
| Existing mobile Playwright suite with temporary isolated config | FAIL: 6 pass, 3 fail in first run | One harness port mismatch and two product/test-contract failures, detailed below. |
| User-isolation case rerun with compatible API port | PASS: 1 test | Removes the harness-only failure; not a rerun of the two remaining failed tests. |

Counts are Go JSON test-result events, not line/branch coverage. Two remaining Go skips: `TestRedisRoundTripWhenConfigured`, `TestS3SmokeObjectLifecycle`. Their live Redis/S3 environments were not supplied. Existing HTTP mocks or S3 safe-error tests do not replace real provider smoke proof.

Executed DB-backed Go invocation (JSON output was summarized in memory):

```sh
rtk proxy env 'MYPOCKET_TEST_DATABASE_URL=postgres://postgres@127.0.0.1:64738/mypocket_verify_parity_20260908?sslmode=disable' go test -json -p 1 ./... -count=1
```

The repository test helpers reset `public` schema; `-p 1` prevents different packages racing on that disposable schema. Do not substitute a normal application or production database URL.

## Browser results and failure analysis

Temporary config imported `frontend/playwright.config.ts`, replaced its database URL with the disposable E2E database, changed web port `4173 → 4187`, and initially changed API `18173 → 18473`. Existing test bodies were not changed. It used the existing mobile Pixel 5 / Chromium project, workers=1, max-failures=3.

```sh
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm run test:e2e -- --config playwright.parity-audit.config.ts --project mobile --workers 1 --max-failures 3
```

Passed in first run: forbidden-auth rendering; finance CRUD; offline reconnect-once; stale-edit conflict/keep-server; event/debt linking and schedule creation; service-worker offline reload.

1. **Logout obstruction — reproduced UI defect.** `auth-shell.spec.ts:19` timed out clicking the visible logout button. Playwright repeatedly reported `pwa-install-prompt` content intercepting pointer events, with dock descendants intercepting in some attempts. This is not a credential failure. Relevant code: App's unconditional install aside and fixed dock, Account's logout control, styles.css positioning. The regression acceptance is an ordinary mobile click completing logout without force-click or hiding overlays in the test.
2. **Budget name contract — reproduced mismatch.** `planning-automation.spec.ts:41` could not find the entered budget name. Browser snapshot showed a successfully created row with the category name, 500,000 limit, 410,000 spent and 80% alert. `BudgetsScreen.tsx` renders `categoryNames`, not `budget.budget.name`. The save operation was not disproved; named-budget lookup and the subsequent edit/archive assertions were not completed. Also note its focusable `role=button` row has no keyboard handler, and the live page displays a sample-data notice despite real API data.
3. **Isolation harness — not an app failure.** `user-isolation.spec.ts` hardcodes API port 18173, while the initial audit server used 18473. Connection was refused before authorization could be tested. The temporary config was corrected to use 18173, then only that test was rerun unchanged and passed in 69 ms (3.6 s total run).

```sh
rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm run test:e2e -- --config playwright.parity-audit.config.ts --project mobile --workers 1 --grep 'two real users'
```

Thus **7 distinct mobile cases have passing evidence across the runs; 2 remain failed**. This must not be reported as a green 9-test suite. Desktop, WebKit/iPhone device, real OAuth, push, OCR and production UAT were not executed in this audit.

## Code-only findings: confidence boundary

The matrix records direct implementation observations (missing route/input field/handler, unmounted UI, SQL predicates). New golden-data tests were not added for each gap. Specifically, included-wallet report scope, category-parent budgets and prior-period calendar arithmetic require independent SQL fixtures before fixing; a passing existing suite does not verify these untested semantics.

Analytics currently has two helper tests in `internal/analytics/repository_test.go`, not real SQL proof for every Money Lover formula. The current 60 frontend tests do not exercise every visible control or the full Account portfolio flow.

## Docs review

- Source catalog covers 96 unique IDs and links every article to at least one matrix row; editorial/historical entries explicitly remain reference-only.
- Research summaries are original concise descriptions; full source article bodies/images are not copied into the repo.
- Unsupported portal export/AI/OCR/confirm/reset claims are marked; verified file paths, analytics query/envelope fields and category-create limitations corrected.
- Old research retained with supersession note. Context/backlog/validation/changelog receive additive audit evidence; human acceptance and exclusions are not silently changed.
- New research local links and whitespace are validated at handoff; portal build evidence is recorded in final verification notes below.
- Docs-only UAT is not required; functional UAT remains open. No ADR because no runtime/domain decision was changed.

## Cleanup

Temporary Playwright config is removed after the run; disposable PostgreSQL is stopped and its tmpfs test data disappears. Production state is untouched. The audit findings and reproducible commands above are the durable evidence, not a promise of retained temporary Playwright traces.

## Final verification notes

- PASS: portal `rtk proxy env PATH=/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin npm run build` from `frontend/docs`: Docusaurus client/server compiled successfully; generated static site in `build`.
- PASS: normal frontend build repeated after E2E to restore the default API-base build instead of leaving the temporary fixture-server URL in local `dist`; same asset filenames as the pre-E2E build.
- PASS: Node read-only validation counted 96 article rows, 96 unique source IDs, 50 assessment rows, no unknown mapped IDs and zero broken local links in all four new audit documents.
- PASS: `rtk git diff --check` across the working tree.
- PASS: `rtk proxy docker ps -a --filter name=mypocket-parity-audit-20260908-pg --format '{{.Names}} {{.Status}}'` returned no containers after stop; `--rm` and tmpfs removed only the audit container/test data.
- Application code and existing test bodies were not edited. Source docs changed locally; no publish/deploy. The two failed mobile checks remain unresolved and are the first proposed remediation.
