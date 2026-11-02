# Report findings implementation plan

**Goal:** Replace generic saved AI prose with useful, verifiable, scoped findings.
**Architecture:** Canonical queries → typed candidates → provider selection → server-rendered claims → shared native finding cards.
**Tech Stack:** Go/GORM/PostgreSQL, SwiftUI, existing advisor provider.
**Spec:** ../../work/tickets/REPORTS-AI-USEFULNESS-20261102-DETAIL_DESIGN.md

**Execution status:** Implemented locally; owner accepted 2026-11-02 (done). Tasks
1–4 integrated with targeted test, three-review and live provider/evidence/handoff
proof. The overbroad optional Assistant transaction card is a separate backlog
follow-up. The full backend/web suites are not green. See the detail design and
`docs/work/VALIDATION_MATRIX.md`; no release/merge/push is implied.

## Global Constraints

Preserve owner data and unrelated dirty work. Work locally on codex/report-findings-20261102. Preserve existing base styles. No production changes. Implement all slices before one combined test run. Financial claims and targets originate in owner-scoped backend calculations. No fabricated causality. Old payloads remain decodable. Do not invoke subagents from workers.

## Task 1: Canonical budget scope

Own backend monthly_report_postgres.go budgetPreviews, additive budget preview period fields, Assistant budget read tool/schema and targeted tests. Reuse canonical budget period and carryover semantics; overlap interval, effective limit, precise evidence targets; requested budget/month filtering before list display. Coordinate entity/monthly_report.go edits with controller (only ReportBudgetPreview fields). Do not change report_analysis or report_insight files. Report changed paths, choices and tests added; don't run tests yet.

## Task 2: Finding contract and analysis pipeline

Controller owns entity report findings/analysis, provider prompt, analysis service/worker and persistence. Build typed candidates with distinct IDs, exact labels/metrics/ranges and source drilldowns; validate selection/action intent, deduplicate, enforce comparison eligibility; server renders. Version fingerprint; persist provenance and fallback. Add focused regression tests.

## Task 3: Native presentation and follow-up

Own native report DTOs, shared charts/finding bases, Reports feature and Assistant handoff/presentation. Consume additive fields and retain decoding of historical results. Show useful titles/metrics/evidence actions, avoid duplicate saved copy and Insider duplication, preserve snapshot/analysis/month/wallet/finding scope in follow-up. Scope-aware card list overflow. Preserve shared styles, bubble rules and user draft. Add meaningful targeted native tests if contract warrants; no test run yet.

## Task 4: Integrated review and validation

Three independent source reviews covering all changed report behavior. Run affected backend packages and native build once integrated, resolve failures. Rebuild/restart local backend workers preserving env/DB, signed install simulator candidate, verify November overrun 350000 and weekly budget, October eligible comparison, analysis caching/provenance and drilldown/handoff. Reconcile design/context/backlog/validation/changelog. No automatic merge/push/deploy.

## Review Focus

Owner scoping, exact evidence intervals, carryover, immutable snapshot fidelity, comparison eligibility, stale result visibility, safe fallback, historical decoding and handoff identity.
