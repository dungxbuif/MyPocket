# Feedback d098ff72: Face ID missing

Status: in_review (local verification only; production feedback remains unresolved).
Owner request: verify first and preserve behavior that already works.

## Report and boundary

- ID: `d098ff72-bdf8-4e23-92c6-1e00caed1071`
- Verbatim title: `FaceID`
- Verbatim description: `Mất mất chuecs năng faceid rồi`
- Reported symptom: Face ID appears to be missing. The report does not identify
  whether this means the Account enrollment control, login action, or unlock
  prompt, nor does it provide device enrollment/session state.
- Severity: user-visible sign-in capability; priority: focused verification.
- In scope: current native Account/auth controls, remembered-device state,
  simulator reproduction, and read-only deployed-build evidence. A confirmed
  bounded defect may change only the auth/biometric module and its screen docs.
- Out of scope: backend/API/security-model expansion, credentials, production
  mutation/deployment, and shared `CONTEXT`, `BACKLOG`, `VALIDATION_MATRIX`.

## Intended behavior and existing contract

After password sign-in, Account settings offers Face ID enrollment with password
confirmation. Once a valid device credential is enrolled, logout retains it;
the login screen offers Face ID and biometric authentication creates a fresh
session. A valid stored session may instead present the biometric/passcode
unlock screen. The option is absent when no valid enrollment metadata exists.
The protected credential is local to this device. See the [public guide](../../guides/face-id.md),
[Account screen contract](../../design/pages/account-settings/README.md), and
[auth API reference](../../architecture/AUTH_API.md).

Current source has the Account toggle and password sheet in `AccountView`, the
login button in `AuthView`, and guarded restore/logout/device sign-in in
`AuthViewModel` and `KeychainDeviceSignInStore`. Existing
`DeviceSignInTests` cover logout retention, expired sessions, cancellation,
revocation, and network failure. The 2026-10-20 [UAT audit](../audits/FEEDBACK-UNRESOLVED-UAT-20261020.md)
classifies this report as needing physical Face ID and signed-device Keychain
acceptance. These observations do not prove the current UI on a device.

## Reproduction and verification plan

1. Build/test current source on a dedicated iPhone 16 simulator with isolated
   DerivedData and result bundles. Verify the Account entry path and Face ID
   control with synthetic authenticated UI state; verify the password login
   fallback and the no-enrollment state. Run focused device sign-in contracts.
2. If a disposable local API and simulator biometric enrollment are available,
   run the existing end-to-end enrollment/logout/device-login UI case. Do not
   point the fixture at production or enter owner credentials.
3. Inspect installed physical app version and read-only screen state only if
   available without owner authentication. No automated biometric match,
   logout, enrollment, app reinstall, or credential interaction on the owner's
   phone. Leave the sensor/Keychain gate for owner UAT if it cannot be observed.
4. For any reproducible bounded defect, record expected/actual steps and root
   cause here, add a failing focused test before code, then make only the
   auth/biometric fix and reconcile the Account screen and public Face ID guide
   as needed. If behavior is present, add focused regression evidence only.

Expected: Account security control is reachable; the login action appears when
a valid enrollment exists; logout preserves that enrollment. No API, schema,
config, or security-contract change is intended.

## Fresh result, 2026-10-26

- Dedicated `MyPocket-FaceID-d098ff72` iPhone 16 simulator, iOS 18.5,
  `28EB9153-CE80-4C02-843D-C940037D0AD3`; isolated
  `/tmp/mypocket-faceid-d098ff72-derived`. Current-source Debug app build
  succeeded. Its packaged `NSFaceIDUsageDescription` is present and API base
  URL is simulator loopback `http://127.0.0.1:8080/api/v1`.
- On a clean install with no enrollment, the login accessibility tree showed
  email/password and no Face ID button. This is the expected state, not a
  reproduction of a lost enrolled credential.
- With the DEBUG-only synthetic authenticated UI fixture, a direct
  accessibility walk found `Account` tab -> `more-profile-account` -> Account
  settings -> `Bảo mật` -> `account-face-id-toggle` labeled
  `Bật đăng nhập Face ID`. Tapping it opened the password-confirmation sheet
  with `face-id-enrollment-password` and disabled
  `face-id-enrollment-confirm` until input. Screenshot:
  `/tmp/mypocket-faceid-d098ff72-sheet.png`. No password was entered or
  enrollment attempted.
- Focused device sign-in tests passed 5/5 in
  `/tmp/mypocket-faceid-d098ff72-unit-retry.xcresult`. This covers logout
  retention and fresh device sign-in,
  cancellation, revocation versus network failure, expired session, and
  enrollment/forget transport contracts with synthetic credentials.
- Selected UI command with the same destination/DerivedData and
  `-only-testing` for `testAppLaunches`,
  `testFeedbackBubbleDoesNotCoverAccountSecurityDescription`, and
  `testMoreProfileCardOpensAccountSettings`: 2 passed, 1 failed in
  `/tmp/mypocket-faceid-d098ff72-ui.xcresult`. The failing test expected a
  separate `Hồ sơ tài khoản` static subtitle on the More card; it still
  navigated successfully to Account settings. This is a shared UI assertion
  outside this Face ID scope. An initial focused unit attempt also stopped
  before tests on a concurrently edited `DesignSystemTests.swift`
  `groupsDigits` compile mismatch; the retry above passed after that shared
  code changed. Neither issue is treated as a Face ID defect.
- Read-only `devicectl device info apps` reports paired physical MyPocket
  `com.mypocket.ios` version `1.2.1 (4)`, newer than the October 20 UAT note.
  `mobilecli apps foreground` reports SpringBoard. No physical app navigation,
  biometric operation, credential read, install, or production request was
  performed. Physical signed-build enrollment/Keychain/Face ID and owner
  acceptance remain unverified.
- The opt-in `testAPIBackedFaceIDEnrollmentSurvivesLogout` was not run:
  `lsof -nP -iTCP:8080 -sTCP:LISTEN` found no disposable local API listener,
  and no enrolled simulator biometric or test fixture was prepared. Running
  that case without its local-only prerequisites would only skip it. This is
  the remaining local end-to-end proof gap, separate from owner device UAT.

Exact build and automated test commands (run from repository root):

```sh
rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -configuration Debug -destination id=28EB9153-CE80-4C02-843D-C940037D0AD3 -derivedDataPath /tmp/mypocket-faceid-d098ff72-derived CODE_SIGNING_ALLOWED=NO build
rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -configuration Debug -destination id=28EB9153-CE80-4C02-843D-C940037D0AD3 -derivedDataPath /tmp/mypocket-faceid-d098ff72-derived -resultBundlePath /tmp/mypocket-faceid-d098ff72-unit-retry.xcresult -parallel-testing-enabled NO -only-testing:MyPocketIOSTests/DeviceSignInTests CODE_SIGNING_ALLOWED=NO test
rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -configuration Debug -destination id=28EB9153-CE80-4C02-843D-C940037D0AD3 -derivedDataPath /tmp/mypocket-faceid-d098ff72-derived -resultBundlePath /tmp/mypocket-faceid-d098ff72-ui.xcresult -parallel-testing-enabled NO -only-testing:MyPocketIOSUITests/LaunchTests/testAppLaunches -only-testing:MyPocketIOSUITests/LaunchTests/testMoreProfileCardOpensAccountSettings -only-testing:MyPocketIOSUITests/LaunchTests/testFeedbackBubbleDoesNotCoverAccountSecurityDescription CODE_SIGNING_ALLOWED=NO test
```

Conclusion: the reported missing control is not reproduced in current-source
simulator UI; the intended auth contract still passes its focused unit suite.
The report lacks the enrolled-device state needed to attribute a cause, so no
production auth change is justified. A possible cause is absence or expiry of
the local enrollment metadata, but this is only a hypothesis and must not be
presented as the owner's root cause. Keep production feedback unresolved for
owner UAT: on the installed signed build, check Account -> Bảo mật, enroll with
the owner's own password if desired, then verify logout -> Face ID login and
fresh launch -> unlock. If a specific step fails, capture the visible error,
device Face ID enrollment state, and app version without sharing credentials.

Docs review: this note adds the internal verification record. The public Face
ID guide and agent-readable auth API reference already match the observed
contract; neither needs a behavior update. No auth code, API, schema, config,
deployment, shared tracking file, or production feedback status was changed.

Trace: [native phase](../phases/PHASE-002-native-ios-swiftui.md),
[feedback classification](../tickets/FEEDBACK-20261019-CLASSIFICATION.json),
[UAT harness](../tickets/FEEDBACK-20261019-UAT-HARNESS.json),
[physical UAT](../../operations/IOS-PHYSICAL-UAT-20261019.md),
[validation matrix](../VALIDATION_MATRIX.md).
