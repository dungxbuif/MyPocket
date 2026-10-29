---
artifact_type: test_verification
id: BUSINESS-LOGIC-AUDIT-2026-09-10
status: in_progress
owner: ai
updated: 2026-09-10
---

# Business Logic Audit — 2026-09-10

## Scope and sources of truth

Scope: wallets, transaction accounting, transfers, cash-flow reports, recurring schedules, receipt references, optimistic versions, Vietnam dates, sync and obligations. Sources: finance/planning domain code, HTTP adapters, PostgreSQL integration tests, frontend service tests and Playwright flows.

Reusable audit contract: [repository skill](../../../.agents/skills/auditing-mypocket-business-logic/SKILL.md). Public companion: `frontend/docs/docs/skills/business-logic-audit.mdx`.

## Confirmed findings and corrections

| Severity | Invariant | RED result | Correction | State |
| --- | --- | --- | --- | --- |
| Critical | Balance arithmetic must not wrap `int64` | Income, expense, transfer destination and adjustment accepted overflow | Checked add/sub and validation errors | GREEN local |
| Critical | Edit/archive reversal must not overflow | Archiving an old expense wrapped a maxed balance | Checked reversal arithmetic and error propagation | GREEN local |
| High | Direct transaction update must reject stale version | HTTP/frontend discarded `base_version`; repository was last-write-wins | DTO/input/repository conflict check and stable `VERSION_CONFLICT` | GREEN local |
| High | Recurring payload must create a valid future transaction | Invalid category/destination/type combinations were accepted | Shared validation plus category-kind verification | GREEN local |
| High | Editing must preserve receipt reference | Update validation cleared `receipt_object_id` | Preserve and trim receipt reference | GREEN local |
| High | Idempotency hash covers all business fields | Changing only receipt reused the old response | Receipt reference added to canonical request hash | GREEN local |
| Medium | Selected calendar date represents HCM business date | Device timezone changed the persisted instant | Explicit `+07:00` noon conversion | GREEN local |
| High | Repayment must match obligation direction and remaining principal | Wrong-direction transactions and cross-obligation reuse could be counted | Row locks, direction rule, one-obligation rule and atomic remaining-principal check | GREEN local |
| High | Budget percentage and analytics aggregation must not wrap | Extreme values produced an invalid percentage or could overflow Go aggregation | `math/big` percentage plus checked money conversion/addition | GREEN local |
| High | Wallet writes must reject stale clients | Direct update, archive and default-AI commands could overwrite newer state | Required `base_version`, locked/conditional writes and `409 VERSION_CONFLICT` | GREEN local |
| High | Transaction archive must reject stale clients | Archive could reverse a transaction changed after the client read it | Required `base_version` checked before reversal | GREEN local |
| Critical | Revoked API key must never be authorized by stale Redis state | A cached user id could survive DB revocation or cache-delete failure | PostgreSQL revalidates every Bearer request; Redis is only a hint | GREEN local |
| High | Invalid Bearer must not fall back to a valid browser cookie | A malformed/revoked Authorization header silently inherited cookie authority | Authorization header is authoritative and denial is audited | GREEN local |
| High | Sync must surface a race discovered during domain write as conflict | Version could change after precheck and return a generic rejection | Re-read authoritative state and persist `conflict` result | GREEN local |
| High | Planning/category writers must not silently overwrite newer data | Budget, event, obligation, recurring archive and category archive/update were inconsistent | Required `base_version`, conditional/locked writes and stable `409 VERSION_CONFLICT` | GREEN local |
| High | Recurring catch-up must be bounded per worker run | A long-offline daily schedule could generate years of drafts in one DB transaction | Shared worker limit now bounds occurrence attempts and advances incrementally | GREEN local |
| High | Offline wallet/category activation must not compare an unrelated category version or corrupt the category mirror | Queued activation used `base_version=0`, always conflicted with an existing category, and its setting payload was then eligible for category upsert | Treat PUT as idempotent absolute state; skip category-version precheck and never upsert the setting payload as a category | GREEN local |

## Browser proof

- `business-core.spec.ts`: wallet create/rename/archive; income/expense create/edit/archive with exact reversals; transfer create/edit/archive; transfer excluded from report.
- Core business set (`business-core`, `accounting-correctness`, `finance-crud`): 18/18 passed across desktop Chromium, mobile Chromium and mobile WebKit with one worker.
- Real-user cookie/API-key isolation: 3/3 passed across desktop Chromium, mobile Chromium and mobile WebKit.

## Focused backend/frontend proof

- Finance unit and PostgreSQL integration tests for overflow, reversal, stale wallet/transaction versions, receipt preservation and idempotency pass.
- Planning PostgreSQL integration tests for recurring transaction-shape/category rules, repayment direction, transaction reuse and concurrent-safe principal checks pass.
- Analytics tests cover reversed date ranges, checked decimal-money conversion and aggregate overflow; auth tests cover invalid-Bearer precedence and stale Redis hints.
- Sync tests cover the race between optimistic precheck and domain mutation.
- Transaction input date test passes under both `TZ=America/Los_Angeles` and `TZ=Asia/Ho_Chi_Minh`.

Fresh verification evidence:

- Backend: `go test -p 1 ./... -count=1` passed against dedicated PostgreSQL database `mypocket_verify_business_audit`.
- Race detector: finance, planning, sync and HTTP API packages passed with `go test -race -p 1 ... -count=1`.
- Static Go analysis: `go vet ./...` passed.
- Frontend baseline: 19 files / 157 tests passed serially; TypeScript and Vite production build passed. The follow-up below supersedes this count. Serial execution does not establish the cause of earlier parallel flakiness.
- Docusaurus: production build passed.
- Repository whitespace check: `git diff --check` passed.

Integration packages must run serially because their helpers reset a shared test schema; parallel package execution against one database is invalid evidence. No production release is implied.

## Remaining risks under active audit

1. [ADR-008](../../decisions/ADR-008-atomic-offline-sync-commit.md) is implemented locally: domain writes, feed and receipt share a transaction; direct finance/portfolio and draft-confirmation writes emit changes. Apply migration `0012` before the new binary. Existing feed omissions are not backfilled: clients must perform a full resync after upgrade.
2. Physical-device Safari/PWA, provider acceptance and broader release gates remain separate; these tests do not establish whole-app completion.

Resolved contract checks:

- `include_in_total=false` excludes a wallet only from wallet/net-worth totals. Cash-flow reports intentionally remain transaction-scoped and include reportable income/expense from that wallet; `excluded_from_reports` controls report inclusion.
- Wallet/category activation is an idempotent absolute-state PUT. Concurrent opposite writes are last-write-wins until the setting gains its own version.
- Archiving a wallet retains referenced history. Since no restore command exists, its transactions are read-only after archive; ordinary lists hide the wallet and new/reversal accounting cannot target it.

## Release state

Follow-up 2026-09-11: portfolio and offline presentation corrections below are local-only. Full release verification is not green merely because focused suites pass.

`local-only / atomic-sync slice in_review` — automated corrections are green; broader audit and release gates remain open. No deployment was performed for this slice.

## Atomic-sync follow-up proof

- Observed RED: receipt failure left wallet balance at -100 instead of 0; direct wallet creation had no change feed; concurrent same-mutation replay hit a duplicate primary key; ambiguous frontend response incorrectly fell back to direct create.
- GREEN: shared command transaction/savepoints, per-user advisory lock before accounting entity locks, atomic cursor/feed writes, and original mutation retention. Tests cover injected feed/receipt failure with full rollback, safe retry, concurrent replay, canonical portfolio feed, category parent clearing, default-wallet no-op versions, and draft finalization rollback.
- Fresh frontend: 19 files / 158 tests; production TypeScript/Vite build passed. Earlier 157-test result above is superseded.
- Race detector: sync, finance, planning, portfolio and HTTP API passed serially against dedicated PostgreSQL.
- Current browser set: 24/24 for offline-sync, business-core, accounting-correctness and user-isolation across desktop Chromium/mobile Chromium/mobile WebKit; additional committed-response-loss retry scenarios passed 3/3.
- The injected response-loss test blocks service workers only for that scenario because WebKit service-worker requests bypassed page interception. Ordinary reconnect tests retain real service workers. This is test-harness evidence, not a production service-worker fix.
- Final backend rerun after the receipt conflict-projection adjustment: `MYPOCKET_TEST_DATABASE_URL=<dedicated test database> go test -p 1 ./... -count=1` passed (finance 5.791s, planning 5.223s, portfolio 1.022s, sync 2.241s). A preceding run used the wrong `DATABASE_URL` variable and skipped DB tests; that run is not integration evidence.
- Final `go vet ./...`, Docusaurus production build and `git diff --check` passed. Docusaurus warns only about last-update dates on untracked files.

## Portfolio and offline presentation follow-up — 2026-09-11

- Scope: [portfolio design](../phases/PORTFOLIO-MONEY-DETAIL_DESIGN.md), [offline reconciliation design](../phases/OFFLINE-RECONCILIATION-DETAIL_DESIGN.md). Existing accounting/auth/schema contracts unchanged.
- Critical RED: `go test ./internal/portfolio -run 'TestReplayLedger(RejectsMoneyOverflow|AcceptsMaximum)' -count=1` failed five overflow cases. Quantity × max price yielded -2 cost basis; fee/basis additions wrapped negative; a loss became +3 and accumulated profit became -4.
- Fix: checked arbitrary-precision conversions/add/sub at portfolio ledger, valuation and summary boundaries; existing validation error propagation and atomic command rollback. Neighbor search covered rounded cost removal, realized-total readback and aggregate market value.
- GREEN: focused unit suite and `MYPOCKET_TEST_DATABASE_URL=<dedicated test DB> go test -race -p 1 ./internal/portfolio ./internal/analytics ./internal/sync -count=1` passed. Real PostgreSQL assertions prove rejected trade/price leave trade count, price history, position version and sync cursor unchanged. Aggregate overflow rejects rather than wraps.
- Full release script baseline: Go race suite, TypeScript and 158 frontend tests passed; browser suite failed with 79/81 pass. Mobile lost-response test displayed two rows while captured server response contained one transaction. Desktop immediate-offline category selection timed out; cause remains under investigation.
- High RED: three pending-row merge tests fail with copied production merge logic. It used mutation IDs and rendered all entity types as transactions. Fix retains entity metadata and overlays by entity ID/type/operation; ten focused frontend tests pass, including original outbox tests.
- Test-harness RED: repeated concurrent browser runs produced identical six-digit timestamp names for shared fixture users, causing six selector failures (12/18 pass). UUID fixture names replace timestamps; assertions remain strict.
- Anti-patterns: never convert `big.Int` to `int64` before `IsInt64`; never concatenate pending and confirmed records without entity identity; never use truncated wall-clock time for concurrent test identity.
- Final backend: full `MYPOCKET_TEST_DATABASE_URL=<dedicated test DB> go test -race -p 1 ./... -count=1` and `go vet ./...` passed after portfolio changes. Docusaurus build passed; whitespace check clean.
- Frontend category test first failed (160/161): its fake GET toggled active state after the first read without any PUT. Corrected fixture keeps state until successful PUT. Final `npm test -- --run`: 20 files / 161 passed. This explains that observed fixture failure; it does not prove all historical flakiness has the same cause.
- Long 36-character UUID names exposed mobile pointer interception in the category form (12/18 browser passes, six mobile failures). Ordinary fixtures now use eight random characters; long-name mobile layout remains an explicit separate defect, not resolved by shorter fixtures.
- Final browser rerun after corrections: `npx playwright test --workers=6` = **67 passed / 14 failed**, all mobile Chromium. Desktop/mobile WebKit lost-response and conflict cases pass. Traces identify PWA install-prompt elements intercepting Account/Add transaction clicks; exact geometry cause is not yet established. This supersedes the 79/81 baseline as current release evidence.
- Corrective UI work paused under the repository loop guard because repeated verification revealed a broader shared-layout issue. Preserve all assertions, existing dataset and traces; next design must isolate prompt geometry and category-form interception without masking failures through banner dismissal or forced clicks. No deploy.

## Mobile layout root-cause follow-up — 2026-09-11

- Owner resumed the bounded investigation. Both zero-wallet and 100-wallet fixtures reproduce the original failure: seven chart money labels force a 434px content column, document width 442px on a 393px mobile viewport. Layout height grows to 818px while visual height stays 727px, displacing fixed hit targets by 91px. Removing HTML scroll padding had no effect and was reverted.
- Root fix is in shared `ComparisonBarChart`, not PWA offsets: shrinkable equal grid columns and bounded labels, matching baseline alignment, full values retained in text/title. No accounting, auth, API, schema or runtime changes.
- Initial mobile regression plus existing install-prompt tests pass 4/4; measured document width returns to 393px and visual offset to zero. No forced clicks, banner hiding or dataset deletion.
- First full run after the fix: all 81 existing browser tests pass, including mobile category forms, lost-response recovery and conflict handling. Two of six new controlled-data layout cases fail on WebKit because service-worker requests bypass page routes (snapshot contains real wallet data instead of the zero-wallet fixture). Restrict worker blocking to this layout-only spec, as already done for response-loss injection; ordinary PWA and offline tests retain real workers.
- The earlier long-name hypothesis is not established as the cause of historical form failures. New layout fixtures include 100 long-name wallets; no claim about every possible long-name form interaction is inferred from this coverage.
- Final fresh verification: Go vet and full `go test -race -p 1 -count=1 ./...` pass with `MYPOCKET_TEST_DATABASE_URL` pointing to dedicated `mypocket_verify_business_audit`; 20 frontend files / 161 tests pass; TypeScript/Vite production build passes; `npx playwright test --workers=6` passes **87/87 in 36.9s**. This supersedes the historical 67/81 browser gate. Docusaurus production build passes after public-doc reconciliation (only untracked-file update-date warning); whitespace check passes. No deployment. Physical-device/provider acceptance and migration 0012/full-resync rollout preparation remain open.
