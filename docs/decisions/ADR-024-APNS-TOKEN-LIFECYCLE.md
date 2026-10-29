# ADR-024: owner-scoped APNs token lifecycle

Status: Accepted for local registration only (2026-10-19)

## Context

In-app notifications are persisted and read by owner, but native push delivery
needs a durable token registration. Registering tokens without owner checks could
send a future notification to the wrong account. Provider delivery is not yet
configured.

## Decision

Store APNs token registrations in a dedicated owner-linked table. Protect
registration and revocation with the existing JWT/full-owner API-key middleware.
Return metadata only; never return or log the token. Enforce global token
uniqueness by platform/environment and owner-scoped revocation. Re-registration
by the same owner/device is an upsert.

## Alternatives rejected

- Storing tokens only in iOS Keychain: the backend could not deliver after a
  process restart or to multiple devices.
- Hashing only the token: a future APNs adapter could not address the device.
- Adding APNs calls to the worker now: signing, provider credentials, retries and
  production entitlement behavior are not defined and would make delivery look
  complete when it is not.

## Consequences

The database contains a sensitive provider identifier, so it is excluded from
JSON and logs and requires normal database access controls. The API is usable by
native clients now, while delivery remains an explicit deployment/provider gate.
Future provider work must add encryption-at-rest or equivalent secret handling,
retry/dead-letter policy, token invalidation on provider response, and signed
device acceptance before enabling push.

Links: [detail design](../work/tickets/APNS-DEVICE-TOKENS-20261019-DETAIL_DESIGN.md), [API](../architecture/API.md), [ERD](../architecture/ERD.md).
