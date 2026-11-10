---
artifact_type: release_note
version: 1.3.9
status: ready_for_release
date: 2026-11-11
---

# MyPocket 1.3.9: Finance Assistant migration recovery

## Intended behavior

The production migration must create `assistant_delegations` against the
existing MyPocket owner table without changing any other application schema or
shared release routing. The table uses the same text owner identifier as the
authoritative `"user"(id)` column.

The release also provides a narrow repair command for the failed 1.3.8 attempt.
It is a no-op for a clean migration history. It may reset only version 54 when
the history is dirty and `public.assistant_delegations` does not exist; any
other dirty version or partial table fails closed.

## Root cause and decision

The first Finance Assistant migration declared `owner_id uuid` and referenced
`users(id)`. MyPocket actually stores owner identifiers as `text` in the quoted
singular table `"user"`. The immutable `prod-1.3.8` image therefore started
correctly but PostgreSQL rejected migration `000054` before any application
service was updated.

Because production never completed migration `000054`, version 1.3.9 corrects
that migration in place. PostgreSQL rolled back its DDL, while golang-migrate
may retain the pre-execution `54 dirty` marker. The protected deploy flow must
therefore run the guarded repair before retrying `up`. The fix is deliberately
limited to the MyPocket source image; it does not broaden shared CI rules or
alter another application's deployment path.

## Verification plan

- Add a migration contract test that requires `owner_id text` and the exact
  `REFERENCES "user"(id) ON DELETE CASCADE` constraint.
- Test clean-history no-op, exact `54 dirty` recovery, partial-table rejection
  and rejection of every other dirty version.
- Run the backend migration and package tests.
- Build and publish exactly one immutable `prod-1.3.9` source tag from current
  `main`, then let the generated ops receipt perform migration and deployment.
- Verify the ops deployment, route health and both immutable image digests.

## Implementation verification

- Backend race suite: 881 tests passed across 20 packages.
- Finance Assistant integration suite: 39 tests passed against disposable
  PostgreSQL 17.
- Disposable PostgreSQL 16 reached migration version 54 with `dirty: false`.
- A simulated `54 dirty` state with an existing table was rejected; after
  removing the rolled-back table artifact, the guarded repair reset to 53 and
  `up` completed at clean version 54.
- The MyPocket-only ops suite passes all 25 release, isolation, secret and
  rollback checks. Production deployment and digest verification remain the
  release gates.

