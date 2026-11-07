---
artifact_type: detail_design
id: MCP-API-PARITY-OWNER-COVERAGE-20261107
status: approved
owner: shared
updated: 2026-11-07
---

# MCP owner API coverage

## Intent

Make MCP expose every authenticated owner operation that a valid JWT or
`mpk_` API key can call. Keep credential-establishment, public, service-token,
and MCP transport routes excluded.

## Scope

- Add the 15 currently omitted owner routes for reports, advisor entries/actions,
  and portfolio catalog/unit/price-source operations.
- Add the REST `transactions.list` `jar_id` query parameter to the MCP schema and
  dispatch mapping.
- Replace the current parity test allowance so owner routes cannot remain hidden
  behind an exclusion entry.
- Reconcile the agent-readable MCP contract and generated API references.

## Non-goals

- No production deployment, release tag, route exposure change, or OpenClaw setup.
- No change to JWT/API-key authorization semantics or database schema.

## Decision

Reuse the canonical REST router through the existing MCP operation catalog. Each
descriptor keeps the same method/path, request schema, query/path encoding,
owner checks, and idempotency behavior as REST. The MCP contract will report
116 authenticated owner operations; only non-owner routes remain excluded.

## Verification

Run the focused MCP route/catalog/OpenAPI tests, the authenticated middleware
tests, and the backend package tests covering the changed handlers. Production
parity remains a later step requiring a valid owner credential.
