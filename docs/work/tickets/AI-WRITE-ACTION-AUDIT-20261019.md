---
artifact_type: audit_note
id: AI-WRITE-ACTION-AUDIT-20261019
status: bounded_v1_verified
updated: 2026-10-19
---

# Finance Assistant write-action audit

## Scope

This audit reconciles the production feedback asking for broader Finance
Assistant actions with the registered tool catalog, confirmation dispatcher,
owner authorization and native confirmation-card contract. It intentionally
does not invent accounting semantics for transaction edits, portfolio trades,
payments, bank transfers or category rules.

## Verified bounded contract

The only supported model-facing write-adjacent tools are proposals:

- `propose_create_wallet`
- `propose_create_jar`
- `propose_update_jar`
- `propose_create_budget`
- `propose_update_budget`

Each proposal validates its schema and returns a persisted action card; it does
not mutate a financial row. `POST /api/v1/ai/advisor/actions/{action_id}/confirm`
accepts only the conversation ID and generation. The server resolves the
owner-bound proposal and delegates to the ordinary owner-scoped handler with
the proposal action ID as the idempotency key. Stale generations, foreign
owners, unknown actions and client-supplied financial fields fail closed.

The capabilities response advertises `write:false` and exactly the five
confirmable actions. This means that the model cannot write directly; it may
prepare a reviewable action for an explicit owner confirmation.

## Deliberately not implemented

`create/update/delete_transaction`, payment or bank-transfer actions,
category rules, and portfolio trade corrections remain out of the registry.
Portfolio correction requires an accounting decision (compensating trade,
replacement pair, or hard edit/delete) before an API can be safely added; see
`FEEDBACK-20261019-PORTFOLIO-CORRECTION-CONTRACT.md`. Release-aware external
RAG also remains disabled until an authenticated source and refresh owner are
chosen; see `FEEDBACK-20261019-PRODUCT-HELP-RAG-CONTRACT.md`.

## Verification plan

- Unit-test the exact registry and capability action set.
- Unit-test proposal schema/owner-field rejection and no-write behavior.
- Unit-test confirmation dispatch for every bounded action and stale/foreign
  action rejection.
- Run the backend race suite and reconcile this note with the shipped docs.

