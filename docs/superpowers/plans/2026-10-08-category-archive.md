# Category Archive Implementation Plan

> **For agentic workers:** Execute inline with `superpowers:executing-plans`; keep every test-first step and run all listed gates. Do not commit or deploy during this task.

**Goal:** Add reversible, owner-scoped archive state to personal categories and make web/iOS management and selectors honor it without changing existing financial history.

**Architecture:** Backend stores `is_active` and enforces owner-scoped status changes plus assignment validation. Web and iOS consume all visible categories for historical resolution, but filter inactive categories out of new-selection affordances and expose archive/restore in management.

**Tech Stack:** Go/GORM/PostgreSQL migrations and tests; React/TypeScript; SwiftUI/XCTest; generated Swagger/OpenAPI.

**Spec:** `docs/superpowers/specs/2026-10-08-category-archive-design.md`

## Global Constraints

- Existing categories migrate as active; no production DB or migration is touched.
- Historical transactions, reports, month summaries, saved budgets and recurring schedules remain unchanged.
- Archive/restore is personal-category-only and owner-scoped; system categories remain active.
- New category selection excludes inactive categories; existing saved references may be retained as specified.
- Preserve current uncommitted changes and do not stage or commit.
- Reconcile internal and public human/AI-readable docs and generated OpenAPI.

## Review Focus

- An archived parent with an active child remains hierarchy context, not a selectable category.
- A transaction editor can retain its unchanged archived category but cannot switch to another archived category.
- A pending AI proposal cannot create a transaction after its category is archived.
- Existing budget and recurring records keep working with archived category references.
- Another owner or a system category cannot be archived or restored, and the response does not disclose a foreign category.

### Task 1: Backend schema and category status contract

**Files:** `backend/migrations/000030_category_active.up.sql`, `backend/migrations/README.md`, `backend/internal/entity/wallet.go`, `backend/internal/repository/category.go`, `backend/internal/infrastructure/repository/category_postgres.go`, `backend/internal/usecase/category.go`, `backend/internal/controller/http/category_handler.go`, category tests, `backend/internal/controller/http/routes_category.go`, Swagger outputs.

**Interfaces:** Add `Category.IsActive bool`; repository `SetActive(ownerID, id string, active bool) (*entity.Category, error)`; interactor `SetActive(ownerID, id string, active bool) (*entity.Category, error)`; status route `PATCH /api/v1/categories/:id/status` with `{is_active:boolean}`. Keep list behavior returning all visible records.

**Response compatibility:** The status endpoint returns a complete category model, including `wallet_ids` as a JSON array (empty when unrestricted/no explicit links), so native clients can decode and reuse the response without a follow-up read.

- [x] Write tests proving idempotent status updates, foreign/system isolation, and `category_inactive` validation at transaction, credit, budget, recurring and AI-approval boundaries.
- [x] Run affected Go tests and verify RED specifically on the missing active field/status route/validation; resolve positive-fixture defaults revealed by the new `IsActive` field.
- [x] Implement additive migration, entity/repository/use-case/handler/route and assignment validation.
- [x] Add PostgreSQL assertions for the migration's active default, owner/system isolation, idempotent archive/restore and unchanged transaction references; run against a temporary isolated cluster.
- [x] Run `go test ./...`; regenerate Swagger and inspect route/schema diff.

### Task 2: Web management, selectors and API docs

**Files:** `app/src/services/categories.ts`, `app/src/atomic/organisms/GroupManagementPanel.tsx`, `app/src/atomic/organisms/GroupEditorPage.tsx`, shared category tree/selector and every consumer in transactions, credit, budgets, recurring and AI entry; web tests; `docs/architecture/API.md`, public docs index/agent surface, category design docs, changelog.

**Interfaces:** Extend `Category` with `is_active`; add `setCategoryActive(id, active)`; management fetch continues to resolve all categories; selectors exclude inactive nodes but preserve inactive ancestors as context.

- [x] Add behavior tests for archive visibility, hierarchy context and selection exclusion.
- [x] Implement management/editor actions using existing base components and API client.
- [x] Add regression tests for retaining an existing archived category and excluding it from new selections.
- [x] Run focused category selection tests, design/type checks and production build.
- [x] Update human-facing and AI-readable public API documentation and regenerate docs artifacts.

### Task 3: Native iOS category management and selectors

**Files:** `mobile-ios/MyPocketIOS/Core/Models/FinanceModels.swift`, `Core/API/FinanceAPI.swift`, `Features/Secondary/SecondaryFeatureViews.swift`, shared category tree/picker bases, impacted transaction/budget/recurring/credit/AI-entry views, XCTest/UI tests and native design docs.

**Interfaces:** Decode `CategoryModel.isActive`; expose `FinanceAPI.setCategoryActive(id:active:)`; use one shared active-selection policy for all category picker flows.

- [x] Add model/policy tests proving inactive exclusion, ancestor context and archived edit retention.
- [x] Add API request contract for status endpoint and category management filter/archive/restore controls.
- [x] Implement shared filtering and management actions using existing base components.
- [x] Add tests for transaction/budget/recurring/credit/AI picker policy and existing archived-category edit retention.
- [x] Extend `testAPIBackedCategoryCreateUpdateArchiveRestoreDeleteFromNativeUI` to archive a created category, verify default filtering and archived-state rendering, restore it through the native editor, and verify both status transitions from the API.
- [x] Run the archive/restore UI flow on iPhone 16 Simulator against a disposable local API/database; both API-backed status transitions and visible filtering passed. The selected UI test remains red only on its independent feedback-bubble overlap assertions; base-component, project verification and Release warnings-as-errors gates pass in the validation matrix.

### Task 4: Integration and docs reconciliation

**Files:** `docs/work/PORT-100-AUDIT.md`, `docs/work/VALIDATION_MATRIX.md`, `docs/work/BACKLOG.md`, `docs/CONTEXT.md`, category screen/base specs, `docs/releases/CHANGELOG.md`, public API/AI-readable docs, and `backend/migrations/README.md`.

- [x] Add isolated-database handler-to-PostgreSQL proof in `backend/internal/controller/http/category_handler_test.go`: archive through the category status route, confirm the archived category remains visible in category reads, reject a transaction create with `category_inactive`, preserve transaction/budget/recurring category references, run an existing recurring schedule while archived, restore through the route, and allow a new transaction. Ran with the full migration chain on a temporary PostgreSQL cluster.
- [x] Run backend full suite (265 tests/16 packages, including isolated PostgreSQL integration), web focused category tests and production build, and native focused Simulator tests (4/4). Full native UI suite remains a separate gate.
- [x] Reconcile design, API, Swagger, changelog, audit and validation docs; no production/remote data was touched.
- [x] Leave changes uncommitted and undeployed for owner review.
