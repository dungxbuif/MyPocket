---
artifact_type: adr
id: ADR-007
status: accepted
owner: shared
date: 2026-09-22
---

# Use AWS SDK for Go v2 to presign private S3 reads

## Context

The handwritten SigV4 query signer for private attachment GETs produced URLs rejected by the configured S3-compatible service with `403 SignatureDoesNotMatch`, while the same adapter's signed PUT succeeded. The owner approved using the AWS S3 SDK. See [BUG-001](../work/bugs/BUG-001-s3-presigned-get-signature.md) and [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md).

## Decision

Use `github.com/aws/aws-sdk-go-v2/service/s3` and its presign client for `SignedGet`. Supply the existing static credentials, configured signing region, custom base endpoint, bucket and path-style setting, and preserve the five-minute expiry. Keep the current PUT and DELETE request implementation unchanged in this fix.

## Alternatives

- Continue maintaining the handwritten presigner: rejected because its canonicalization has already failed against the real configured service and duplicates security-sensitive protocol code.
- Replace the entire storage adapter with the SDK: deferred; only presigned GET is implicated, and broadening PUT/DELETE increases change surface without evidence of a defect there.

## Consequences

The backend now has a direct AWS SDK v2 S3 dependency and its required module upgrades. Presigning behavior comes from the SDK, while endpoint compatibility remains a live integration requirement. PUT/DELETE still use existing code and should be migrated only under separately tested scope.

## Trace and verification

- Trigger: [BUG-001](../work/bugs/BUG-001-s3-presigned-get-signature.md)
- Work: [AI-ENTRY-02](../work/tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md)
- Validation: [matrix](../work/VALIDATION_MATRIX.md#ai-entry-02-live-provider-evaluation-2026-09-22)
- Code: `backend/internal/infrastructure/storage/s3.go`
