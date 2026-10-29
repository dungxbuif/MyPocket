# MyPocket 1.3.0 release candidate

Date: 2026-10-27
Native build: `1.3.0 (5)`
Status: installed release candidate; owner acceptance pending

## Included

- Native Money Lover/REA reference redesign across the shell, daily ledger,
  wallets, organization, planning/reports and MyPocket-only extensions.
- Native transaction search within the selected wallet and period.
- Feedback-driven fixes for transaction presentation, Feedback dock,
  notification duplication, transaction entry, Portfolio, Face ID, API-key
  handling and AI proposal review.
- One-time API-key privacy hardening: no ordinary text-selection copy,
  local-only expiring clipboard and automatic background remasking.
- Review-only structured XLSX ingestion candidate with deterministic ownership
  of date, amount, direction, currency and source identity.
- October owner-data reconciliation evidence. Production data changes are
  operational records, not part of the application binary.

## Verified

- Signed native unit target: 422/422, zero failures or skips.
- API-backed one-time secret privacy regression: 1/1.
- Native Release simulator build with warnings treated as errors.
- Backend: 729 tests; focused race suite: 636; `go vet` clean.
- Private XLSX truth set: 118/118 rows and source identities.
- Web design guard, 7/7 design tests, typecheck and production build.
- Physical iPhone 15 Plus Release build, signature and entitlements verified.
- `1.3.0 (5)` installed and launched successfully with bundle identifier
  `com.mypocket.ios` and production API
  `https://money.dungxbuif.com/api/v1`.

## Boundaries

- PDF/OCR remains review-only. A private gate returned 33 proposals for 31
  October rows; prompt tightening did not restore source/date parity.
- XLSX/backend changes are not production behavior until the API and worker are
  separately deployed and smoke-tested.
- Installing build 5 does not resolve production Feedback records. Resolution
  requires owner acceptance on the physical device.
- Face ID, offline behavior and populated-screen owner acceptance remain
  explicit follow-up checks. Installation and launch are complete.

## Device evidence

- Build artifact:
  `/tmp/mypocket-1.3.0-build5-device-20261027/Build/Products/Release-iphoneos/MyPocketIOS.app`
- Signing team: `AU442S3MN3`; designated requirement and strict deep signature
  verification passed.
- Personal-build entitlements contain the application/team identifiers and
  development debugging entitlement only; no associated-domain entitlement.
- Installed on device `00008120-001C28C02131A01E`; `devicectl` reported a
  successful installation and process launch on 2026-10-27.

## Rollback

Reinstall the previous signed `1.2.1 (4)` application artifact. No schema
migration is required for the native rollback. Backend production remains on
its current release unless a separate controlled rollout is approved.
