# Docs-to-router drift checklist

Date: 2026-09-11

Status: automated local contract pass; production URL smoke pending.

The checklist is generated and enforced per route descriptor by `backend/internal/platform/httpapi/openapi.go` and `openapi_test.go`. Every registered `/api/v1` method/path must expose all of the following in `/api/v1/openapi.json`:

- stable `operationId` and summary;
- named success-response schema and named request schema for mutations;
- Docusaurus `externalDocs` URL mapped to the owning domain page;
- matching `x-curl-example` containing the exact versioned route;
- cookie/Bearer security alternatives and CSRF requirement where applicable;
- stable `400`, `401`, `404`, `409`, `429`, and `503` error envelopes;
- `Retry-After` for rate limits and `X-Correlation-ID` for successful responses;
- required `Idempotency-Key` parameter for retryable logical mutations;
- path parameters for every `{id}`-style segment.

`TestOpenAPICoversEveryVersionedRoute` fails if a router descriptor lacks the operation/security/idempotency contract. `TestOpenAPIOperationsUseNamedRequestAndResponseSchemas` fails if any descriptor lacks named schemas, a public guide, or an exact curl example. This makes the OpenAPI document itself the exhaustive per-operation checklist instead of duplicating a manually drifting table.

The Docusaurus domain pages provide exact payload examples, version semantics and workflow constraints. A client must treat an OpenAPI/page mismatch as drift and stop the affected mutation rather than guessing a field.

Current intentional boundary: no bank/webhook route and no generic OCR proxy are registered or documented. OCR Platform is called server-side only for an owned receipt attached to an Agent run.

## Stale-phrase audit

The required search for `OCR deferred`, `bank webhook`, and `implementation pending` has no unexplained live-contract match. Remaining matches are either:

- dated, immutable historical design/verification/research artifacts describing the earlier plan or failure at that time;
- the current completion spec/requirements explicitly stating that bank integration is out of scope;
- the plan line containing the audit command itself.

Current architecture, requirements, public Docusaurus pages, OpenAPI and the reusable API skill all use the 2026-09-11 boundary.
