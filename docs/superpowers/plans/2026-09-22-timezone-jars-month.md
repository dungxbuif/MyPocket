# Implementation Plan: account timezone, jars, and automatic month summary

Design authority: [CORE-03-TIME-JARS-MONTH](../../work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md)
Approved scope: TICKET-01-01 timezone; TICKET-04 jars; TICKET-07-04 automatic monthly overview/note.

## Execution order

1. [x] Persist and expose validated account timezone; initialize new-account zone once; add editable Account setting and cache invalidation. Automated backend/API tests pass; owner settings UAT remains pending.
2. [x] Add shared account-zone calendar conversion and make transaction, AI review, and budget calendar fields use it; convert date-only DB/API fields while preserving labels. Migration and automated calendar/backend tests pass; owner boundary UAT remains pending.
3. [x] Implement jar identity/month snapshots, one-time prior-month copy, owner-safe APIs, ordinary expense assignment and calculated monthly/cumulative reports. PostgreSQL ownership/concurrency and summary tests pass; owner jar history/CRUD UAT remains pending.
4. [x] Integrate optional jar choice into manual and AI proposal review; add jar month/cumulative screen and Overview entry point. Selector fixture and build pass; owner interaction UAT remains pending.
5. [x] Implement monthly report boundaries/totals and independent owner/month note API; add automatic current/complete status. PostgreSQL note tests and calendar tests pass; owner timezone/boundary UAT remains pending.
6. [x] Add Overview month summary and month detail/note screen, using existing bases and real APIs. Design checks/build pass; owner visual and data UAT remains pending.
7. [ ] Reconcile API/ERD/business/report/screen docs, validation, release notes, backlog and context; perform completion audit against all three requirements.

## Interfaces

- `UserProfile` owns `timezone` and `timezone_confirmed`; `PATCH /api/v1/auth/profile` validates an IANA name. Initial browser timezone write is conditional and idempotent.
- Calendar-only API fields are `YYYY-MM-DD`; month selectors use `YYYY-MM`. Timestamp APIs remain offset-bearing RFC3339 instants.
- Transaction write payload adds optional `jar_id`; omission/null clears on explicit full-form save. AI draft editing carries optional `jar_id`, but extraction never auto-selects it.
- Jar report/config APIs are owner-scoped; month report response explicitly returns account timezone, date label, boundary instants, automatic completion state, derived figures and user note.
- Monthly completion is computed from current local month on every read; there is no manual close, immutable snapshot, or scheduled job.

## Constraints

- Preserve all existing owner changes in the dirty worktree.
- Keep existing accounts in `Asia/Ho_Chi_Minh`; do not silently adopt browser timezones for them.
- Keep UTC source timestamps immutable and date labels stable through migrations/timezone changes.
- Reuse design-system bases; no screen-local palette/control/card styling.
- Keep three functions in scope; do not pull in recurring budgets, transfers, debt/credit, AI narratives, or additional AI providers.

## Verification checklist

- [x] Review migration backfill, constraints and rollback semantics against current DB state; migration 15 is applied clean and data-destructive down paths refuse populated jar/note tables.
- [x] Check affected module contracts and code paths with focused unit/integration tests and local runtime health checks.
- [x] Run app design checks and build; all documented CORE-03 frontend suites pass.
- [ ] Exercise account-zone change, boundary transactions, manual/AI jar assignment, prior-month copy, cumulative report, note independence, and automatic completion without modifying owner data.
- [x] Update validation evidence; CORE-03 remains open until owner UAT and docs/release completion audit.
