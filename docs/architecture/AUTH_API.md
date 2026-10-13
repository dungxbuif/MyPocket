# Authentication and refresh-token contract

MyPocket keeps the short-lived JWT access token for API calls and adds a
rotating opaque refresh token for long-lived sessions. Refresh tokens are never
JWTs and are never logged or returned after the initial login/refresh response.

## Contract

- `POST /api/v1/login` and Google login responses return `token`, `expires_at`,
  `refresh_token`, and `refresh_expires_at`.
- `POST /api/v1/auth/refresh` accepts `{ "refresh_token": "..." }` and returns
  the same login output shape with a new access token and a new refresh token.
- Each successful refresh deletes the previous refresh-token record before
  issuing the replacement. Replaying a rotated token returns `401`.
- Refresh records are Redis entries keyed by a SHA-256 digest and scoped to the
  stored user/session. The raw token is held only by the client.
- `REFRESH_TOKEN_TTL_SECONDS` controls the lifetime and defaults to 30 days.
  Access-token lifetime remains `JWT_TTL_SECONDS` (one hour by default).

The frontend retries one failed authenticated request after refreshing, then
clears the session if refresh is rejected. API-key requests are never refreshed.

## Remembered native device sign-in

Upgrade compatibility: old Redis sessions/refresh records remain readable.
Because old records contain no family/device linkage, their first logout can
also sign out the same owner's other pre-upgrade sessions once. Fresh password
or device logins use independent families and are not affected. Refresh issuance
rechecks revocation after storing replacement credentials, before returning
them, so a refresh overlapping logout cannot resurrect a revoked family.

Device credentials are independent of access/refresh sessions. Enrollment
requires a normal authenticated session and confirmation of the account
password; API keys are rejected by the JWT-only route middleware.

| Endpoint | Authentication | Input / result |
| --- | --- | --- |
| `POST /api/v1/auth/devices` | JWT session + account password | `{device_id,device_name,password}` → `{id,secret,expires_at}`; secret returned once |
| `POST /api/v1/auth/device-login` | Device credential | `{id,secret}` → ordinary `LoginOutput` with a new access/refresh family |
| `DELETE /api/v1/auth/devices/{id}` | Owner JWT session | Revoke only that owner's enrollment; `204`, unknown/other owner `404` |
| `POST /api/v1/auth/logout` | JWT session | `{refresh_token}` → `204`; revoke current session family, preserve enrolled devices |

Enrollment confirmation with a wrong password returns `403`
`device_enrollment_password_invalid` (not an expired-session `401`). Invalid,
revoked, password-invalidated or expired device credentials return `401`
`device_credential_invalid`. Infrastructure failures return `503` and must not
erase a valid local enrollment.

The backend generates a random 256-bit `mpd_` secret and stores only its SHA-256
digest in PostgreSQL `device_credentials`. Enrollment lasts 180 days and
re-enrolling the same owner/device replaces its secret. A digest of the user's
current password hash invalidates enrollment after a password change. The
server verifies possession of a credential; it does not receive or attest a
face scan. iOS protects the secret with Keychain `biometryCurrentSet` and
`WhenPasscodeSetThisDeviceOnly`; only non-secret display metadata is in
UserDefaults, namespaced by backend URL.

Access/refresh records now carry a session-family ID in Redis. Refresh keeps
the family; logout marks it revoked and deletes the supplied current refresh
record. Session verification rejects earlier access tokens in the same family
and refresh rejects a revoked family, including a concurrently rotated token.
Legacy user-ID-only session values and older refresh records remain readable.
Migration `000031_device_credentials.up.sql` is additive; deploy/migrate backend
before installing an iOS build that uses these endpoints. No financial records
are changed or reset.

Native cancellation keeps enrollment, while a definitive `401` clears it.
Logout/revocation requires a confirmed server response; transient network errors
keep the active session and show a retry message. See the
[Face ID user guide](../guides/face-id.md) for the native actions.

## Native iOS email/password login

The native iOS client currently uses the existing password contract rather than
the web-only Google flow:

```http
POST /api/v1/login
Content-Type: application/json

{"email":"user@example.com","password":"<user-supplied-password>"}
```

The response is the normal `LoginOutput` shape above. The app stores the access
and refresh tokens in Keychain, never stores the password, and refreshes the
session through `POST /api/v1/auth/refresh` after an expired access token. The
identifier labelled “username” in the native UI is the account email; no new
`username` column or alias is introduced.

The native login screen no longer presents a Google button. Google routes remain
available during migration for existing browser users and rollback; they must
not be disabled until every affected account has a verified password setup or
reset path and a production rollback window has been approved.

## Native iOS Google PKCE exchange

`POST /api/v1/auth/mobile/google/exchange` is a mobile-only public auth route.
The iOS client sends `{code, code_verifier, redirect_uri, platform:"ios"}`
after an `ASWebAuthenticationSession` S256 PKCE flow. The backend exact-matches
the configured iOS client/redirect pair, exchanges the code with Google, checks
the verified Google email and returns the same `LoginOutput` shape as web login.
The callback URL contains only the authorization code and state; access and
refresh tokens are never placed in the deep-link URL. `GOOGLE_IOS_CLIENT_ID`,
`GOOGLE_IOS_REDIRECT_URL` and the optional backend-only
`GOOGLE_IOS_CLIENT_SECRET` are environment settings. The current browser
Google callback remains unchanged.
