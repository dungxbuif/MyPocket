---
artifact_type: detail_design
id: IOS-RELEASE-ARTIFACT-PREFLIGHT-20261019
status: verified
owner: shared
trace:
  context: ../../CONTEXT.md
  backlog: ../BACKLOG.md
  validation: ../VALIDATION_MATRIX.md
  rollout: ../../operations/ROLLOUT_READINESS.md
  physical_uat: ../../operations/IOS-PHYSICAL-UAT-20261019.md
---

# iOS release-artifact preflight

## Problem

The project guard checks source configuration, but a Release archive or `.app`
can still be inspected with the wrong bundle version, API base URL, orientation
allow-list, or bundle identifier. The connected phone currently remains on the
older `1.1.0 (2)` build, so these checks must be read-only and must not imply a
successful device install.

## Decision

Add a shell preflight that inspects a built app's resolved `Info.plist` and
requires the release contract: bundle identifier `com.mypocket.ios`, version
`1.1.0`, build `2`, production HTTPS API base URL, and portrait-only supported
orientations. It also checks that no tracked secret-like files are introduced.
The script accepts an explicit `APP_PATH`; it never signs, installs, mutates
production, or reads credential values. Source-only validation remains the
responsibility of `verify-project.sh`.

## Alternatives rejected

- Treating an unsigned simulator build as proof of a signed physical release:
  simulator binaries do not prove provisioning or device installation.
- Parsing Xcode build logs instead of the resolved plist: logs can contain
  unresolved settings and do not prove the packaged artifact.
- Automatically invoking `devicectl install`: installation is an external
  signing/device operation and must remain an explicit operator step.

## Acceptance and verification

- Valid Release `.app` passes all checks without network or filesystem writes.
- Wrong bundle/version/build/API URL/orientation fails with a specific message.
- The test fixture covers these failures and never exposes secrets.
- Run the project/base guards, the artifact preflight against the fresh Release
  simulator artifact, and `git diff --check`.

## Reconciliation

Update `ROLLOUT_READINESS.md`, `IOS-PHYSICAL-UAT-20261019.md`,
`VALIDATION_MATRIX.md`, `BACKLOG.md`, `CONTEXT.md`, and `CHANGELOG.md` with the
command and result. Physical biometric/offline/APNs acceptance remains a
separate owner/device gate.

## Current verification — 2026-10-20

The packaged physical-device candidate was rechecked read-only after the
Personal Team install. `codesign --verify --deep --strict` succeeds for the
latest feedback-race source artifact at
`/tmp/mypocket-release-device-latest-1792431904/Build/Products/Release-iphoneos/MyPocketIOS.app`,
and `release-artifact-preflight.sh` reports:

```text
bundle=com.mypocket.ios version=1.1.0 build=2 orientation=portrait
```

`xcrun devicectl device info apps` confirms the same bundle, version and build
on the paired iPhone 15 Plus (`00008120-001C28C02131A01E`). The artifact is
verified as a signed production-API Release install; this does not close the
separate owner-driven Face ID, offline, APNs or full touch-screen acceptance
gates.
