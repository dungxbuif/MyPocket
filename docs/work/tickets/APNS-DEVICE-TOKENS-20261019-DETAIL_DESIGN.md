---
artifact_type: detail_design
id: APNS-DEVICE-TOKENS-20261019
status: verified
owner: AI
---

# APNs device-token registration

## Scope

The notification inbox and budget producer already exist, but the native app has
no owner-scoped registration lifecycle for an APNs token. This slice adds only
the local registration/revocation contract. It does not contact APNs, enqueue a
provider request, or claim that push delivery works.

## Contract

Protected routes accept the same JWT or full-owner API-key middleware as other
account routes:

- `POST /api/v1/notifications/devices`
  - body: `{device_id, device_name, platform, environment, token}`
  - `platform` is currently `ios`; `environment` is `sandbox` or `production`.
  - response returns metadata and a server id, never the token.
- `DELETE /api/v1/notifications/devices/:id`
  - revokes only a row owned by the authenticated principal; a foreign/unknown
    id returns `404`.

The token is stored only for a future provider adapter and is never serialized
or logged. A `(platform, environment, token)` uniqueness constraint prevents a
token from being registered to two owners. Re-registering the same owner's
`device_id/platform/environment` updates the token metadata and clears revoke.

## Implementation

- additive migration `000036_push_device_tokens`;
- entity and owner-scoped PostgreSQL repository;
- usecase validation for bounded identifiers, platform/environment and token;
- authenticated notification handler routes;
- Swagger and human/agent API documentation;
- contract/integration tests for owner isolation, idempotent re-registration,
  token redaction, revocation and malformed input.

## Explicit external gate

APNs entitlements, signed-device permission, APNs key/certificate, environment
selection at deployment, provider retry/dead-letter policy, and actual delivery
remain external gates. No APNs provider call is made in this slice.

## Native client boundary

`FinanceAPI.registerPushToken(deviceID:deviceName:token:environment:)` sends
the token only after iOS has delivered a token to the app delegate. The native
client treats notification permission and server registration as separate
operations:

1. If permission is denied, keep the in-app inbox available and do not call
   the registration route.
2. If APNs returns a token, register it with a stable installation/device id;
   reuse one idempotency key while retrying the same request.
3. On logout or device removal, call `revokePushToken` and clear only the
   local registration id. A push registration failure must not log the user
   out or clear the authenticated session.
4. A `401`, `404`, or `409` is a registration-state result, not evidence that
   APNs delivery works. Keep the in-app inbox usable and expose a retry state.

The current app ships these API-client contracts and tests but deliberately
does not request permission, add an APNs entitlement, or claim remote delivery
until a supported signing profile and provider configuration are available.

## Verification

Focused usecase, handler and PostgreSQL repository tests pass under `go test
-race`, including migration `000036`, owner isolation, same-owner upsert,
foreign-token conflict, token redaction and revocation. Swagger was regenerated
and `git diff --check` passes. The parent integrated backend race suite still
owns the final combined evidence and shared validation/backlog reconciliation.

Links: [API architecture](../../architecture/API.md), [worker contract](../../architecture/WORKER.md), [remaining feedback map](FEEDBACK-20261019-REMAINING.md), [validation matrix](../VALIDATION_MATRIX.md).
