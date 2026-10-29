---
artifact_type: decision_note
id: AI-USAGE-01
status: done
owner: ai
links:
  backlog: ../BACKLOG.md
  ai_entry: TICKET-09-01-ENTRY-DETAIL_DESIGN.md
  ai_entry_02: AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md
  api: ../../architecture/API.md
  integrations: ../../architecture/INTEGRATIONS.md
  validation: ../VALIDATION_MATRIX.md
---

# AI-USAGE-01 - No User AI/OCR Usage Limit For Now

## Scope

Record the current owner decision for the existing one-shot AI entry flow: **do not limit user usage at this time**. This covers text extraction and OCR-backed file extraction through `POST /api/v1/ai/entry/process`.

Out of scope:

- Agent chat/advisor history and tool calling.
- Public third-party API key scopes/OpenAPI contract overhaul.
- Recurring transactions, offline sync, or new worker runtime.
- Provider billing integration.

## Problem

The repository previously had a hardcoded `20` requests per rolling 24 hours inside `AIEntryPostgresRepository.BeginMessage`. The owner clarified that MyPocket should not enforce a user usage limit now.

## Decision

Remove the hardcoded per-user AI request limit and do not add a replacement quota field.

- `ai_entry_requests` remains the idempotency/audit table for one-shot processing.
- Idempotent replay still does not repeat provider work.
- No `user.ai_daily_limit` column, quota API fields, or rate-limit error are introduced.
- Cost consent, provider pricing, and future usage budgets stay in [AI-ENTRY-03](AI-ENTRY-03-PROVIDER-SELECTION-DETAIL_DESIGN.md).

## Touched Files

- `backend/internal/repository/ai_entry.go`
- `backend/internal/infrastructure/repository/ai_entry_postgres.go`
- `backend/internal/controller/http/ai_entry_handler.go`
- tests for repository and HTTP capability behavior
- docs: API, integrations, validation matrix, context, backlog/changelog

## Risks

- Provider usage can grow until a future provider/billing policy is approved.
- There is still one in-flight request guard per process/session; this is not a usage cap.
- No user-visible remaining quota can be shown because no quota exists.

## Verification Plan

- RED/GREEN repository test: prior requests do not block a new process.
- Regression test: idempotent replay still does not repeat provider work.
- Backend test command: `go test ./internal/infrastructure/repository ./internal/controller/http ./internal/usecase`.
- Docs reconciliation for API and runtime config surfaces.
