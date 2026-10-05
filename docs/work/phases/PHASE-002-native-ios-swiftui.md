---
artifact_type: phase
id: PHASE-002-native-ios-swiftui
status: in_progress
owner: shared
created: 2026-10-05
updated: 2026-10-05
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

This is a checkpoint, not the phase exit gate: the feature-parity screens and
their API-backed view models still need to be implemented before claiming the
native port is complete.

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
suite passed: 11 unit/contract tests and 3 UI smoke tests (auth launch,
authenticated primary tabs, More route + floating feedback overlay).

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
