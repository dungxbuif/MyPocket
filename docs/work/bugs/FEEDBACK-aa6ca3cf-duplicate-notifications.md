---
artifact_type: bug
id: FEEDBACK-aa6ca3cf
status: in_review
owner: native-notifications
updated: 2026-10-26
---

# Duplicate notification entry

## Source and scope

Production read-only `GET /api/v1/agent/feedback?limit=100` on `vm100`, using
the runtime service token without printing or storing it, returned
`aa6ca3cf-2fe4-4eb3-b9ae-39b63e4f2edd`, title `Thông báo`, description
`Thông báo có ở trang chủ rồi. Bên tài khoản lại có tiếp`, `status=open`,
`resolved=false`. No production state was changed. This concerns duplicate
native navigation entry points, not duplicate notification delivery.

## Pre-change design and reproduction

- Intended behavior: the Overview bell is the single visible entry to the
  notification inbox. Account/More must not offer a second notification row.
- Current flow: Overview's `onOpenNotifications` selects Account with an
  internal `.notifications` destination. `MoreReferencePresentation.sections`
  omits notifications; `SecondaryScreen` retains the internal route for the bell
  and deep links. `NotificationsView` still owns inbox loading and read state.
- Reused bases: `MPCenteredScreenHeader`, `MPHeaderIconButton`, `MPMenuRow`,
  `MPSurfaceCard`, and shell navigation. No new base or API/schema/config change.
- Historical reproduction: the 2026-10-14 native UI audit recorded both
  Overview bell and More row, then a RED-to-GREEN duplicate-route UI check after
  removing the More row. Current focused UI tests are being rerun now.
- Verification: run the More no-duplicate and Overview-bell navigation UI tests;
  inspect current sections and route; run design checks/build if a screen spec
  changes. Physical owner UAT remains separate from simulator evidence.
- Docs: reconcile `docs/design/pages/notifications/README.md`, whose pending
  removal statement is stale, and record results here and in the validation
  matrix. No public API or setup documentation changes are expected.

## Post-change reconciliation

No native code change was necessary. The current Account sections already omit
the visible notification row, while the Overview bell and internal destination
still open `NotificationsView`. The first two-test simulator run passed the
Overview bell case; the More case was killed during app launch, before any
assertion. Isolated rerun of the More case passed 1/1, zero skips, on iPhone 16
Simulator (iOS 18.5):
`/tmp/mypocket-feedback-aa6ca3cf-retry.xcresult`. The combined first run is
`/tmp/mypocket-feedback-aa6ca3cf.xcresult` and is not claimed as 2/2 green.

The notification screen spec was corrected from "queued for removal" to the
current single-entry behavior. No API, schema, config, push delivery, production
data, or deployment changed. This is locally verified only; the unresolved
production feedback remains open for owner acceptance on the affected physical
build. Public API/setup docs and release notes do not change because this pass
ships no new user-facing behavior. No ADR is needed for an existing decision.

Post-documentation checks passed: `mobile-ios/scripts/check-base-components.test.sh`,
`mobile-ios/scripts/verify-project.sh`, `git diff --check`, and `app/` commands
`npm run check:design`, `npm run test:design`, and `npm run build`. The build
reported only its existing large-chunk advisory. Human screen guidance and the
plain-Markdown agent-readable screen spec are the same file; no separate public
reference surface is affected.

## Trace

[Backlog](../BACKLOG.md) · [Prior audit](../tickets/NATIVE-UI-SYSTEM-AUDIT-20261014.md)
· [Validation](../VALIDATION_MATRIX.md) · [Screen spec](../../design/pages/notifications/README.md)
· [Production UAT audit](../audits/FEEDBACK-UNRESOLVED-UAT-20261020.md)
