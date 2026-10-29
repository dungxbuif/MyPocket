# ADR-020 — Encrypted native outbox with transactional leases

Date: 2026-10-14. Status: accepted within owner-approved offline/full-gap scope;
implementation locally verified, not activated in production or the app UI.

## Context and decision

[Task4](../superpowers/plans/2026-10-06-offline-financial-sync.md) needs durable,
account-isolated intents before any network send. Use system SQLite3 behind the
OutboxStore actor, WAL/synchronous FULL and BEGIN IMMEDIATE. Each record is
AES-GCM sealed using a per-server/account256-bit Keychain key, random nonce and
AAD binding format1/namespace/mutationID. SQLite exposes only ordinal, random ID,
SHA256 namespace and operation/key hash. Payload, state, retry/lease metadata and
acknowledged resource IDs stay encrypted; auth/provider credentials are not fields.

Keychain keys are AfterFirstUnlockThisDeviceOnly, non-synchronizing. A missing key
with existing rows fails closed; never regenerate or wipe unreadable ciphertext.
Directory uses completeUntilFirstUserAuthentication and backup exclusion. App
Support location is supplied by the later app composition layer, not hardcoded
to a temporary directory. No runtime store is instantiated by current screens.

SQLite serializes claims across store actors. Only first eligible FIFO intent
can lease; future retry/live lease/blocked auth prevents later sends. Expired
leases recover using the same idempotency key. Outcome updates require current
owner+mutation+leaseID+sending state. Auth blocking invalidates outstanding leases;
only explicit same-account resume releases blocked records. Typed markOutcome
unifies planned markRetry/markApplied/markConflict/markRejected without four
duplicated persistence paths. Terminal outcomes do not auto-retry.

## Alternatives and consequences

- Plain Codable/UserDefaults: rejected, exposes finance data and lacks cross-store
  atomicity. SwiftData: feasible, but SQLite's explicit transaction/unique index
  directly expresses queue claiming and has no added dependency.
- Whole-file encryption: feasible, but rewrite/crash coordination is more complex.
- Retaining only an in-memory key: rejected, loses intents after app termination.

Ciphertext still reveals queue size/order and hashed identifiers, not universal
metadata secrecy. Uninstall/device loss/key loss can make unsent intents
unrecoverable; this is not a cloud backup. Simulator verifies API attributes and
encryption, not physical hardware lock protection. Store currently scans the
owner queue inside transactions; future retention/scale work must preserve
idempotency and encrypted metadata. It performs no networking or confirmed-balance
projection. Cache/sync/discard UI/auth integration remain separate gates.

## Evidence and remaining gates

[Ticket](../work/tickets/OFFLINE-ENCRYPTED-STORE-20261014.md),
[native architecture](../architecture/MOBILE_IOS.md),
[public/agent guide](../guides/offline-sync-api.md),
[validation](../work/VALIDATION_MATRIX.md). Runtime RED6/6 persistence and4/4
lease tests; final209/209 native unit tests include16store cases. No migration,
prod deployment, signed installation or release. Before sync, freeze entry-time
Travel association instead of inheriting the active trip at delayed send time.
