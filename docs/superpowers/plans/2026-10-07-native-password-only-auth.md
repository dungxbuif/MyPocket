# Native password-only authentication plan

## Intent

Make the iOS login screen follow the owner's latest auth decision: sign in with
account email and password. The current source contradicts the architecture
doc by still rendering a Google button and exposing a native PKCE action.

## Contract

- Keep email/password login, Keychain session storage, refresh-token restore,
  logout and optional Face ID unlock.
- Remove Google sign-in from the native user interface and production
  authentication implementation, including its client/redirect build settings
  and URL scheme registration.
- Keep the backend's Google routes independent; the DEBUG-only simulator fixture
  may continue exchanging its synthetic local code to create a disposable API
  session.
- No data, API, migration, or production configuration changes.

## Verification

- First add a UI assertion that Google login and its divider are absent, and
  observe it fail against the current screen.
- Then remove the UI and native production PKCE implementation/configuration;
  verify unauthenticated email/password screen and existing auth/session tests.
- Reconcile native auth docs, port audit, validation and changelog. Run native
  Release warnings-as-errors and project/base-component/diff checks.

## Non-goals

- Do not remove server-side Google OAuth for existing web users.
- Do not use real user credentials or production data during verification.

## Result

- The launch assertion failed before implementation because both Google UI
  controls existed; it passes after the password-only change.
- The unauthenticated UI test passes 1/1 at
  `/tmp/mypocket-auth-green.xcresult`; the native unit/contract suite passes
  127/127 at `/tmp/mypocket-auth-unit-rerun.xcresult`.
- Release warnings-as-errors passes at `/tmp/mypocket-auth-release`;
  `verify-project.sh`, base-component guards/tests, plist validation and
  `git diff --check` pass.
- The simulator capture at `/tmp/mypocket-auth-password-only.png` shows only
  email/password login. No backend/API/schema/migration change was made.
- Physical iPhone build was attempted with the installed Apple Development
  identity. Xcode found the iPhone but could not sign/install because no Xcode
  account is configured for the team and no development provisioning profile
  exists for `com.mypocket.ios`.
