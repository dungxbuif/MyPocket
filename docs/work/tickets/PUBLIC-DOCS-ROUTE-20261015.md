# Public docs route

Status: locally verified; rollout remains open.

## Decision

Keep the generated Swagger UI at `/api/v1/docs/index.html` and add a public
`/docs/` entry point that redirects there. The redirect is deliberately
unauthenticated and contains no API credentials or user data. Both `/docs` and
`/docs/` are accepted so copied links and browser normalization behave the
same way.

## Verification

- Unit-test both paths and require a `302` location of
  `/api/v1/docs/index.html`.
- Run the backend test/vet gates and diff check.
- This change does not publish the route or alter production data; deployment
  remains a separate release gate.

Evidence: `TestPublicDocsRedirectsToVersionedSwaggerUI` passes, and the full
backend `go test ./...` plus `go vet ./...` gate passes locally.
