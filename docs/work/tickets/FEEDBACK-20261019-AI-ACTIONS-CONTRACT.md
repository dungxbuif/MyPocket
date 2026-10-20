---
artifact_type: decision_contract
id: FEEDBACK-20261019-AI-ACTIONS-CONTRACT
status: bounded_v1
updated: 2026-10-19
---

# Finance Assistant action capability contract

Feedback `afee0f86` requests broader AI actions. The safe boundary for the
current release is **prepare, review, confirm**. A model tool never mutates
the ledger directly, even when the caller uses a full-owner API key.

## Capability matrix

| Capability | Provider tool | Mutation at tool time | Confirmation |
| --- | --- | ---: | --- |
| Read transactions/reports/wallets/budgets/goals/jars | `get_*`, `search_transactions`, `compare_spending_periods` | No | No |
| Product navigation/help | `get_product_help` | No | No |
| Create wallet | `propose_create_wallet` | No | Required |
| Create/update jar | `propose_create_jar`, `propose_update_jar` | No | Required |
| Create/update budget | `propose_create_budget`, `propose_update_budget` | No | Required |
| Create/edit/delete ordinary transaction | `propose_create_transaction`, `propose_update_transaction`, `propose_delete_transaction` | No | Required |
| Internal wallet transfer | `propose_transfer_wallet` | No | Required; atomic paired rows |
| Create/update category and recategorize ordinary transaction | `propose_create_category`, `propose_update_category`, `propose_recategorize_transaction` | No | Required |
| Credit payment, bank/external transfer, category rule | Not registered | N/A | Requires a separate domain contract |

The registered action set is intentionally explicit. Unknown tool names,
model-supplied owner IDs, malformed arguments, stale generations, foreign
owners, duplicate confirmations and confirmations without persisted action
context must fail closed.

## Adding an action

Before adding a tool, record: capability name, domain handler, input schema,
owner lookup, preview card, confirmation endpoint, idempotency identity,
replay behavior, cancellation behavior, audit event, and native/provider
evaluation cases. The tool must first return a persisted proposal. The
confirmation endpoint derives financial fields from that proposal and delegates
to the ordinary owner-scoped domain handler.

## Current proof

`ai_advisor_tools_test.go` asserts the registered set and rejects escape/write
tools. Proposal tests assert validation, owner override rejection, durable
action IDs and no ledger write. Provider tests assert proposal descriptions
retain the non-mutating boundary. The confirmation endpoint dispatches only
the persisted proposal view to the ordinary owner-scoped transaction,
transfer, or category handler; request fields and model owner IDs are ignored.
Credit/payment/external transfers remain outside this contract. `expected_version`
is now a persisted row-version contract for transaction and category
edits/deletes/recategorization. Migration `000040` initializes transaction and
category versions to `1`; the canonical repository performs an owner-scoped
compare-and-swap and increments the version only after a successful mutation.
Missing or stale versions are rejected with a conflict, including when the
proposal is confirmed through the advisor action endpoint. Create proposals do
not require an existing row version, but their persisted action/confirmation
identity still provides the review-before-write and replay boundary. Transfer
and credit guards plus owner/type validation remain enforced by canonical
handlers.
