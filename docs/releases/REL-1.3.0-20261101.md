---
artifact_type: release_note
id: REL-1.3.0-20261101
status: done
version: 1.3.0
build: 7
date: 2026-11-01
deployment: api_worker_production_and_simulator
---

# MyPocket 1.3.0 build 7

Status: released. The production API/worker rollout and native simulator
installation completed under the owner's explicit release instruction. The
native artifact is installed on the owner's existing iPhone 16 simulator; no
physical-device install is claimed here.

## Included

- Unified full-screen Assistant with larger composer, typed-entry batch review,
  report/asset cards and attachment diagnostics.
- Shared native typography/density bases while retaining the approved light
  cards, colors and floating Feedback bubble.
- Reports/Insider summaries and drilldown state, plus accumulated-asset units,
  price-source selection and edit-friendly trade presentation.
- Bundled Phosphor icon catalog with semantic role fallback and license notice.
- Privacy-safe per-file AI attachment diagnostics in process/recovery responses
  and native review UI.
- Real APNs token-authenticated transport and durable delivery worker queue;
  APNs is disabled by default until Apple secrets, entitlements and device UAT
  are available.

## Verification

- Backend: `rtk proxy go test ./... -count=1` — PASS.
- Native signed unit target: `442/442` with zero failures, skips or runtime
  warnings; result bundle `/tmp/mypocket-release-unit-signed-20261101.xcresult`.
- Native Release simulator build with Swift warnings treated as errors — PASS.
- Artifact preflight — PASS for `com.mypocket.ios`, `1.3.0 (7)`, portrait-only
  and `https://money.dungxbuif.com/api/v1`.
- Project, shared-base, icon-catalog and diff guards — PASS.
- Installed and launched on `MyPocket-iPhone-16` simulator
  (`EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`).
- Aggregate UI scheme run is not claimed: the simulator UI runner did not
  produce a completion result after more than eleven minutes and was
  interrupted. Focused visual/device acceptance remains open.

## Deployment record

- Source commit: `fe5344fc053a1c26ea7283b96174414290861b59`
- Image: `registry.dungxbuif.com/mypocket-api:release-1.3.0-build7-fe5344fc`
- Immutable digest: `sha256:f0bf37e65759d8d6fef61389b93acdb1e7154151ca1966f2d4d011da3cdf7a44`
- Platform: `linux/amd64`, API and worker use the same digest
- Backup: `/var/backups/mypocket/release-1.3.0-build7-20261031T214222Z/pre-release.dump`
- Backup SHA-256: `48be77ce25b2169927b242fe73ab48dd5e165fd264b0b2da3c7e7a57df6f4cc4`
- Migration readback/apply/readback: `version: 52 dirty: false`; `up` was a
  clean no-op because the target was already current
- Services: `mypocket-api 1/1`, `mypocket-worker 1/1`
- Health: `https://money.dungxbuif.com/api/v1/health` returned HTTP `200`,
  `data.status=ok`
- Read-only OpenAPI parity: `96 live / 96 local`
- Worker post-rollout scan: no panic, fatal, missing schema or migration-dirty
  pattern in the observed five-minute log window
- Native install/launch: `MyPocket-iPhone-16`
  (`EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`), `xcrun simctl launch` returned PID
  `89323`

## Known gates

- APNs live delivery is not enabled or claimed; `APNS_ENABLED=false` remains
  the safe default.
- Physical-device, owner visual acceptance, Dynamic Type and provider freshness
  remain separate UAT gates.
- The aggregate native UI scheme run was interrupted after the simulator UI
  runner produced no completion result for more than eleven minutes. Signed
  native unit coverage is complete at 442/442; this is a remaining UI-runner
  environment gate, not a reported product assertion failure.
- Rollback: restore the previous API/worker immutable image. Native rollback is
  reinstalling the previous signed artifact; no destructive database reset is
  part of this release.

## Traceability

- [Changelog](CHANGELOG.md)
- [Validation matrix](../work/VALIDATION_MATRIX.md)
- [Implementation reconciliation](../work/audits/IMPLEMENTATION-RECONCILIATION-20261101.md)
- [AI diagnostics/APNs plan](../superpowers/plans/2026-11-01-ai-attachment-diagnostics-apns.md)
