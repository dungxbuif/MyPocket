# MyPocket 1.3.0 build 6 follow-up

Date: 2026-10-29
Native build: `1.3.0 (6)`
Status: signed device artifact ready; physical installation pending

## Included

- Feedback now renders in a transparent lower-trailing overlay above the native
  route instead of reserving a 60pt content lane.
- Existing keyboard, screenshot-capture, feedback-sheet, and trailing-anchor
  behavior remains unchanged.
- No API, schema, finance flow, production feedback status, or deployment
  behavior changed.

## Verified

- Feedback route-height regression: PASS.
- Reference/feedback unit slice: 19/19 PASS.
- Native shell/editor UI regression on iPhone 16 Simulator: 1/1 PASS.
- Base-component guard, project verification, and `git diff --check`: PASS.
- Release build with Swift warnings treated as errors: PASS.
- Artifact preflight and strict deep code-sign verification: PASS.

## Device artifact

`/tmp/mypocket-1.3.0-build6-device-20261029/Build/Products/Release-iphoneos/MyPocketIOS.app`

Bundle: `com.mypocket.ios`; production API; portrait-only; Apple Development
signature for team `AU442S3MN3`.

The connected iPhone was offline, so installation and physical owner acceptance
remain pending. Reconnect the device, then install this artifact from Xcode or
with the normal device deployment flow.
