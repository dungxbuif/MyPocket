---
artifact_type: adr
id: ADR-008
status: proposed
owner: shared
trace:
  requirement: ../../requirements/BUSINESS_RULES.md#time-thời-gian-và-phân-kỳ
  phase: null
  tickets_or_bugs:
    - ../work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md
    - ../work/tickets/TICKET-01-01-thiet-lap-ca-nhan.md
    - ../work/tickets/TICKET-04-hu-chi-tieu.md
    - ../work/tickets/TICKET-07-04-tong-ket-thang-ai.md
  detail_design: ../work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md
  master_docs:
    - ../architecture/API.md
    - ../architecture/ERD.md
    - ../../requirements/BUSINESS_RULES.md
    - ../../requirements/REPORTS.md
  release_notes: ../releases/CHANGELOG.md
---

# ADR-008 — Lưu UTC instant riêng với ngày lịch theo timezone tài khoản

## Status

Proposed for owner ADR review. The implementation scope is approved in the CORE-03 conversation; this ADR's final status remains owner-owned.

## Context

Transactions and audit events represent instants; budgets, wallet targets, jar configuration months, and month notes represent calendar labels. Treating both as UTC timestamps can move a selected date across days when viewed in another timezone. Server-local timezone also cannot provide consistent account reports. Account timezone changes must not rewrite source instants or user-entered date/month labels.

## Decision

- Persist instants as PostgreSQL `timestamptz` in UTC and return offset-bearing RFC3339 values.
- Persist calendar dates as SQL `date`/API `YYYY-MM-DD`; persist month keys as first-of-month SQL `date`/API `YYYY-MM`.
- Store a validated IANA timezone per account. Existing accounts use confirmed `Asia/Ho_Chi_Minh`; a new account may initialize from the browser timezone once and can later be edited.
- Derive month queries from local month start and next-month start, converted to a half-open UTC interval `[start, next_start)`.
- A timezone change rebuckets timestamp-based reports but never rewrites UTC instants, date-only labels, month-note keys, or saved monthly configuration.
- Month completion is derived from the current account-local month. Reports remain live and recalculable; no persisted/frozen close snapshot or scheduled close job is introduced by this decision.

## Alternatives considered

- Store every calendar date as UTC midnight: rejected because clients can display a different date after timezone conversion.
- Use server-local timezone: rejected because results depend on deployment location and cannot represent each account's calendar.
- Rewrite timestamps when account timezone changes: rejected because it changes the recorded instant instead of only changing its calendar interpretation.
- Freeze monthly totals at close: rejected for this slice because edits/backfills must recalculate the report and would require stale-snapshot and adjustment workflows.

## Consequences

- Positive: date labels stay stable, boundaries work across DST, and the same UTC ledger can be viewed in the account's current timezone.
- Negative: every calendar query must explicitly load the account timezone; form editors must distinguish wall time from UTC instant.
- Neutral: migrations `000013`–`000015` introduce account timezone/date-only columns, jars, and month notes. Automated tests cover conversion, report boundaries, note isolation, jar copy/idempotency and owner scope; browser UAT remains pending.

## Linked work

- [CORE-03 design](../work/tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md)
- [Validation matrix](../work/VALIDATION_MATRIX.md)
- [Changelog](../releases/CHANGELOG.md)
