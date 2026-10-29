# Stage v1 migrations

Release 1.2.1: migration 000041 defaults deadline_at to database now + 240
seconds so old API instances can keep inserting during rolling update. This
default was added before 41's first production deployment. Fresh migrations
and an old-writer regression pass on disposable PostgreSQL. Financial tables
and values are untouched; never rerun a previously applied migration.

## Current candidate rollout

The current candidate is additive after the deployed version-34 baseline and
must be applied in order: `000035` → `000036` → `000037` → `000038` → `000039` →
`000040` → `000041`. Verify a validated
custom-format backup and `dirty=false` before applying the forward path. The
API image must be deployed with the matching worker only after version 041 is
clean. See [`docs/operations/ROLLOUT_READINESS.md`](../../docs/operations/ROLLOUT_READINESS.md)
and run the read-only repository preflight:

```sh
./scripts/mypocket-rollout-preflight.sh
```

## 000035 — recurring budgets (local candidate)

Adds owner-scoped series/revision tables, optional series/carry snapshot fields
and a unique generated-period index. Wallet/category deletion triggers stop
series only when the deleted scope is needed by the next ungenerated period
or a later revision. Historical superseded references do not stop unrelated
future configuration. Existing ledger/budget rows are not rewritten.

Apply after a validated production backup before starting the new API/worker;
production remains at version 34 until this rollout is performed. Fresh
disposable PostgreSQL 17 and real repository/HTTP race tests exercise version 35.
Do not edit migration 035 after production has applied it; later fixes require
a new forward migration. [ADR-023](../../docs/decisions/ADR-023-ledger-savings-recurring-budgets.md).

- `000032_feedback_owner_resolution.up.sql` adds independent owner confirmation
  (`resolved` false by default, nullable `resolved_at`); no existing feedback is
  automatically accepted and dev status/changelog/financial records are unchanged.
  Deployment requires validated backup and before/after content checksums.

Stage v1 starts with one baseline migration pair and uses additive forward migrations for post-baseline changes:

- `000001_stage_v1.up.sql` creates the complete schema and canonical seed data.
- `000001_stage_v1.down.sql` is intentionally non-destructive; user data must not be dropped by an automated rollback.
- `000002_ai_entry_error_code.up.sql` adds safe provider diagnostics to one-shot AI entry records.
- `000003_feedback_screenshot.up.sql` adds private screenshot metadata for owner-scoped feedback.
- `000020_ai_entry_reply.up.sql` persists the bounded model clarification returned with a one-shot AI entry result. The number follows the current deployed schema version so existing version-19 databases can apply it additively.
- `000021_ai_model_usage.up.sql` adds internal, non-public JSON usage metadata for AI entry and Finance Assistant model calls. It contains provider request/latency/token metadata only; prompts, OCR text and credentials are not stored.
- `000022_transaction_mutation_idempotency.up.sql` adds owner-scoped idempotency records for transfer mutations.
- `000023_transaction_adjustments.up.sql` adds immutable explicit balance-adjustment direction metadata.
- `000024_recurring_transactions.up.sql` adds recurring schedules and idempotent due-occurrence markers.
- `000025_credit_ledger.up.sql` adds credit entry kind/payment pairing metadata and indexes.
- `000026_travel_mode.up.sql` adds owner-scoped travel events, a one-active-event invariant, and nullable transaction links.
- `000027_portfolio_ledger.up.sql` adds owner-scoped immutable buy/sell history and optional manual prices for the portfolio ledger.
- `000028_transaction_category_required.up.sql` backfills ordinary income/expense rows with the canonical fallback groups and enforces a non-null category for ordinary and credit ledger rows; adjustment rows remain the explicit uncategorized exception.
- `000030_category_active.up.sql` adds a reversible active/archive flag to categories; existing categories remain active, and the migration does not alter transaction history.
- `000031_device_credentials.up.sql` adds the independent owner-scoped device sign-in table and index. It does not reset/change financial rows or existing passwords; deploy backend before enabling native remembered Face ID sign-in. Back up and verify the live migration version before applying 30/31. Never force or rerun a baseline on production.
- `000033_advisor_idle_reset.up.sql` adds nullable server-authoritative last-user activity metadata and a partial expiry-scan index. It backfills from user messages only and never deletes advisor or financial data; the application/worker must be deployed and verified separately before enabling cleanup.
- `000034_advisor_reset_requests.up.sql` adds an owner-scoped, content-free idempotency ledger for explicit Finance Assistant “Chat mới” resets. It does not touch wallets, transactions, reports or confirmed actions.
- `000036_push_tokens.up.sql` adds owner-scoped iOS push-token registration metadata. It does not call APNs, send notifications, alter financial rows or expose the token in API responses. Apply only after backup/version verification; provider entitlements and delivery remain a separate gate.
- `000037_audit_outbox.up.sql` adds a private redacted audit delivery buffer with idempotent stream/event keys and leased worker retries. It contains no prompts, OCR text, credentials or financial rows; apply only after backup/version checks and worker rollout readiness.
- `000038_portfolio_trade_corrections.up.sql` adds optimistic portfolio asset versions and immutable owner-scoped replace/void revisions. Existing trades and wallet balances remain unchanged; apply only after backup/version verification.
- `000039_push_deliveries.up.sql` adds the durable owner-scoped APNs delivery hand-off. It is deduped per notification/token/environment, stores only an app-owned route payload, and does not call APNs by itself; apply only after backup/version checks and matching worker rollout.
- `000040_row_versions.up.sql` adds positive, non-null optimistic-lock versions to ordinary transactions and personal categories. Existing rows start at version `1`; the migration does not rewrite financial values or category relationships. Apply only after backup/version checks and deploy the matching API before enabling clients that send `expected_version`.
- `000041_advisor_runtime.up.sql` adds the server-authoritative `advisor_runs.deadline_at` backfill/check and reaper index, then verifies duplicate assistant completions before installing an idempotent partial uniqueness guard. Existing advisor/financial rows are preserved; apply only after backup/version checks. Its down file is disposable-fixture-only.

The previous incremental migration files were squashed for the stage reset. Existing databases whose migration table is at version 19 must be recreated or explicitly reset to version 1 in a disposable environment before running this baseline. Do not force or drop a production database.

Local reset example (only for a disposable database):

```sh
dropdb --if-exists mypocket_stage_v1
createdb mypocket_stage_v1
DATABASE_URL='postgres://dev:password@127.0.0.1:5432/mypocket_stage_v1?sslmode=disable' MIGRATIONS_PATH=backend/migrations go run ./backend/cmd/migrate up
```

The application and tests continue to use `MIGRATIONS_PATH` when a non-default migration directory is needed.
