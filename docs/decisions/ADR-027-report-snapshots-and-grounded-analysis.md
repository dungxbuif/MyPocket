# ADR-027: Durable report snapshots and on-demand grounded analysis

- Status: accepted; local candidate implemented
- Date: 2026-11-01
- Scope: Reports & Money Insider backend

## Decision

Monthly reports are produced from a repeatable-read, owner-scoped accounting
query. The query excludes adjustments, internal transfers, archived/excluded
rows, and legacy system transfer categories. Surplus remains income minus
expense; it is not spendable balance or goal-wallet savings.

The backend maintains one owner-wide report revision. Transaction, wallet,
category/category-wallet, budget/series, jar/month, and account-timezone writes
bump that revision in the same database transaction. A report snapshot is
reusable only for the same owner, month, wallet scope, timezone, granularity,
rule version, and data revision. Current-month snapshots expire at the next
account-local midnight; retained snapshots are bounded by the 90-day policy.

Deep analysis is an explicit asynchronous request against an immutable report
snapshot. Ordinary report reads never call the provider. The worker reconstructs
at most 30 typed facts from the snapshot and sends no raw ledger, transaction
notes, images, or full transaction list. Provider output is limited to three
observations and three suggestions, is validated against fact references, and
falls back to deterministic server-generated insight copy without a repair call.
Provider failures leave the deterministic report available and are not
automatically retried.

Retention deletes analyses before snapshots in bounded batches and never
deletes ledger rows. Stale running jobs are fenced and marked failed rather
than silently requeued.

## Migration allocation

Migration `000043` and the existing dirty-workspace migrations `000044` through
`000046` belong to other work. This feature therefore uses `000047` for report
revisions/snapshots and `000048` for report analyses.

## Consequences

- API and worker wiring must deploy with migrations 000047 and 000048.
- Generated Swagger and native clients require a later reconciliation pass.
- Savings report semantics remain implemented by the existing savings reader;
  this read model does not replace or reinterpret that query.
