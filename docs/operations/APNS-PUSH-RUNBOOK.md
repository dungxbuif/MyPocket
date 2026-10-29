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

**Current implementation boundary (2026-10-20):** the worker currently runs
the in-app notification producer and audit/outbox passes only. It does not
lease `push_deliveries`, construct an APNs HTTP/2 request, or call a remote
provider. `APNS_ENABLED` is therefore a reserved configuration value for the
future adapter, not an activation switch. Do not set it in production and
infer that push is enabled; no delivery claim is valid until the provider
adapter, queue lease/revoke integration, and signed-device acceptance are
implemented and verified together. This explicit boundary prevents a false
green rollout when migration `000039` is present but no delivery worker is
running.

Apply migration `000039_push_deliveries` only after a validated backup and the
matching API/worker image. The migration is additive and has a disposable down
file; it is not a production rollback mechanism.

## Release gates

The Go contract tests cover success, stale-token revocation, retry scheduling,
attempt bounds, and invalid credential/topic responses using an injected mock
transport. They prove classification only; they do not prove that the worker
calls that transport. Real APNs delivery still requires an Apple
capability/profile and a provider key. A local notification or a registered
token is not evidence of remote delivery. See the Apple [capability
reference](https://developer.apple.com/help/account/reference/supported-capabilities-ios)
and [signing workflow](https://help.apple.com/xcode/mac/current/en.lproj/dev60b6fbbc7.html).
