# R0 — Stabilize existing functions without redesign

Status: in_progress. Approval: user approved the functionality-first plan and said “Làm” on 2026-09-09. This artifact records that approved direction; it does not promote deferred subsystems.

## Scope and trace

Sources: [audit](../../research/moneylover/PARITY-AUDIT-2026-09-08.md), [prior evidence](../test-verification/MONEYLOVER-PARITY-2026-09-08.md), [backlog](../BACKLOG.md), [requirements](../../requirements/SPEC.md), [validation](../VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md).

Keep the current shell, colors, typography and card style. Work directly in the existing dirty checkout, preserving pre-existing edits; do not stage unrelated work. No production changes, API/schema/auth changes, new dependencies or bulk component migration.

## First independently verifiable slice

1. PWA prompt: preserve bottom-corner installation help; add dismiss and installed-state handling. Reserve measured overlay space in page/scroll layout so the last action can scroll above prompt and dock. Reuse base buttons; isolate prompt lifecycle from App. Existing mobile logout test is the regression test, without forced clicks or dismissing the prompt to conceal obstruction.
2. Budgets: render the persisted budget name plus category context; use a native-button base card for keyboard/offline semantics. Pass real zero/remaining values to the existing gauge; do not substitute design fixtures. Remove the misleading sample-data notice. Empty/loading states must not invent balances or remaining days.
3. Audit currently mounted actions, record working/unsupported/unreachable operations and their next slices. Do not mount prototype screens with fake data simply to make a button respond.

## Root cause and alternatives

PWA: fixed prompt and dock overlay normal content; prompt has no dismissal or standalone gate and page bottom padding covers only the dock. Prior live mobile E2E records pointer interception. Test measured scroll clearance, rather than hiding logout or using force-click.

Budgets: API supplies `budget.name`, but row renders only joined category names. The gauge uses positive-number conditionals that substitute 65M/19.92M/45.075M for legitimate zero values. Existing CRUD test cannot reach edit/archive. Fix rendering rather than changing test selectors to ignore the missing name.

Rejected: full redesign; duplicating financial APIs; changing the accepted 80% threshold; silently considering disabled future features complete.

## Implementation sequence

- [x] Re-run live auth and budget E2E on isolated disposable PostgreSQL; record red.
- [x] Add component tests for persisted name, keyboard activation, offline disable, zero/unspent/exhausted budgets and PWA dismiss/installed lifecycle; record red.
- [x] Implement shared prompt/action-card behavior and minimal App/Budgets wiring.
- [x] Run unit, typecheck/build and live browser regressions; inspect mobile output. WebKit offline-reload failure remains recorded separately.
- [x] Record [mounted-action inventory](../test-verification/R0-MOUNTED-ACTIONS-2026-09-09.md) and [verification](../test-verification/R0-FUNCTIONAL-STABILITY-2026-09-09.md); first slice is in review, full R0 remains in progress.

## Verification and reconciliation

Use `rtk` command prefix. Frontend Node PATH must include `/Users/dungxbuif/.nvm/versions/node/v24.0.2/bin`. PostgreSQL is temporary container `mypocket-r0-20260909-pg`, loopback port 64739, database `mypocket_r0_e2e`; never use the production database. Opt-in `frontend/playwright.r0.config.ts` derives from the regular config, replaces only test DB and web port, keeps API port 18173 required by existing isolation test, and disables server reuse.

Acceptance: auth login/reload/logout works with install prompt visible; budget create/edit/archive works; zero balances stay zero; keyboard activates the correct budget and offline cannot edit. Human production UAT remains separate.

API, ERD, architecture and ADR: no changes required for this frontend-only slice. Update context/backlog/validation/changelog and attach verification here. Later R0 actions need their own bounded implementation/test slices; this document is not certification that all R0 actions already work.
