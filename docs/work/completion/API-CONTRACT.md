# Public API Contract Evidence — 2026-09-11

Status: implementation and local verification complete; production URL smoke remains a release gate.

## Delivered

- Public unauthenticated OpenAPI 3.1 at `GET /api/v1/openapi.json`.
- Explicit route inventory for finance, planning, reports, portfolio, notification, sync, audit, import/export and account lifecycle.
- Machine checks for operation coverage, cookie+CSRF and Bearer security, path parameters, idempotency headers, stable error envelopes, correlation IDs, `429`, `Retry-After`, and retryable `503`.
- Redis fixed-window limiter at the authenticated Bearer boundary, default 120 requests/minute and bounded to 1–10,000.
- Fail-closed Bearer behavior when Redis cannot enforce a limit; cookie UI traffic remains independent.
- Repository skill `skills/mypocket-api` plus a public Docusaurus skill page and API quick reference.

## Authorization conclusions

- User identity always comes from the verified cookie or API key, never request payload.
- Invalid/revoked API keys return `401` and never fall back to a cookie in the same request.
- User-owned repository lookups include owner scope and externally hide foreign records as `404` where an object identifier is addressed.
- Cookie mutations require CSRF; Bearer mutations do not.
- Reset/delete are browser-only and require recent authentication, a signed preview, exact typed confirmation and idempotency.
- Audit reads additionally enforce `AUDIT_VIEWER_EMAIL`, independent of whether the authenticated principal used cookie or API key.

## Verification

| Gate | Result |
| --- | --- |
| OpenAPI coverage and authorization inventory tests | pass |
| Redis limiter enforcement, isolation and fail-closed tests | pass |
| Configuration default/boundary tests | pass |
| Go vet | pass |
| `mypocket-api` official skill validator | pass |
| Docusaurus production build | pass |
| Secret-pattern scan | no secret found |

The OpenAPI document and skill use `<user-api-key>` only. Neither contains a real API key, credential, private key, cookie, user payload, audit fingerprint or production database value.
