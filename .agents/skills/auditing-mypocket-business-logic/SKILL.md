---
name: auditing-mypocket-business-logic
description: Use when auditing or changing MyPocket finance, wallets, transactions, transfers, reports, planning, sync, API-key integrations, authorization, Redis-backed identity, or any bug that could corrupt balances, ownership, idempotency, versions, dates, or aggregates.
---

# Auditing MyPocket Business Logic

## Core contract

Prove business invariants through public behavior and persisted state. Reading code is reconnaissance, not proof. For every confirmed defect: reproduce it with the smallest failing automated test, fix the invariant at the lowest shared boundary, and rerun both focused and neighboring workflows.

Read these files before acting:

1. `../../../AGENTS.md`
2. `../../../docs/CONTEXT.md`
3. `../../../docs/work/BACKLOG.md`
4. `../../../docs/standards/DEBUGGING.md`
5. `references/invariants.md`

## Audit workflow

1. **Define the invariant.** Write the expected ledger, ownership, version, idempotency, time, or aggregate relationship in one sentence.
2. **Trace one vertical path.** Request DTO → authorization → validation → repository transaction/locks → audit/change feed → response → frontend/offline caller.
3. **Create RED proof.** Add the smallest unit, repository, HTTP, or browser test that fails for the business reason. Do not weaken an assertion to match current behavior.
4. **Fix the shared boundary.** Prefer domain validation and repository atomicity over UI-only guards. Keep money as checked `int64` VND values.
5. **Run GREEN proof.** Run the focused test, the package suite, then the relevant API-key/browser flow.
6. **Check neighboring invariants.** Create/edit/archive/replay/conflict must preserve the same accounting effect. A transfer must be tested as one atomic two-wallet operation.
7. **Reconcile docs.** Update public API docs, `docs/work/VALIDATION_MATRIX.md`, an evidence record, `docs/CONTEXT.md`, backlog status, and changelog. Say `local-only` until deployment is verified.

## Required audit surfaces

| Surface | Required proof |
| --- | --- |
| Money arithmetic | Overflow/underflow rejection; exact reversal on edit/archive |
| Transaction lifecycle | Create, edit, archive, idempotent replay, stale `base_version` conflict |
| Transfer | Source debit + destination credit in one DB transaction; edit/archive reverse both |
| Reports | Transfers and excluded rows omitted where contracted; date/wallet filters and HCM boundaries |
| Wallet/category | User ownership, active state, category kind, optimistic update/archive and wallet activation |
| Planning | Optimistic writes; obligation repayment direction/limit; schedule payload parity and bounded catch-up; draft confirm/reject |
| Sync | Mutation receipt and domain change are at-most-once; conflict is stable, retry-safe and user-scoped |
| Third-party API | Same business behavior via `Authorization: Bearer mpk_...`; no CSRF for bearer; revoked key denied |
| Authorization/audit | Cross-user IDs never disclose or mutate data; successful sensitive mutations emit safe audit records |
| Redis | Cache failure falls back safely; revoke invalidates cached identity; Redis is not the source of truth |

## Test ladder

Run commands from the repository root. Use the project test database, never production.

```bash
cd backend
MYPOCKET_TEST_DATABASE_URL='postgres://mypocket:mypocket@127.0.0.1:55433/mypocket_verify_business_audit?sslmode=disable' go test -p 1 ./... -count=1
```

The integration helpers reset the shared test schema, so keep `-p 1` when packages use the same database. Parallel package execution against one database is invalid evidence because the test fixtures can drop each other's schema.

```bash
cd frontend
npm test -- --run
npm run build
npx playwright test e2e/business-core.spec.ts e2e/accounting-correctness.spec.ts --project=desktop --project=mobile --project=webkit-mobile --workers=1
```

Use narrower commands during RED/GREEN. Before a release claim, use the repository release verification command documented in `scripts/verify-beta.sh` and preserve its actual output in an evidence file.

## Defect classification

| Severity | Meaning |
| --- | --- |
| Critical | Cross-user access, double application, silent ledger corruption, lost committed mutation |
| High | Wrong balance/report, broken conflict/idempotency, invalid repayment or schedule creates data |
| Medium | Valid operation rejected, misleading empty state, timezone/category mismatch |
| Low | Non-destructive contract or feedback inconsistency |

## Common mistakes

- Treating a passing handler test as accounting proof while repository state is unchecked.
- Testing only create and missing edit/archive reversal.
- Accepting `int64` wraparound in balances, deltas, percentages, or aggregates.
- Calling two DB transactions “atomic” because both usually succeed.
- Testing cookie auth but not bearer API keys for third-party routes.
- Assuming Redis contains authoritative authorization state.
- Using device-local or UTC calendar dates for Vietnam business-day defaults.
- Claiming docs or code are deployed because a local build passed.

## Stop conditions

Stop and record a design decision when the intended accounting meaning is ambiguous, a schema migration is required, or a safe fix changes a public contract beyond approved scope. Otherwise, confirmed correctness defects stay in scope: fix them and attach proof.

## Output contract

Every audit record contains, in order:

1. Scope and sources of truth
2. Confirmed findings with severity and exact invariant
3. RED evidence
4. Fix and affected boundaries
5. GREEN commands/results
6. Remaining risks or decisions
7. Release state: local, staged, or production-verified

Use `references/evaluations.md` to verify that another agent can retrieve and apply this skill.
