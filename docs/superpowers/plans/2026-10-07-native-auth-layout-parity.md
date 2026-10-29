# Native auth layout parity

## Intent

Match `AuthGateScreen` and the supplied mobile reference: centered circular
`M` mark, direct title/subtitle on the canvas, flat shared fields, primary
full-width sign-in action, divider, and compact secondary Google action.

## Contract

- Preserve email/password, Google PKCE, loading/error, session and biometric
  behavior; this slice changes presentation only.
- Compose shared native bases (`MPText`, `MPTextField`, `MPSecureField`,
  `MPButton`, `MPStatusMessage`) and no screen-local colors or shapes.
- Keep the existing `auth-email`, `auth-password`, `auth-google` identifiers.

## Verification plan

- Add a copy/layout contract test and capture RED.
- Implement the canvas/flat-form anatomy and compact secondary action.
- Run focused/full native tests, unauthenticated UI smoke, Release build and
  project/base/web guards.
- Reconcile iOS design docs, validation matrix, port audit, context and
  changelog.

## Status

- [x] Contract test added and red run captured
- [x] Native auth anatomy implemented with shared bases
- [x] Native unit/UI, Release and guards pass
- [x] Documentation reconciled

## Evidence

- RED: focused contract initially failed because `AuthReferenceCopy` was not
  defined.
- GREEN: `testAuthCopyMatchesWebMobileContract` passed 1/1 at
  `/tmp/mypocket-auth-layout-green/Logs/Test/Test-MyPocketIOS-2026.10.07_15-54-26-+0700.xcresult`.
- UI: unauthenticated `testAppLaunches` passed 1/1 at
  `/tmp/mypocket-auth-ui-parity/Logs/Test/Test-MyPocketIOS-2026.10.07_15-54-42-+0700.xcresult`.
- Full native unit/contract target passed 109/109 at
  `/tmp/mypocket-auth-full-tests/Logs/Test/Test-MyPocketIOS-2026.10.07_15-55-29-+0700.xcresult`.
- Release warnings-as-errors build passed at `/tmp/mypocket-auth-release`.
- Project/base-component/diff guards, web typecheck, Vite production build and
  design/design-doc guards all pass with the bundled Node runtime.
- Visual evidence: `/tmp/mypocket-auth-parity.png` shows the native circular
  mark, direct canvas title/subtitle, flat fields, full-width primary action,
  divider and compact Google action matching the web AuthGate anatomy.
