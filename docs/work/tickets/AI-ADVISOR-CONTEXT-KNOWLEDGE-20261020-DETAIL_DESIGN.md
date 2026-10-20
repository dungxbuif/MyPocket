---
artifact_type: detail_design
id: AI-ADVISOR-CONTEXT-KNOWLEDGE-20261020
status: approved_for_local_implementation
owner: ai
updated: 2026-10-20
links:
  architecture: ../../architecture/AI_ADVISOR_TOOLS.md
  backlog: ../BACKLOG.md
  validation: ../VALIDATION_MATRIX.md
---

# Finance Assistant knowledge/context improvement

## Goal

Make the Finance Assistant understand the current MyPocket product surface and
account-local calendar context more reliably, without adding a keyword router,
vector database, hidden user-memory store, or direct database access.

## Design

1. Keep the bounded, source-controlled handbook and `get_product_help` tool as
   the product-knowledge boundary. Expand it from the currently partial
   navigation/jar entries to every mounted native secondary destination and
   the supported proposal actions. Each entry states the route, purpose,
   read/write boundary and confirmation requirement.
2. Keep persisted conversation context bounded to the latest 12 visible text
   messages. Do not retain hidden summaries or inject old card payloads as
   facts. Financial numbers remain tool-grounded.
3. Add an account-local “today” line to the system context. The formatting
   helper accepts an explicit clock instant for deterministic tests and uses
   `time.Now()` only as the production default. Convert the instant through the
   validated account IANA timezone; if timezone is absent/invalid, omit the
   date rather than guessing from the server locale.
4. Preserve the existing 30-minute idle generation reset and all owner/auth
   boundaries. This change does not alter retention, routes, schemas,
   proposals, provider selection, or confirmation behavior.

## Verification

- Unit tests cover every supported secondary route/action name in handbook
  output, jar/budget/wallet/transaction proposal boundaries, and no direct
  writes.
- Clock tests cover `Asia/Ho_Chi_Minh`, a date crossing UTC, and invalid or
  missing timezone handling.
- Existing context limit/order and product-help ranking/no-match tests remain
  green.
- Run `go test ./internal/usecase -run 'Advisor|AIEval' -count=1` and then the
  full backend suite.

## Non-goals

- No vector/RAG infrastructure or network retrieval.
- No hidden memory beyond the existing persisted visible conversation window.
- No changes to AI provider/client, public API, database schema, or production
  rollout.
