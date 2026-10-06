---
artifact_type: phase
id: PHASE-002-native-ios-swiftui
status: in_progress
owner: shared
created: 2026-10-05
updated: 2026-10-06
trace:
  backlog: ../BACKLOG.md
  roadmap: ../ROADMAP.md
  detail_design: ../../superpowers/specs/2026-10-05-native-ios-swiftui-design.md
  plan: ../../superpowers/plans/2026-10-05-native-ios-swiftui-port.md
  adr: ../../decisions/ADR-010-native-ios-swiftui-client.md
  validation: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Phase 002 — Native iOS SwiftUI parity

## Goal

Deliver a first-party SwiftUI renderer for the currently mounted MyPocket web
features, with base-component ownership and an iOS Simulator-first proof.

## Scope

- New `mobile-ios/` target on `codex/native-ios-port`.
- Typed API/auth core, mobile Google PKCE exchange, Keychain session and Face
  ID unlock.
- Overview, ledger, wallet, budget, report, jar/month, category, recurring,
  travel, portfolio, AI entry, Finance Assistant, feedback and account flows.
- XCTest, API contract tests, static base-component guard and Simulator UI
  smoke proof.

## Out of scope

- Deleting/moving web or Expo code.
- App Store/TestFlight release, physical signing, push/background sync and
  unapproved offline ledger writes.
- New database/finance rules unrelated to the mobile auth boundary.

## Entry and exit gates

- Entry: detail design and implementation plan are approved; isolated worktree
  is created; web branch remains clean; Xcode SDK is available.
- Exit: every parity row has a native screen/view model, base guard passes,
  backend/frontend regression suites pass, and a Simulator run is recorded. If
  no Simulator runtime is installed, phase remains `blocked` with the exact
  Xcode component prerequisite rather than being marked complete.

## Verification checkpoint — 2026-10-05

The simulator prerequisite is now available locally: iOS 18.4/18.5 runtimes
were detected and an iPhone 16 device was created as
`EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`. The current auth/base-component slice
was built, installed, launched, and screenshot-checked on that device.

Passing commands:

```bash
xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -destination 'id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' \
  test CODE_SIGNING_ALLOWED=NO
bash mobile-ios/scripts/check-base-components.test.sh
bash mobile-ios/scripts/verify-project.sh
go test ./...
go generate ./cmd/api
```

This is a historical checkpoint, not the phase exit gate. The feature-slice
checkpoint below records the API-backed screens added after this first smoke;
the remaining exit work is the authenticated staging/production validation
called out in the latest verification section.

## Feature-slice checkpoint — 2026-10-05

The native branch now contains API-backed SwiftUI routes for overview,
transactions/manual entry, wallets, budgets, reports, categories, jars, month
summary/notes, recurring schedules, Travel Mode, portfolio, credit statements,
Finance Assistant, AI entry review, feedback/changelog and account settings.
All user controls in these routes compose the project-owned DesignSystem; the
static guard and Simulator tests remain passing. The remaining exit work is
an authenticated end-to-end smoke run against a fixture/staging API and a
final manual review of backend-specific report payloads. Native UI coverage
now includes deep CRUD/link flows, replay-safe multipart auth, persisted
assistant/feedback entry points, Face ID restore gating, and typed Finance
Assistant cards.

The Debug-only `-UITestAuthenticatedAPI` launch argument now exchanges the
local fixture OAuth payload, persists the returned session in Keychain and
renders the authenticated overview through the real API client. This closes
the simulator smoke gap caused by the older fake-token fixture without
introducing a production bypass; the local validation screenshot is recorded
in the validation matrix.

The report contract follows the mounted backend/web Money Insider payload:
category totals, top expense rows, previous-period delta and spending/income
ratio. Merchant/cashflow cards are not claimed because they are not part of
that API contract.

Wallet parity review requires the native editor to preserve goal target dates,
expose credit statement balance/statement-day/due-day metadata, and keep wallet
type plus opening balance immutable after creation. Selecting a wallet must
open an owner-scoped detail view rather than only showing a list card.

Transaction parity also requires account-timezone week/custom filtering and
local-date grouping. Opening Edit from a transaction detail must retain the
selected transaction and issue a PATCH, never silently create a new row.

Release configuration is part of the parity gate: XcodeGen must attach the
tracked Debug/Release xcconfig files, and the Google client ID plus callback
URI must be exported into the app plist. A Release build must therefore use
the production API base URL and fail visibly when an iOS OAuth client ID is
still a placeholder, rather than silently falling back to localhost or an
empty client configuration.

Category editing must preserve the tree contract: a user can choose a parent
category when creating or editing a child. Recurring schedules must preserve
note text and submit account-local date/time as RFC3339 instants; date-only
strings are not a valid backend schedule input.

Credit entry/payment forms must preserve the web contract: an entry chooses its
kind and an applicable category, while a payment chooses a non-credit source
wallet; both actions accept amount and optional note. The native screen must not
silently drop category or note fields when calling the credit APIs.

Recurring schedule parity also follows the web constraints: credit wallets are
not valid recurring sources, category options are filtered to the selected
expense/income type, and switching type clears a category that no longer
matches the selected type.

Shared selectors must present localized semantic labels while retaining their
raw API enum values for request payloads. Jar allocation mode changes must clear
the previous mode's draft and reject a zero percentage; the valid percentage
range is greater than zero through 100, matching the web editor.

Ordinary transaction and AI proposal editors must never offer credit wallets.
Changing the wallet or income/expense type clears an incompatible category, and
credit rows opened from aggregate ledger detail are read-only; entry/payment
changes use the dedicated credit statement flow.

Budget editors expose only expense categories and non-credit wallets; balance
adjustments use the same non-credit wallet boundary.

An ended budget may be viewed and deleted but cannot be edited or saved; active
budgets retain the normal edit flow.

Portfolio trade dates are account-local calendar dates: the selected day is
serialized at local noon, not by blindly using the device's current instant.

Selecting a credit wallet from the main wallet list must route to the same
statement plus entry/payment surface as the dedicated credit route; a read-only
credit detail is not sufficient for parity.

The shared native API client must preserve query strings when a feature builds
an `APIRequest` path. A query delimiter belongs in `URLComponents`' query
field, not in the escaped path; the contract is covered by an XCTest because
credit statements, reports, jars and Assistant history all depend on it.

Basic and goal wallet details must preserve the web detail behavior: show
account-timezone monthly transaction groups with category/note labels and
income/expense totals, and allow adding or editing ordinary transactions from
the detail screen through the shared transaction editor. Credit remains on its
dedicated statement/payment panel.

The transaction ledger must keep transfer/adjustment actions behind one shared
overflow menu, support switching date/category grouping, include wallet and
category context in rows, and avoid showing ordinary week/custom controls for
goal or credit selections. Selecting a goal wallet opens its detail surface;
selecting a credit wallet opens the statement with entry/payment actions rather
than exposing the ordinary transaction editor.

Ledger loading treats the transaction and wallet lists as the required core
payload while category labels and profile timezone are optional metadata. A
metadata failure must not make a valid transaction list unusable.

Portfolio parity requires the summary to expose cost basis, current market value
and unrealized P/L when the API provides them. Current-price drafts are keyed
per asset and may be cleared to `nil`; portfolio trades must preserve the user
selected trade date rather than silently replacing it with the current instant.

AI entry accepts either a text instruction or one or more validated image/PDF
attachments, uses the account timezone for extraction context, and keeps every
proposal in review until an explicit approve/reject decision.

Category parent selection must preserve a tree: editing a category cannot select
itself or any descendant as its parent, preventing cycles before the API call.

AI entry must preserve the one-shot request contract: a processing response is
polled by its request ID, reload uses GET only, and the UI never submits the
same multipart payload again automatically. Transaction/report dates displayed
or filtered by native screens must use the account timezone rather than the
simulator/device timezone.

Identity-bearing selectors use the shared `MPOptionSelector` base component
with stable wallet/category IDs; duplicate display names no longer select an
arbitrary record. Credit statement dates use the account timezone, and wallet
detail resolves the selected wallet by ID after refresh.

## Parity completion pass — 2026-10-05

The simulator-first parity pass closed the highest-risk interaction gaps:

- transaction edit, travel-event link/unlink, and transfer-safe detail actions;
- jar create/update/remove and cumulative report navigation;
- portfolio buy/sell trade entry with quantity/price validation;
- AI entry proposal editing before approve/reject and image/PDF thumbnails;
- authenticated UI fixtures for primary routes, feedback overlay and regression
  coverage for the shared base-component contract;
- Face ID preference persistence and lock/unlock after session restore.

These changes preserve the existing API owner scope and do not add offline
ledger writes. Verification is the full iOS Simulator XCTest suite, static
base-component guard, project verification, and backend contract tests when
shared API contracts are touched.

## Simulator verification — 2026-10-06 (latest)

The iPhone 16 simulator `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D` is booted and
the latest native build was installed and launched successfully. The full
suite passed: 48 unit/contract tests and 14 signed API-backed UI tests
(auth launch, authenticated primary tabs, all secondary destinations, AI and
Assistant surfaces, feedback, wallet/transaction mutations, transfer pair
deletion and credit entry/payment refresh). The shared
`MPButton` now explicitly owns its rectangular rendered hit shape; this fixed
compact More rows that were visible in the accessibility tree but did not
consistently route taps. Ledger filters reject malformed timestamps instead of
including them in a selected period; wallet detail renders every row returned by
the API; and a Finance Assistant run remains cancellable after the 30-second
local polling window instead of being silently detached.

The latest run also verified category-parent cycle prevention, stable ID-based
wallet/category selectors, account-timezone boundary formatting, recurring
RFC3339/note payloads, typed API error presentation and release configuration
changes. The app was relaunched after the test run with bundle
`com.mypocket.ios` for manual simulator inspection.

The post-commit rerun on 2026-10-06 07:13–07:23 ICT repeated the complete
cross-surface gate: backend race tests, web design/type/build/API/AI/transfer
checks, native 48/48 unit tests, signed `LOCAL_API_UI` UI 14/14 in 293.319s,
the Release warnings-as-errors build, base-component/project guards, and a
200 response from `/api/v1/health`. The local disposable fixture was empty
after teardown. Provider-backed AI/OCR, real Google OAuth, authenticated
staging/production and physical-device validation remain external gates.

The follow-up system-category scope fix on 2026-10-06 corrected native parity
with the web/backend contract: system metadata is read-only, delete is hidden,
and wallet applicability uses `PATCH /api/v1/categories/{id}/wallets`.
The native unit suite passes 49/49 and the signed simulator route sweep passes
all 13 More destinations (1/1 test, 121.909s). A second full API-backed UI
rerun was not claimed after one simulator diagnostic-service hang; the prior
14/14 proof remains valid for the unchanged API-backed flows.

The follow-up Hũ chi tiêu scope fix on 2026-10-06 restored the bounded
account-local month selector, actual/allocated/unassigned metrics, advisory
over-allocation warning, active/history action gating, and selected-jar
cumulative report. Empty month responses clear the selected report so stale
data cannot cross month boundaries. Native XCTest passes 50/50, the Release
warnings-as-errors build passes, and the signed More-route sweep passes 1/1 in
123.999s. The prior complete API-backed suite remains 14/14; a full rerun
after this UI-only change was not claimed because the Xcode diagnostic service
hung once.

```bash
xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -destination 'id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' \
  CODE_SIGNING_ALLOWED=NO
bash mobile-ios/scripts/check-base-components.test.sh
bash mobile-ios/scripts/verify-project.sh
cd backend && go test ./... && go generate ./cmd/api
```

The authenticated UI run uses the DEBUG-only `-UITestAuthenticated` fixture;
it proves native navigation, state rendering and feedback entry without
placing a real JWT or API secret in the app. A staging/production authenticated
API smoke test remains a release prerequisite.

The DEBUG-only `-UITestAuthenticatedAPI` launch argument was also exercised
after resetting the simulator. It exchanged fixture PKCE fields against the
local API, persisted the disposable JWT in Keychain and rendered the overview
through live URLSession requests; no production credentials or data were used.

## Native parity follow-up — 2026-10-06

Two audit findings were fixed before the simulator test pass:

- `AuthViewModel.restore()` now checks access and refresh expiry. A valid refresh
  token is consumed once through `POST /auth/refresh` before Face ID can lock the
  refreshed session; an invalid refresh state clears Keychain and biometric
  preference rather than presenting an authenticated shell.
- Category management now composes the shared `MPMultiOptionSelector`. Editing
  a category round-trips every `wallet_ids` value through the existing category
  API; an empty set continues to mean all owner wallets. This closes the native
  parity gap for categories shared by multiple wallets.
- If a request receives `401` and the rotating refresh endpoint also rejects the
  token, `APIClient` clears the Keychain session and notifies `AuthViewModel` so
  the UI returns to the login gate instead of retaining a stale authenticated
  shell.
- Wallet detail month headers now use the shared `MPFormat.monthLabel`, with
  shape/range validation instead of leaking interpolation source text.
- Travel Mode drafts now trim required text and reject malformed or reversed
  optional calendar ranges before an API request.

Verification on the same simulator:

```bash
xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -destination 'id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' \
  -only-testing:MyPocketIOSTests -enableCodeCoverage NO
xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -destination 'id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' \
  -only-testing:MyPocketIOSUITests -enableCodeCoverage NO
```

Both commands pass (41 unit/contract tests and the full UI suite: 10 tests, 5
passed and 5 explicit API-fixture skips). The same pass
also covers typed `204 No Content` mutation responses and native bulk-delete
selection/confirmation/error retention. Provider-backed
AI/OCR, real Google iOS OAuth, authenticated staging/production and physical
device checks remain external release gates.

The Finance Assistant history/overview parity was then verified with an opt-in
API-backed UI test on that simulator:

```bash
xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -destination 'id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' \
  SWIFT_ACTIVE_COMPILATION_CONDITIONS='DEBUG LOCAL_API_UI' \
  -only-testing:MyPocketIOSUITests/LaunchTests/testAPIBackedAssistantLoadsOverview
```

The test passed against the disposable local `:8080` API. It launches with the
DEBUG-only `-UITestAuthenticatedAPI` fixture, opens Finance Assistant, verifies
the 200 overview/capability contract and asserts that no overview error is
shown. The default UI suite keeps this opt-in test skipped unless the explicit
`LOCAL_API_UI` compilation condition is supplied.

The same opt-in configuration also ran
`testAPIBackedWalletDetailExposesTransactionEditor` with a disposable basic
wallet. The smoke opened the wallet detail, verified the add-transaction editor
and its category/note controls, and removed the seeded wallet with a 204 API
response after the run.

The opt-in API-backed UI fixtures are now self-seeding: wallet-detail and
credit-ledger tests exchange the local fixture session, create their own
owner-scoped wallet through `POST /wallets`, and register async teardown via
`DELETE /wallets/:id`. The full `LOCAL_API_UI` UI suite passed 10/10 with no
skips; a post-run local API check found zero `Fixture*` wallets. This remains
local disposable proof, not production validation.

AI-entry capability parity is also covered: the shared document picker remains
disabled until the API reports private OCR/storage readiness, while text-only
entry remains enabled. The pure policy regression and the opt-in AI-entry UI
test both pass against the capability-disabled local fixture.

The native ledger week-period audit then removed the incorrect all-time tab and
added the shared account-local Monday–Sunday week navigator, bounded at the
current week, plus the web-matched `Tiền vào`/`Tiền ra` cashflow card. The pure
filter/label tests and an opt-in iPhone 16 UI assertion pass; goal and credit
scopes still route to their specialized bodies.

The goal-wallet target-date audit also corrected date-only parsing and
serialization to use the account IANA timezone. The regression test confirms a
goal date round-trips to the same calendar day independently of device timezone.

## Disposable local API smoke — 2026-10-05

Before any release decision, the typed client contracts were exercised against
a fresh local PostgreSQL database (`mypocket_ios_test`) with the normal Go API
and Redis running. The smoke covered password login/profile and refresh-token
rotation, wallet variants (basic/goal/credit), category CRUD, transaction
create/update/list, atomic transfer update/delete with idempotency keys,
adjustments, budgets, jars and cumulative reports, recurring schedules,
Travel Mode lifecycle, portfolio assets/prices/trades, credit entries and
payments/statements, user feedback, changelog, API-key create/list/revoke and
AI capability responses. Every asserted request returned the expected 2xx
contract; unauthenticated wallet access returned 401. AI/OCR and Advisor were
correctly reported disabled because no provider secrets were injected into the
disposable environment.

This smoke is intentionally non-production and does not replace provider-backed
AI/OCR verification or a real authenticated staging/production iOS run.

## Latest cross-surface regression — 2026-10-05

The simulator-first pass was followed by a fresh local integration run. The Go
backend passed `go test -race ./... -count=1` against disposable PostgreSQL and
Redis. The FE-proxy transfer harness now verifies the complete mutation surface:
paired creation, balance/report invariants, idempotent PATCH, and atomic pair
DELETE. The real Chromium 390×844 flow passed against the same API after its
cleanup was corrected to remove linked transfer rows by `transfer_id` rather
than attempting invalid single-row deletes.

The feedback flow was exercised end to end with a JWT, a scoped user API key,
the service-token agent routes, lifecycle transitions and changelog publication;
the resulting Redis feedback audit stream entries were observed. The local test
feedback/changelog records were removed afterwards. This closes local contract
coverage but does not claim provider-backed AI/OCR, real Google OAuth, a
physical-device run or authenticated staging/production validation.

The web AI-entry browser regression also passed after correcting its fixture to
derive the current account-local week. The prior failure was calendar drift in
the test data (rows dated outside the period selector), not a production UI or
API defect.

## Simulator rerun — 2026-10-06

The iPhone 16 simulator run was repeated after the goal target-date timezone,
recurring draft-validation, jar allocation contract, portfolio trade
validation, request-correlation, transfer wallet/date and wallet-detail state
fixes. Unit/contract XCTest passed 47/47. The opt-in API-backed UI suite passed
10/10 in 200.110 seconds, including the AI-entry capability gate, Assistant
overview, credit ledger, week/cashflow ledger, wallet transaction editor, all
13 secondary destinations, feedback entry, primary tabs and transfer/adjustment
overflow actions. The recurring editor's zero/invalid amount, interval,
metadata and reversed end-date guards, jar safe-integer/finite-precision
guards, portfolio trade numeric/side guards, API request-ID reuse across
refresh, transfer wallet/date boundaries and wallet summary refresh are covered
by the tests. The
disposable fixture check returned zero `Fixture*` wallets.

The first diagnostic invocation incorrectly added `CODE_SIGNING_ALLOWED=NO` to
the UI test command. The local API returned 200, but Keychain persistence failed
with `KeychainError` and the app stayed on the login gate. Re-running with the
normal simulator ad-hoc signing path passed. Keep `CODE_SIGNING_ALLOWED=NO`
limited to pure unit/build checks; API-backed UI tests require the signed
simulator test runner because they persist the disposable session in Keychain.

The rerun also hardened numeric draft parsing. The shared parser rejects
malformed integer input instead of silently filtering characters across
transaction, transfer, adjustment, wallet, credit, jar, budget, assistant
proposal and portfolio price/trade editors. The shared numeric amount field
now exposes a keyboard “Xong” action so credit notes remain reachable while a
number pad is open. Native XCTest passed 47/47. The latest signed mutation
smoke covers wallet create/update, transfer create/delete and credit
entry/payment; the complete signed API-backed UI suite passed 14/14 in 293.192
seconds. The app launched successfully on the iPhone 16 simulator;
physical-device, real Google OAuth, provider-backed AI/OCR and authenticated
staging/production gates remain external checks.

The follow-up goal-wallet savings-summary fix on 2026-10-06 restored the web
contract for reached/remaining amount, account-local target-day status,
bounded progress, and explicit target amount through shared base components.
Native XCTest passes 52/52, the Release warnings-as-errors build passes, and
the signed More-route sweep passes 1/1 in 124.194s. The complete API-backed
suite remains 14/14 from the preceding gate; the targeted route covers this
goal-detail presentation change.

The follow-up portfolio draft/refresh fix on 2026-10-06 now trims asset
symbol/name values, rejects whitespace-only drafts, and exposes the shared
header refresh action. Native XCTest passes 53/53, the Release
warnings-as-errors build passes, and the signed More-route sweep passes 1/1 in
124.144s. The complete API-backed suite remains 14/14 from the preceding gate;
the targeted route covers this portfolio presentation/input change.

The follow-up feedback draft-normalization fix on 2026-10-06 now trims
title/description and rejects unsupported types before JSON or screenshot
multipart requests. Native XCTest passes 54/54, the Release warnings-as-errors
build passes, and the signed feedback-entry UI test passes 1/1 in 9.821s.
