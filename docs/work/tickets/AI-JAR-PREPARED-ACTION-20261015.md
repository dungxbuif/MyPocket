# Finance Assistant jar prepared action

Status: implemented locally for create and monthly update; production rollout
and device UAT remain separate.

## Goal

Allow the Finance Assistant to prepare a new spending-jar proposal without
letting the model mutate financial data. The native user reviews the jar name,
month and allocation rule, then explicitly confirms through the owner-scoped
backend action dispatcher.

## Contract

- Read-only advisor turns may call `propose_create_jar`.
- The tool returns `jar_creation_proposal` with a server-generated `action_id`.
- Persisted proposal context includes owner-bound conversation ID, generation and
  run ID, like wallet proposals.
- Confirmation accepts only `{conversation_id,generation}` and the path action
  ID. The backend reads the persisted proposal as source of truth, validates
  the jar contract and executes the existing owner-scoped jar creation contract
  with the action ID as the idempotency key.
- Stale generation, missing action, altered client fields and cross-owner data
  fail closed. No transaction, transfer, payment, SQL or external URL access is
  added.
- The existing `transaction_mutations` table is reused with operation
  `jar.create`; the ordinary `POST /api/v1/jars` handler accepts the same
  idempotency key, and advisor confirmation derives that key from `action_id`.

## Monthly update

- Read-only advisor turns may call `propose_update_jar` with a UUID `jar_id`,
  month, name and mode-specific allocation.
- The proposal is rendered as the same confirmation card vocabulary with
  update-specific copy. Confirmation delegates to the owner-scoped
  `PUT /api/v1/jars/{jar_id}/months/{month}` handler, so it replaces the
  selected month configuration and cannot create a second jar.
- Invalid IDs/periods, stale generations, cross-owner proposals and malformed
  client fields fail closed. No migration or provider contract changed.

## UI and verification

The maintained native structured-card registry exposes `jar_confirmation`; the
card shows name/month/allocation and has one explicit confirm action. Existing
read-only `jar_progress` is unchanged. The archived web surface is not part of
this native slice. Backend tool/schema, persistence, owner-isolation,
stale/tamper/replay HTTP tests are green and the native simulator build
compiles the card/API contract. The fresh unit-test runner was blocked by
CoreSimulator diagnostics; no migration or environment rollout was performed.
