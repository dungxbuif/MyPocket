---
id: NATIVE-UI-SESSION-ISOLATION
status: regression_verified
scope: mobile-ios
---

# Native UI test session isolation

## Symptom

`LaunchTests.testAppLaunches` expects the unauthenticated Google sign-in action,
but a previous simulator run can leave a valid Keychain session for
`com.mypocket.ios`. The app then restores the authenticated shell and the test
fails even though the app is behaving correctly for a returning user.

## Reproduction evidence

- Simulator UI smoke run on 2026-10-06 failed only `testAppLaunches`.
- The failure was `Auth screen should expose the Google sign-in action after launch.`
- The same run opened the authenticated overview, proving a persisted session
  was restored before the assertion.
- `KeychainSessionStore` is intentionally persistent and `AuthViewModel.restore`
  has no debug-only unauthenticated override.

## Intended behavior

UI tests that assert the auth surface must be able to start from a deterministic
unauthenticated state without changing normal release behavior or deleting a
real user's session.

## Verification plan

1. Add a UI test launch argument for an unauthenticated debug fixture.
2. Verify the test still fails before the app honors the argument.
3. Implement the smallest DEBUG-only restore branch that clears only the test
   Keychain item and sets `.unauthenticated`.
4. Re-run native unit tests, standard UI smoke, simulator launch, and release
   build/guards.

## Resolution

The DEBUG-only `-UITestUnauthenticated` launch argument now clears only the
simulator Keychain session and biometric preference before setting the auth
phase to `.unauthenticated`. `testAppLaunches` uses the argument, so it does
not depend on whichever user session a previous simulator run left behind.

Red proof: the test failed with the argument before the restore branch existed.
Green proof: the targeted test passed, followed by the full default UI suite
(9 tests, 5 passed, 4 explicitly skipped API-fixture tests, 0 failures) on
iPhone 16.
