---
artifact_type: audit
id: FEEDBACK-20261019-PLATFORM
status: in_review
owner: shared
---

# Platform contract audit

Trace: [backlog](../BACKLOG.md), [worker intake](WORKER-01-cron-service.md),
[worker contract](../../architecture/WORKER.md),
[offline guide](../../guides/offline-sync-api.md),
[validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md).

## Scope and decision

Read-only audit of `backend/cmd/worker`, auth/session/device use cases,
full-owner API-key service/middleware/repository and offline durable receipt
tests. No production code/test changes proposed: inspected core behavior has
existing meaningful regression coverage. This is not an assertion that every
platform failure mode or all product requests are complete.

Worker implements recurring transactions, budget-series materialization and
advisor expiry as independent passes. Default interval is 60 seconds; positive
integer overrides are accepted; `WORKER_ONCE=true` runs one pass. Repository
operations own duplicate protection. The old WORKER-01/backlog claim that no
worker exists is stale. Month-close artifact semantics, bank jobs, notification
producer/delivery and physical APNs acceptance remain separate product work.

Full-owner API keys intentionally ignore legacy scope arrays. Tests exercise
financial writes/replay, isolated owner reads, child-key creation/revocation,
one-time secrets, metadata redaction and self-logout. Session/device tests cover
owner/family logout isolation, remembered device retention, absolute refresh
family expiry and the deterministic refresh/logout race. Offline integration
checks actual PostgreSQL writes/receipts plus real Redis/JWT authentication,
concurrent single-write behavior, rollback and paired excluded transfers.

## Documentation defect

The worker guide's `go run ./backend/cmd/worker` example assumes a root module,
but this checkout's Go module is in `backend/`. Read-only reproduction:
`rtk proxy go list ./backend/cmd/worker` from repository root fails with
`cannot find main module`. Parent owns WORKER.md reconciliation; change examples
to `cd backend` followed by `go run ./cmd/worker` (including once-mode example).
No worker runtime change is necessary to repair that instruction.

## Verification

Ran the following against disposable PostgreSQL 55436 and Redis 49412; all four
packages passed with no skips in the selected integration cases:

```sh
rtk proxy env 'TEST_DATABASE_URL=postgres://mypocket_test:local-fixture-only@127.0.0.1:55436/mypocket_savings_budget?sslmode=disable' 'TEST_REDIS_URL=redis://127.0.0.1:49412/0' go test ./cmd/worker ./internal/usecase ./internal/controller/http ./internal/infrastructure/repository -run 'Test(WorkerIntervalDefaultsAndOverrides|WorkerOnceFlag|FullAPIKeyFinancialWritesManagementAndOwnerIsolation|OfflineCreateHTTPWithRealJWTSessionRedisAndAPIKeyBoundary|OfflineCreateReceiptReplayConflictAndOwnerIsolation|OfflineCreateReceiptConcurrentSingleWriteAndRollback|OfflineTransferCreateReceiptReplaysExactlyTwoExcludedRows|UserAPIKeyAuthenticateIgnoresLegacyScopesButKeepsOwnerAndRevocation|UserAPIKeyCreateNeedsNoScopesAndIgnoresObsoleteInput|UserAPIKeyValidatePrincipalRejectsRevokedKey|LogoutRevokesRotatedSessionFamilyButKeepsDeviceEnrollment|LogoutCannotRevokeAnotherOwnerOrAnotherSessionFamily|RefreshRotationDoesNotExtendFamilyExpiry|ConcurrentRefreshCannotResurrectFamilyAfterLogoutMarkerExpires)$' -count=1
```

Parent owns combined suite and shared context/backlog/validation/changelog.
No schema, API, runtime behavior, credentials or production state changed.
No new requirements/ADR/ERD/API reference needed for this read-only audit;
the worker command correction is the only proposed doc amendment.
