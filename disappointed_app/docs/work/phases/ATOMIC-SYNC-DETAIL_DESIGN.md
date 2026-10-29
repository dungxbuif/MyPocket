# Atomic sync correction

Status: in_review (local automated proof; not deployed). Approval: owner requested “tiếp tục fixx” after reviewing ADR-008 and its blocker. This authorizes implementation of the shared transaction boundary.

Scope: finance and portfolio commands, draft-confirm accounting, sync receipts/change feed and consistent resync. Public request formats stay compatible. Integration testing exposed missing asset entity/operation CHECK values in the existing feed schema; additive migration 0012 admits the already supported API operations without rewriting historical migrations or data.

Root cause: domain repositories commit before the sync service appends its change and receipt; direct commands omit changes. Existing caller-owned draft confirmation demonstrates that finance accounting can run in one outer transaction.

Implementation: introduce an explicit transaction-bound database handle with nested savepoints. Finance/portfolio command entry points use one transaction, serialize per-user writes before domain locks, and append canonical entity changes before commit. Accounting emits affected wallet changes plus the transaction. Draft confirmation calls the same transaction-owned accounting path. Sync binds its repositories to an outer transaction, locks the user before receipt lookup, executes the command and stores the receipt before commit; audit runs afterward. Domain errors roll back to savepoints before deterministic rejection/conflict receipts are stored. Resync reads one repeatable-read snapshot; portfolio readers must close row streams before dependent reads on the same connection.

Verification: PostgreSQL failure injection at change/receipt insertion, concurrent identical/different mutation IDs, direct wallet/transaction/portfolio changes and draft confirmation, exact ledger rollback; full Go suite serially, race tests, existing browser flows. Reconcile ADR-008, architecture/API, public sync docs, repo skill, audit evidence, context, backlog, validation matrix and changelog. Release remains local until separately verified.

Links: [ADR-008](../../decisions/ADR-008-atomic-offline-sync-commit.md), [audit](../test-verification/BUSINESS-LOGIC-AUDIT-2026-09-10.md), [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md).

## Anti-patterns and verification

Do not commit a domain write before its sync feed/receipt, allocate cursors outside the command transaction, or recover a lost response by issuing a fresh direct create. A batch can have a committed prefix: replay original mutation IDs and payloads. Integration tests must set `MYPOCKET_TEST_DATABASE_URL` (not `DATABASE_URL`) and run packages serially because test helpers reset the shared schema.

Local evidence: injected receipt/feed failures and concurrent replay tests; 158 frontend tests; 24 core/offline/isolation browser scenarios plus 3 committed-response-loss retries; race checks include portfolio. See the audit record for exact scope and release limitations. Docusaurus build passes; untracked documentation files produce a non-fatal last-update-date warning.
