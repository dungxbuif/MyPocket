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
