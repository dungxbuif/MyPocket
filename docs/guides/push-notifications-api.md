# Native push-token registration

MyPocket currently materializes budget alerts in the authenticated in-app
inbox. Native iOS may register its APNs address so a future provider adapter
can deliver the same owner-scoped events.

## Register

```http
POST /api/v1/notifications/devices
Authorization: Bearer <JWT-or-full-owner-api-key>
Content-Type: application/json

{
  "device_id": "stable-installation-id",
  "device_name": "iPhone",
  "platform": "ios",
  "environment": "production",
  "token": "<APNs-device-token>"
}
```

The response contains `{data: {id, device_id, device_name, platform,
environment, created_at, updated_at}}`; the token is never returned. Registering
the same owner's device/environment again updates the token and clears a prior
revocation. A token already registered to another owner returns `409
PUSH_TOKEN_CONFLICT`.

## Revoke

```http
DELETE /api/v1/notifications/devices/{id}
Authorization: Bearer <JWT-or-full-owner-api-key>
```

Only the owning account can revoke a registration. A foreign or unknown id is
reported as `404 PUSH_TOKEN_NOT_FOUND`.

## Delivery boundary

These routes persist registration metadata only. They do not call APNs or send
push notifications. Provider credentials, Apple entitlements, environment
selection, retry/dead-letter handling, token invalidation from provider errors
and signed-device delivery acceptance are separate deployment gates.

AI-agent contract: use the exact routes and fields above; never infer that a
successful registration means delivery succeeded. See [OpenAPI](../architecture/API.md)
and [ADR-024](../decisions/ADR-024-APNS-TOKEN-LIFECYCLE.md).
