# ADR-019 — Financial SQL diagnostic redaction

Date:2026-10-14. Status: locally verified; rollout/review pending.
Work: [SQL privacy](../work/tickets/SQL-LOG-PRIVACY-20261014.md),
[offline prerequisite](../work/tickets/OFFLINE-CREATE-RECEIPTS-20261014.md).

## Decision

Shared NewPostgres logger emits only allowlisted statement class, elapsed time,
rows and sanitized SQLSTATE (or generic failure/known record-not-found). It never
forwards SQL text, bound parameters, raw driver error strings or generic logger
format arguments. LogMode preserves this wrapper. Application query results and
returned errors are unchanged; API/auth/transaction semantics are unaffected.

## Why

GORM default Info logs bound financial values. ParameterizedQueries/ParamsFilter
alone is insufficient: GORM1.25.12 Scan swaps in its recorder and expands SQL
before original Trace. Driver errors can independently echo rejected values.
Two real PostgreSQL tests reproduced both leaks; class-only Trace plus safe error
classification fixes them, with Row/event/not-found regression proof.

## Alternatives and costs

Silent logging would remove useful latency/row/error-class diagnostics. Regex
scrubbing SQL cannot safely distinguish arbitrary data/literals and is rejected.
Full query templates are intentionally unavailable; debug via source, request
trace and statement class rather than live financial payload logs. GORM generic
event message/arguments are also omitted, trading some startup detail for safety.
This is not an audit of arbitrary application logging or removal of historic
logs. No env escape hatch, dependency, schema migration or production rollout.
