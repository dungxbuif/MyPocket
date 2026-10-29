---
artifact_type: detail_design
id: UI-BASE-01
status: done
owner: shared
approval: authorized_by_user_normalize_guardrail_refactor_2026_09_13
---

# Base UI normalization and design enforcement

Owner explicitly requested normalization, guardrails and refactoring in this conversation (FB-002). This revision supersedes the historical screen-removal proposal: preserve current routes/features, normalize shared primitives. Extract seven image/HTML exports into specifications, then remove redundant exports. Record screen-specific behavior as screens are implemented.

## Root cause and approach

styles.css did not import ui/theme.css; atom tokens.ts duplicated primitive colors; screens and molecules recreated controls or overrode base visuals. No build check enforced base ownership. Source inspection and initial failing guard reproduce this.

Use one theme, named variants, domain tone mappings, base-owned controls/text/cards/progress and composition at screen level. Keep the recently reduced 12px corners, explicit circle/pill exceptions and named brand/action/accent roles. Preserve route/API/auth/data behavior. Read gateway/context/backlog/standards, current UI artifact, component source, seven screenshots and relevant architecture/product contracts.

## Execution plan completed

- [x] Add AST guardrail and negative fixtures; run current source to record red.
- [x] Import theme, migrate literal colors, consolidate variants and compatibility re-export.
- [x] Refactor controls into shared bases; use wallet FormField/input/select/checkbox; named inline inputs, disabled/focus and keyboard tab behavior; avoid nested interactive link/button.
- [x] Compose shared cards/text/status/navigation/progress/gauge/charts; sheet owns focus/close/scroll behavior.
- [x] Replace export-driven docs with tokens, behavior, per-component specs, source extraction and honest partial/planned inventory. Delete exactly the extracted exports.
- [x] Verify source/docs guard, regression tests, SSR contracts, typecheck/build and browser fixture; reconcile standards/ADR/context/backlog/matrix/changelog.

## Scope and alternatives

Touched: app/src, app/scripts, app/tests, docs/design and directly related workflow/master docs. No new dependency. No API, database, security, auth or deployment change. Existing incomplete mock/keypad/chart interaction behavior remains explicitly partial rather than being presented as product completion.
Docs-only enforcement was previously bypassed; rejected. Deleting working screens was superseded by the latest owner direction. A new component-library dependency was unnecessary.

## Trace and reconciliation

[Feedback FB-002](../FEEDBACK_LOG.md) · [Backlog](../BACKLOG.md) · [Design gateway](../../design/README.md) · [Architecture](../../architecture/ARCHITECTURE.md) · [ADR-002](../../decisions/ADR-002-design-contract-enforcement.md) · [Verification / tests / docs review](UI-BASE-01-VERIFICATION.md) · [Validation matrix](../VALIDATION_MATRIX.md) · [Context](../../CONTEXT.md) · [Release](../../releases/CHANGELOG.md).
Phase/roadmap: no product phase; framework maintenance. SDD absent; architecture updated. Requirements/API/ERD require no change because product contracts are preserved. Shared browser acceptance and exact commands are in verification; no new product UAT is claimed.
