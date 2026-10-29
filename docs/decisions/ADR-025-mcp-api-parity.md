---
artifact_type: architecture_decision
id: ADR-025
status: accepted
owner: shared
updated: 2026-10-20
---

# ADR-025 — MCP preserves authenticated REST owner authority

## Context

The owner approved full implementation of [MCP/API parity](../work/tickets/MCP-API-PARITY-20261020-DETAIL_DESIGN.md)
on 2026-10-20 and selected the official Go SDK. [ADR-021](ADR-021-full-owner-api-keys.md)
already establishes identical same-owner user authority for JWTs and API keys,
without scopes. Personal agents need the same operations while preserving
canonical validation, receipts, compare-and-swap and audit privacy.

## Decision

Use `github.com/modelcontextprotocol/go-sdk/mcp` v1.8.0 for stateless JSON
Streamable HTTP at authenticated `/api/v1/mcp`. The SDK owns protocol framing,
negotiation, initialize/list/call and protocol errors. MyPocket owns named
operation descriptors, bounded encoding/results and in-process dispatch through
the canonical REST router with the original bearer. `NewApplicationRouter` is
the shared production/test route composition.

Every authenticated user operation belongs to the descriptor catalog. The
current inventory is 101 tools; 19 excluded method/path entries classify
credential bootstrap/exchange, service-token routes, public surfaces and the
transport itself. Tests compare catalog, actual registered routes and OpenAPI,
failing on unclassified additions. Runtime `tools/list` publishes authoritative
tool names and closed input schemas; Swagger documents the authenticated
transport and canonical REST contracts without duplicating every tool schema.

JWT, API key and MCP calls have equal same-owner authority. No scopes or new
MCP-only approval policy are introduced. Ordinary user writes execute directly;
conversational Assistant prepared actions still require their existing
server-persisted, owner/conversation/generation-bound confirmation endpoint.
AI-entry proposals retain their ordinary review/approve/reject flow.

Pass idempotency metadata into the canonical header, preserve endpoint-specific
receipt support and version checks, and never retry writes automatically.
Normalize async handles and finite advisor replay without promising token
streaming. Convert AI-entry base64 files to bounded canonical multipart and
attachment redirects to ephemeral MCP resource links plus authenticated REST
fallback. Audit dedicated `mcp.*` metadata and bounded byte counts through a
nonblocking best-effort 64-event queue and the existing sink/outbox. Full queues
drop events; durability starts at outbox persistence, not queue acceptance.
Exclude payloads, secrets, caller contexts and signed URLs.

The server marks audited in-process MCP dispatch with a private context key,
so canonical advisor authentication remains active while its duplicate REST
access audit is suppressed in favor of the outer MCP event. Direct REST access
audit remains synchronous. A copied feedback service at application composition
uses a separate bounded queue for MCP feedback-created/resolution domain events;
direct REST retains its original sink contract. Both queues have 64 pending
slots and at most one drain worker, with drop-on-full semantics. Feedback PNG
multipart and screenshot ResourceLinks follow the canonical 1 MiB/2 MiB limits
and owner checks. Category replay uses the existing durable mutation ledger,
with category, wallet links and response receipt in one database transaction.

## Alternatives

- Direct repository/usecase MCP calls duplicate owner checks, DTO validation,
  receipts and domain error mapping; rejected in favor of canonical dispatch.
- An arbitrary `call_rest(method,path,body)` tool allows unsafe surface growth
  and poor discoverability; rejected in favor of named closed schemas.
- Separate MCP scopes or universal confirmation would contradict full-owner
  parity and split user authority; rejected. Existing operation-specific
  confirmation and password requirements still apply.
- A hand-written JSON-RPC transport duplicates SDK protocol responsibilities;
  rejected in favor of the owner-selected official dependency.

## Consequences

One REST authority owns domain behavior across transports and route drift is
tested. The adapter is coupled to router assembly and must evolve alongside
new routes/DTOs. Full-owner keys are powerful; integrations can perform ordinary
financial writes and key lifecycle actions. Key expiry, revocation, owner
binding, provider-run validation and redacted audits remain required.

Stateless JSON has no persistent server session or live advisor token stream.
Outer JSON, multipart, response and deadline bounds may reject requests that a
different REST transport could carry; clients must follow canonical recovery
and receipt rules after ambiguous failures. Binary attachments remain private
resource links rather than in-band bytes. The adapter adds no database schema,
quota or deployment by itself. Staging/production/provider/device acceptance is
not inferred from local code or generated docs.

## Trace and verification

[Task plan](../superpowers/plans/2026-10-20-mcp-api-parity.md),
[API](../architecture/API.md), [auth](../architecture/AUTH_API.md),
[API keys](../guides/api-keys.md), [Advisor](../architecture/AI_ADVISOR_TOOLS.md),
[human/agent contract](../ai/mcp-api-parity.md),
[validation](../work/VALIDATION_MATRIX.md), [changelog](../releases/CHANGELOG.md).
Catalog/auth/write/AI/audit/docs tests provide local proof; rollout gates are
listed in the integration contract and remain distinct from this decision.
