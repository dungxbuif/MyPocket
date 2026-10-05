---
artifact_type: phase
id: PHASE-002-native-ios-swiftui
status: in_progress
owner: shared
created: 2026-10-05
updated: 2026-10-05
trace:
  backlog: ../BACKLOG.md
  roadmap: ../ROADMAP.md
  detail_design: ../../superpowers/specs/2026-10-05-native-ios-swiftui-design.md
  plan: ../../superpowers/plans/2026-10-05-native-ios-swiftui-port.md
  adr: ../../decisions/ADR-010-native-ios-swiftui-client.md
  validation: ../VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# Phase 002 — Native iOS SwiftUI parity

## Goal

Deliver a first-party SwiftUI renderer for the currently mounted MyPocket web
features, with base-component ownership and an iOS Simulator-first proof.

## Scope

- New `mobile-ios/` target on `codex/native-ios-port`.
- Typed API/auth core, mobile Google PKCE exchange, Keychain session and Face
  ID unlock.
- Overview, ledger, wallet, budget, report, jar/month, category, recurring,
  travel, portfolio, AI entry, Finance Assistant, feedback and account flows.
- XCTest, API contract tests, static base-component guard and Simulator UI
  smoke proof.

## Out of scope

- Deleting/moving web or Expo code.
- App Store/TestFlight release, physical signing, push/background sync and
  unapproved offline ledger writes.
- New database/finance rules unrelated to the mobile auth boundary.

## Entry and exit gates

- Entry: detail design and implementation plan are approved; isolated worktree
  is created; web branch remains clean; Xcode SDK is available.
- Exit: every parity row has a native screen/view model, base guard passes,
  backend/frontend regression suites pass, and a Simulator run is recorded. If
  no Simulator runtime is installed, phase remains `blocked` with the exact
  Xcode component prerequisite rather than being marked complete.

## Verification checkpoint — 2026-10-05

The simulator prerequisite is now available locally: iOS 18.4/18.5 runtimes
were detected and an iPhone 16 device was created as
`EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`. The current auth/base-component slice
was built, installed, launched, and screenshot-checked on that device.

Passing commands:

```bash
xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS \
  -destination 'id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' \
  test CODE_SIGNING_ALLOWED=NO
bash mobile-ios/scripts/check-base-components.test.sh
bash mobile-ios/scripts/verify-project.sh
go test ./...
go generate ./cmd/api
```

This is a checkpoint, not the phase exit gate: the feature-parity screens and
their API-backed view models still need to be implemented before claiming the
native port is complete.
