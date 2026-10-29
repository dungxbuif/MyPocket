# iOS physical-device UAT — 2026-10-19

Status: partial / manual-device gates open. This is an acceptance record
for the connected iPhone; it does not change production data, app data, or
feedback status.

## Personal signing renewal — 2026-10-20 evening

Owner reports expired installation and requests renewal. Build current stable
native source as Release with automatic provisioning for the same Personal Team
and bundle `com.mypocket.ios`, then install over the existing app without
uninstalling or resetting its container. Artifact directory:
`/tmp/mypocket-renew-device-20261020`.

Before installation verify code signature, embedded profile expiry/device,
production API endpoint and portrait-only metadata. After installation verify
device inventory and launch. The native agent reports 374/374 unit tests
passing for the included chat race fixes; this is not full physical UAT.
Production backend changes are outside this installation operation.
Outcome: Release build, signature verification and artifact preflight PASS.
`devicectl device install app` succeeded on the paired iPhone 15 Plus, followed
by successful `device process launch` for `com.mypocket.ios`. No uninstall or
data reset was used. Embedded profile expiration is `2026-10-27T12:50:24Z`
(19:50:24 Asia/Ho_Chi_Minh). Version remains 1.1.0 (2); this is an owner-requested
Personal Team renewal, not a published release. Xcode emitted the existing
full-screen/orientation warning; portrait artifact checks passed.

## Current authoritative status — 2026-10-20

The latest read-only device inventory confirms the Personal Team Release
candidate is installed and available:

```text
MyPocket  com.mypocket.ios  1.1.0  2
```

The candidate artifact passes strict code-signature verification and release
metadata preflight for the production API URL and portrait-only orientation.
The latest feedback-race source build was also installed successfully from
`/tmp/mypocket-release-device-latest-1792431904/Build/Products/Release-iphoneos/MyPocketIOS.app`;
after the owner unlocked the phone, `devicectl` launched the bundle
successfully and captured `/tmp/mypocket-iphone-latest-20261020.png` (PNG
1290x2796). The captured frame is the real production-configured login screen.
This proves install/launch/screenshot on the physical device; touch, Face ID,
and full screen acceptance remain separate manual gates.
The historical table below records earlier attempts for audit traceability;
its earlier signing-blocked row is superseded by the Personal Team install
recheck documented later in this file. Remaining gates are manual biometric/
touch acceptance, physical offline interruption and APNs/provider validation;
none is inferred from install alone.

### Latest source build installed — 2026-10-20 00:54 +0700

The current source (including the portfolio correction adapter) was built with
Xcode Release configuration and Personal Team signing:

```text
Artifact: /tmp/mypocket-release-device-final-20261020/Build/Products/Release-iphoneos/MyPocketIOS.app
Bundle: com.mypocket.ios
Install: PASS on 00008120-001C28C02131A01E
Launch: PASS after device unlock
Screenshot: /tmp/mypocket-iphone-final-20261020.png (1290x2796)
```

This is a device-install verification only; it does not claim App Store,
TestFlight, APNs, biometric or full touch-matrix acceptance.

### Latest backlog build installed — 2026-10-20 01:01 +0700

After the portfolio correction UI and AI capability/error hardening landed, a
fresh Release build was signed and installed over the same paired device:

```text
Artifact: /tmp/mypocket-release-device-backlog-20261020/Build/Products/Release-iphoneos/MyPocketIOS.app
Install: PASS
Launch: PASS after device unlock
Screenshot: /tmp/mypocket-iphone-backlog-20261020.png (1290x2796)
```

The phone was unlocked for a fresh Mobile MCP retry. The installed Device Kit
reports version `0.0.34`, but the UI snapshot still times out waiting for
WebDriverAgent. This is a tooling/device-automation blocker, not an app launch
failure; `devicectl` launch and screenshot remain green.

### Authenticated foreground capture — 2026-10-20 01:04 +0700

After the owner unlocked the phone, `devicectl` relaunched the installed
candidate and captured an authenticated, populated transaction screen:

```text
Install: MyPocket 1.1.0 (2)
Launch: PASS
Screenshot: /tmp/mypocket-physical-native-uatevidence-20261020.png (1290x2796)
```

The frame shows the wallet selector, week filter, transaction card, shared
bottom navigation, add action and feedback bubble in the production-configured
native shell. This proves authenticated rendering against the owner account;
it does not replace manual touch, Face ID, offline or full route acceptance.

## Device and candidate

- Device: iPhone 15 Plus, UDID `00008120-001C28C02131A01E`, paired and
  available through `devicectl`.
- OS: `26.6.1` (`23G83`), Developer Mode enabled.
- Installed app: `com.mypocket.ios`, version `1.1.0`, build `2`.
- Existing installed app was retained. No uninstall, reset, logout, or
  financial mutation was performed.

## Evidence

### Launch and persisted lock state — pass

The existing installed app was terminated and relaunched with:

```sh
mobilecli apps terminate --device 00008120-001C28C02131A01E com.mypocket.ios
mobilecli apps launch --device 00008120-001C28C02131A01E com.mypocket.ios
```

`devicectl` then captured the app screen successfully. The first post-launch
frame displayed the native `Mở khóa MyPocket` screen with:

- `Xác thực để mở khóa` (Face ID/keychain unlock action)
- `Đăng xuất`
- copy stating that the saved session remains protected and requires Face ID
  or the passcode/key.

This proves the app launches and preserves the locked-session state. A prior
foreground capture showed the Finance Assistant screen and owner-scoped data
after the session had already been unlocked.

### Installed app and screenshot capture — pass

```sh
xcrun devicectl device info apps --device 00008120-001C28C02131A01E --search MyPocket
xcrun devicectl device capture screenshot \
  --device 00008120-001C28C02131A01E \
  --destination /tmp/mypocket-physical-relaunch.png
```

The installed app was reported as developer-built and the screenshot was
saved at 1290×2796. Repeated immediate captures after launch stabilized on
the same frame after the initial lock screen; no full-screen loading overlay
was observed in this run.

### Release build/install — pass (Personal Team)

The current source was built for the physical device using the available
Personal Team signing identity:

```sh
xcodebuild -project MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -configuration Release \
  -destination 'id=00008120-001C28C02131A01E' \
  -derivedDataPath /tmp/mypocket-physical-uat-20261019 \
  DEVELOPMENT_TEAM=AU442S3MN3 CODE_SIGN_STYLE=Automatic build
```

The build succeeded and was installed/launched with `devicectl`. The generated
`Info.plist` was verified as `MP_API_BASE_URL=https://money.dungxbuif.com/api/v1`,
version `1.1.0`, build `2`. Screenshot evidence is
`/tmp/mypocket-prod-release-20261019.png`.

For subsequent Xcode runs, keep the scheme's Run configuration on **Release**
for the physical UAT device (or pass the Release destination explicitly). The
repository's Debug configuration intentionally points at
`http://127.0.0.1:8080/api/v1` for local development; using the default Xcode
Run action with Debug can therefore launch an app that only shows previously
cached data and cannot reach production. Before installing a device candidate,
verify the resolved app bundle with
`mobile-ios/scripts/release-artifact-preflight.sh` and confirm the production
HTTPS base URL. This is a build-selection guard, not a signing or API change.

Associated Domains is intentionally absent from this Personal Team build
because Apple Personal Teams cannot provision that capability; Face ID remains
enabled. A distribution/TestFlight profile is still required for a signed
release with universal links.

### Face ID interaction — blocked by tooling, not classified as product pass

The device reports Face ID hardware support, and the app presented the
unlock action. However, `devicectl` cannot simulate a physical Face ID match,
and MobileNext `mobilecli` cannot start WebDriverAgent on this device:

```text
failed to start agent on device 00008120-001C28C02131A01E:
failed to wait for agent: timed out waiting for WebDriverAgent to be ready
```

`devicectl device settings biometrics` also reports that biometrics cannot be
controlled through this host. A human must tap the unlock action and approve
Face ID/passcode on the physical phone to complete this gate.

### Offline/loading — not run

No network-disabling operation was performed because the physical device is
using its own cellular/Wi-Fi connectivity and the available `devicectl`
settings API has no airplane-mode/network control. The local unit and cache
tests cover the offline/loading contract; physical offline acceptance remains
open until the owner disables network on the device or an approved network
proxy/fault fixture is attached.

### Push permission/token lifecycle — not run

This gate requires a Release-signed binary with the `aps-environment`
entitlement. The current local build intentionally does not request push
permission or register an APNs token, so no permission prompt was expected in
the installed candidate. When signing is available, verify the following in
order without mutating financial data:

1. Deny notifications: the in-app inbox remains usable and no device-token
   registration request is sent.
2. Allow notifications: the app receives an APNs token, calls
   `POST /api/v1/notifications/devices`, and never renders or logs the raw
   token.
3. Relaunch/retry: the same device/environment updates one registration and
   does not create duplicates.
4. Logout/device removal: `DELETE /api/v1/notifications/devices/{id}` is
   owner-scoped; a registration failure does not log the user out.

These checks prove registration state only. They do not prove APNs delivery;
that requires provider credentials, entitlement, and an explicit notification
send fixture.

## Result

| Gate | Result | Evidence / next step |
|---|---|---|
| App installed and launchable | PASS | `devicectl` inventory + launch |
| Existing session lock screen | PASS | physical screenshot after relaunch |
| Face ID unlock | BLOCKED | requires human Face ID/passcode; WDA timeout |
| Latest current-source Release build/install | PASS | `/tmp/mypocket-release-device-backlog-20261020/Build/Products/Release-iphoneos/MyPocketIOS.app` installed and launched after unlock; see latest backlog build section above |
| Offline cached refresh | NOT RUN | no safe network fault control available |
| Loading flash regression | PARTIAL PASS | no full-screen overlay in immediate relaunch captures; interactive sweep pending |
| Push permission/token lifecycle | NOT RUN | current binary has no APNs entitlement/provider; follow checklist above |

The older signing-blocked row from the initial recheck is retained only as an
audit trail; it is superseded by the Personal Team Release install recorded
above and by the later backlog build evidence. It must not be interpreted as
the current install status.

## Current Mobile MCP/device recheck — 2026-10-19 23:49 +0700

The connected physical iPhone was rediscovered with `xcrun devicectl` and the
installed `com.mypocket.ios` candidate was launched and captured:

```text
MyPocket  com.mypocket.ios  1.1.0  2
Screenshot: /tmp/mypocket-physical-uat-current.png (1290x2796)
```

The frame is still protected by the app's Face ID/keychain lock and is
intentionally blurred. A fresh current-source Release build could not be
created because this Mac has no Apple account/profile for team `AU442S3MN`.
No financial mutation was made. The Mobile MCP/WebDriverAgent executable is
not installed or exposed in this environment (`mobilecli` is unavailable), so
full touch automation cannot be claimed; human Face ID unlock remains
required before the screen matrix can continue.

### Personal Team Xcode run — 2026-10-19 23:50 +0700

Xcode's Signing & Capabilities UI resolved **Dũng Bùi (Personal Team)** with
`Apple Development: bui.dung.2k@icloud.com (GYRA8D4M54)` and reported the run
finished on the connected iPhone. This confirms the personal account can
install/run a development build through Xcode. The Xcode Run action uses the
Debug configuration, whose API base is intentionally `127.0.0.1:8080`; it is
not a production UAT build. Production UAT must use Release configuration and
requires the personal account's managed profile to be available to the
command-line build as well.

## Offline contract regression — 2026-10-19 23:41 +0700

The simulator offline contract suite ran on the booted `MyPocket-iPhone-16`
destination with `SyncEngineTests`, `OutboxStoreTests` and
`FinancialSnapshotCacheTests` selected. XCTest reported 56 passed, 0 failed,
0 skipped. The suite covers FIFO drain, duplicate idempotency keys, dropped
responses, 401 owner/session blocking, 408/429/503 backoff, 409 conflict,
422 rejection, logout cancellation, wrong-owner isolation and persisted
snapshot recovery. Evidence:
`/tmp/mypocket-offline-contract-20261019/Logs/Test/Test-MyPocketIOS-2026.10.19_23-41-17-+0700.xcresult`.

This closes the deterministic simulator contract portion only. Physical
airplane-mode/reconnect and full touch coverage remain manual gates because
the available device tooling cannot control network state or reliable
biometric/touch automation on the connected iPhone.

The production API rollout and isolated API UAT are recorded below. No
production financial data was reset or deleted by this UAT.

## Production API UAT — 2026-10-19

The API/worker `release-20261019-uat2` release was deployed separately before the physical-device
attempt. The production database was backed up and migrated from version 34 to
37 with `dirty=false`; the API and worker are both `1/1` on the immutable
`release-20261019-uat2` image. A dedicated UAT owner was created without
touching existing financial data. Authenticated smoke checks passed for
profile, wallets, categories, home, savings report and refresh. The UAT owner
also created two isolated wallets, an expense transaction and an atomic wallet
transfer; the transfer returned two linked rows with the same transfer ID.

The strict read-only route audit passed (`88` live paths = `88` local paths).
The public health endpoint and Swagger document returned successfully. No
APNs delivery claim is made.

### Authenticated production regression — 2026-10-19 23:05 +0700

The dedicated UAT owner was exercised with the GET-only regression harness
`scripts/mypocket-production-uat-regression.sh`. Login and every populated
read surface returned HTTP 200: profile, home, wallets, categories,
transactions, savings/insider/month reports, jars, recurring, travel events,
portfolio, notifications, API-key listing, feedback, changelog, sync
capabilities, advisor capabilities/overview, AI-entry capabilities,
budget-series and budgets. The harness also probes credit statements, jar
reports/months and budget reports/transactions when the UAT owner has those
fixtures; absent fixtures are skipped rather than manufactured. Advisor
conversation history is skipped until a real conversation ID exists.

The run is recorded in `/tmp/mypocket-production-uat-regression-20261019.log`:
all attempted checks passed and no production mutation was attempted. This is
API read-surface regression evidence, not a substitute for signed-device
Face ID, offline, APNs or human visual acceptance.

### Read-only regression recheck — 2026-10-19

After the current worktree changes, the strict public route audit again
reported health `ok`, OpenAPI `1.1.0`, and `88` live paths matching `88`
local paths. The dedicated UAT owner again returned HTTP 200 for every
attempted authenticated GET surface in
`scripts/mypocket-production-uat-regression.sh`; no production mutation was
attempted. Local provider/OCR/AI tests and the disposable-database Go race
suite also passed. These results do not close the remaining physical offline,
APNs, or full touch-regression gates.

## Final physical-device attempt — 2026-10-19 23:01 +0700

The paired iPhone was available and the installed `com.mypocket.ios` app was
launched with `xcrun devicectl`; a 1290×2796 screenshot was captured at
`/tmp/mypocket-uat-physical-final.png`. The captured frame is privacy-blurred
by the locked-session screen, so this proves launchability but not an unlocked
interactive sweep.

A fresh Release device build was attempted with both the release team and the
local Apple Development identity. Xcode failed before compilation because no
Apple account or provisioning profile is available for team `AU442S3MN` and
bundle `com.mypocket.ios`; `-allowProvisioningUpdates` reported the same
missing-account/profile condition. Therefore no new binary could be installed
over the existing app through that release-team path.

## Personal Team install recheck — 2026-10-19 23:04 +0700

Xcode was opened with the paired iPhone selected as the run destination. The
Personal Team was selected in Signing & Capabilities and the unsupported
Associated Domains capability was disabled for this local personal-team build;
the repository entitlement file remains unchanged for the normal release
configuration. Xcode completed build and install and reported the app running
on the physical device.

Evidence:

```text
xcrun devicectl device info apps --device 00008120-001C28C02131A01E --search MyPocket
MyPocket  com.mypocket.ios  1.1.0  2
xcrun devicectl device capture screenshot --device 00008120-001C28C02131A01E
Dimensions: 1290x2796 pixels
```

The installed app is currently at the protected “Mở khóa MyPocket” screen.
Interactive Face ID and the populated-screen regression still require the
owner to tap the unlock action on the physical phone; this host cannot emulate
a biometric match on a real device.

The owner subsequently unlocked the installed app on the physical phone. A
fresh screenshot at `/tmp/mypocket-uat-physical-current.png` shows the native
Account screen with the shared navigation, feedback bubble, add button, wallet,
credit, portfolio, budget, jar and month entry points visible. This closes the
Face ID unlock/display gate for this run; offline, APNs and full touch-by-touch
regression remain separate gates.

## Release recheck — 2026-10-19 22:38 +0700

The device remained paired and available. `xcrun devicectl device info apps`
still reports `com.mypocket.ios` at version `1.1.0`, build `2`; no new binary
was installed. A fresh source build for the physical destination was attempted
with the available development team and failed before compilation because
Xcode has no account for team `GYRA8D4M54` and no provisioning profile matching
`com.mypocket.ios`. A fresh unsigned Release simulator build passed with
warnings treated as errors. This confirms the source/build path but does not
close physical signing or installation acceptance.

The fresh simulator artifact was also inspected after packaging:

```sh
APP_PATH=/tmp/mypocket-release-artifact-preflight-20261019/Build/Products/Release-iphonesimulator/MyPocketIOS.app \
  mobile-ios/scripts/release-artifact-preflight.sh
```

Result: `release artifact preflight passed` for bundle `com.mypocket.ios`,
version `1.1.0`, build `2`, production HTTPS API URL and portrait-only
orientation. This validates packaged metadata only; it does not change the
physical device or waive the signing/profile blocker above.

## Native regression audit — 2026-10-19 23:04–23:06 +0700

The current source was rechecked after the production rollout. The native unit
target completed **364 tests, 0 failures** at
`/tmp/mypocket-ios-regression-audit-20261019.xcresult`. The full UI target was
started on the dedicated iPhone 16 simulator; the exercised account/add,
feedback bubble, transaction editor, portrait lock, assistant keyboard and
navigation flows passed. Provider-backed and fixture-dependent UI cases remain
explicitly skipped by their existing guards when no disposable fixture is
configured; this is not a claim of provider or physical-device acceptance.

A fresh physical destination build was also attempted at
`/tmp/mypocket-physical-audit-20261019`. It stopped before compilation with
the same external signing condition: no development provisioning profile for
`com.mypocket.ios` is available to Xcode. No device data or installed binary
was changed by this attempt.

## Mobile MCP/WDA recheck — 2026-10-20 00:09 +0700

The physical iPhone was rechecked without changing application data:

```text
Device: iPhone 15 Plus (iPhone15,5)
UDID: 00008120-001C28C02131A01E
iOS: 26.6.1 (23G83)
Developer Mode: Enabled; pairing: paired; transport: wired
Installed app: MyPocket / com.mypocket.ios / 1.1.0 (2)
```

`xcrun devicectl device process launch` successfully launched the installed
bundle, and `xcrun devicectl device capture screenshot --destination ...`
successfully captured a 1290x2796 PNG at
`/tmp/mypocket-mobile-mcp-recheck-20261020/screen.png`. The captured screen
is the normal email/password login screen, so the app is launchable and no
Face ID lock obscured this run.

Mobile MCP's `mobilecli` binary is not on `PATH`, but the previously downloaded
binary at
`/Users/dungxbuif/.npm/_npx/9f1df1bfb1203d38/node_modules/@mobilenext/mobilecli-darwin-arm64/mobilecli-darwin-arm64`
is available. `agent status` reports Device Kit `0.0.34` installed. A read-only
`mobilecli snapshot` succeeded and exposed the SpringBoard lock screen. A
`mobilecli apps launch` attempt timed out because the phone was locked; no
application data was changed. Automated touch/Face ID and the remaining
interactive screen matrix therefore remain **BLOCKED** until the owner unlocks
the phone and the app can be foregrounded. This recheck proves device-kit
availability and accessibility-tree read capability, not full app acceptance.

## Native transaction/report visual acceptance recheck — 2026-10-20 00:11 +0700

The read-only simulator UI slice was rerun on `MyPocket-iPhone-16` (iOS 18.5)
against the current source. The result bundle is
`/tmp/mypocket-visual-device-20261020-1792429747.xcresult`; the simulator
capture is `/tmp/mypocket-visual-device-20261020.png` (1179×2556).

The selected screen contracts passed **5/5** with zero failures and no runtime
warnings: transaction editor/date picker base controls, ledger wallet-scoped
header, transaction overflow actions, reports month filter/reload, and the
secondary-route shell inset/feedback boundary.

Three provider/fixture-backed cases were intentionally skipped by their
existing guards because `LOCAL_API_UI` was not configured: transaction
presentation consistency across overview/ledger/wallet, report insider
details, and report trend values. This is UI/navigation/geometry coverage only;
it does not claim populated report totals or live API readback.

## Mobile MCP/WDA recovery and touch probe — 2026-10-20 01:07–01:11 +0700

The earlier WDA timeout was reproduced while the MobileCLI daemon had a stale
runner/tunnel state:

```text
failed to start agent on device ...: failed to wait for agent:
timed out waiting for WebDriverAgent to be ready
```

No product source, app install, or financial data was changed. Restarting the
local daemon restored the agent connection:

```sh
mobilecli daemon stop
mobilecli daemon start
mobilecli dump ui --device 00008120-001C28C02131A01E --format text
```

The final command succeeded and returned the full accessibility tree for the
running MyPocket app, including stable identifiers for the wallet selector,
add-transaction button, feedback bubble, transaction rows and all shell tabs.
`mobilecli apps foreground` also reported `com.mypocket.ios` version `1.1.0`.

A safe navigation probe tapped the logical account-tab coordinate `(367,870)`
and the subsequent UI dump reported `bottom-tab-account (selected)`. The
device orientation was set to portrait using `devicectl` before the probe;
this avoids the landscape-offset screenshot seen while the app is portrait
locked. This proves WDA accessibility and one route transition, but is not a
complete manual screen/Face ID/offline acceptance run.
### Follow-up — auth field reproduction, 2026-10-20

MobileNext reproduced a real-device layout defect after entering a long email:
the email clear action moved outside the viewport and the password control
disappeared from the accessibility tree. The shared `MPTextField`/`MPSecureField`
base was constrained to the available row width and the focused native design
suite passed 98/98. The Release device build completed, but `devicectl install`
was blocked because the iPhone was locked and iOS could not mount the developer
disk image. After the owner unlocked the device, Release `1.2.0` installed
successfully. Mobile MCP logged into the isolated production UAT account and
traversed Overview, Transactions, Account and the floating feedback entry
point. The long-email field stayed bounded, the password field stayed visible,
and password login completed. Face ID/offline fault acceptance remains a
separate manual gate.

