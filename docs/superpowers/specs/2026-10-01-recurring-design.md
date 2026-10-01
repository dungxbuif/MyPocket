# Recurring transactions — implementation contract

## Scope

Deliver owner-scoped recurring transaction schedules for ordinary income/expense
rows. A schedule stores the next account-local due instant, frequency and
source fields; generated rows are ordinary transactions that users can edit or
delete independently.

## Decisions

- Frequencies are `daily`, `weekly`, `monthly`, and `yearly`, with a positive
  interval. Monthly/yearly schedules preserve their requested day when it
  exists and clamp 29–31 to the last day of the target month.
- `next_run_at` is stored as an RFC3339 instant calculated in the owner's IANA
  timezone. A run creates at most one row for each `(schedule_id, due_at)`.
- A due run writes the transaction and occurrence marker in one database
  transaction. Repeating the run after a timeout/restart is a no-op for an
  already marked occurrence. Catch-up is bounded to 100 occurrences per call.
- Pausing/deleting a schedule affects future occurrences only. Existing
  generated transactions have no cascade link that would make them disappear.
- Generated notes default to `Giao dịch định kỳ — {name}`; an explicit schedule
  note wins. Recurring rows never receive Travel Mode automatically.
- The first web slice exposes explicit `run-due` for a worker/cron to call;
  no request path silently runs unbounded work.

## Verification

Test schedule validation, owner isolation, date clamping, duplicate run
prevention, pause/delete behavior, note defaults, and ordinary transaction
effects. Run backend tests, frontend type/design checks and production build.
