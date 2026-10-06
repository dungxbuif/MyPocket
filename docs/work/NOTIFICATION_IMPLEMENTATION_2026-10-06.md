# Notification slice — implementation note

## Intended behavior

The goal-wallet notification switch is a real persisted account preference,
not a disabled placeholder. The first delivery is an owner-scoped in-app
inbox. It deliberately does not claim APNs or scheduled budget fan-out yet.

## API/schema contract

- `wallets.goal_notifications_enabled` is a boolean, default `false`.
- `PATCH /api/v1/wallets/:id` accepts that field only for goal wallets and
  keeps owner scoping from the existing wallet repository.
- `GET /api/v1/notifications` returns newest-first in-app items for the current
  owner.
- `POST /api/v1/notifications/:id/read` is idempotent and owner-scoped.
- The HTTP handler forwards `c.Request.Context()` into repository operations so
  cancellation, deadlines and request tracing are preserved.
- Notification timestamps are rendered in the authenticated account timezone,
  with a safe device-timezone fallback for sessions that omit the value.
- Notification generation is a separate worker concern and must use a dedupe
  key; ordinary transaction writes do not synchronously fan out notifications.

## Verification plan

- apply migration on a disposable PostgreSQL database;
- test owner isolation and idempotent read behavior;
- test wallet preference roundtrip for goal/non-goal wallets;
- render the preference with shared/base controls on web and native;
- run backend tests, web design/type/build checks, and iOS simulator tests.

## Explicit non-goals

Push/APNs registration, scheduled budget evaluation, and notification fan-out
are not marked complete by this slice.
