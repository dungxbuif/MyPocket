---
artifact_type: architecture_reference
id: MOBILE_IOS
status: implementation_in_progress
owner: shared
updated: 2026-10-08
trace:
  detail_design: ../superpowers/specs/2026-10-05-native-ios-swiftui-design.md
  plan: ../superpowers/plans/2026-10-05-native-ios-swiftui-port.md
  adr: ../decisions/ADR-010-native-ios-swiftui-client.md
---

# Native iOS client

## Current status

Newest transport/session checkpoint: sendEnvelope retains meta, expectedOwnerID
guards local queued session identity, refresh atomically replaces/clears only
matching credentials across Keychain instances. Full219/219(218unit+API-key E2E),
0skip. [ADR022](../decisions/ADR-022-owner-bound-native-refresh.md). Receipt/resource
validator/sync actor/cache/UI remain open. Local fixture API rebuilt/restarted at
127.0.0.1:8081; production and physical installation unchanged.

2026-10-14 authoritative checkpoint: native-only MyPocket UI/base consolidation
and full-gap completion remain in progress. Archived web is not maintained or a
visual authority. See [screen registry](../design/platforms/ios-screen-registry.md)
and [completion work](../work/tickets/NATIVE-FULL-COMPLETION-20261014.md). Older
counts/reference-parity notes below are historical, not current acceptance.
Latest209/209unit gate includes encrypted local outbox persistence/leases; it
does not prove live provider/UI/prod/physical Face ID or signed delivery.

### Offline storage boundary

Core/Offline owns SQLite3 through OutboxStore, CryptoKit AES-GCM and
OutboxKeyProviding. The actor is the only table consumer; caller supplies private
Application Support directory/server URL/account identity. Schema1 keeps ordinal,
randomID, hashed namespace and hashed operation/key in clear; financial payload,
state, attempts, retry timestamp, leaseID/expiry and server IDs are sealed.
Account/server isolation, key loss refusal, ciphertext authentication, FIFO,
exclusive claims and stale response rejection are verified with real temporary
SQLite and Simulator Keychain. See [ADR020](../decisions/ADR-020-encrypted-native-outbox.md).
Typed markOutcome replaces four duplicated state setters; expired lease recovery
retains key and attempts. Auth block/resume APIs exist but are not wired to session
events. No screen/sync engine currently instantiates the store; reads cache,
retry policy, discard UI and delayed-Travel context still need implementation.

## Historical migration checkpoint — 2026-10-08

The SwiftUI client is implemented on the isolated `codex/native-ios-port`
branch. It does not replace the web client or the historical Expo client. The
first platform gate is an iOS Simulator build/test; physical-device signing
and authenticated production sign-in are separate gates.

The current Stitch-parity checkpoint is 127 native unit/contract tests, 32
bounded UI tests, Debug/Release simulator builds, project/base-component
guards and the Reports reference-card/trend smoke path green. Feedback/changelog
description/status parity, amount keypad arithmetic and all `MPDateField`
date-only inputs now use the shared transaction-style calendar base.
Authenticated live data, provider/OCR calls, pixel comparison against the
supplied Stitch screenshots and physical-device signing remain explicit release
gates. Budget start/end
periods now use the same shared account-timezone calendar as transactions,
wallets, transfers and portfolio forms. Date+time fields in recurring and AI
entry also use the opt-in time section of that base, preserving the existing
instant API contract. The transaction custom-period range forwards that same
timezone to both date fields.

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

The app icon is packaged through `MyPocketIOS/Assets.xcassets` as the
`AppIcon` asset catalog entry. `scripts/generate-app-icon.sh` derives its
square-background 1024×1024 PNG from the canonical web PWA wallet mark; run it
when `app/public/pwa-icon.svg` intentionally changes. Keep `project.yml`'s
asset-catalog resource and `ASSETCATALOG_COMPILER_APPICON_NAME` setting in
sync. `scripts/verify-project.sh` regenerates and byte-compares the image,
checks dimensions, and Release build verification confirms it is compiled into
`Assets.car`.

`MPAmountField` remains a project-owned base component. Its direct number-pad
input uses a UIKit `UITextField` bridge with a UIKit `inputAccessoryView` for
`Xong` instead of SwiftUI's `.toolbar(placement: .keyboard)`. The field also
opens the shared `MPAmountKeypad`, whose checked non-`eval` parser accepts only
integer-safe non-negative VND results. The bridge preserves the binding,
accessibility label, rounded-border style and keyboard dismissal behavior
without entering SwiftUI's invalid-frame toolbar path.

`MPTextField` is likewise a project-owned base component. Its text input uses a
small UIKit `UITextField` bridge with a delegate-backed binding so per-keystroke
updates do not resign the field or restore a stale editing buffer. The shared
clear button resets both the binding and native field state; feature editors
must pass a stable accessibility identifier when a field needs deterministic
UI-test targeting (for example, `recurring-name`).

The base-component guard also rejects raw feature-level `Text` and `Image`
declarations. Feature rows use `MPText` and `MPIcon`; decorative icons are
hidden from accessibility while interactive rows expose one stable semantic
label. This keeps the design-system ownership rule from being bypassed by
small visual primitives.

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

The global native feedback bubble hides itself for one rendered frame before
capturing the current window, so the attached context is the underlying screen
and never the feedback control. The capture is a bounded PNG (at most 1 MiB)
and the optional `screenshot` multipart part is sent as
`screen.png`/`image/png`, matching the backend's private feedback attachment
contract. Text-only feedback remains the fallback when a window snapshot
cannot be produced within that bound.

`MPBottomNavigation` has an explicit intrinsic height and owns the centered
56pt add-transaction FAB as an overlay of that bounded bar. It must not use a
flexible safe-area height: doing so lets the bar's background intercept taps
over More-menu rows on compact devices. More routes use a typed
`NavigationStack` path so Overview deep links and the shell back action share
one route source of truth.

Adaptive chat exception (owner-approved 2026-10-14): while Assistant keyboard is
visible, MainTabView omits navigation and its height reservation. Assistant moves
overview access into header, requests keyboard dismissal through composer-owned
FocusState, preserves shell draft and restores tabs on keyboard hide. Shared
MPBottomNavigation uses MPMenuSelector with full destination labels for
accessibility Dynamic Type; account-tab hit-test guard applies only to normal
tab layout, never the menu. Existing normal navigation/actions remain unchanged.
No API/schema/provider changes. See NATIVE-CHAT-ADAPTIVE-20261014 work ticket.

When the account tab is selected, the shell navigation is disabled for the
long More menu and a narrow transparent hit-test guard is placed only over the
selected account-tab footprint inside the safe-area inset. This prevents a
repeat tap on the selected tab from falling through to the nearest menu row
while leaving centered More rows (including Hũ chi tiêu) tappable. The menu
also reserves the matching bottom inset and the authenticated route sweep
verifies all 15 destinations.

The wallet-management route is also part of that path: it exposes the shared
back affordance when opened from More, while tapping a wallet row hands the
selected owner wallet back to the shell and recreates `TransactionsView` with
that wallet scope. The explicit “Chi tiết” action remains the entry point for
goal and credit detail surfaces.

The centered add action keeps the web contract: a tap opens the ordinary
transaction editor, while a 500ms long press routes to the AI-entry review
screen without also presenting the manual editor.

The transactions screen treats transactions and wallets as required data, but
loads category labels and account timezone as optional metadata. A metadata
request failure therefore leaves the ledger usable instead of replacing valid
rows with a generic load error.

Portfolio UI keeps current-price drafts keyed by asset ID, mirrors the nullable
price contract, and sends the selected trade date through the existing API.

Recurring schedules exclude credit wallets and filter category choices by the
selected expense/income type. Changing the schedule type clears an incompatible
category before a create/update request is sent, matching the web form.

`MPWalletSelector` keeps API enum values unchanged in bindings and payloads but
renders the same Vietnamese semantic labels as the web UI (for example,
`expense` → `Chi`, `credit` → `Tín dụng`, and `percent` → `Phần trăm thu
nhập`). Unknown values remain visible rather than being discarded.

Jar allocation validation matches the web contract: fixed amounts must be
positive, percentage allocations must be greater than zero and at most 100, and
switching allocation mode clears stale values from the previous mode.

The ordinary transaction editor and AI proposal editor exclude credit wallets,
and changing wallet or income/expense type clears a category that no longer
applies. Credit-ledger rows remain read-only in aggregate transaction detail;
credit mutations go through the dedicated statement actions.
Credit-entry category selection uses the same searchable hierarchy as
transactions, budgets and recurring schedules, rendered as an in-place step
inside the existing credit action sheet. It reuses the server-provided
category tree and existing entry-kind/credit-wallet eligibility; category ID,
payload, owner scope and server validation remain unchanged. Payment forms do
not use or enter the category-selection step.

Budget category scope is expense-only and budget/adjustment wallet selectors
exclude credit wallets, matching the web restrictions for those ordinary ledger
operations.

Ended budgets are immutable in the native editor: the screen keeps their
history and delete action visible but does not offer a save/edit path, matching
the web's view-or-delete behavior.

Wallet detail rows use the same action policy as the web ledger. Ordinary
income/expense rows open a detail screen with edit/delete actions; transfer rows
open paired transfer edit/delete actions; balance-adjustment rows are
read-only; and credit rows remain in the dedicated credit statement. The
policy also restricts Travel Mode linking to ordinary non-credit rows. The
detail screen receives the complete owner wallet set so a paired transfer can
be edited without losing its destination wallet, and adjustment display signs
follow `adjustment_direction` (`increase` or `decrease`).

Portfolio trade date pickers use the account timezone and serialize the selected
calendar day at local noon before sending the instant, preventing a device
timezone offset from moving a trade to the previous or next account day.

Finance Assistant history uses one UserDefaults conversation ID keyed by the
authenticated owner ID. The native screen reloads that conversation's latest
50 messages and the read-only monthly overview on entry; a later request can
page older messages with `before_seq`. UserDefaults stores only the opaque
conversation ID, never message text or financial data.

When a conversation is empty, the native assistant renders the same three
first-use suggested questions as the web client. They are ordinary
`MPButton` instances that only populate the composer; they never submit a
provider request by themselves. Report month input, notification mark-read,
and API-key create/revoke actions expose stable accessibility identifiers so
authenticated UI tests and device UAT can target the owner-scoped actions
without relying on localized text.

The native client prefers the existing `POST /api/v1/login` contract, so no
new backend route is required for password sign-in. The existing
`POST /api/v1/auth/mobile/google/exchange` route is referenced only by the
DEBUG-only local API fixture, which submits synthetic fields to mint a
disposable test session. It is not part of native sign-in or Release behavior.

## Environments

`Debug.xcconfig` and `Release.xcconfig` contain the API base URL and non-secret
build settings. No Google client or redirect identifiers are required by the
native target. No `.xcconfig` containing a secret is committed.

For a paired physical iPhone, signing requires an authenticated Xcode account
and a provisioning profile for `com.mypocket.ios`. The API URL must also be
overridden to a host-reachable HTTPS/staging URL; the simulator's
`127.0.0.1` Debug default is intentionally not usable from a physical device.
Use the selected environment's xcconfig or an `MP_API_BASE_URL` build-setting
override instead of committing a developer-specific LAN address.

For a Personal Team build, choose the current developer's `Personal Team` in
the app target's **Signing & Capabilities** pane and leave **Automatically
manage signing** enabled. Do not infer the team from an older certificate or
hard-code another developer's Team ID in command-line build settings. Xcode
creates/refreshes the managed profile for the connected device; on first
launch, the device owner must trust their own development certificate under
**Settings → General → VPN & Device Management**. Free Personal Team profiles
expire and must be refreshed periodically. This is a device-development
workflow, not a production distribution/signing setup.

## SwiftUI hot reload / live preview

The Debug configuration keeps `SWIFT_OPTIMIZATION_LEVEL = -Onone` and
`ENABLE_PREVIEWS = YES`, so Xcode can attach SwiftUI Live Preview and rebuild
only changed view files while the Simulator remains open. Open
`mobile-ios/MyPocketIOS.xcodeproj`, choose the `MyPocketIOS` scheme and the
iPhone 16 Simulator, then run Debug. Canvas previews are presentation-only;
API-backed behavior still requires the normal Debug run and XCTest commands.
Release keeps optimized compilation and the production API URL.

## Shell hit-testing contract

The persistent shell navigation is a shared base component, but secondary
account destinations own the bottom rows of their long menu. `MainTabView`
therefore keeps the shell above feature overlays on primary tabs and disables
its hit-testing while the account destination list is active. Wallet detail
uses a nested safe-area CTA with a 64px cushion above the shell. Do not replace
these with ad-hoc per-screen offsets: if a new secondary screen has a bottom
action, use the same safe-area/base-component pattern and add a compact-phone
UI test.

For API-backed simulator smoke, Debug builds expose the launch argument
`-UITestAuthenticatedAPI`. It submits synthetic fixture fields to the local
development endpoint, stores the returned disposable session in the in-memory
test session store, and then exercises the same authenticated URLSession path
as the app. The argument is DEBUG-only and must not be enabled in Release
builds or pointed at production.

## Native authentication

The native login screen uses the existing `POST /api/v1/login` email/password
contract. “Username” means the normalized account email; the password is sent
only over the configured HTTPS API and is never persisted by the app. A
successful response is stored in Keychain as the access/refresh session.

On restore, an expired access token is refreshed once when the rotating refresh
token is still valid. Only the refreshed valid session can enter the optional
Face ID unlock gate; an expired or rejected refresh token clears only the active
Keychain session. A valid independent device enrollment remains available on the
login screen. Password-confirmed enrollment stores a separate 180-day credential
behind biometryCurrentSet/ThisDeviceOnly; Face ID exchanges it for a fresh server
session. Logout revokes the active session family, not the enrollment. Disable
and forget-device explicitly revoke enrollment. See [auth contract](AUTH_API.md)
and [user guide](../guides/face-id.md). Face ID is not a replacement for server
authentication, and fresh login still needs connectivity.

The native UI uses email/password login only. It does not compile a Google
OAuth client, register a Google callback URL scheme or require Google client
identifiers. Backend Google routes remain available to existing web clients. A
DEBUG-only simulator fixture can exchange a synthetic local code to mint a
disposable test session; it is not included in Release.

## HTTP diagnostic privacy — local verified

APIClient DEBUG request/response diagnostics emit only the HTTP method, status
(once received), and locally generated request ID. They do not interpolate URL
paths/query strings or response bodies, which may contain notes, prompts, or
credentials even on errors. Typed ProblemDetails still reaches UI/error policy;
redaction does not alter response decoding or retry/auth behavior. Multipart
transport emits no payload diagnostics. Two actual-client stdout regressions and
the full native unit suite verify the boundary.
[HTTP evidence](../work/tickets/NATIVE-HTTP-LOG-PRIVACY-20261014.md).

Assistant capabilities/overview, Budgets save and the local auth fixture now use
`APIError.diagnosticDescription` instead of arbitrary error reflection. It
allowlists failure class and numeric HTTP status only; unknown Error becomes
`unexpected`. No ProblemDetails fields, associated transport/decoding descriptions
or NSError metadata are printed. The local simulator fixture debug error label
also uses this summary; normal typed userMessage semantics stay unchanged.
Formatter stdout regressions and4chat UI regressions pass locally; this is not
all-screen fault injection, historic-log deletion or a production rollout.
[Feature evidence](../work/tickets/NATIVE-FEATURE-LOG-PRIVACY-20261014.md).

## Native scope selectors

Category applicability uses the same API contract as the web client: an empty
`wallet_ids` array means all owner wallets, while a non-empty array may contain
multiple wallet IDs. `MPWalletApplicabilityCard` and `MPCheckboxRow` own the
visible multi-select interaction in the DesignSystem. Category create/edit
passes the exact selected set and never collapses an existing multi-wallet
category to its first wallet. `CategoryModelInput` pins the encoder's wire key
to `wallet_ids`; avoid synthesizing it from the acronym property `walletIDs`.
Category management rows open the dedicated editor as a full-row action;
wallet-scope subtitles are derived from the same response, so this presentation
does not require a second endpoint or backend-specific client logic.

## Simulator commands

## Native deep-link routes and auth gate

Native scenes use the stable route contract in
[`2026-10-15-native-routing-auth-design.md`](../superpowers/specs/2026-10-15-native-routing-auth-design.md).
Supported links use either `mypocket://app/<path>` or
`https://money.dungxbuif.com/app/<path>` and include overview, transactions,
new transaction, wallets, wallet detail, budgets, reports, assistant, account,
and feedback, plus categories, jars, months, recurring transactions, Travel
Mode, portfolio, credit, notifications and AI transaction entry. The app
parser rejects foreign hosts, credentials, fragments and sensitive query
values; `wallet_id` is the only accepted query hint.

Routes are held until `AuthViewModel` reaches an authenticated phase. A locked
session must unlock first, while unauthenticated/expired sessions go through
the normal email/password flow. Pending routes are consumed once. Wallet IDs
are only used to select a row from the current owner's wallet response; the
route cannot override owner scope or grant access to another user's data.

The `mypocket` scheme is registered in the native Info.plist. Universal links
also require the production `apple-app-site-association` file and a signed
build with the associated-domain entitlement; adding the parser alone is not
evidence that production universal links are live.

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

Physical-device prerequisite check:

```bash
xcrun devicectl list devices
security find-identity -v -p codesigning
xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj \
  -scheme MyPocketIOS -destination 'id=<paired-device-udid>' \
  -allowProvisioningUpdates \
  DEVELOPMENT_TEAM=<your-team-id> \
  MP_API_BASE_URL=<host-reachable-api-base-url> build
```

This is a prerequisite command, not simulator evidence. Never commit the
developer team, OAuth client secret or a local LAN address to shared config.

If `simctl list runtimes` is empty, install an iOS Simulator runtime from
Xcode Settings → Components before claiming a booted-app test. A generic SDK
build is not evidence that the app launched in Simulator.

## Feature parity

The authoritative route/behavior matrix is in
[the native iOS detail design](../superpowers/specs/2026-10-05-native-ios-swiftui-design.md).
Screen semantics continue to follow the existing web page contracts under
`docs/design/pages/`; iOS-specific layout and accessibility notes belong in
`docs/design/platforms/ios.md`.
