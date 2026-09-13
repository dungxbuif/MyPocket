---
artifact_type: test_verification
id: VERIFY-02-01
status: verified
owner: ai
trace:
  ticket: TICKET-02-01-ghi-thu-chi.md
  detail_design: TICKET-02-01-DETAIL_DESIGN.md
  ui_spec: ../../design/screens/transactions/README.md
  validation_matrix: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Verification — Ledger thu/chi cơ bản

Verified 2026-09-13 for the approved basic income/expense slice.

## Proven behavior

- Authenticated users can list, create, edit and delete owner-scoped income/expense transactions.
- Amount is a positive integer; direction applies through type. Wallet must belong to the owner and be `basic` or `goal`; credit waits for its own ledger semantics.
- Category is optional. A chosen category must be visible, match income/expense kind, and apply to the selected wallet when restricted.
- Note, local datetime converted to RFC3339 UTC, and report inclusion round-trip through the editor.
- The real list groups newest rows by local date, displays signed group totals and opens the same base-composed editor for changes.
- Global add, transaction edit and wallet changes refresh Header, Overview, Transactions and Wallet management through one refresh contract.

## Evidence

- TDD red/green: restricted category initially returned `201`; opening balance/type mutation and stale cross-panel refresh paths were reproduced and locked to the final immutable-ledger contract.
- `rtk go test ./...` in `backend/`: pass, 19 tests across 12 packages after this slice.
- Real HTTP UAT against dev PostgreSQL: restricted category/wrong wallet returned `400`; balance create/update/delete sequence was 115,000 → 113,000 → 93,000; permanent wallet deletion cascaded the remaining transaction.
- Chrome UAT: created a 25,000 expense from the global FAB, saw the real row and 75,000 balance, edited it to 35,000 and saw row/Header update to -35,000/65,000. Test wallet deletion cleaned the fixture ledger to zero.
- `npm run test:transactions`: pass (category applicability, signed amount and wallet total rules).
- `npm run test:design` and `npm run build`: pass; the design docs guard recognizes the new transaction screen spec.
- `go generate ./cmd/api`: pass; generated Swagger reflects `current_balance`.

## Acceptance boundary

The basic ledger slice is verified. Receipt attachments/OCR and jar assignment remain explicit later slices, so the parent BA ticket is not marked `done` and no human acceptance sign-off is invented.
