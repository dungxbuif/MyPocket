---
artifact_type: detail_design
status: approved_for_local_implementation
owner: ai
scope: wallet behavior model cleanup
release_state: local-only
---

# Wallet Behavior Model Design — 2026-09-12

## Goal

Replace pseudo wallet types with a behavior model that is honest in code, API and docs:

- `basic` — a normal money container. This covers previous cash, bank, e-wallet and generic debt-labelled wallets.
- `goal` — a savings/goal container with optional target and deadline metadata.
- `credit` — a credit-card behavior with optional credit limit, statement day and payment due day metadata.

The public API keeps the JSON field name `type` for compatibility, but the allowed values become behavior values only. Debt/loan remains a planning obligation/category concept, not a wallet type.

## Existing state

The current system stores and exposes `cash`, `bank`, `credit`, `e_wallet`, `savings` and `debt`. Only `credit` has behavior. The other values are labels, so the app lets users pick types that do not change finance behavior and are easily confused with budgets, goals and obligations.

## Data migration

Add a migration after the current local migrations:

| Old `wallets.type` | New `wallets.type` |
| --- | --- |
| `cash` | `basic` |
| `bank` | `basic` |
| `e_wallet` | `basic` |
| `debt` | `basic` |
| `savings` | `goal` |
| `credit` | `credit` |

Add optional columns:

- `goal_target_vnd bigint NULL CHECK goal_target_vnd IS NULL OR goal_target_vnd > 0`
- `goal_deadline_on date NULL`

Constraints:

- `type IN ('basic', 'goal', 'credit')`
- Credit metadata (`credit_limit_vnd`, `statement_day`, `payment_due_day`) is allowed only for `credit`.
- Goal metadata (`goal_target_vnd`, `goal_deadline_on`) is allowed only for `goal`.
- Existing balances and transactions are preserved.

## Backend contract

`finance.WalletType` exposes only:

- `WalletBasic`
- `WalletGoal`
- `WalletCredit`

`Wallet`, create request/response and sync/offline payloads include goal metadata. Create validates:

- Name is required.
- Type must be `basic`, `goal` or `credit`.
- Credit fields are rejected unless type is `credit`.
- Goal fields are rejected unless type is `goal`.
- Credit limit cannot be negative.
- Goal target must be positive.
- Statement and due day must be 1..31 when present.

Update wallet remains limited to name/include-in-total in this slice; editing goal or credit metadata can be a later explicit form because current UI only edits those existing fields.

## Frontend behavior

Wallet create UI offers exactly:

- `Cơ bản`
- `Mục tiêu`
- `Tín dụng`

When the selected type is `goal`, the form shows optional target amount and deadline inputs. When `credit`, it may send optional credit metadata only when the UI exposes those fields; this slice does not need to invent a credit-card management form.

Icons/labels use the behavior values. Tests and fixtures must stop creating `cash`, `bank`, `e_wallet`, `savings` or `debt` wallets.

## Documentation

Update internal and public docs to show only the behavior enum. Public wallet docs, ERD, API architecture, release notes and the `mypocket-api` skill must no longer present fake wallet types as live contract.

Historical planning docs may keep old references only as history, not as current API truth.

## Verification

- RED/GREEN backend validation test for accepted/rejected wallet behavior types and metadata rules.
- RED/GREEN migration test proving old data maps into `basic/goal/credit`.
- RED/GREEN frontend test proving wallet create offers the three behavior options and sends goal metadata.
- Full backend integration suite.
- Full frontend unit suite and production build.
- Business Playwright E2E suite for wallet CRUD, transfers, budgets, recurring and Agent remains green with `basic` wallet fixtures.
- Docusaurus build.
