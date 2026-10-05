---
artifact_type: architecture_reference
id: MOBILE_IOS
status: design_approved_for_scaffold
owner: shared
updated: 2026-10-06
trace:
  detail_design: ../superpowers/specs/2026-10-05-native-ios-swiftui-design.md
  plan: ../superpowers/plans/2026-10-05-native-ios-swiftui-port.md
  adr: ../decisions/ADR-010-native-ios-swiftui-client.md
---

# Native iOS client

## Current status

The SwiftUI client is implemented on the isolated `codex/native-ios-port`
branch. It does not replace the web client or the historical Expo client. The
first platform gate is an iOS Simulator build/test; physical-device signing
and production OAuth are separate gates.

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

`APIClient` treats an HTTP `204 No Content` response as a successful
`EmptyResponse` for typed delete/bulk-delete mutations; it never tries to
decode an empty body as an envelope. This is required by the existing wallet,
transaction, transfer, category, budget, jar, recurring, travel, feedback and
API-key revoke contracts.

`APIClient` also parses a request path's query component into
`URLComponents.percentEncodedQuery`. Feature APIs may continue to construct
semantic paths such as `/credit/wallets/{id}/statement?from=...`; the client
must never percent-encode the `?` into the resource path. The URL contract is
covered by native XCTest.

The transactions screen treats transactions and wallets as required data, but
loads category labels and account timezone as optional metadata. A metadata
request failure therefore leaves the ledger usable instead of replacing valid
rows with a generic load error.

Portfolio UI keeps current-price drafts keyed by asset ID, mirrors the nullable
price contract, and sends the selected trade date through the existing API.

Finance Assistant history uses one UserDefaults conversation ID keyed by the
authenticated owner ID. The native screen reloads that conversation's latest
50 messages and the read-only monthly overview on entry; a later request can
page older messages with `before_seq`. UserDefaults stores only the opaque
conversation ID, never message text or financial data.

The only native-specific API addition is
`POST /api/v1/auth/mobile/google/exchange`; it validates the iOS PKCE code and
redirect allowlist before returning the existing login/refresh response.

## Environments

`Debug.xcconfig` and `Release.xcconfig` contain non-secret API base URL and
Google client/redirect identifiers only. Values are supplied per environment;
no `.xcconfig` containing a secret is committed. Simulator development should
use a local or staging API with a fixture OAuth path when real Google OAuth is
not configured.

For API-backed simulator smoke, Debug builds expose the launch argument
`-UITestAuthenticatedAPI`. It exchanges fixture PKCE fields against the local
development endpoint, stores the returned disposable session in the Keychain,
and then exercises the same authenticated URLSession path as the app. The
argument is DEBUG-only and must not be enabled in Release builds or pointed at
production.

## Native authentication

The app uses `ASWebAuthenticationSession` and S256 PKCE. The callback includes
an authorization code and state only. The app sends that code to
`POST /api/v1/auth/mobile/google/exchange`; the Go API validates the exact
platform/client/redirect allowlist and returns the normal short-lived JWT plus
rotating opaque refresh token. Tokens are stored in Keychain. On restore, an
expired access token is refreshed once when the rotating refresh token is still
valid. Only the refreshed valid session can enter the optional Face ID unlock
gate; an expired or rejected refresh token clears the Keychain session and
biometric preference. Face ID is not a replacement for server authentication.

## Native scope selectors

Category applicability uses the same API contract as the web client: an empty
`wallet_ids` array means all owner wallets, while a non-empty array may contain
multiple wallet IDs. `MPMultiOptionSelector` owns this interaction in the
DesignSystem. Category create/edit passes the complete selected set and never
collapses an existing multi-wallet category to its first wallet.

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
