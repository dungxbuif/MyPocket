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

## Simulator verification — 2026-10-05 (latest)

The iPhone 16 simulator `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D` is booted and
the latest native build was installed and launched successfully. The full
suite passed: 24 unit/contract tests and 4 UI smoke tests (auth launch,
authenticated primary tabs, More route + floating feedback overlay, and a
fresh-launch sweep through all 13 secondary destinations). The shared
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

Verification on the same simulator:

```bash
xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -destination 'id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' \
  -only-testing:MyPocketIOSTests -enableCodeCoverage NO
xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -destination 'id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' \
  -only-testing:MyPocketIOSUITests -enableCodeCoverage NO
```

Both commands pass (24 unit/contract tests and 4 UI smoke tests). Provider-backed
AI/OCR, real Google iOS OAuth, authenticated staging/production and physical
device checks remain external release gates.

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
