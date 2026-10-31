# Implementation Plan: AI Attachment Diagnostics and APNs Transport

Date: 2026-11-01
Status: implemented locally; external acceptance open
Scope: close backlog item #8 (privacy-safe per-file AI attachment diagnostics) and connect the existing push-delivery boundary to a real APNs transport and durable worker queue.

## Constraints

- Preserve backward compatibility for existing AI-entry responses and APNs-disabled deployments.
- Per-file diagnostics expose only a stable input position, safe stage/status, and allow-listed error code. Never expose OCR text, source bytes, object keys, provider URLs, device tokens, private keys, or raw upstream errors.
- A failed attachment is diagnosable without automatic re-upload/retry; successful evidence remains visible.
- APNs uses token-authenticated ES256 JWT over Apple's sandbox/production HTTP endpoint. Secrets are configuration-only and never persisted or logged.
- Push delivery remains disabled by default until `APNS_*` configuration and Apple entitlements are provisioned.

## Workstreams

1. AI attachment diagnostics: persist/map per-file processing outcomes, add a privacy-safe API field, update native models/presentation, and add deterministic mixed-success tests.
2. APNs transport: implement JWT/HTTP transport and response classification, add durable `push_deliveries` repository/worker leasing and notification handoff, and add deterministic HTTP/queue tests.

## Verification

- Backend unit/race tests for both workstreams and the full `go test ./... -count=1` suite.
- Native iOS unit/build tests covering attachment-to-diagnostic mapping where native code changes.
- `git diff --check`, API/generated contract checks, migration/project guards, and a disposable Postgres worker/queue check where available.
- Live APNs delivery is not claimed without Apple credentials and device entitlement; deterministic HTTP tests prove request signing/header/body and all provider status paths.

## Reconciliation

After implementation, update the backlog, context, API/worker architecture docs, validation matrix, and changelog with exact test evidence. Record any deployment-only gate separately from code completion.
