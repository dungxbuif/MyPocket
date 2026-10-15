# SQL logging privacy prerequisite

Status: in_review (locally verified, not deployed). Owner's approved full-gap/offline scope requires diagnostics
without financial payloads. Parent: [offline receipt integration](OFFLINE-CREATE-RECEIPTS-20261014.md).

Cause: NewPostgres installs GORM default Info logger, whose ParamsFilter leaves
bound values in SQL traces. Database error text can also echo an invalid value,
so ParameterizedQueries alone does not satisfy privacy. API/worker/cleanup share
this constructor. Existing AI entry uses a silent session but other financial
queries do not. No production logs or user data need reading to reproduce.

Current design: wrap the logger at database boundary, retain level changes and
allowlisted SQL statement class/elapsed/row metadata, never full SQL text;
enforce ParamsFilter returning no bound values as defense-in-depth.
Trace errors expose only safe SQLSTATE (five uppercase ASCII letters/digits) or
generic failure, preserving known record-not-found semantics. Other GORM event
methods must not forward arbitrary formatted arguments. No financial data,
tokens, DSN, or raw server error details in logs. Returned application errors stay
unchanged, so error handling/HTTP status/auth/transactions are unaffected.
Raw interpolated SQL is not made safe by a parameter filter; app queries must
remain parameterized. No new dependency/env flag or DB migration.

TDD: controlled test logger writer captures real NewPostgres SELECT with synthetic
bound note/amount and a failing integer cast echoing a synthetic secret. Tests use
dedicated local TEST_DATABASE_URL and restore global logger, no t.Parallel/no DB
writes. RED must contain fake private values; GREEN retains SQL/rows and SQLSTATE
without either marker. Verify LogMode does not drop the wrapper. Then full Go
suite/race/vet, docs/runbook reconciliation. No production deployment or historic
log deletion; those require separate rollout/retention review.

First wrapper attempt fails both real tests. Root traced to GORM1.25.12 Scan:
it substitutes logger.Recorder, expands SQL without ParamsFilter, then sends that
expanded string to original Trace. Filtering normal callbacks cannot protect
this path. Revised design before attempt2: Trace discards full SQL text and emits
only allowlisted statement class (SELECT/INSERT/UPDATE/etc), elapsed/rows and
sanitized SQLSTATE. ParamsFilter remains defense-in-depth for normal callbacks.
No regex-based attempt to reconstruct private SQL. Cost: SQL text/template no
longer available in logs; use source/query context and request spans for debug.
This meets financial privacy across Scan as well as standard query callbacks.

Evidence: original constructor validRED2/2 (bound note/amount and PostgreSQL cast
error value). First filter-only wrapper remainedRED2/2; traced Scan recorder
source before attempt2. Final class-only traces pass4 actual DB tests: Scan result
unchanged/no values, failed cast returns original application error but logs only
SQLSTATE22P02 (LogMode retained), Row/generic Info/Warn/Error and unknown Trace
redacted, record-not-found sentinel retained without lookup identity.
`TEST_DATABASE_URL=<dedicated local55434/mypocket_feedback> go test -race ./... -count=1`
final6719 exit0 across all packages after test additions; earlier85614 also pass.
`go vet ./...` and `git diff --check` pass. Queries are read-only; no migrations,
data cleanup, API restart, prod rollout or historic log removal.

Docs reconciled: ADR019, deployment diagnostics, API logging status, parent offline
ticket/ledger, CONTEXT/BACKLOG/VALIDATION_MATRIX. No route/response/auth/schema or
client UI change, so Swagger/native screen docs unchanged. This verifies shared
SQL boundary, not every arbitrary application log call. Final branch review and
production activation remain pending; no release/feedback acceptance publication.
