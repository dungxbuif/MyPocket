---
artifact_type: release_notes
id: REL-1.1.0-native-ios
status: verified
version: 1.1.0
release_approval: pending
deployment: api_worker_production
linked_phase: ../work/phases/PHASE-002-native-ios-swiftui-port.md
linked_plan: ../superpowers/plans/2026-10-15-release-v1.1.0.md
---

# MyPocket v1.1.0

## Scope

This release freezes the current MyPocket product snapshot and aligns the
native iOS client, web package metadata, and public OpenAPI version. It also
packages the current Finance Assistant capability hardening and the native
base-component, offline-contract, feedback, reports, wallet, jar, and
transaction work already present in this checkout.

## Release contract

- Semantic version: `1.1.0`.
- iOS bundle short version: `1.1.0`.
- iOS bundle build: `2`.
- No database migration is part of this release.
- API and worker must run the same immutable production image digest.
- The web surface is not redeployed by this native-focused release.

## Verification record

Fresh release-gate evidence for tag `v1.1.0`:

- Backend: `rtk proxy go test ./...` passed and `rtk proxy go vet ./...` passed.
- Web: `npm run build` and the full design contract suite passed; Vite emitted
  only the existing chunk-size advisory.
- Native: base-component/project guards passed; the Release simulator build
  passed with `SWIFT_TREAT_WARNINGS_AS_ERRORS=YES` for simulator UUID
  `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`; the Release `iphoneos` build also
  passed with signing disabled at
  `/tmp/mypocket-ios-release-1.1.0-device-unsigned/MyPocketIOS.app`.
- Production: API and worker both converged on
  `registry.dungxbuif.com/mypocket-api:release-1.1.0-20261015-2212@sha256:bef8040e7a372774234fd96d69e690215dd5d823f4520f901fd5cf1638fb37d2`;
  `/api/v1/health` returned `200` and public Swagger reports version `1.1.0`.
- Owner-device smoke: a Personal Team-signed Release artifact embedding
  `https://money.dungxbuif.com/api/v1` installed and launched on the connected
  iPhone as `com.mypocket.ios` version `1.1.0 (2)`. This is a development
  install only; it does not claim App Store/TestFlight distribution.
- Data safety: no migration was run as part of this release and no financial
  data was deleted or rewritten.

Passing a local build does not imply a signed iPhone install or owner-level
production UAT.

## Known gates

- The current physical install uses the owner's Personal Team development
  profile. Associated Domains are not available to that profile, so universal
  links remain a separate signing gate.
- Owner-authenticated production UAT remains required for the advisor jar
  capability and other configuration-dependent provider routes.
- APNs and provider-configured routes remain environment-dependent.

