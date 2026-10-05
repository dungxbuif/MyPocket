---
artifact_type: architecture_reference
id: MOBILE_IOS
status: design_approved_for_scaffold
owner: shared
updated: 2026-10-05
trace:
  detail_design: ../superpowers/specs/2026-10-05-native-ios-swiftui-design.md
  plan: ../superpowers/plans/2026-10-05-native-ios-swiftui-port.md
  adr: ../decisions/ADR-010-native-ios-swiftui-client.md
---

# Native iOS client

## Current status

The SwiftUI client is planned on the isolated `codex/native-ios-port` branch.
It does not replace the web client or the historical Expo client. The first
platform gate is an iOS Simulator build/test; physical-device signing and
production OAuth are separate gates.

## Repository boundary

```text
mobile-ios/
  MyPocketIOS.xcodeproj
  MyPocketIOS/
    App/             # app root and navigation
    Config/          # non-secret build configuration
    Core/            # URLSession, auth, Keychain, DTOs
    DesignSystem/    # all user-facing SwiftUI bases
    Features/        # route-level view models and composition
  MyPocketIOSTests/
  MyPocketIOSUITests/
```

The native client consumes the existing Go API and never opens PostgreSQL,
Redis or S3 directly. Feature code is owner-scoped by the JWT issued by the
backend. Provider secrets remain server-side.

The only native-specific API addition is
`POST /api/v1/auth/mobile/google/exchange`; it validates the iOS PKCE code and
redirect allowlist before returning the existing login/refresh response.

## Environments

`Debug.xcconfig` and `Release.xcconfig` contain non-secret API base URL and
Google client/redirect identifiers only. Values are supplied per environment;
no `.xcconfig` containing a secret is committed. Simulator development should
use a local or staging API with a fixture OAuth path when real Google OAuth is
not configured.

## Native authentication

The app uses `ASWebAuthenticationSession` and S256 PKCE. The callback includes
an authorization code and state only. The app sends that code to
`POST /api/v1/auth/mobile/google/exchange`; the Go API validates the exact
platform/client/redirect allowlist and returns the normal short-lived JWT plus
rotating opaque refresh token. Tokens are stored in Keychain. Face ID is an
optional unlock gate for an existing valid session, not a replacement for
server authentication.

## Simulator commands

```bash
xcode-select -p
xcrun simctl list runtimes
xcrun simctl list devices available
xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj \
  -scheme MyPocketIOS -sdk iphonesimulator -configuration Debug \
  build CODE_SIGNING_ALLOWED=NO
xcodebuild test -project mobile-ios/MyPocketIOS.xcodeproj \
  -scheme MyPocketIOS -destination 'platform=iOS Simulator,name=<available iPhone>'
```

If `simctl list runtimes` is empty, install an iOS Simulator runtime from
Xcode Settings → Components before claiming a booted-app test. A generic SDK
build is not evidence that the app launched in Simulator.

## Feature parity

The authoritative route/behavior matrix is in
[the native iOS detail design](../superpowers/specs/2026-10-05-native-ios-swiftui-design.md).
Screen semantics continue to follow the existing web page contracts under
`docs/design/pages/`; iOS-specific layout and accessibility notes belong in
`docs/design/platforms/ios.md`.
