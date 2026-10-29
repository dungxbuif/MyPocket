---
artifact_type: operational_guide
id: IOS-BUILD-HANDOFF
status: active
owner: shared
human_fields:
  - apple_team_selection
  - device_trust_and_provisioning
ai_fields:
  - repository_build_contract
  - environment_contract
  - verification_steps
---

# MyPocket iOS: clone, Personal Team build, install and handoff

This is the reproducible handoff for a developer who wants to clone the native
iOS client, sign it with their own free Apple Personal Team, install it on a
connected iPhone, and point it at an already-running MyPocket API. It does not
copy certificates, private keys, provisioning profiles, passwords, API keys,
or backend provider credentials.

## Current build contract

| Setting | Repository value |
| --- | --- |
| Xcode project | `mobile-ios/MyPocketIOS.xcodeproj` |
| Scheme / app target | `MyPocketIOS` |
| Deployment target | iOS 17.0 |
| Bundle identifier | `com.mypocket.ios` |
| Debug API base | `http://127.0.0.1:8080/api/v1` |
| Release API base | `https://money.dungxbuif.com/api/v1` |
| Public API health | `https://money.dungxbuif.com/api/v1/health` |
| Swagger UI | `https://money.dungxbuif.com/api/v1/docs/index.html` |
| URL scheme | `mypocket://` |
| Orientation | Portrait only |
| Version/build in the checked-in plist | `1.2.0 (3)` |

The Release configuration is the one to use for a physical iPhone connected
to the production API. A Debug build keeps the loopback URL intentionally; an
iPhone cannot reach the Mac's `127.0.0.1`. For local backend testing, override
`MP_API_BASE_URL` with an HTTPS host reachable by the phone, or use a simulator
with the local API fixture.

The current checked-in entitlement files are empty. The Personal Team file is
kept as an explicit build override at
`mobile-ios/MyPocketIOS/Signing/MyPocketIOS.personal.entitlements`. No
Associated Domains/Universal Links capability is required by this local build;
the custom `mypocket://` URL scheme remains in `Info.plist`.

## Prerequisites

1. macOS with a supported Xcode installation (`xcodebuild -version`).
2. An Apple ID signed in to Xcode under **Settings → Accounts**. A paid Apple
   Developer membership is not required for a personal device build, but Apple
   controls the short-lived provisioning profile and capability limitations.
3. An unlocked, trusted iPhone with Developer Mode enabled and a USB/Wi-Fi
   connection to the Mac.
4. A reachable MyPocket API. The app never connects directly to PostgreSQL,
   Redis, S3, OCR, or the model provider.

## Clone and open

```sh
git clone <repository-url> MyPocket
cd MyPocket
open mobile-ios/MyPocketIOS.xcodeproj
```

The checked-in project is usable directly. `mobile-ios/project.yml` is the
XcodeGen source; regenerate only when intentionally changing project structure:

```sh
xcodegen generate --spec mobile-ios/project.yml
```

Do not put a personal Team ID into the shared `project.yml` or commit a local
bundle-ID change. If Apple reports that `com.mypocket.ios` is unavailable for
your account, use a local uncommitted bundle identifier and keep the API
contract unchanged.

## Configure the Personal Team in Xcode

1. Select the `MyPocketIOS` target, open **Signing & Capabilities**, select
   your Apple ID team, and let Xcode manage signing.
2. Keep the application bundle ID as `com.mypocket.ios` when it is available;
   otherwise choose your own unique reverse-DNS ID locally.
3. If Xcode rejects the normal entitlement file, select the empty Personal Team
   file or use the command-line override below. Do not add Associated Domains,
   APNs, App Attest, or other paid capabilities to this personal build.
4. Trust the developer profile on the iPhone if iOS asks for it. Never import a
   certificate or private key copied from another developer or machine.

The Team ID in the commands below is intentionally a placeholder. Obtain it
from Xcode's account/team picker; never put a secret or private signing key in
this repository or in chat.

## Build a signed Release `.app`

Set the values locally, then run from the repository root:

```sh
TEAM_ID="<YOUR_PERSONAL_TEAM_ID>"
DEVICE_UDID="<YOUR_CONNECTED_DEVICE_UDID>"
BUILD_DIR="/tmp/mypocket-ios-personal-release"

xcodebuild build \
  -project mobile-ios/MyPocketIOS.xcodeproj \
  -scheme MyPocketIOS \
  -configuration Release \
  -destination "platform=iOS,id=${DEVICE_UDID}" \
  -derivedDataPath "${BUILD_DIR}" \
  -allowProvisioningUpdates \
  -allowProvisioningDeviceRegistration \
  DEVELOPMENT_TEAM="${TEAM_ID}" \
  CODE_SIGN_STYLE=Automatic \
  CODE_SIGN_IDENTITY="Apple Development" \
  CODE_SIGN_ENTITLEMENTS=MyPocketIOS/Signing/MyPocketIOS.personal.entitlements \
  SWIFT_TREAT_WARNINGS_AS_ERRORS=YES \
  MP_API_BASE_URL="https://money.dungxbuif.com/api/v1"
```

The final artifact is normally:

```text
${BUILD_DIR}/Build/Products/Release-iphoneos/MyPocketIOS.app
```

Verify the resolved artifact before installing. Use the checked-in expected
metadata only when the version/build has not changed:

```sh
APP_PATH="${BUILD_DIR}/Build/Products/Release-iphoneos/MyPocketIOS.app" \
  EXPECTED_BUNDLE_ID="com.mypocket.ios" \
  EXPECTED_VERSION="1.2.0" \
  EXPECTED_BUILD="3" \
  EXPECTED_API_URL="https://money.dungxbuif.com/api/v1" \
  mobile-ios/scripts/release-artifact-preflight.sh

codesign --verify --deep --strict "$APP_PATH"
codesign -dv --verbose=4 "$APP_PATH" 2>&1 | sed -n 's/^Identifier=\|^TeamIdentifier=\|^Authority=//p'
```

The preflight is read-only. It checks bundle ID, version, production HTTPS API,
portrait orientation, and that the app bundle does not contain obvious secret
files. It does not prove that the API is healthy or that Apple provisioning
will remain valid.

## Install and launch on the iPhone

With the phone unlocked and trusted:

```sh
xcrun devicectl device install app \
  --device "${DEVICE_UDID}" \
  "$APP_PATH"

xcrun devicectl device process launch \
  --device "${DEVICE_UDID}" \
  com.mypocket.ios
```

If `devicectl` is unavailable, open the project in Xcode, choose the connected
phone and the `MyPocketIOS` scheme, then press **Run** after selecting the same
Personal Team. Xcode's Devices and Simulators window can also install the
resolved `.app`.

## First-run verification

1. Confirm the app reaches the production health/API host rather than
   `127.0.0.1`.
2. Sign in with the existing email/password contract:
   `POST /api/v1/login`. The native app does not need Google credentials and
   does not show the browser Google flow.
3. Verify Keychain-backed refresh using `POST /api/v1/auth/refresh`, then enable
   Face ID. Face ID unlocks the still-valid stored session; it is not a password
   transport and does not bypass server expiry or logout.
4. Exercise one read-only screen, one ordinary transaction draft, and logout.
   Do not use a production account for destructive tests without an explicit
   owner-approved test plan.

## Backend configuration contract

The native bundle receives only `MP_API_BASE_URL`. Provider/database secrets
stay on the Go API/worker. Local API startup reads the ignored
`backend/.env.local` (process environment wins); production receives the same
secret names from its deployment secret store. Names currently defined by
`backend/.env.example` and `backend/internal/config/config.go` are:

```text
APP_ENV HTTP_ADDR DATABASE_URL REDIS_URL
JWT_SECRET JWT_TTL_SECONDS REFRESH_TOKEN_TTL_SECONDS
AI_BASE_URL AI_API_KEY AI_MODEL AI_STORE_USAGE
OCR_API_URL OCR_API_KEY
S3_ENDPOINT S3_REGION S3_BUCKET S3_PREFIX S3_FORCE_PATH_STYLE
S3_ACCESS_KEY_ID S3_SECRET_ACCESS_KEY
GOOGLE_CLIENT_ID GOOGLE_CLIENT_SECRET GOOGLE_REDIRECT_URL
GOOGLE_IOS_CLIENT_ID GOOGLE_IOS_CLIENT_SECRET GOOGLE_IOS_REDIRECT_URL
FEEDBACK_AGENT_TOKEN
APNS_ENABLED APNS_TEAM_ID APNS_KEY_ID APNS_TOPIC APNS_PRIVATE_KEY APNS_ENVIRONMENT
CORS_ALLOWED_ORIGINS ALLOWED_LOGIN_EMAILS
SEED_USER_EMAIL SEED_USER_NAME SEED_USER_PASSWORD
```

Use the appropriate local secret file or deployment secret manager as the
source. Never paste values into this guide, an iOS `.xcconfig`, an app bundle,
or an API request made from the client. Production outside development refuses
to start without private attachment storage and `FEEDBACK_AGENT_TOKEN`; APNs
delivery is a separate capability and is not enabled by iOS installation.

## What this build is and is not

- This is a locally signed Personal Team development installation, not an App
  Store/TestFlight distribution artifact.
- Apple may require periodic re-signing/reinstallation and limits Personal Team
  capabilities; the app cannot extend those limits in code.
- No certificate, provisioning profile, private key, API key, OCR key, model key,
  database URL, Redis URL, or user password is copied by this workflow.
- MCP/mobile-MCP integration is not part of this guide; its implementation and
  device automation remain a separate pending work item.

## Agent-readable command summary

```text
bundle_id=com.mypocket.ios
deployment_target=iOS17.0
release_api=https://money.dungxbuif.com/api/v1
health=https://money.dungxbuif.com/api/v1/health
docs=https://money.dungxbuif.com/api/v1/docs/index.html
debug_api=http://127.0.0.1:8080/api/v1
personal_entitlements=mobile-ios/MyPocketIOS/Signing/MyPocketIOS.personal.entitlements
build_gate=mobile-ios/scripts/verify-project.sh
artifact_gate=mobile-ios/scripts/release-artifact-preflight.sh
```

