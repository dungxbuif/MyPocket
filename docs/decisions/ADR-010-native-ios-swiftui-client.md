---
artifact_type: adr
id: ADR-010
status: accepted_for_scaffold
owner: shared
updated: 2026-10-05
trace:
  detail_design: ../superpowers/specs/2026-10-05-native-ios-swiftui-design.md
  plan: ../superpowers/plans/2026-10-05-native-ios-swiftui-port.md
  architecture: ../architecture/ARCHITECTURE.md
  api: ../architecture/API.md
  auth: ../architecture/AUTH_API.md
  release_notes: ../releases/CHANGELOG.md
---

# ADR-010 — Native SwiftUI iOS client boundary

## Context

MyPocket has a Vite/React web client and an existing Expo/React Native
migration branch. The owner now wants the frontend ported to a native iOS app
and tested in Xcode Simulator before physical-device work. Web DOM and Expo
views cannot be reused as native visual components without creating a second,
unclear styling contract. Native login also needs a mobile OAuth callback and
secure token storage rather than the browser cookie flow.

## Decision

- Add `mobile-ios/` as a SwiftUI/Xcode client on branch `codex/native-ios-port`.
- Keep `app/` web code, the existing web Google callback, and the Expo branch
  unchanged while native parity is built and verified.
- Use a project-owned SwiftUI `DesignSystem` as the only source of user-facing
  controls, cards, navigation, loading/error states and AI/feedback cards.
- Share semantics through API contracts and documentation, not by sharing DOM,
  React Native or CSS components.
- Add a mobile-only Google Authorization Code + PKCE exchange endpoint and
  keep Google secrets on the backend. Store MyPocket JWT/refresh session data in
  Keychain and gate optional Face ID unlock behind a valid session.
- Treat iOS Simulator build/tests as the first release gate. Physical signing,
  TestFlight and App Store distribution are later work.

## Alternatives considered

1. **Continue Expo as the native target** — rejected because the requested
   deliverable is a SwiftUI/Xcode app and Simulator workflow.
2. **Move or delete the web/Expo clients now** — rejected; it increases
   rollback risk before feature parity and would make comparison impossible.
3. **Adopt a third-party SwiftUI design system** — rejected for now; base
   ownership and visual semantics must remain auditable against the existing
   design contract.
4. **Reuse browser OAuth callback** — rejected because it relies on browser
   cookies and redirect URLs that are not a native session boundary.

## Consequences

Positive:

- Native safe-area, navigation, accessibility and Face ID behavior are owned by
  the iOS target.
- Web and native can be verified independently against the same API/business
  contracts.
- A base-component guard can prevent feature screens from drifting into raw
  controls.

Negative:

- SwiftUI is a second renderer, so every feature needs view-model and UI proof.
- A mobile OAuth API/configuration surface must be maintained alongside web
  OAuth.
- Simulator proof does not establish physical-device provisioning or production
  OAuth readiness.

## Linked design and verification

- [Native iOS detail design](../superpowers/specs/2026-10-05-native-ios-swiftui-design.md)
- [Implementation plan](../superpowers/plans/2026-10-05-native-ios-swiftui-port.md)
- [Authentication contract](../architecture/AUTH_API.md)
- [Feedback API](../architecture/FEEDBACK_API.md)
