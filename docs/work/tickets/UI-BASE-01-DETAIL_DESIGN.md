---
artifact_type: detail_design
id: UI-BASE-01
status: in_progress
owner: shared
approval: authorized_by_user_do_it
---

# Base UI normalization and screen removal

User approved fixing the reviewed base defects with current design as authority and deleting screens that bypass base components. This is the execution artifact for this bounded UI cleanup; no product phase applies.

## Scope and decisions

- Use the current base-cards export for card radius, shadow, typography and semantic tones. Primary button uses system design #4caf50; emerald action #059669 and switch #10b981 have distinct roles. Preserve original exports as reference; record precedence in the base contract.
- CSS theme is the single primitive token source. Typed global component variants refer to semantic utilities; business mappings live separately.
- Eliminate conflicting variant classes, implement keyboard selection, configurable loading, icon disabled state and stat-card slots.
- Delete OverviewPanel, TransactionsPanel, BudgetsPanel, ReportsPanel, GroupManagementPanel, AccountPanel and QuickAddSheet. Remove obsolete header/navigation/shell. Keep API clients and auth flow; replace screen entry points with an honest base-composed unavailable state, not fake financial data.
- Keep existing routes resolving so bookmarked URLs remain reviewable. No API, DB, OAuth or container changes.
- Verify SSR contracts and browser keyboard/computed style behavior; typecheck. Preserve test fixtures for replay.

## Trace and reconciliation

- [Backlog](../BACKLOG.md), [validation and docs review](../VALIDATION_MATRIX.md), [context](../../CONTEXT.md), [release](../../releases/CHANGELOG.md).
- [Base design](../../design/atoms/base-cards/code.html), [base contract](../../design/system/BASE_COMPONENTS.md), [screen specification](../../design/screens/README.md).
- [Decision](../../decisions/ADR-001-base-design-reconciliation.md).
- Tests: app/tests/base.test.tsx and app/tests/base.html. UAT: owner visual review pending; do not claim product screens implemented.
- Known limits: no full 28-component inventory delivery; retained molecules outside the changed base contracts remain subject to their own design review.

## Completion evidence

Pending implementation and verification.
