---
artifact_type: work_ticket
id: AI-ENTRY-IMAGE-CONTEXT-20261016
status: implemented_local
owner: shared
---

# AI entry multi-image context

## Problem

Overlapping bank-statement screenshots can produce wrong dates, wallets or
duplicate proposals. The current flow assigns random attachment IDs and sorts
the OCR inputs by those IDs, then concatenates provider text without a document
boundary. That removes the user's upload order and makes account/date context
ambiguous before the model sees it.

## Intended behavior

- Preserve the original attachment order from the request through OCR, model
  input and persisted attachment text.
- Give every OCR block an explicit, backend-owned document marker and filename
  so the model can reset date/account context at each image boundary.
- Keep OCR text untrusted and never let markers change the existing schema or
  approval boundary.
- Preserve the existing 20-file limit, owner scoping, idempotency and review-
  first behavior.

## Scope and decision

This is a context assembly fix, not a provider switch. It does not claim that
character recognition is perfect. The raw OCR provider still returns text only;
the model remains responsible for row interpretation and fuzzy overlap review.
Deterministic exact-draft merging remains unchanged in this slice.

## Verification plan

1. Add a failing unit test proving two images are sent to OCR/model in request
   order even when their generated attachment IDs sort differently.
2. Add a failing contract test proving the model source contains one marker per
   image with stable ordinal and filename, without embedding image bytes.
3. Run the focused AI client/usecase tests, then `go test ./...`, race and vet.
4. Re-run the supplied seven screenshot fixture with a configured OCR/model
   provider and compare each OCR block and reviewed proposal against the source
   rows; do not auto-approve any proposal. This live fixture gate remains open
   because the current local worktree has no OCR credential.

## Public/API impact

No route, request schema, response schema, database migration or credential
behavior changes. The internal model prompt receives additive source markers;
public OCR and AI-entry docs now describe the boundary.
