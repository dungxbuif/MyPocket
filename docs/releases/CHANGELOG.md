---
artifact_type: changelog
id: CHANGELOG
status: active
owner: shared
human_fields: [release_approval]
ai_fields: [change_entries, linked_releases]
shared_fields: [status]
---

# Changelog

## Field Ownership

- Human owns release approval.
- AI maintains change entries and links to release notes.

This file is the first-release baseline for the current MyPocket snapshot. Earlier
development notes remain available in Git history and are intentionally not presented
as active release entries here.

## [Unreleased] - 2026-10-06

### Fixed

- Native category management now matches the backend/web scope contract:
  system categories cannot be deleted or have metadata rewritten, while
  applicable wallet IDs can be updated through the dedicated wallet-scope
  endpoint. Personal categories keep full editing controls.

- Native wallet editing now carries an explicit create/edit presentation payload,
  so an existing wallet reliably sends `PATCH /wallets/:id` instead of opening
  a create draft. The shared amount-field base adds a keyboard “Xong” action so
  credit notes remain reachable while the number pad is open.
- Added signed iPhone 16 simulator business-mutation coverage for wallet
  create/update, atomic transfer create/delete and credit entry/payment
  refresh. The complete `LOCAL_API_UI` suite passes 14/14 and leaves zero
  disposable wallets, transactions or feedback rows.
- Native transaction detail now selects Travel Mode events by stable ID rather
  than display name, and balance adjustments now preserve the account-local
  calendar day instead of the device timezone. Native XCTest is 48/48 and the
  signed simulator UI suite remains 14/14.
- Native Wallet Detail now has a signed simulator mutation smoke that creates a
  real expense through the API, verifies the persisted row, edits its note, and
  verifies the updated row. The complete API-backed iPhone 16 UI suite is now
  11/11; the disposable fixture wallet is removed in teardown.
- Native financial draft fields now reject malformed integer input instead of
  silently stripping non-digit characters. A shared strict parser covers
  transaction, transfer, adjustment, wallet, credit, jar, budget, assistant
  proposal and portfolio price/trade paths while preserving explicit recurring
  display normalization. The regression suite is now 47/47 unit tests and
  10/10 signed iPhone 16 API-backed UI tests.
- Re-ran the native iPhone 16 simulator regression after the goal target-date,
  recurring, jar-allocation and portfolio-trade validation fixes: 45/45 unit
  tests and 10/10 API-backed UI tests pass, with no
  disposable `Fixture*` wallets left behind. API-backed UI tests retain normal
  simulator ad-hoc signing so Keychain fixture sessions can be persisted;
  `CODE_SIGNING_ALLOWED=NO` remains for pure unit/build checks only.
- Native recurring editors now reject non-positive amount/interval values and
  unsupported type/frequency or reversed end-date drafts before sending an API
  request; the contract suite now passes 45/45 with the portfolio validator.
- Native jar allocation drafts now mirror the backend contract: fixed values
  stop at the JavaScript-safe integer ceiling and percentage values must be
  finite, positive, at most 100% and no more precise than two decimals.
- Native portfolio trade drafts now validate buy/sell, positive quantity
  precision and non-negative integer VND price/fee values locally, avoiding
  silent digit stripping before the API request.
- Native API requests now carry a UUID `X-Request-ID`; a JWT refresh retry
  keeps the same correlation ID so server errors can be traced end to end.
- Native transfer forms now exclude credit wallets and serialize the selected
  transfer day in the account timezone instead of the device timezone.
- Native wallet detail now refreshes the wallet summary after loading and keeps
  valid transaction rows visible when optional category metadata is unavailable.
- Native UI smoke tests now start from a deterministic unauthenticated state
  with the DEBUG-only `-UITestUnauthenticated` fixture, clearing only simulator
  Keychain/biometric state. The iPhone 16 suite passes 10 tests (5 passed, 5
  explicit API-fixture skips).
- Opt-in native API-backed UI tests now seed and tear down their own basic and
  credit wallet fixtures through the local API; the full `LOCAL_API_UI` suite
  passes 10/10 without relying on leftover simulator or database state.
- Native AI-entry attachment selection now stays disabled until private
  OCR/storage capability is configured; text-only input remains available.
- Native aggregate/basic ledgers now match the approved Week/Custom contract:
  Monday–Sunday account-local ranges, bounded previous/next week navigation,
  no all-time tab, and shared `Tiền vào`/`Tiền ra` cashflow summary.
- Native goal-wallet target dates now round-trip as account-local calendar dates,
  avoiding UTC/device timezone day shifts in the editor.
- Native session restore now refreshes an expired access token before enabling the
  Face ID lock gate. If the refresh token is expired or rejected, Keychain state
  and the biometric preference are cleared instead of opening an invalid session.
- Native category editing now preserves all applicable wallets. The shared
  multi-select base uses an empty `wallet_ids` selection for the documented
  all-wallet scope, so editing a multi-wallet category no longer silently keeps
  only the first wallet.
- Shared native buttons now apply a rectangular hit shape to their rendered
  control bounds, fixing compact “Thêm” rows that could appear visible but fail
  to open a destination in simulator/device UI automation.
- Native ledger period filters now exclude rows with malformed timestamps instead
  of accidentally including them, wallet detail no longer truncates the API's
  transaction history at 50 rows, and Finance Assistant keeps a still-running
  run cancellable when its local polling window expires.
- Debug simulator builds now expose `-UITestAuthenticatedAPI`, which exchanges
  the local fixture OAuth payload and persists a disposable session so an
  authenticated overview smoke uses the real URLSession/API path instead of a
  fake token.
- Native transactions now support the web parity bulk-delete flow with base
  selection controls, 100-row/immutable-row guards, atomic confirmation and a
  visible error that preserves the selection. Typed `204 No Content` responses
  are accepted for delete mutations instead of being decoded as JSON.
- Native Finance Assistant now restores one owner-keyed conversation, loads the
  latest 50 messages and read-only monthly overview on entry, pages older
  messages, and renders structured overview/source data without storing message
  text in UserDefaults. The API-backed simulator smoke covers the disabled
  capability response and the live overview contract.
- Native credit ledger entry/payment forms now match the web fields: entry kind,
  applicable category, payment source wallet, amount and optional note are all
  preserved in the credit API payload.
- Native basic/goal wallet detail now matches the web detail flow with
  account-timezone monthly groups, category/note labels, income/expense totals,
  and shared transaction add/edit actions.
- Native transaction ledger now uses the shared overflow menu for period,
  grouping, transfer, adjustment and bulk actions; rows show category/wallet
  context and signed amounts, while goal/credit scopes hide ordinary period
  controls.
- Selecting a credit wallet from the main native wallet list now opens the same
  credit statement, entry and payment actions as the dedicated credit route.
- Fixed native API query construction so credit statements, reports, jar views
  and Assistant history send query parameters as URL queries instead of
  accidentally escaping the delimiter into the path.
- Native transaction filtering now switches selected goal and credit wallets to
  their dedicated detail surfaces, so credit wallets no longer expose the
  ordinary transaction editor from the ledger tab.
- Native ledger loading now keeps valid rows visible when optional category or
  profile metadata fails; only the required transaction/wallet payload blocks
  the screen.
- Native portfolio now shows current/unrealized metrics, keeps price drafts
  independent per asset, allows clearing an unknown price, and preserves the
  selected trade date.
- Native recurring schedules now exclude credit wallets and filter categories
  by expense/income type, clearing stale category selections when the type
  changes.
- Native selectors now render localized semantic labels while preserving raw API
  enum values, and jar allocation mode changes clear stale drafts and reject
  zero-percent allocations to match the web validation contract.
- Native ordinary transaction and AI proposal editors now exclude credit
  wallets and clear incompatible categories on wallet/type changes; aggregate
  credit rows remain read-only and route mutations through the credit ledger.
- Native budget and balance-adjustment selectors now exclude credit wallets, and
  budgets only offer expense categories, matching the web scope rules.
- Native ended budgets are now view/delete-only; the editor no longer exposes a
  stale save path after the budget's end date.
- Native portfolio trade dates now use the account timezone and local-noon
  serialization, preventing device-timezone day shifts.
- Native wallet detail month headers now render real calendar labels, and Travel
  Mode rejects whitespace-only names, malformed dates and reversed date ranges
  before submission.

### Verification

- Native iOS simulator regression passed on iPhone 16: 45 unit/contract tests and
  5 standard authenticated UI smoke tests, with base-component and Xcode project
  guards. The Release simulator build also passed with warnings treated as
  errors, and the installed app launched successfully.
- An opt-in API-backed wallet-detail UI smoke passed with a seeded disposable
  wallet, opened the transaction editor fields, and deleted the fixture with a
  204 response afterward.
- The opt-in API-backed Assistant UI smoke passed on the same iPhone 16 using
  `SWIFT_ACTIVE_COMPILATION_CONDITIONS=DEBUG LOCAL_API_UI`, the disposable local
  API, and the `-UITestAuthenticatedAPI` fixture; no production credentials or
  data were used.
- The API-backed simulator fixture rendered the authenticated overview against
  local `:8080` with empty owner data; no production credentials or data were
  used.
- The authenticated UI sweep opened and returned from all 13 destinations in the
  “Thêm” menu; this caught and fixed the compact-row hit-target regression.
- The opt-in API-backed credit-wallet smoke passed on the iPhone 16 simulator,
  including the Giao dịch-tab and Ví-tab routes plus statement action buttons.
  Its disposable credit fixture was deleted with an HTTP 204 response after
  the run.
- Refresh rejection during an authenticated request now clears the native
  Keychain session and biometric preference through an auth invalidation hook;
  the regression suite covers the 401 → refresh 401 path.
- Backend race/integration verification passed against disposable PostgreSQL and
  Redis. Transfer browser/integration harnesses now cover linked update/delete
  semantics and idempotency-key cleanup.
- Feedback was verified through JWT, scoped user API key, service-token agent
  lifecycle and changelog publication; Redis feedback audit events were observed.
- AI-entry browser regression fixtures now generate ledger rows in the current
  account-local week, preventing calendar drift from hiding the expected grouped
  rows in future runs.
- Backend test fixtures were normalized with `gofmt`; `go vet ./...` and the
  warning-as-error Swift simulator build remain green.

### Release gates still open

- Provider-backed AI/OCR, real Google iOS OAuth credentials, authenticated
  staging/production smoke and physical-device validation remain required before
  a production release claim.

## [1.0.8] - 2026-10-01

### Fixed

- Finance Assistant now accepts a bounded batch of up to eight read-only tool
  calls from one model response. This prevents valid multi-query answers from
  failing when the provider returns five or more parallel finance lookups.

### Verification

- Added a regression test covering five tool calls in one provider response;
  backend tests and the frontend release gate remain required before rollout.

### Deployment

- API image `registry.dungxbuif.com/mypocket-api:release-46c48f91`
  (`sha256:2be9c9a189e7caee55f8b6fdd9e16a0a3dec1eaa4fff55f10501f65dcfded226`)
  is running in production; the service is 1/1 and `/api/v1/health` returns 200.

## [1.0.7] - 2026-10-01

### Added

- Travel Mode now has owner-scoped event CRUD, one active event per account,
  ordinary transaction auto-linking, and explicit transaction link/unlink.
  Deleting an event clears links without deleting ledger rows or changing
  report inclusion. Recurring, transfer, adjustment and AI approval paths do
  not auto-attach events.
- Credit entry categories now enforce applicable-wallet scope in both the
  handler and repository transaction.

### Verification

- `go test ./...`, `go generate ./cmd/api`, `npm run check:design`,
  `npm run typecheck`, `npm run test:design`, `npm run test:transactions`,
  `npm run test:calendar`, `npm run test:transaction-jars`, and `npm run build`
  pass locally. PostgreSQL integration tests remain environment-gated by
  `TEST_DATABASE_URL`.

### Deployment

- No production deployment is claimed by this local implementation commit.

## [1.0.6] - 2026-09-30

### Fixed

- Transaction ledger wallet scope now follows the supplied wallet-selector reference:
  the compact wallet capsule opens the shared `Chọn Ví` sheet with aggregate,
  included/excluded wallets, real balances, add-wallet and disabled-link states.
  Selecting a wallet keeps the existing account-local week/custom ledger behavior.

### Verification

- `npm run test:wallet-scope`, `npm run check:design`, `npm run typecheck`,
  `npm run test:design`, `npm run test:transactions`, `npm run test:calendar`,
  `npm run test:transaction-jars`, and `npm run build` pass locally.

### Deployment

- Production web service converged on
  `registry.dungxbuif.com/mypocket-web:feedback-ui-wallet-design-20260930-0700`
  (`sha256:933dd62630e3ee5ae7917c6ddaab36b3d858bc58f2fdb0abb87cc2dd5e36d80e`).
  Public HTML serves the new bundle and `/api/v1/health` returns 200.

## [1.0.4] - 2026-09-30

### Fixed

- AI transaction-entry proposals now normalize date-only model output (`dd/MM/yyyy`,
  `dd-MM-yyyy`, or `yyyy-MM-dd`) in the account timezone instead of silently replacing
  a recognized date with today; the missing-time assumption is shown for review.
- Exact duplicate proposals within one OCR/text submission are merged while preserving
  review questions from the duplicate rows.

### Deployment

- API image `registry.dungxbuif.com/mypocket-api:feedback-api-ai-date-20260930-0011`
  was rolled out to production and verified healthy.

## [1.0.5] - 2026-09-30

### Added

- Login and Google OAuth responses now include rotating refresh credentials;
  `POST /api/v1/auth/refresh` consumes a refresh token once and issues a new
  access/refresh pair.
- The web client retries one expired-session request through the refresh route
  and keeps API-key requests out of the refresh flow.

### Deployment

- API `feedback-api-refresh-20260930-0041` and web
  `feedback-ui-refresh-20260930-0032` were rolled out and their public health
  and invalid-refresh checks passed.

## [1.0.2] - 2026-09-30

### Fixed

- Increased the frontend API proxy read/send timeouts to 240 seconds so the
  bounded OCR/model processing window does not turn into an HTML 504 response.

## [1.0.3] - 2026-09-30

### Fixed

- Category search now promotes a parent group when the query matches a child
  category, so matching groups are visible at the top of the result.

## [1.0.1] - 2026-09-29

### Fixed

- Recent transactions on the overview now show the account-local date and time
  before wallet and note metadata.
- Feedback routes accept owner-scoped API keys with `feedback:read` and
  `feedback:write` scopes.

### Deployment

- Production API and web services were rolled out to the verified registry
  images after the backend and frontend checks passed.

## [1.0.0] - 2026-09-24

### Deployment note

- Homelab `money.dungxbuif.com` was redeployed on 2026-09-24 from the
  `feature/wallet-management` branch. Production PostgreSQL schema was reset
  and migrated to version 21; Redis DB 1 was flushed. A pre-reset PostgreSQL
  dump was kept on Pi5 under `/home/dungxbuif/backups/mypocket/`.
- Production now uses Docker images `homelab/mypocket-api:0ff84828-deploy-20260924015529`
  and `homelab/mypocket-web:0ff84828-deploy-20260924015529`. The old
  `mypocket-worker` Swarm service is scaled to zero because this code snapshot
  has no `cmd/worker` binary.
- Production AI entry, receipt OCR, and Finance Assistant are enabled through
  server-side environment variables. MyPocket reuses the shared homelab LLM and
  OCR credentials; no provider keys are exposed to the browser or committed docs.
- The production web proxy accepts the same 101 MiB AI-entry upload envelope as
  the backend so receipt images/PDFs are not rejected before OCR processing.
- The production client now normalizes account and AI-entry timezones before
  submission, falling back to `Asia/Ho_Chi_Minh` when a browser/webview reports
  an empty or non-IANA timezone such as `Local`.

### Added

- Authenticated wallet, category, transaction, budget, jar, account-timezone, feedback,
  and local Finance Assistant slices are present in the current Stage v1 codebase.
- AI transaction entry supports text and receipt review proposals with wallet/category
  suggestions, duplicate handling, explicit user review, and bulk save controls.
- Finance Assistant exposes owner-scoped read-only finance tools, conversation history,
  typed cards, local API-key scopes, and best-effort Redis access audit metadata.
- Transaction ledger supports aggregate/basic-wallet weekly and custom periods, savings
  wallet history, provisional credit history, wallet filtering, and internal transfer
  entry from the transaction options menu.
- Design documentation now follows `system / atoms / molecules / organisms / templates /
  pages / references`; owner PNG/HTML bundles are retained as non-runtime evidence.

### Changed

- The active release narrative is reset to `1.0.0`; old release prose is retained only
  in Git commits, not carried forward as a second active release history.
- Design contracts use page terminology for route-level specifications and a dedicated
  references tree for visual evidence.
- AI provider/model configuration remains environment-driven; credentials are not stored
  in documentation or source control.

### Fixed

- Corrected wallet-scoped weekly ledger filtering so totals and rows use the same
  account-local period predicate.
- Corrected AI-entry empty/503 failure paths for bounded model output and nullable usage
  initialization; incomplete proposals remain reviewable instead of being auto-saved.
- Corrected OCR attachment handling for validated image/PDF inputs and bounded wait/poll
  behavior without exposing provider secrets or financial payloads in logs.

### Documentation

- Added the approved release-baseline detail design and inventory.
- Reconciled design indexes, page contracts, system behavior links, work tickets, backlog,
  validation matrix, context, and agent-readable transaction guidance.

### Known release gates

- Production deployment and public release approval are not claimed by this entry.
- Finance Assistant SSE/reconnect, durable Redis audit delivery/alerting, configured-
  provider browser proof, and full owner UAT remain open where the validation matrix says
  pending or not run.
- Credit-card statement/payment accounting, recurring transactions, advanced reports,
  and Travel Mode runtime remain separate incomplete scopes.
- Frontend `npm run check:design`, lint, build, and tests must be rerun in an environment
  with Node installed; this cleanup environment reported `env: node: No such file or
  directory`.
