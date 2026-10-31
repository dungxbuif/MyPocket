# APNs push delivery runbook

## Scope

The in-app notification inbox remains the source of truth. `push_deliveries`
is a durable, owner-scoped hand-off for optional APNs delivery; it never
changes a wallet, transaction, report, or notification row. Jobs are deduped by
`(notification_id, push_token_id, environment)` and are leased by a worker.

The provider boundary classifies responses as:

| APNs result | Queue action |
| --- | --- |
| 2xx | mark `sent` |
| 410 | mark `revoked` and revoke only that token registration |
| 429/5xx | `retry`, honor `Retry-After`, stop after five attempts |
| 400/403 | `failed`, operator-visible; do not retry invalid payload/credentials/topic |

An ambiguous transport failure is retryable, but the queue must keep the same
notification ID and collapse ID. Exactly-once banners are not promised.

## Configuration

The configuration loader accepts `APNS_ENABLED`, `APNS_TEAM_ID`,
`APNS_KEY_ID`, `APNS_TOPIC`, `APNS_PRIVATE_KEY`, and `APNS_ENVIRONMENT`
(`sandbox` or `production`). Keep the key in the deployment secret store;
never put it in a migration, API response, log, or client bundle. Disabled
APNs is valid and is the expected Personal Team/local mode.

The worker now leases `push_deliveries`, constructs a token-authenticated APNs
HTTP/2 request, and calls the configured sandbox or production endpoint.
`APNS_ENABLED` is an explicit activation switch: it remains false by default,
and enabling it validates all credentials at worker startup. A successful
worker/provider request still does not prove a signed device has received the
push; that remains a separate external acceptance gate.

Apply migration `000039_push_deliveries` only after a validated backup and the
matching API/worker image. The migration is additive and has a disposable down
file; it is not a production rollback mechanism.

## Release gates

The Go tests cover JWT signing/request headers through an `httptest` APNs
endpoint, Retry-After parsing, success, stale-token revocation, retry
scheduling, attempt bounds, and invalid credential/topic responses through the
worker's injected transport. Live APNs delivery still requires an Apple
capability/profile, provider key and registered device. A local notification or
a registered token is not evidence of remote delivery. See the Apple [capability
reference](https://developer.apple.com/help/account/reference/supported-capabilities-ios)
and [signing workflow](https://help.apple.com/xcode/mac/current/en.lproj/dev60b6fbbc7.html).
