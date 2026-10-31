---
artifact_type: release_note
id: REL-1.3.0-20261101
status: release-candidate
version: 1.3.0
build: 7
date: 2026-11-01
deployment: pending
---

# MyPocket 1.3.0 build 7

Status: release candidate. Production image, migration readback, service
convergence and health evidence are filled after the approved rollout. The
native artifact is a simulator Release build installed on the owner's existing
iPhone 16 simulator; no physical-device install is claimed here.

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

Filled after rollout:

- Source commit: pending
- Image: pending
- Digest: pending
- Backup and SHA-256: pending
- Migration version/dirty state: pending
- API/worker convergence and health: pending

## Known gates

- APNs live delivery is not enabled or claimed; `APNS_ENABLED=false` remains
  the safe default.
- Physical-device, owner visual acceptance, Dynamic Type and provider freshness
  remain separate UAT gates.
- Rollback: restore the previous API/worker immutable image. Native rollback is
  reinstalling the previous signed artifact; no destructive database reset is
  part of this release.

## Traceability

- [Changelog](CHANGELOG.md)
- [Validation matrix](../work/VALIDATION_MATRIX.md)
- [Implementation reconciliation](../work/audits/IMPLEMENTATION-RECONCILIATION-20261101.md)
- [AI diagnostics/APNs plan](../superpowers/plans/2026-11-01-ai-attachment-diagnostics-apns.md)
