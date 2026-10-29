# MyPocket Gap Closure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the remaining release gaps without changing financial data contracts or claiming production/device verification that has not run.

**Architecture:** Keep the Go API as the source of truth and keep iOS owner-scoped. Production UAT is read-only and credential-gated; native changes are isolated by feature area. New behavior is added only where an existing API/entity contract already exists; unsupported product scope is documented as partial instead of being guessed.

**Tech Stack:** Go, PostgreSQL, SwiftUI, XCTest, OpenAPI, existing OCR/AI provider adapter.

**Spec:** Current gap audit in `docs/CONTEXT.md`, `docs/requirements/REPORTS.md`, `docs/guides/offline-sync-api.md`, and `docs/work/VALIDATION_MATRIX.md`.

## Global Constraints

- Preserve owner scoping and API-key/JWT authorization.
- Do not reset, delete, or rewrite financial data.
- Do not introduce a new migration unless an existing contract requires it.
- Do not treat simulator signing as physical-device release proof.
- Do not expose provider credentials or user financial data in logs.

## Review Focus

- Missing/expired owner credentials must fail closed in production probes.
- A month-summary failure must not render a misleading partial report.
- Offline retry must remain idempotent across reconnect/restart.
- AI/OCR fixtures must preserve duplicate rows and missing-date review state.
- Budget/savings behavior must not invent recurrence or forecast semantics absent from the current contract.

### Task 1: Production read-only UAT and release evidence

**Files:**
- Modify: `docs/operations/DEPLOYMENT.md`
- Modify: `docs/CONTEXT.md`
- Test: production read-only HTTP probes (no source mutation)

- [ ] Run health, Swagger, route-method, migration, worker-log, and Redis probes.
- [ ] Run authenticated GET probes only if a safe owner credential is already available.
- [ ] Record 401/blocked cases without printing credentials.
- [ ] Update deployment evidence and remaining gates.

### Task 2: Offline/sync contract verification

**Files:**
- Modify: `docs/guides/offline-sync-api.md` only if behavior differs from code.
- Test: existing Go offline/sync tests and signed native XCTest.

- [ ] Verify capability response, receipt, idempotency, reconnect, retry, and owner isolation.
- [ ] Add only missing regression tests for existing behavior.
- [ ] Record production/device UAT as pending when credentials or device signing are unavailable.

### Task 3: Reports hardening

**Files:**
- `mobile-ios/MyPocketIOS/Core/API/FinanceAPI.swift`
- `mobile-ios/MyPocketIOS/Features/Secondary/SecondaryFeatureViews.swift`
- `mobile-ios/MyPocketIOSTests/FinancialSnapshotCacheTests.swift`

- [x] Bind `expectedOwnerID` and fail-safe report loading when month summary fails.
- [x] Run focused signed XCTest and update validation/changelog docs.

### Task 4: Savings and budget scope audit

**Files:**
- Modify docs under `docs/requirements/` and `docs/work/`.
- Modify code only when the existing entity/API contract has a concrete missing path.

- [ ] Confirm goal-wallet progress, transfer, notifications, and savings report contracts.
- [ ] Confirm current budget interval semantics and explicitly mark recurrence/forecast/drill-down as unsupported when no contract exists.
- [ ] Add tests/docs for any concrete existing-path fix; do not invent new financial semantics.

### Task 5: AI/OCR evaluation fixtures

**Files:**
- Modify: `evals/` or existing AI test fixtures only.
- Modify: `docs/architecture/AI_ADVISOR_TOOLS.md` and release notes if coverage changes.

- [ ] Add deterministic fixtures for blurry/duplicate/missing-date/wallet-inference cases.
- [ ] Assert tool arguments, owner scope, duplicate review state, and no auto-approval.
- [ ] Keep live provider execution opt-in and redact all credentials/output.

### Task 6: iPhone release gate

**Files:**
- Modify: `docs/operations/DEPLOYMENT.md` only for evidence.

- [ ] Run signed simulator tests with signing enabled.
- [ ] Check connected-device team/profile state.
- [ ] If provisioning is absent, report the exact blocker; do not weaken entitlements or claim installation.

## Verification Gate

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- Focused signed iOS XCTest plus build-for-testing.
- `git diff --check` and docs reconciliation.
