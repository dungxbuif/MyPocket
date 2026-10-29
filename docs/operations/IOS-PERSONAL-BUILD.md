# Personal Team device build

Owner repeatedly selected free personal development signing for their own
iPhone. A Release device build on 2026-10-16 found that the ordinary target's
Associated Domains entitlement is unsupported by Personal Team AU442S3MN3.
The existing signed MyPocket profile is valid but has no associated-domains
permission; this is an entitlement mismatch, not a missing Apple login.

## Personal variant

For a personal build only, override `CODE_SIGN_ENTITLEMENTS` with an empty
entitlements file. Keep the ordinary target's entitlement file for paid-team
builds. This disables Universal Links in the personal artifact; declared custom
URL schemes, API authentication, Keychain, Face ID and refresh logic remain.
No auth checks or app-side ownership guards are removed.

```sh
xcodebuild -project MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -configuration Release \
  -destination 'platform=iOS,id=<connected-device>' \
  -derivedDataPath /tmp/mypocket-savings-device-20261016 \
  -allowProvisioningUpdates -allowProvisioningDeviceRegistration \
  DEVELOPMENT_TEAM=AU442S3MN3 \
  CODE_SIGN_ENTITLEMENTS=MyPocketIOS/PersonalTeam.entitlements \
  SWIFT_TREAT_WARNINGS_AS_ERRORS=YES build
```

Verification: build success plus `codesign` inspection of the final artifact;
installation/launch and owner Face ID remain separately recorded. Personal
signing still expires according to Apple's development provisioning limits.
Use the main target's ordinary entitlements when Universal Links are required.

This developer-only variant requires no financial API/public behavior change;
this Markdown is the human/agent operational reference. The native screen
survey continues separately and must not be interrupted by app installation.

## Verification result — 2026-10-16

The command above completed with `BUILD SUCCEEDED`. Artifact:
`/tmp/mypocket-savings-device-20261016/Build/Products/Release-iphoneos/MyPocketIOS.app`.
Log: `/tmp/mypocket-savings-personal-build-20261016.log`.
`codesign --verify --deep --strict` succeeds; inspected identifier is
`com.mypocket.ios`, team `AU442S3MN3`, with application/team identifiers and
development debugging entitlements only. No associated-domains entitlement is
present. This artifact has not been installed or owner-accepted.
