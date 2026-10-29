# Data Safety Evidence

Status: local implementation; production backup/restore evidence pending.

## Migration and resync invariants

- Migration runners hold a dedicated PostgreSQL advisory lock for the complete discovery/check/apply sequence.
- Applied migration bodies are pinned by SHA-256 checksum and modified history is rejected.
- Migration 0012 is repeatable against a fixture migrated through 0011, preserves existing finance/portfolio rows, and enables canonical asset change operations.
- Sync snapshots identify `server_epoch=atomic-sync-v1`.
- A user-scoped browser mirror replaces authoritative stores and records the epoch in the same IndexedDB transaction while preserving pending outbox mutations, conflicts, receipt bytes, device identity, and mutation sequence.
- Epoch reconciliation is retryable and cannot block ordinary live finance reads when the sync endpoint is temporarily unavailable or older than the client.

## RED evidence

- Two concurrent migrators failed with PostgreSQL duplicate catalog-key state before serialization was added.
- The sync package did not expose a server epoch.
- The client had no epoch reconciliation function and its new test failed before implementation.
- The first direct release-script invocation failed because the file lacked executable permission; the repository mode is now corrected.

## Local GREEN evidence

The first local drill used Homebrew PostgreSQL client 18 against PostgreSQL 16. Restore rejected the generated `transaction_timeout` setting, revealing an invalid recovery toolchain. The scripts now require matching major versions.

Using PostgreSQL 16 client tools against the local PostgreSQL 16 service:

- backup completed with a custom dump, row-count evidence, S3 inventory, object copy and SHA-256 manifests;
- restore completed into `mypocket_restore_task2`, never the source database;
- the S3 copy was restored only beneath `restore-drill/task2-local-pg16`;
- all public-table row counts and downloaded object SHA-256 values matched.

Focused verification from the Task 2 tree:

- `go test -p 1 ./internal/platform/db ./internal/sync ./internal/platform/httpapi -count=1` — pass;
- `npm test -- --run src/offline src/app/App.test.tsx` — 3 files, 51 tests passed;
- `npm run typecheck` — pass;
- `npm run test:e2e -- offline-sync.spec.ts` — 9/9 passed across mobile Chromium, desktop Chromium, and mobile WebKit;
- `backup-production.sh` plus `restore-drill.sh` using matching PostgreSQL 16 tools — pass.

This is local recovery evidence only. Production remains unchanged and still requires a fresh encrypted backup plus restore drill before deployment.
