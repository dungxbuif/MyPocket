# ADR-021 — Full-owner API keys without scopes

Date2026-10-14. Status: accepted by explicit owner decision; implemented/tested
locally, production activation pending. [Work](../work/tickets/API-KEY-FULL-ACCESS-20261014.md).

Owner rejected adding transactions:write and requested removal of the entire
fine-grained key permission model. Common RequireAuth now accepts JWT/session or
validated mpk_ key. All user operations inherit the key's server-derived owner.
Advisor and feedback reuse common auth without endpoint-specific scope checks.
Secret hash, expiry/revocation, owner isolation and provider-run revalidation stay.
JWT Redis validation/refresh and device password confirmation are unchanged.

Existing scope arrays remain in PostgreSQL to avoid unnecessary destructive
migration; they have no authority. Creation no longer asks for scopes. Scope
normalizers/HasScope/principal scope state are removed. Deprecated internal
input/argument/labels tolerate old callers but are ignored. Metadata responses
expose scopes:[] solely for installed-client decoder compatibility, not rights.
Native UI removes selector and labels full-account read/write authority clearly.

JWT logout revokes its session family; API-key logout revokes itself, without
affecting unrelated sessions/owners. Login/refresh/device-login are credential
exchanges, not interchangeable bearer operations. Internal/admin service-token
routes are not JWT user routes and remain separate. Model tool whitelist and
explicit financial confirmation do not expand merely because bearer auth does.

Alternative scoped keys were explicitly rejected by owner. Consequence: **every
existing valid key gains full account read/write rights at rollout**; integrations
must revoke untrusted keys and keep secrets outside logs/URLs/source. There is no
read-only key option. No server schema/data wipe, key regeneration or quota needed.

Proof: four auth/service runtime REDs→GREEN; real PostgreSQL HTTP creates/replays
ordinary transaction and transfer (exact3rows/2excluded pair), forbids other-owner
read/write/revoke, manages keys and self-revokes on logout. Real JWT/Redis expiry/
logout and key expiry/revocation tests still pass. Full Go race/vet and native
212/212unit+selectedUI pass. Screenshot inspection finds known Feedback bubble
covering helper text; do not claim full visual acceptance. [Validation](../work/VALIDATION_MATRIX.md),
[human/agent guide](../guides/api-keys.md), [auth](../architecture/AUTH_API.md).
