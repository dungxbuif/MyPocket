# Encrypted account-scoped outbox store

Status: in_progress. Implements persistence prerequisite of approved full offline
scope/[Task4](../../superpowers/plans/2026-10-06-offline-financial-sync.md), not UI
activation or complete sync. Existing worktree/inline execution retained.

Design before implementation: native SQLite3 (system module, no dependency),
actor-owned connection, WAL+synchronousFULL+BEGIN IMMEDIATE for atomic enqueue.
Schema1 rows contain ordinal,random mutation ID,server/account namespace hash,
operation/key intent hash and AES-GCM sealed Codable record. No amount/note/raw
account ID/token plaintext. CryptoKit random nonce; associated data binds namespace
and mutationID. Per-server/account256-bit key in Keychain, AfterFirstUnlockThisDeviceOnly,
non-synchronizing; duplicate-key creation re-reads winning key. Never regenerate a
missing key when encrypted rows already exist. No unencrypted fallback or wipe
on decrypt/unknown-format error. Protected/excluded-from-backup Application Support
directory; app reinstall without uninstall preserves it, unsynced rows are not a
cloud backup. Auth token/provider secrets cannot enter typed transaction/transfer
payload fields. No raw payload/SQL logging.

Initial slice: enqueue/items/pendingCount, stable UUID key across restart, FIFO
by SQLite ordinal, owner+server isolation and idempotent duplicate-enqueue with
body conflict. Models reserve recovery states for later lease/retry/apply actor
methods. Lease/restart sending recovery/cache/sync/UI remain Task4/5 gates, so do
not mark store task complete on enqueue alone. No user usage quota introduced.

Before sync activation, resolve mutable Travel default: server currently links
active trip when input travel_event_id is nil. Queue must freeze entry-time
context or explicitly opt out; add a tested create-context contract/capability
before sending, not silently use the trip active at delayed sync time. This
slice has no networking/financial writes and cannot enable that unsafe path.

Tests: real temporary SQLite files + CryptoKit, restart/key/FIFO/isolation,
same-key conflict/concurrent two-store enqueue, invalid payload and wrong/missing
key refusal without destroying original ciphertext. Actual Simulator Keychain
test checks stable32byte key/accessibility/sync flags with exact service cleanup.
Start fail-closed/transient scaffold after tests for runtime RED, then replace
with persistence; compile errors are not behavioral RED. Regenerate Xcode project
from YAML preserving unrelated settings; units/guard/project/diff proof. Internal
architecture/spec/plan and public human+agent offline guide/current docs updated.
No API/prod data migration/deploy/install/release or owner feedback acceptance.

Persistence checkpoint: runtimeRED70442 (6/6 tests failed as expected), SQLite/
CryptoKit/Keychain implementation passes6/6 then security expansion11/11. Initial
two compile failures were test-only Swift6 enumeration/non-Sendable teardown
captures, corrected without disabling concurrency checks. Full unit gate running.

Next beforecode design: keep state metadata encrypted in the same record; add
attempt count, next retry timestamp, random leaseID/expiry, allowlisted error code
and acknowledged resource IDs. BEGIN IMMEDIATE protects read-modify-write across
store actors. leaseNext returns earliest nonterminal item only: sending lease or
future retry blocks later intents (strict FIFO). Default60s lease; renewals are
owned by sync engine later. Expired leases recover to pending retaining exact key
and attempt count; live leases cannot be stolen by another store/restart. Every
outcome update requires matching owner/mutation/leaseID and sending state; stale
responses throw leaseLost instead of overriding a newer outcome. Applied,
conflict,rejected,cancelled are never automatically retried. Explicit blockForAuth
revokes live leases and retains pending/retry items as blockedAuth; resumeAuth
only releases same-account blocked items to pending. No network/auth bypass.
Tests: two-store race single lease, FIFO with retry delay, lease expiry/restart,
stale response refusal, terminal state persistence, auth isolation and re-login.
Do not enable sync or claim full offline feature from store proof alone.

Result: lease runtimeRED43035 four new tests fail against nil/no-op scaffold;
63674 focused15/15pass. Added negative owner/invalid outcome/terminal overwrite
regression; final88836 exit0,209/209native units (16store),0skip, bundle
`/tmp/mypocket-outbox-store-final-20261014.xcresult`. Base guard, guardtests,
verify-project and diff-check pass. Scope remains in_progress: no session wiring,
sync/cache/discard UI/Travel-safe network send or production/install/release.
Docs reconciled: ADR020/native architecture/public human+agent guide/CONTEXT/
BACKLOG/VALIDATION_MATRIX/parent plan. OpenAPI/backend ERD unchanged because no
API/server-schema change in this slice; release withheld until owner acceptance.
