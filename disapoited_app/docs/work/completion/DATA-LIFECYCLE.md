# Data Lifecycle Evidence — 2026-09-11

Status: implementation and local verification complete; production deployment remains gated by the final release task.

## Contracts

- Imports accept a directly uploaded private CSV or an existing object beneath `users/{userID}/imports/`.
- CSV parsing is UTF-8, header-stable, RFC3339-aware, integer-VND-only, limited to 15 MiB and 10,000 rows, and resolves wallet/category identifiers against records available to the authenticated owner.
- Parsing produces row-numbered errors and a review state. No transaction is posted before a current-version confirmation, and any preview error makes the import non-confirmable.
- Confirmed rows use stable `import:{jobID}:{row}` idempotency keys and execute inside one database transaction through existing finance commands.
- Exports select wallets, user categories, and transactions at a server-recorded snapshot cutoff; CSV ordering and headers are deterministic. Credentials, API key hashes, audit fingerprints, and provider payloads are outside the snapshot schema.
- Export objects live at `users/{userID}/exports/{jobID}.csv`; download URLs expire after five minutes.
- Reset and delete require a recent browser authentication, CSRF, a fresh signed affected-count preview, exact typed confirmation, and an idempotency key.
- Delete disables the user and revokes API keys before asynchronous cleanup. Cookie cache hits are revalidated against the active user record, so cached sessions cannot bypass disablement.
- Cleanup derives the S3 prefix from the authenticated user ID and rejects arbitrary prefixes. Jobs are claimed with `FOR UPDATE SKIP LOCKED`, versioned, resumable after transient failure, and terminal after a bounded attempt count.

## Verification

| Gate | Result |
| --- | --- |
| Lifecycle parser, formatter, repository, worker, HTTP, identity, and S3 integration tests | pass |
| PostgreSQL two-user isolation, idempotent replay, versioned confirmation, reset scope, and immediate disable tests | pass |
| LocalStack bounded put/get/delete proof | pass |
| Full backend race suite | pass |
| Account screen lifecycle tests | pass: 11 tests in the screen file |
| Full frontend unit suite | pass: 20 files, 165 tests |
| Frontend production build | pass |
| Focused account/import/export browser tests | pass: 6 tests |
| Full Playwright suite | pass: 96 tests across mobile Chromium, desktop Chromium, and mobile WebKit |

The browser tests mount the real Account screen, call the real API, upload CSV through the lifecycle endpoint, verify job progress is visible, prove the preview/cancel path leaves wallet data unchanged, and use ordinary clicks without force overrides. Worker integration tests separately execute preview, versioned confirmation, exact-once posting, and private export creation against PostgreSQL.
