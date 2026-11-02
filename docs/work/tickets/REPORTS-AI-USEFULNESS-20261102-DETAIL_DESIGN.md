---
artifact_type: detail_design
status: done
owner: shared
---

# Evidence-backed report findings

Approved by owner: “làm luôn đi”, following the [runtime audit](REPORTS-AI-USEFULNESS-20261102-AUDIT.md). Execute locally; preserve rich fixture and current shared base styling. No production deployment.

## Contract

- Canonical budget periods and carryover determine report budget spending, effective limit and remaining. Include overlapping weekly/midmonth budgets; evidence must use the same interval, wallet and category as the calculation.
- Backend builds a bounded, ranked set of typed finding candidates from immutable report evidence. Provider selects candidate IDs and supported action intent, never authors financial facts. Distinct meaningful findings, maximum three, no filler quota. Early-month non-comparable data never becomes a trend claim.
- Preserve public observations/suggestions wire fields. Add optional evidence metrics/period/action fields to ReportInsight, consumed by shared native finding base. Names/amounts/ranges and target IDs are server-owned. Existing saved payloads remain readable.
- Finding kinds: budget overrun, eligible matched-period category change, category concentration, cashflow deficit. Reuse source targets, no invented transaction causality. Reserve candidate diversity rather than truncating flat facts. Larger-purchase attribution requires actual transaction evidence and is not inferred from category totals.
- AI may rank findings and choose an allowed next action. Backend renders the statement and bounded action copy. Invalid, duplicate or ineligible selection falls back to deterministic candidates with truthful provenance. No-data and insufficient-evidence results are explicit.
- Persist summary, provider_called and error_code along with findings. Bump rule/prompt versions; old saved analyses become stale and remain visible. Opening never auto-generates.
- Shared MPReportInsightCard/MPReportAnalysisArticle render specific title, metrics, period, explanation and evidence action; same base styles. Insider stays brief; detailed findings own full explanation and actions. Follow-up carries report month/wallet/timezone, analysis/snapshot/finding identity and selected evidence, without displaying raw identifiers.
- Assistant budget read tool filters requested budget/month before presentation; unrelated historical items cannot crowd out requested budget. Remaining relevant items are disclosed.

## Files and verification

Backend: entity/report_analysis.go, entity/report_insight.go, entity/monthly_report.go; usecase/report_analysis.go; infrastructure/ai/report_analysis.go; repository/report_analysis_postgres.go, monthly_report_postgres.go; Assistant budget tool and schemas.
Native: Core report models, DesignSystem/MPReportCharts.swift, Features/Reports, Assistant handoff path/card list.

Write targeted regressions for candidate eligibility, invalid provider selection, persistence, budget periods/carryover and filter scope. Run one integrated verification after implementation (owner preference), then signed simulator build and live local checks with seeded November/October data. Three independent final reviews. Do not claim full-app UAT.

Reference mapping: screenshot repeated “Quan sát từ dữ liệu” → named finding title; generic advice bullets → specific action per finding; duplicate saved labels → one status; unscoped Hỏi tiếp → scoped handoff. Shared bases: MPSurfaceCard, MPButton, MPText/MPFont/MPSpacing tokens, MPReportInsightCard, MPReportAnalysisArticle. No screen-local visual tokens.

Trace: [backlog](../BACKLOG.md), [validation](../VALIDATION_MATRIX.md), [context](../../CONTEXT.md), [base contract](../../design/system/BASE_COMPONENTS.md), [changelog](../../releases/CHANGELOG.md). Existing unrelated draft-safety edits remain intact. API changes additive; no new database columns expected, verify existing schema. No new library/model dependency.

## Implementation and validation — 2026-11-02

All four implementation slices are integrated locally. Three independent source
reviews accepted the final candidate after correcting candidate diversity,
same-day fingerprint stability, historical suggestion follow-up and matched-period
endpoints. Native selected suites passed 123 tests, then five focused report
analysis tests after the neutral handoff prompt correction. Targeted backend
report suites passed. Broader backend/web failures remain separately recorded in
the validation matrix; this is not a release approval.

Live November analysis calls the real provider and identifies the 350,000 VND
shopping overrun. October identifies a 5,544,000 VND shopping overrun and compares
3,905,000 versus 6,935,000 VND food spending across 30 matched days. Both saved
analyses survive day/week changes without another provider call. No new migration
is required. Owner visual acceptance remains open.

Runtime UAT exposed two additional bugs within these flows:

- Weekly aggregation of an unfinished day never advanced its final day index.
  A two-second regression timeout reproduced the loop in `seriesForRange`.
  Stop when the emitted bucket reaches the exact report endpoint. Repository
  `TestMonthly*` tests pass, including partial first/final day and midnight
  boundaries; real weekly requests now complete in 35–40 ms in this local sample.
- Empty retained Assistant history returned `data: null`, breaking native array
  decoding. Normalize successful empty message results to `[]` at the HTTP
  boundary, preserving errors, ownership and metadata. A failing HTTP regression
  reproduced null; empty-history and older-proposal tests pass after the fix.

The host restarted during UAT, clearing temporary build/log files. Existing DB
and installed signed app survived; the local API/worker were rebuilt and restarted
against the same fixture. Earlier test results are recorded observations; temporary
xcresult/log artifacts from before the restart are no longer available.

Docs reconciliation reviewed: requirements, additive API and tool contracts,
report/base specifications, context, backlog, validation matrix and unreleased
changelog now reflect the implemented candidate and its limits. No schema or
deployment changes were introduced. Native runtime follow-up completed with the
selected finding and one matching budget card; an additional broadly queried
transaction card remains an explicit acceptance gap in the backlog. The simulator
is left on the new saved findings for owner review. This status is `in_review`,
not a full-app or production sign-off.

## Owner acceptance — 2026-11-02

Owner accepted the saved November findings and asked to close this slice. Status
is `done` for the local slice. Acceptance/UAT: owner accepted (targeted report
findings, drilldown and follow-up flows; not full-app UAT). Residual items move to
separate backlog follow-ups: overbroad optional Assistant transaction card and
pre-existing full backend/web suite failures. No merge, push, migration or
production deployment is implied.
