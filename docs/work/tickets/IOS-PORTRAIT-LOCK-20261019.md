# iOS portrait lock

## Decision

MyPocket iOS is portrait-only. The app delegate returns
`UIInterfaceOrientationMask.portrait` for every window so the shell, bottom
navigation, compact cards, and keyboard forms keep one supported layout.

## Verification

- `DesignSystemTests.testApplicationSupportsPortraitOnly`
- Release simulator build with `CODE_SIGNING_ALLOWED=NO`
- Physical-device signing remains an environment gate: the local Xcode project
  has no configured development team or provisioning profile.

## Rollout note

This is an app-only change. It does not alter backend contracts, migrations, or
production data. A signed device build requires an Apple Developer account/team
configured in Xcode for bundle identifier `com.mypocket.ios`.
