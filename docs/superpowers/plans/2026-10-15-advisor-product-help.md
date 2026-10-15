# Advisor Product Help Retrieval Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a bounded, versioned product-help tool and source-identifiable card without granting the assistant financial or external-data access.

**Architecture:** An embedded handbook section catalog is searched deterministically by the advisor tool registry. The result travels through the existing `FinanceResult`/structured-part contract and is rendered by the existing web/iOS assistant card registries.

**Tech Stack:** Go, existing advisor tool loop, SwiftUI, React/TypeScript, existing base components.

**Spec:** `docs/work/tickets/NATIVE-ADVISOR-PRODUCT-HELP-20261015.md`

## Global Constraints

- No SQL, owner financial data, arbitrary URLs, writes, migrations, provider/configuration or deployment changes.
- Query is bounded to 240 bytes; result count is bounded to five.
- No-match is explicit and must not fabricate a navigation answer.
- Existing unknown-card fallback and owner/auth/audit boundaries remain intact.

## Review Focus

- Diacritic/case variants must match the same handbook section — Go search test.
- Empty/oversized/unknown JSON fields must fail closed — decoder/tool test.
- No-match must not leak all handbook text — no-match test.
- Malformed card values must not render unsafe source text — native/web contract tests.
- Old clients must continue to handle `meta`/unknown structured parts — existing fallback/build suite.

### Task 1: Handbook catalog and tool

**Files:**
- Create: `backend/internal/usecase/ai_advisor_help.go`
- Test: `backend/internal/usecase/ai_advisor_help_test.go`
- Modify: `backend/internal/usecase/ai_advisor_tools.go`
- Modify: `backend/internal/usecase/ai_advisor_tools_test.go`

- [x] Write failing tests for matching, no-match, bounds and registry exposure.
- [x] Run the focused Go tests and verify the expected missing-tool failure.
- [x] Implement the versioned section catalog, normalized token matcher and
  `get_product_help` tool with strict JSON decoding.
- [x] Run focused tests and then `go test ./...`.

### Task 2: Structured response mapping

**Files:**
- Modify: `backend/internal/usecase/ai_advisor_parts.go`
- Test: `backend/internal/usecase/ai_advisor_parts_test.go`
- Modify: `backend/internal/controller/http/ai_advisor_handler.go`
- Modify: `docs/architecture/AI_ADVISOR_TOOLS.md`

- [x] Add failing part-mapping test for `product_help`.
- [x] Map the tool result and advertise it in capabilities/docs.
- [x] Run Go tests; Swagger annotations did not change.

### Task 3: Native and web card

**Files:**
- Modify: `mobile-ios/MyPocketIOS/Features/Assistant/AssistantViews.swift`
- Modify: `mobile-ios/MyPocketIOSTests/DesignSystemTests.swift`
- Modify: `app/src/atomic/molecules/FinanceAssistantPart.tsx`
- Test: existing web design/type/build and targeted native tests.

- [x] Add failing type/title/source contract assertions.
- [x] Render bounded source rows through existing card/base components.
- [x] Run native build/tests, web build/design/type checks and diff validation.

### Task 4: Docs and reconciliation

**Files:**
- Modify: `docs/guides/finance-assistant-ios.md`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/work/tickets/NATIVE-RECONCILIATION-20261015.md`
- Modify: `docs/releases/CHANGELOG.md`

- [x] Record local evidence and remaining environment gates.
- [x] Confirm no production/deployment claim.
