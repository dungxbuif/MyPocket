# AI budget create action implementation plan

> Local-only plan. No production rollout, migration or device install.

## Goal

Allow Finance Assistant to prepare a validated budget proposal and let the
owner confirm it once. The model never receives a budget write tool; the
confirmation dispatcher resolves the persisted proposal and calls the existing
owner-scoped budget handler through a durable idempotency boundary.

## Contract

- Read-only tool: `propose_create_budget`.
- Required: `name`, positive `limit_amount`, `start_date`, `end_date`.
- Optional: `wallet_id`, `category_id`; IDs are UUIDs and owner-scoped at the
  existing handler.
- Structured view: `budget_creation_proposal` with action `create_budget`.
- Confirmation accepts only `conversation_id` and `generation`.
- Replays with the same action ID return the original budget; changed payloads
  conflict. No new migration: reuse `transaction_mutations` with
  `budget.create`.

## Verification-first tasks

1. Add red registry/schema/parts/dispatcher/idempotency tests.
2. Implement strict tool validation and structured mapping.
3. Add budget repository mutation interface/implementation and make POST
   budget use the shared idempotent path.
4. Add dispatcher routing and native confirmation-card decoding/copy using
   existing base components.
5. Update public/internal docs and run backend/native/full guard suites.

Out of scope: budget update/delete proposals, arbitrary transaction writes,
provider changes, web-surface work, deployment and database migration.

## Reconciliation — 2026-10-15

Implemented locally with red/green registry, structured-part, dispatcher and
replay tests. Budget POST now reuses the existing owner validation and records
confirmed creates in `transaction_mutations` under `budget.create`; native
uses the shared Assistant card/button/state bases. Backend focused tests and
native DesignSystemTests 90/90 pass. Full backend/native regression and
production/device/provider gates remain separate.
