# AI Jar Update Action Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Let Finance Assistant prepare an owner-scoped monthly jar update proposal that the user explicitly confirms, without exposing direct model writes or changing the existing jar API shape.

**Architecture:** Add a read-only `propose_update_jar` tool that emits the same structured confirmation envelope as jar creation plus `jar_id` and `action:update_jar`. The existing generation-bound advisor dispatcher resolves the persisted proposal and delegates to the existing owner-scoped `PUT /jars/{jar_id}/months/{month}` route; PUT replacement semantics make confirmation retries safe without creating rows. Native reuses the existing jar confirmation card and changes copy/validation based on the action.

**Tech Stack:** Go/Gin, PostgreSQL-backed jar repository, SwiftUI native card, XCTest, Swagger/Markdown docs.

**Spec:** `docs/work/tickets/NATIVE-FEEDBACK-NEW-20261015-DETAIL_DESIGN.md` (AI jar actions row and confirmation requirements).

## Global Constraints

- The model may prepare proposals but never receives SQL, credentials, owner IDs or direct write tools.
- Confirmation is owner/conversation/generation bound and derives financial fields from the persisted proposal.
- Existing `PUT /api/v1/jars/{jar_id}/months/{month}` validation and owner scope remain the source of truth.
- No database migration, provider change, production rollout or web-surface work.
- Native card must remain on shared `MPAssistantCard`, `MPButton`, `MPText` and existing confirmation state bases.

## Review Focus

- Invalid/foreign/missing `jar_id` must fail closed before confirmation; test schema and persisted-action lookup.
- `none`, `fixed` and `percent` allocation modes must preserve exact validation and clear incompatible fields; test tool and native decoder.
- A stale conversation generation must return conflict and must not call PUT; retain existing dispatcher tests.
- Repeating the same confirmed update must leave one deterministic month configuration; test request replay through PUT semantics.
- Legacy `jar_confirmation` create cards must continue to say/create, while update cards say/succeed as update; test both native action aliases.

### Task 1: Backend proposal and dispatcher

**Files:**
- Modify: `backend/internal/usecase/ai_advisor_tools.go`
- Modify: `backend/internal/usecase/ai_advisor_parts.go`
- Modify: `backend/internal/controller/http/ai_advisor_handler.go`
- Test: `backend/internal/usecase/ai_advisor_tools_test.go`
- Test: `backend/internal/usecase/ai_advisor_parts_test.go`
- Test: `backend/internal/controller/http/ai_advisor_handler_test.go`

- [x] Write failing schema/tool/part/confirm tests for `propose_update_jar`, `jar_update_proposal`, and `update_jar` dispatch.
- [x] Run the focused Go tests and record the expected missing-tool/action failure.
- [x] Implement bounded argument decoding (`jar_id`, `month`, `name`, allocation mode/fields), structured proposal mapping and dispatcher routing to the existing PUT handler.
- [x] Run focused tests, then `go test ./...` and `go vet ./...`.

### Task 2: Native confirmation card

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift`
- Test: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`
- Test: `mobile-ios/MyPocketIOSTests/APIClientTests.swift`

- [x] Add red decoder/card assertions for `action:update_jar`, valid `jar_id`, and update copy.
- [x] Extend the existing jar card decoder and confirmation copy without creating a new screen component.
- [x] Run focused XCTest and the full DesignSystem target.

### Task 3: Contract/docs reconciliation

**Files:**
- Modify: `backend/docs/docs.go`, `backend/docs/swagger.json`, `backend/docs/swagger.yaml` if generated output changes.
- Modify: `docs/architecture/AI_ADVISOR_TOOLS.md`
- Modify: `docs/guides/finance-assistant-ios.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`, `docs/work/BACKLOG.md`, `docs/CONTEXT.md`, `docs/releases/CHANGELOG.md`

- [x] Document tool input/result, confirmation action, owner/generation boundary and PUT retry semantics.
- [x] Record exact test/build evidence and the remaining production/device gates.
- [x] Run `git diff --check` and the relevant backend/native guards.

## Reconciliation — 2026-10-15

Implemented locally. Focused red tests failed before the tool/dispatcher/card
changes and passed after them. Backend `go test ./...`, `go vet ./...`, the
focused handbook race test, native DesignSystemTests 90/90, signed full native
XCTest 294/294, warnings-as-errors simulator build, base-component guards and
project verification all pass.
The result is not deployed and no migration, provider or device state changed.
Swagger generated files were not changed: this slice adds a capability/tool
payload and reuses the already documented confirmation route and jar PUT
route; the public Markdown catalog carries the new structured view contract.

## Verification Commands

```sh
rtk proxy go test ./... -count=1
rtk proxy go vet ./...
rtk proxy xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -destination 'platform=iOS Simulator,id=182D41A2-4FC6-445B-A592-70FDCB5C662E' \
  -only-testing:MyPocketIOSTests/DesignSystemTests -collect-test-diagnostics never \
  SWIFT_TREAT_WARNINGS_AS_ERRORS=YES CODE_SIGNING_ALLOWED=NO
rtk proxy git diff --check
```

No production or device claim is allowed from these local checks alone.
