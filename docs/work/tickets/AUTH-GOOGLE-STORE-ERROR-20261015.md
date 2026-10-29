# Google login persistence error classification

Status: implemented locally; rollout and device verification remain open.

## Problem

`POST /api/v1/auth/mobile/google/exchange` converted every error from
`FindOrCreateGoogleUser` into `GOOGLE_MOBILE_LOGIN_INVALID` (401). A database
outage or unavailable local fixture therefore looked like a bad Google
credential and made the real cause difficult to diagnose.

## Contract

- Invalid/missing OAuth profile remains `401 GOOGLE_MOBILE_LOGIN_INVALID`.
- A user-store lookup/create failure returns `503
  GOOGLE_USER_STORE_UNAVAILABLE` with a generic user-safe detail.
- The underlying persistence error remains available to server-side logs and
  `errors.Is` callers, but is never returned in the HTTP problem detail.
- No auth permissions, token shape, user scope or production configuration
  changes are introduced.

## Verification

- Use-case regression proves the store sentinel and underlying error are both
  preserved.
- HTTP regression proves the 503 code and verifies that database details are not
  leaked.
- Full backend tests/vet and generated documentation checks must remain green.
