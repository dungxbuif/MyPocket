---
artifact_type: release_notes
id: REL-1.1.0-native-ios-simulator-draft
status: superseded
version: 1.1.0
release_approval: pending
deployment: superseded_by_REL-1.1.0-native-ios
linked_phase: ../work/phases/PHASE-002-native-ios-swiftui.md
linked_plan: ../superpowers/plans/2026-10-05-native-ios-swiftui-port.md
---

# Native iOS SwiftUI parity — superseded draft

This simulator-only draft was superseded by the verified release note
[`REL-1.1.0-native-ios.md`](REL-1.1.0-native-ios.md), which records the
production API/worker deployment, fresh release gates, and the remaining
physical-device signing gate.

## Scope

Adds the first-party SwiftUI iOS client under `mobile-ios/`, using the shared
base-component design system and the existing owner-scoped API. The mounted
finance routes include overview, wallets, transactions, reports, budgets,
categories, recurring schedules, jars, travel, portfolio, credit, AI entry,
Finance Assistant, feedback/changelog and account settings.

## Verification

- 19 unit/contract XCTest cases passed on iPhone 16 Simulator.
- 3 UI smoke tests passed for auth launch, authenticated primary tabs, and the
  More route with the floating feedback entry.
- Stable-ID wallet/category selectors, account-timezone date boundaries, and
  credit statement rendering were covered by the latest parity pass.
- Base-component guard and project/config verification passed.
- Release configuration build passed and resolves the production API URL.

## Known release gates

- An iOS Google OAuth client ID must replace the tracked placeholder before a
  real sign-in build is distributed.
- Provider-backed AI/OCR and authenticated staging/production API smoke are not
  proven by this Simulator-only run.
- No production deployment, App Store/TestFlight signing, or production data
  mutation is claimed by this draft.
