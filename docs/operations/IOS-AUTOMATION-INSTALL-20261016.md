# Physical iPhone automation helper — 2026-10-16

Status: verified. The owner explicitly requested installation of the
automation agent on the connected physical iPhone using the existing Personal
Team. This note records the approved operational design before installation.

## Scope and decision

Install the MobileNext DeviceKit XCTest runner on iPhone 15 Plus
`00008120-001C28C02131A01E`, using team `AU442S3MN3` and a separate helper
provisioning profile. Keep the installed `com.mypocket.ios` app and its profile
intact. Do not change product source, financial data, device pairing, keychain,
certificates, or global MCP configuration. Temporary signing/build artifacts
live outside the repository. No commit or staging is included.

The local Mobile MCP 1.0.11 server is at `http://127.0.0.1:19347/mcp` with
telemetry disabled. Device discovery works, while UI inspection reports that
the helper is not installed. The helper runner suffix is
`com.mobilenext.devicekit-iosUITests.xctrunner`; MobileNext recognizes a signed
runner with this suffix. Obtain a dedicated development profile with Xcode
automatic signing, then install the verified helper IPA using `mobilecli apps
install --force-resign --provisioning-profile ...`. The upstream `agent install`
path downloads the original bundle identity without rewriting it, so the
temporary helper IPA uses the Personal Team-prefixed identity matching its
dedicated profile. Do not use the MyPocket-specific profile.

## Verification plan

1. Confirm helper bundle identity and signing requirements against upstream.
2. Create a dedicated helper profile with the configured Apple account and
   existing signing identity; do not replace or revoke certificates.
3. Install the runner and check installed-app inventory/status.
4. Read screen elements through Mobile MCP on this exact device.
5. Record any iOS trust/confirmation or provisioning blocker precisely.

See [initial probe](MOBILE-MCP-DEVICE-PROBE.md) and
[current project context](../CONTEXT.md). This is developer tooling; public
product/API/setup documentation and financial schemas are unaffected. This
Markdown note is the human and agent-readable operational reference.

## Evidence

Verified at approximately 21:04 +0700 on 2026-10-16:

- Existing Apple Development identity was valid. No new certificate or keychain
  change was needed.
- `xcodegen generate --spec /tmp/mypocket-ios-automation.5P5od5/project.yml`,
  followed by `xcodebuild -project
  /tmp/mypocket-ios-automation.5P5od5/AutomationProvisioning.xcodeproj -scheme
  AutomationProvisioning -configuration Debug -destination
  id=00008120-001C28C02131A01E -derivedDataPath
  /tmp/mypocket-ios-automation.5P5od5/DerivedData -allowProvisioningUpdates
  -allowProvisioningDeviceRegistration build` — pass. The temporary app was
  built to obtain a profile and was not installed.
- Dedicated profile:
  `/Users/dungxbuif/Library/Developer/Xcode/UserData/Provisioning Profiles/df9b0762-fcea-4988-9120-08acdfcfb4c8.mobileprovision`.
  Bundle `AU442S3MN3.com.mobilenext.devicekit-iosUITests.xctrunner`, team
  `AU442S3MN3`, exact target device included, development debugging allowed.
  Expiration: `2026-10-23T14:02:13Z` (renew before reuse after expiration).
- Downloaded upstream DeviceKit runner `0.0.34` from its GitHub release.
  SHA-256 `a39446db09d998289bd2f0821870bb5d9a73330a9f2a3f1f67546b2dd69370bf`
  matches the pinned upstream CLI checksum. Only the temporary app's
  `CFBundleIdentifier` was mechanically changed to the dedicated identity; it
  was repackaged as
  `/tmp/mypocket-ios-automation.5P5od5/devicekit-ios-runner-personal.ipa`.
- `/Users/dungxbuif/.npm/_npx/9f1df1bfb1203d38/node_modules/@mobilenext/mobilecli-darwin-arm64/mobilecli-darwin-arm64
  apps install --device 00008120-001C28C02131A01E --force-resign
  --provisioning-profile <dedicated-profile>
  /tmp/mypocket-ios-automation.5P5od5/devicekit-ios-runner-personal.ipa -v`
  — pass; reports helper `0.0.34`, build `1`, installed.
- `mobilecli agent status --device 00008120-001C28C02131A01E` — pass; reports
  the same installed helper/version. Independent `xcrun devicectl device info
  apps --device 00008120-001C28C02131A01E --search devicekit` agrees.
- MCP `tools/call` → `mobile_list_elements_on_screen`, arguments
  `{"device":"00008120-001C28C02131A01E","format":"text"}` — pass;
  returns 58 accessibility elements from the physical home screen, including
  Device Kit and MyPocket icons. No iOS trust prompt blocked execution. Raw
  screen contents are not copied into project docs.
- Independent `xcrun devicectl device info apps --device
  00008120-001C28C02131A01E --search MyPocket` still reports MyPocket
  `com.mypocket.ios`, `1.1.0 (2)`.

Device UI control was handed to the separately assigned reference survey agent
after this screen-read proof. Product tests/UAT are not implied by helper
installation. The helper runs through the existing local Mobile MCP daemon;
keep that daemon available while surveying the phone. No source code, product
API, schema, production service, financial data, or public docs changed. No
files were staged or committed. Context/backlog/matrix reconciliation is owned
by the parent task; this agent owns only this operational note and tool setup.

## Upstream references

- [Agent installation and checksums](https://github.com/mobile-next/mobilecli/blob/main/commands/agent.go)
- [IPA signing implementation](https://github.com/mobile-next/mobilecli/blob/main/utils/resign.go)
- [Runner discovery and launch](https://github.com/mobile-next/mobilecli/blob/main/devices/ios.go)
