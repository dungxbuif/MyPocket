# Native iOS Deep Links and Auth Gate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add stable deep-link routes for native iOS scenes with an authentication and owner-authorization gate.

**Architecture:** A pure `AppRouteParser` converts supported URL scheme/universal-link URLs into `AppRoute` values. `AppRootView` owns the pending route and passes it through an auth gate; `MainTabView`/`MoreView` apply the route to existing navigation without moving API ownership logic into the router.

**Tech Stack:** Swift 6, SwiftUI, XCTest, existing `AuthViewModel`, `NavigationStack`, and `FinanceAPI` owner-scoped services.

**Spec:** `docs/superpowers/specs/2026-10-15-native-routing-auth-design.md`

## Global Constraints

- Do not accept credentials or tokens in URLs.
- Protected routes require an authenticated owner session; locked sessions must unlock first.
- Resource identifiers from URLs are navigation hints and never override the session owner.
- Do not change backend contracts or production configuration.
- Keep existing `MoreDestination` screens and base components; route logic is an adapter.

## Review Focus

- Malformed or foreign-host URLs must not change navigation (parser tests).
- A link received before restore completes must not be lost or opened unauthenticated (pending-route tests).
- A locked session must not reveal the destination before Face ID succeeds (auth-gate tests).
- A second link must replace a pending link deterministically without replaying the first (one-shot tests).
- Wallet IDs must remain owner-scoped after deep-link navigation (UI/API boundary test).

### Task 1: Route value and parser

**Files:**
- Create: `mobile-ios/MyPocketIOS/Core/Routing/AppRoute.swift`
- Test: `mobile-ios/MyPocketIOSTests/AppRouteTests.swift`
- Modify: `docs/architecture/MOBILE_IOS.md`

**Interfaces:**
- `enum AppRoute: Equatable, Sendable`
- `struct AppRouteParser { static func parse(_ url: URL) -> AppRoute? }`
- `AppRoute` cases include overview, transactions(walletID), newTransaction(walletID), wallets, wallet(id), budgets, reports, assistant, account, feedback.

- [ ] Write failing parser tests for both URL forms, all secondary scenes, wallet query decoding, unknown paths, credentials in URL, and foreign host rejection.
- [ ] Run `xcodebuild test ... -only-testing:MyPocketIOSTests/AppRouteTests` and confirm the new tests fail because the route type is absent.
- [ ] Implement the minimal route enum/parser with strict scheme/host/path validation.
- [ ] Re-run the focused tests and confirm all parser cases pass.
- [ ] Document canonical links and examples in `MOBILE_IOS.md`.

### Task 2: Pending-route auth gate

**Files:**
- Create: `mobile-ios/MyPocketIOS/Core/Routing/AppRouteCoordinator.swift`
- Test: `mobile-ios/MyPocketIOSTests/AppRouteCoordinatorTests.swift`
- Modify: `mobile-ios/MyPocketIOS/App/AppRootView.swift`

**Interfaces:**
- `@MainActor final class AppRouteCoordinator: ObservableObject`
- `@Published private(set) var pendingRoute: AppRoute?`
- `func receive(_ route: AppRoute)` and `func consumeIfAuthenticated(_ phase: AuthPhase) -> AppRoute?`

- [ ] Write failing tests for checking, unauthenticated, locked, authenticated and expired phases, plus one-shot consumption/replacement.
- [ ] Run the focused coordinator tests and verify the expected red failures.
- [ ] Implement the smallest state machine; do not navigate directly from the parser.
- [ ] Wire `AppRootView.onOpenURL` to parse and receive routes, consuming only after the auth phase is ready.
- [ ] Run the focused coordinator tests again.

### Task 3: Apply routes to existing shell navigation

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Shell/MainTabView.swift`
- Modify: `mobile-ios/MyPocketIOS/Features/More/MoreView.swift`
- Test: `mobile-ios/MyPocketIOSUITests/LaunchTests.swift`

**Interfaces:**
- `MainTabView` receives an optional `AppRoute` and exposes a single route-application method.
- Existing `MoreDestination` remains the feature-screen enum; mapping is centralized and tested.

- [ ] Add UI smoke tests for authenticated cold-start URLs for overview, transactions/new, wallet detail and assistant.
- [ ] Verify the tests fail before route application is wired.
- [ ] Map `AppRoute` to shell selection, wallet scope and `MoreDestination` without duplicating feature view logic.
- [ ] Verify the tests pass and that a second URL replaces the current pending route.

### Task 4: Full verification and docs reconciliation

**Files:**
- Modify: `docs/architecture/MOBILE_IOS.md`
- Modify: `docs/guides/face-id.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`

- [ ] Run focused unit tests, native UI route tests, and the existing iOS build/test command.
- [ ] Run `rtk git diff --check` and `mobile-ios/scripts/verify-project.sh`.
- [ ] Record exact commands/results and any simulator/device limitations in the validation matrix.
- [ ] Do not claim production deep links until associated-domain hosting and a signed build are verified.
