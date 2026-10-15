# Native account action contract coverage

## Goal

Close the native port audit gap for the owner-scoped notification
mark-as-read and API-key create/revoke actions. The tests must verify the
actual `FinanceAPI` requests without calling production or a local database.

## Contract

- `POST /notifications/{id}/read` uses the authenticated session and no
  request body.
- `POST /api-keys` sends the key name, scopes and optional expiry as the
  existing JSON contract, with an idempotency key.
- `DELETE /api-keys/{id}` uses the authenticated session and accepts the
  backend's 204 empty response.

## Verification

Implemented with an in-memory `HTTPTransport` capture in `APIClientTests`.
Paths, methods, auth header, JSON fields, idempotency behavior and 204
handling are asserted. The complete native unit suite, simulator build,
project/base-component guards and `git diff --check` passed; exact result
paths are recorded in `docs/work/VALIDATION_MATRIX.md` and the port audit.

## Live fixture extension

The owner-scoped Account surface also has a disposable local-API UI flow for
API-key create/revoke. It uses the DEBUG fixture session, a unique key name,
and teardown deletion by that exact name; production credentials and database
state are never used.

## Stitch/web profile action parity

The Stitch account reference and the web `AccountPanel` both expose a clear
profile-to-account-management affordance. The native More profile surface must
therefore render a shared `MPMenuRow` labelled `Quản lý tài khoản` and route it
to the existing owner-scoped `AccountView`. This is a navigation/UI contract;
it does not add a new API or data field. The UI proof must assert the row is
present and opens the account screen from the More profile card.

The profile card's `Trợ giúp` affordance is also actionable: it routes to the
existing feedback/changelog surface rather than silently discarding a tap. It
uses the existing `MPButton` base, preserves the current feedback API contract,
and is verified by a simulator UI test that opens `Phản hồi & changelog` from
the profile card.

The overview header notification icon must route to the existing
`NotificationsView` destination. The search icon remains presentation-only
until the web/API exposes a search contract; native must not invent a new
query surface for it.
