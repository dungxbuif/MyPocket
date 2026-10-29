---
artifact_type: bug
id: AI-ADVISOR-JAR-LINKED-TRANSACTION-20261020
status: regression_verified
owner: backend
---

# AI confirmation must preserve jar-linked transactions

## Scope

Fix the two confirmed backend defects in AI action confirmation without changing
the public API, database schema, or deployment configuration:

1. `propose_update_transaction` currently omits `jar_id`. The confirmation
   dispatcher sends a complete replacement payload to the ordinary transaction
   handler, so a transaction that already has a jar is saved with a null jar.
2. `propose_recategorize_transaction` preserves the existing `jar_id`, but its
   dispatcher builds a transaction handler without the `Users` and `Jars`
   dependencies required to validate that preserved assignment. Valid
   jar-linked rows therefore fail confirmation.

## Root cause evidence

- `backend/internal/usecase/ai_advisor_tools.go` emits no `jar_id` field for
  `propose_update_transaction`.
- `backend/internal/controller/http/transaction_handler.go` treats the
  replacement `jar_id` field as authoritative and writes it during update.
- `backend/internal/controller/http/ai_advisor_handler.go` preserves `jar_id`
  for recategorization but constructs `NewTransactionHandler` without the
  jar/user repositories.
- `validateTransactionInput` rejects a non-null jar when those dependencies are
  absent.

## Decision

The confirmation boundary remains server-owned. For transaction updates, the
dispatcher loads the owner-scoped existing row and copies its `jar_id` into the
replacement request before invoking the canonical handler. This keeps the
current proposal schema (AI cannot silently change jar membership) and prevents
data loss. For recategorization, the dispatcher passes the already configured
owner/user/jar dependencies into the canonical handler; existing version and
owner checks remain authoritative.

No new endpoint, field, migration, or alternate mutation path is introduced.

## Regression plan

- Red: confirm an update of a jar-linked ordinary expense and assert the
  persisted update input retains the same `jar_id`.
- Red: confirm recategorization of a jar-linked ordinary expense and assert it
  reaches the canonical handler with the jar dependencies and succeeds.
- Green: run the focused HTTP advisor tests, then `go test ./...` and `go vet
  ./...`.

## Reconciliation

The existing AI action contract and transaction/jar business rules remain the
source of truth. This note records only the bug fix and its regression proof;
no API/schema/production behavior outside these confirmation paths changes.

## Implementation and verification evidence

- `ConfirmAction` now loads the owner-scoped existing transaction for
  `update_transaction`, injects its server-owned `jar_id` into the replacement
  payload, and passes `Users`/`Jars` to the canonical transaction handler.
- `recategorize_transaction` now passes the same dependencies, so the existing
  jar assignment is validated rather than rejected as an unavailable feature.
- Red regressions failed before the fix with a cleared `jar_id` and HTTP 400;
  both pass after the fix.
- `go test ./internal/controller/http -count=1`: pass.
- `go test ./... -count=1`: pass.
- `go vet ./...`: pass.
