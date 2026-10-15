# Native Google authentication contract

> Superseded by the owner's later password-only migration decision. The iOS
> app no longer exposes Google login, ships PKCE client code, or requires an
> OAuth callback scheme. Backend Google OAuth for the web client is unaffected.

## Decision

The native auth screen must expose the same Google sign-in capability as the
web client. It uses the existing `GooglePKCEClient` and the backend
`/auth/mobile/google/exchange` contract; the app never embeds a Google client
secret. The client ID and custom callback URI remain build configuration, not
source-controlled credentials.

## Behavior

- Email/password remains the primary form.
- A shared secondary base button starts the PKCE browser session.
- Successful exchange stores the returned access/refresh session in the
  existing session store and enters the authenticated shell.
- Cancelled login leaves the auth screen unchanged.
- Missing or placeholder configuration shows an actionable Vietnamese error
  instead of attempting a malformed OAuth request.
- The callback scheme remains `mypocket://oauth/callback`; Google Cloud must
  register the iOS client and this exact redirect URI before device UAT.

## Scope

Auth view/model only. No API, migration, token-storage, or backend changes.
The existing refresh-token and Face ID flows remain unchanged.

## Verification

Add unit coverage for configuration/error mapping and UI coverage that the
Google action is visible on the unauthenticated screen. Run the native unit
target, project/base-component guards, and a Debug simulator build. A real
Google account exchange remains a credential/device UAT gate.
