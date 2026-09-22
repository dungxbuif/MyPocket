# AI Entry, Account Navigation, and In-App Feedback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make receipt entry reliable and diagnosable, support 20 files with drag/drop, move budgets into Account, normalize transaction styling, and add private screenshot-backed floating feedback.

**Architecture:** Keep the idempotent AI-entry process and add typed, secret-free failure stages plus a stable session error code. Reuse the existing private S3 client for feedback screenshots under a separate key namespace, with owner/agent-scoped signed access. Keep the mobile shell five destinations by making Account the canonical parent for budgets.

**Tech Stack:** Go, Gin, GORM/PostgreSQL, existing S3-compatible storage, Redis audit, React 19, TypeScript, Vite, Tailwind v4, existing base components, `html-to-image` for DOM capture.

**Spec:** `docs/superpowers/specs/2026-09-22-ai-entry-feedback-account-design.md`

## Global Constraints

- Never log or persist API keys, signed URLs, screenshot bytes, raw OCR text, or feedback descriptions in Redis audit events.
- Never automatically resubmit an ambiguous or failed user receipt request.
- Keep the per-file limit at 5 MiB and the new file-count limit at 20; reject invalid input before provider calls.
- Capture only the MyPocket content root; browser chrome, other tabs, and OS-level screen content are out of scope.
- All repeated frontend controls use existing atoms/molecules or a new documented base component.
- Preserve owner scoping and API-key/service-token boundaries for every feedback and screenshot endpoint.
- Every code change updates the matching internal/public documentation and runs the relevant design/type/test checks.

## Review Focus

- OCR provider fails immediately after S3 upload: the process must expose a safe stage/code, mark attachments failed, and create no transaction; covered by Task 1.
- Twenty files exactly versus file 21 and oversized multipart bodies: validation must be deterministic and provider-free; covered by Task 2.
- A dropped file with the wrong MIME or duplicate content: the UI must reject/announce it without corrupting selected files; covered by Task 2.
- A legacy `/budgets` link and Account navigation: both must render the same budget screen without duplicate navigation items; covered by Task 3.
- Screenshot capture failure or cross-owner access: feedback submission must remain usable without an image and signed access must be rejected for the wrong principal; covered by Tasks 4–5.

### Task 1: Make OCR failures diagnosable without leaking provider data

**Files:**
- Modify: `backend/internal/entity/ai_entry.go`
- Modify: `backend/internal/usecase/ai_entry.go`
- Modify: `backend/internal/infrastructure/ai/client.go`
- Modify: `backend/internal/infrastructure/ai/ocr.go`
- Modify: `backend/internal/infrastructure/repository/ai_entry_postgres.go`
- Modify: `backend/internal/controller/http/ai_entry_handler.go`
- Create: `backend/migrations/000002_ai_entry_error_code.up.sql`
- Test: `backend/internal/infrastructure/ai/client_test.go`
- Test: `backend/internal/usecase/ai_entry_test.go`
- Test: `backend/internal/infrastructure/repository/ai_entry_postgres_test.go`

**Interfaces:**
- Produces `AIProviderError{Stage, Code, HTTPStatus, Err}` with `Error()` returning a user-safe message and `Unwrap()` returning the cause.
- Produces `AIEntrySession.ErrorCode string` serialized as `error_code`.
- `FailMessage` accepts both safe error text and the stable code; existing callers without a code use `provider_error`.

- [ ] **Step 1: Write failing unit tests** for OCR submission HTTP 4xx, OCR polling `failed`, model HTTP failure, schema mismatch, and session serialization of `error_code`.
- [ ] **Step 2: Run the focused tests to verify RED.**

  Run: `rtk env TEST_DATABASE_URL='postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable' go test ./internal/infrastructure/ai ./internal/usecase ./internal/infrastructure/repository -run 'AIProvider|OCR|ErrorCode|AIEntry' -count=1`

  Expected: failures for missing typed error and missing database column mapping.
- [ ] **Step 3: Add the typed provider error and map OCR/model/schema boundaries.** Preserve the current safe Vietnamese message, attach only stage/code/status, and emit `slog.Warn` fields `process_id`, `request_id`, `stage`, `provider_status`, and `model` without URLs, keys, OCR text, or response bodies.
- [ ] **Step 4: Add `error_code` to the session table/model/repository and return it in `processResponse`.** Keep existing rows compatible with a default empty string; map unknown legacy errors to `provider_error`.
- [ ] **Step 5: Run the focused tests to verify GREEN.**
- [ ] **Step 6: Run the synthetic live smoke test with a locally generated fixture only.** The command must set `AI_EVAL_LIVE=1` and a temporary synthetic PNG path, never use the user's uploaded files, and record only pass/fail and stage latency.
- [ ] **Step 7: Commit:** `fix: expose safe ai entry failure stages`.

### Task 2: Support 20 receipt files and drag/drop

**Files:**
- Modify: `backend/internal/infrastructure/ai/ocr.go`
- Modify: `backend/internal/controller/http/ai_entry_handler.go`
- Modify: `backend/internal/usecase/ai_entry.go`
- Modify: `backend/docs/docs.go`
- Modify: `app/src/atomic/atoms/BaseFileUpload.tsx`
- Modify: `app/src/atomic/molecules/AssistantComposer.tsx`
- Modify: `app/src/atomic/organisms/AiEntrySheet.tsx`
- Modify: `app/src/services/aiEntryLogic.ts`
- Create: `backend/internal/controller/http/ai_entry_limits_test.go`
- Create: `app/scripts/ai-attachments.test.mjs`
- Modify: `docs/design/atoms/file-upload/README.md`
- Modify: `docs/design/screens/assistant/README.md`

**Interfaces:**
- `MAX_AI_ENTRY_FILES = 20` is exported from the frontend validation module and mirrored by backend constants.
- `BaseFileUpload` accepts `maxFiles`, `onDrop`, and an accessible live status; the component owns no domain-specific validation.
- `validateEntryFiles(files, maxFiles = 20)` returns a stable Vietnamese error string or `null`.

- [ ] **Step 1: Add failing backend tests** for exactly 20 accepted files, 21 rejected files, 5 MiB boundary, and multipart body rejection before provider invocation.
- [ ] **Step 2: Add failing frontend tests** for picker/drop parity, duplicate filenames, MIME mismatch, and file 21 rejection.
- [ ] **Step 3: Run both test files to verify RED.**
- [ ] **Step 4: Raise backend limits and multipart envelope.** Set the count constant to 20, keep per-file 5 MiB, and set the body limit to `20*5 MiB + 1 MiB` with a matching `ParseMultipartForm` memory threshold. Update Swagger comments.
- [ ] **Step 5: Implement drag/drop in `BaseFileUpload` and wire `AssistantComposer`.** Prevent default browser navigation, merge dropped/selected files deterministically, preserve existing files until validation succeeds, and expose rejection text through `aria-live`.
- [ ] **Step 6: Update `AiEntrySheet` copy/capability gating.** Remove the stale “OCR not configured” warning when capabilities report configured; keep the warning only for a false capability and allow the 20-file selection.
- [ ] **Step 7: Run frontend/backend attachment tests, design checks, and typecheck.**
- [ ] **Step 8: Commit:** `feat: support twenty ai receipt attachments`.

### Task 3: Move Budgets into Account and normalize Transactions styling

**Files:**
- Modify: `app/src/atomic/organisms/BottomNavigation.tsx`
- Modify: `app/src/atomic/organisms/AccountPanel.tsx`
- Modify: `app/src/atomic/pages/FinancePrototypePage.tsx`
- Modify: `app/src/router.tsx`
- Modify: `app/src/atomic/organisms/TransactionsPanel.tsx`
- Create: `app/src/atomic/molecules/ScreenHeader.tsx`
- Modify: `app/src/atomic/organisms/BudgetsPanel.tsx`
- Modify: `docs/design/screens/budgets/README.md`
- Modify: `docs/design/screens/transactions/README.md`
- Create: `app/scripts/account-budgets-navigation.test.mjs`

**Interfaces:**
- `ScreenHeader({title, action?, subtitle?})` uses `Text` and `BaseButton` only and becomes the shared header for Transactions and Budgets.
- `tabFromPath('/account/budgets')` returns `account`; both `/budgets` and `/account/budgets` render `BudgetsPanel` during compatibility.

- [ ] **Step 1: Write failing navigation/style tests** asserting no Budgets item in bottom navigation, Account includes Budgets, both routes render the same panel, and Transactions uses `ScreenHeader`.
- [ ] **Step 2: Run the test to verify RED.**
- [ ] **Step 3: Remove Budgets from bottom navigation and add an Account menu row.** Keep the five-column/FAB grid balanced and preserve direct `/budgets` compatibility.
- [ ] **Step 4: Add `ScreenHeader` and refactor Transactions/Budgets to use it.** Replace screen-local header button/card treatment with existing base variants; keep transaction data behavior unchanged.
- [ ] **Step 5: Update route handling and screen specs.** `/account/budgets` is canonical; `/budgets` remains a compatibility render until the next route cleanup.
- [ ] **Step 6: Run navigation tests, `npm run check:design`, `npm run test:design`, and `npm run typecheck`.**
- [ ] **Step 7: Commit:** `refactor: move budgets under account navigation`.

### Task 4: Add private screenshot-backed feedback API

**Files:**
- Modify: `backend/internal/entity/feedback.go`
- Modify: `backend/internal/repository/feedback.go`
- Modify: `backend/internal/infrastructure/repository/feedback_postgres.go`
- Modify: `backend/internal/usecase/feedback.go`
- Modify: `backend/internal/controller/http/feedback_handler.go`
- Modify: `backend/cmd/api/main.go`
- Create: `backend/migrations/000003_feedback_screenshot.up.sql`
- Modify: `backend/docs/docs.go`
- Test: `backend/internal/usecase/feedback_test.go`
- Create: `backend/internal/controller/http/feedback_screenshot_test.go`
- Modify: `docs/architecture/FEEDBACK_API.md`

**Interfaces:**
- `FeedbackInput` gains optional `Screenshot io.Reader`, `ScreenshotMIME string`, and `ScreenshotSize int64`; accepted MIME is `image/png` and size is at most 1 MiB.
- `FeedbackStorage` exposes `Key`, `Put`, `Delete`, and `SignedGet`; the existing S3 implementation satisfies it.
- `GET /api/v1/feedback/{id}/screenshot` returns an owner-scoped short-lived redirect/URL envelope; `GET /api/v1/agent/feedback/{id}/screenshot` is protected by the existing agent middleware.
- Feedback JSON exposes `screenshot_available` but never the object key or signed URL in list responses.

- [ ] **Step 1: Write failing repository/usecase/HTTP tests** for optional screenshot creation, 1 MiB/MIME rejection, storage rollback, owner-scoped access, agent access, and no screenshot behavior.
- [ ] **Step 2: Run focused feedback tests to verify RED.**
- [ ] **Step 3: Add the migration/entity metadata and storage-aware service path.** Call the existing S3 key builder with `batch=feedback` and `attachment=<feedback-id>` so the key is `.../owners/<owner>/batches/feedback/<feedback-id>/screen.png`; delete the object if the database insert fails.
- [ ] **Step 4: Change Create handler to accept both existing JSON and multipart form.** Multipart fields remain `type`, `title`, `description`, and optional `screenshot`; JSON remains backward-compatible without a screenshot.
- [ ] **Step 5: Add owner and agent screenshot endpoints with short-lived signed URLs.** Audit only feedback ID, owner/actor kind, request ID, and action.
- [ ] **Step 6: Regenerate Swagger and run focused Go tests plus `go test -race ./...`.**
- [ ] **Step 7: Commit:** `feat: add private feedback screenshots`.

### Task 5: Add the floating feedback chat and DOM capture

**Files:**
- Modify: `app/package.json`
- Modify: `app/src/atomic/pages/FinancePrototypePage.tsx`
- Create: `app/src/atomic/organisms/FloatingFeedback.tsx`
- Modify: `app/src/atomic/templates/MobileAppShell.tsx`
- Modify: `app/src/services/feedback.ts`
- Create: `app/src/atomic/organisms/FloatingFeedback.test.tsx`
- Modify: `app/src/atomic/organisms/FeedbackPanel.tsx`
- Create: `app/scripts/feedback-screenshot-browser.test.mjs`
- Modify: `docs/design/screens/current-ui.md`

**Interfaces:**
- `createFeedback` accepts `{type,title,description,screenshot?: Blob}` and uses JSON when no screenshot is present, multipart otherwise.
- `FloatingFeedback` receives `captureRoot: HTMLElement | null` and calls `toPng(captureRoot, {cacheBust:true,pixelRatio:1})`; capture failure falls back to text-only submission.
- The modal uses `BaseModal`, `BaseTextInput`, `BaseTextArea`, `BaseSelect`, `BaseButton`, `StatusMessage`, and `SurfaceCard`; no raw control styles are introduced.

- [ ] **Step 1: Add `html-to-image` and write failing UI/browser tests** for floating-button visibility, modal submit, text-only fallback, screenshot removal, and multipart request shape.
- [ ] **Step 2: Run the tests to verify RED.**
- [ ] **Step 3: Implement the service multipart branch and `FloatingFeedback`.** Capture only the `.mypocket-app-root` content wrapper; show “Ảnh màn hình hiện tại sẽ được đính kèm” and a remove control before send.
- [ ] **Step 4: Mount the floating control in `MobileAppShell` for authenticated screens and add a stable capture-root selector.** Do not mount it on the auth gate.
- [ ] **Step 5: Keep Account → Phản hồi as the full history/changelog page and make the floating composer refresh that list after a successful send.
- [ ] **Step 6: Run the browser test against the real Vite + Go API, then run design checks and typecheck.**
- [ ] **Step 7: Commit:** `feat: add floating in-app feedback composer`.

### Task 6: Reconcile docs, release notes, and full verification

**Files:**
- Modify: `docs/architecture/API.md`
- Modify: `docs/architecture/OCR_API.md`
- Modify: `docs/architecture/FEEDBACK_API.md`
- Modify: `docs/design/screens/assistant/README.md`
- Modify: `docs/design/screens/budgets/README.md`
- Modify: `docs/design/screens/transactions/README.md`
- Modify: `docs/standards/DEBUGGING.md`
- Modify: `docs/releases/CHANGELOG.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `backend/docs/docs.go`

- [ ] **Step 1: Reconcile every changed endpoint, limit, route, screenshot retention rule, and user-facing copy with the shipped code.**
- [ ] **Step 2: Run backend verification.**

  Run: `rtk env TEST_DATABASE_URL='postgres://dev:password@127.0.0.1:5432/postgres?sslmode=disable' go test -race ./... -count=1`

- [ ] **Step 3: Run frontend verification.**

  Run: `rtk npm run check:design && rtk npm run test:design && rtk npm run typecheck && rtk npm run build`

- [ ] **Step 4: Run integrated AI, feedback, and navigation browser tests against fresh local servers; assert database state and visible UI state, not only HTTP status.**
- [ ] **Step 5: Run `rtk git diff --check`, inspect migrations in order, and verify no secret/local-env file is tracked.**
- [ ] **Step 6: Commit docs and verification:** `docs: document ai attachments and feedback capture`.

## Self-review

- Spec coverage: OCR diagnostics (Task 1), 20-file/drop flow (Task 2), Account/Budgets and styling (Task 3), private feedback API (Task 4), floating composer (Task 5), docs/full verification (Task 6).
- Placeholder scan: no TBD/TODO or unspecified implementation steps remain; each step names files, interfaces, commands, and expected behavior.
- Type consistency: the screenshot-aware `FeedbackInput` and `FeedbackStorage` interfaces are introduced in Task 4 before the frontend multipart client in Task 5; `ScreenHeader` is introduced before the screen refactors in Task 3.
- Review focus coverage: each failure mode listed above has a named test in its owning task.
