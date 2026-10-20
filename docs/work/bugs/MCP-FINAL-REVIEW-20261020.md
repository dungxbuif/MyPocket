---
artifact_type: bug
id: MCP-FINAL-REVIEW-20261020
status: verified
owner: shared
approval: owner authorized final review correction wave under MCP-API-PARITY-20261020
updated: 2026-10-20
---

# MCP final review correction wave

Parent: [approved design](../tickets/MCP-API-PARITY-20261020-DETAIL_DESIGN.md),
[plan](../../superpowers/plans/2026-10-20-mcp-api-parity.md),
[review](../../../.superpowers/sdd/2026-10-20-mcp-api-parity/final-review.md).

## Reproduction and root causes

The final review reproduces four Important defects on committed `725f0289`:
category replay claims depend on uncommitted handler/usecase/repository code;
canonical advisor middleware synchronously audits inner MCP calls; the feedback
schema omits its existing PNG multipart capability; feedback screenshot 302s
fall through as errors with signed redirect HTML.

## Intended correction

- Integrate and verify the authorized category receipt changes already present.
  Use the existing mutation ledger, atomic owner/operation/key/hash reservation,
  stored response replay, payload conflict and concurrent exactly-once insertion.
  No migration. Keep unrelated advisor/native changes outside this commit.
- Mark only server-created MCP dispatch requests with a private context key.
  When the outer MCP audit queue exists, suppress the duplicate advisor access
  audit on this inner request. Reauthentication and direct REST audit remain.
  A permanent application-router regression also reproduces synchronous
  feedback domain audit blocking the operation deadline. A composition-only
  wrapper copies the supplied feedback handler/service and queues those
  domain events for audited MCP dispatch; direct REST still calls its original
  sink synchronously. Each domain/operation queue has 64 slots and one worker;
  saturation drops excess events. No usecase redesign or domain-event removal.
- Publish optional `feedback.create.screenshot` as a closed
  `{filename,mime_type,base64}` object, PNG only, at most 1 MiB decoded.
  Encode canonical `type`, `title`, `description`, `screenshot` multipart fields;
  keep the existing JSON path when absent. Bound the complete multipart body
  to the canonical 2 MiB parser limit and preserve owner checks.
- Normalize `feedback.screenshot` 302 as a successful SDK ResourceLink with a
  relative authenticated REST fallback. Never return redirect HTML/signed URL
  in structured content, text content or audit.

## Verification and documentation plan

Add permanent failing SDK-to-canonical-parser and application-router tests,
then minimal fixes and focused race checks. Verify category replay/conflict,
owner separation, atomic rollback and concurrency against disposable
PostgreSQL 17 at clean schema 40; run MCP integration with Redis 7. Run full
Go tests/race/vet, formatting/diff checks and regenerate Swagger; remove owned
containers/volumes. Record exact counts in the final-fix report.

Reconcile [human/agent contract](../../ai/mcp-api-parity.md),
[API reference](../../architecture/API.md), [ADR-025](../../decisions/ADR-025-mcp-api-parity.md),
[validation](../VALIDATION_MATRIX.md), [context](../../CONTEXT.md),
[backlog](../BACKLOG.md), [release notes](../../releases/CHANGELOG.md), generated
OpenAPI and the dedicated category bug. Schema/ERD and native documentation
need no changes because storage schema/native behavior do not change.

Staging owner/client, provider, device and deployment acceptance remain open.

## Final local verification

All four Important findings are corrected. The prior category implementation
is explicitly included in this candidate, with its dedicated bug and tests.
The feedback audit wrapper queues only copied, capped scalar domain fields;
a permanent RED/GREEN test catches retention of oversized request identifiers.

The exact staged source was exported to `/tmp/mypocket-mcp-candidate.UUU24K`
(source tree `717606d5a942ab86a3cdacdc24625ed634419090`) to exclude all unrelated
advisor/native worktree changes. With disposable loopback PostgreSQL 17
(`57415`, schema 40 clean) and Redis 7 (`57419`):

| Command from candidate `backend/` | Result |
| --- | --- |
| `rtk proxy env TEST_DATABASE_URL=<fixture> TEST_REDIS_URL=<fixture> go test -json ./... -count=1` | 741 pass, zero fail, 9 explicit live provider/storage skips, 13 tested packages |
| Same command with `-race` | 741 pass, zero fail, same 9 skips, 13 tested packages |
| `go test -race -json ./internal/controller/http ./internal/usecase ./internal/infrastructure/ai ./internal/infrastructure/cache ./internal/infrastructure/repository -run '^TestMCP\|^TestApplicationRouter\|^TestAudit\|^TestRedis.*Audit\|^TestRedisRecord\|^TestPostgres.*Audit\|^TestCreateCategoryIdempotency\|^TestAdvisorConfirmCategoryAction\|^TestCategoryCreateMutation' -count=1` (same fixture environment) | 210 pass, zero fail/skip, five packages |
| `rtk go test ./internal/controller/http -run 'TestMCP.*Docs\|TestMCPAuthenticatedRouteParity\|TestMCPRouteExclusions\|TestApplicationRouter\|TestCandidateRouteContract' -count=1` | 9 pass, zero fail/skip; 101 tools, 19 exclusions, 91 OpenAPI paths |
| `rtk go generate ./cmd/api`; `rtk go vet ./...`; `rtk proxy go mod verify` | Byte-identical Swagger regeneration; vet and modules pass |
| `gofmt -l` on owned Go files; `rtk proxy git diff --cached --check`; `rtk proxy node app/scripts/check-design-docs.mjs` | Clean formatting/diff; 229 valid design links |

Normal/race/focused/docs JSON logs are `/tmp/mcp-final-candidate-{backend,race,focused,docs}.json`.
The earlier shared-worktree suite passed 753 results with 11 opt-in skips;
that is broader dirty-worktree evidence, not the isolated candidate's count.
The final local report in `.superpowers/sdd/2026-10-20-mcp-api-parity/`
records reproduction commands, source ownership, fixture removal and residuals.

## Docs review and acceptance

Human and AI-agent Markdown, generated OpenAPI, ADR, API/feedback references,
approved design/plan, context/backlog/validation and release notes are reconciled.
OpenAPI retains a valid body parameter and uses `x-multipart-form` for the
alternate file variant; runtime discovery publishes the closed screenshot schema.
No ERD/schema, native UI, auth authority, deployment configuration or provider
contract changes were needed. Local owner/SDK acceptance is automated;
staging/client/provider/device acceptance and rollout remain open.
