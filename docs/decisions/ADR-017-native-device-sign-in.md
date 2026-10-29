# ADR-017: Separate remembered device sign-in from access sessions

Status: accepted by owner instruction on 2026-10-13; implementation in progress.

Context: native users want Face ID sign-in after logout and after refresh-token
expiry. Keeping a rotated refresh token across logout conflates ending a session
with remembering an account and prevents clear revocation semantics.

Decision: persist a separately revocable 180-day device credential in PostgreSQL
as a digest. Enroll only with a JWT session plus password confirmation. Protect
the raw credential in the iPhone Keychain using current biometrics, device-only
passcode-backed accessibility, and backend-specific metadata namespacing.
Successful device sign-in creates a new session family. Logout revokes that
family but preserves enrollment; forgetting the device revokes enrollment too.

Alternatives: retaining refresh after logout is simpler but does not satisfy
independent expiry/revocation; saving a password increases password exposure;
Secure Enclave challenge signing is a possible future enhancement but is not
needed for this bounded bearer-credential contract.

Consequences: no biometric data reaches the API, no password is stored locally,
and password changes invalidate enrollment. The API trusts possession of a
protected device secret, not biometric attestation. A migration and backend-first
rollout are required. Redis session-family markers block old rotated sessions;
PostgreSQL enrollment survives Redis session loss. Network failure during
logout/revocation is reported rather than claiming server revocation succeeded.

Links: [design](../work/tickets/NATIVE-FEEDBACK-CONSISTENCY-AUTH-DETAIL_DESIGN.md),
[auth API](../architecture/AUTH_API.md), [validation](../work/VALIDATION_MATRIX.md).
