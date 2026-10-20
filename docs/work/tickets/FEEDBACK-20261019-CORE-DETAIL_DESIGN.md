---
artifact_type: detail_design
id: FEEDBACK-20261019-CORE
status: implemented_local
owner: shared
updated: 2026-10-19
---

# Core wallet offline snapshot consistency

## Finding

The wallet screen keeps an encrypted owner-scoped financial snapshot for
transient/offline reads. A successful wallet deletion must update that snapshot
immediately; otherwise a later offline visit can show a wallet that was already
deleted on the server.

## Contract

`WalletsViewModel.delete` remains server-authoritative. It removes the row from
the in-memory list only after the server deletion succeeds, then persists the
same owner-scoped snapshot while retaining transactions, categories, timezone,
and report fields already present in the snapshot. A failed server deletion
must not mutate the cached wallet list.

## Verification

`FinancialSnapshotCacheTests.testDeletingWalletUpdatesOfflineSnapshotImmediately`
uses an authenticated DELETE 204 fixture, reloads the encrypted snapshot, and
asserts that the deleted wallet is absent while the timezone is retained.
The focused simulator test passed on 2026-10-19. This is local evidence only;
physical-device and production rollout acceptance remain separate gates.
