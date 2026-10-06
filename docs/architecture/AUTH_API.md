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
