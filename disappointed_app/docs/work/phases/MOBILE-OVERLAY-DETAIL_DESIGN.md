# Mobile overlay diagnosis

Status: in_review (local automated proof complete; no deployment). Owner authorized resuming the bounded mobile-layout investigation with “tiếp đii” after the loop-guard handoff.

Scope: PWA prompt, shared navigation/sheet layout and controlled browser fixtures. Preserve API, auth, schema, runtime and current visual concept. First measure viewport and element rectangles/hit targets in the failing mobile profile; compare empty and populated lists. Fix only a demonstrated shared layout cause. Reject forced clicks, hiding the banner or deleting accumulated test data as fixes.

Acceptance: account/add buttons remain clickable with visible prompt; prompt expansion and dismissal work; wallet/category controls remain reachable with populated lists. Regression browser checks plus full frontend/full E2E runs; docs reconciled before completion. Physical-device acceptance and production deployment remain separate.

Confirmed cause: report loading expands the implicit content grid to 434px, overflowing the 393px document to 442px. The shared ComparisonBarChart uses non-shrinking min-content flex children, 24px gaps and unbounded money labels. Mobile layout viewport then grows to 818px height while visual viewport remains 727px; fixed click targets shift by 91px. Same failure with 0 and 100 wallets. Removing root scroll padding did not change the failure and was reverted. Fix the shared chart's column/label shrink constraints and equal column alignment, not popup offsets. Preserve full label text through accessible titles; no values changed.

Links: [prior handoff](OFFLINE-RECONCILIATION-DETAIL_DESIGN.md), [audit](../test-verification/BUSINESS-LOGIC-AUDIT-2026-09-10.md), [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md). API/ERD/ADR need no changes unless investigation establishes a contract change.

Verification: full Go race suite with dedicated PostgreSQL and Go vet pass; frontend 20 files / 161 tests and TypeScript/Vite production build pass. Final `npx playwright test --workers=6`: 87/87 pass (36.9s), including all original mobile failures and six controlled layout cases across Chromium desktop/mobile and WebKit mobile. New layout tests use seven fixed historical report values of 123,456,789 VND and 0/100 long-name wallets, assert no horizontal document overflow and use real Account clicks while the install prompt remains visible. Service workers are blocked only for these mocked layout fixtures to ensure WebKit interception; ordinary PWA/offline suites retain real workers.

Docs reconciliation: context, backlog, validation matrix, audit evidence, changelog and public audit update synchronized. No API/schema/ADR/ERD contract changes. Local-only, not physical-device or production acceptance.

Final docs production build passes (nonfatal untracked-file update-date warning); whitespace check passes. Remaining release work: physical-device/provider checks and migration 0012/full resync for the earlier atomic-sync changes, not for this layout-only correction.
