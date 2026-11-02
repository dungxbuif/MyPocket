---
artifact_type: audit
id: REPORTS-AI-USEFULNESS-20261102
status: done
owner: shared
---

# Report AI usefulness: runtime review and proposed direction

Implementation update: the approved [detail design](REPORTS-AI-USEFULNESS-20261102-DETAIL_DESIGN.md)
is now integrated and running on the local fixture. The audit observations below
describe the original version. Typed findings, exact evidence actions, canonical
budget periods, saved-result reuse and structured follow-up are implemented;
targeted tests and three independent source reviews passed. Native/runtime review
confirmed the new November findings and the one-budget follow-up card. Owner
acceptance is still open. The optional transaction card in the live follow-up is
still too broad; broader backend/web checks also remain non-green. See the
[current validation matrix](../VALIDATION_MATRIX.md) for evidence and limitations.

Owner supplied a screenshot of a saved analysis with generic observations and
advice, called it unhelpful, and asked for app testing and deeper product analysis
after the rich seed. This artifact records findings and a proposal, not an
implemented replacement or a fully approved design.

Trace: [fixture](LOCAL-RICH-DATA-20261102.md), [backlog](../BACKLOG.md),
[validation](../VALIDATION_MATRIX.md), [context](../../CONTEXT.md),
[report design](REPORTS-INSIDER-VISUAL-20261101-DETAIL_DESIGN.md).

## Runtime evidence

Used the signed local simulator app through Device Hub, existing native-fixture
account and actual backend API. No automated suite was run. After observing
owner activity in the composer, stopped further UI interaction to leave it free
for owner testing. This is a targeted Reports/Insider/Assistant review, not whole-app UAT.

- November report displayed income 48,000,000 VND, expense 12,367,000 VND,
  net 35,633,000 VND and 17 eligible transactions. Account has 1,295 ledger rows
  across the year; these counts have intentionally different scopes.
- Insider budget-overrun action opened exactly one transaction: vacuum purchase
  1,850,000 VND on November 1. Live budget endpoint confirmed limit 1,500,000 VND,
  spent 1,850,000 VND, hence overrun 350,000 VND (23.3%).
- Previous-month selection loaded October totals: income 122,960,000 VND,
  expense 44,197,000 VND, net 78,763,000 VND, 137 eligible transactions;
  category chart and matched-30-day contributor chart were populated.
- Saved November AI repeats “Quan sát từ dữ liệu”, generic increase/decrease
  statements, and three non-specific suggestions. Persisted current analysis
  `01f42926-3f9c-4a73-a856-7c4ca3305d86` completed; latest-analysis API restores it.
  Provider usage records 3,202 tokens and 10,811 ms for this generic result.
- “Hỏi tiếp” opened full-screen Assistant with navigation/bubble absent, but
  prefilled only “Tôi muốn hỏi tiếp về bài phân tích báo cáo tháng 2026-10.”
  Its actual response re-summarized October and asked what to explore.
- A subsequent actual UI message asked which November shopping budget purchase
  caused the overrun. Run `be1e02ea-8172-42e5-aa06-1d5338c0e653` completed in
  8.64 seconds and correctly answered limit/spent/overrun/date/purchase.
  No financial write was requested or approved.
- That answer also returned a `budget_status` containing **all 27 budgets**.
  Native `visibleItems` shows its first eight: January–April historical budgets,
  omitting the November target. The accompanying transaction list had one correct
  transaction. Specific prose and unrelated card scope coexist in the same answer.
- Live `/budgets` includes the weekly food budget for November 2–8 (limit
  1,300,000 VND, spent 30,000 VND), while `/reports/monthly?month=2026-11`
  omits that budget from `budgets`.

## Confirmed causes

1. `backend/internal/infrastructure/ai/report_analysis.go` explicitly prohibits
   numbers, percentages, transaction names, notes and causes in text.
   `entity/report_analysis.go` rejects provider text containing ASCII digits.
2. `BuildReportFactBundle` drops names, budget limits, comparison eligibility,
   matched-day counts and exact windows. Its sequential 30-fact limit can let
   contributors crowd out relevant budgets/jars.
3. `usecase/report_analysis.go` hardcodes the same observation title. It attempts
   to reuse deterministic insight copy only for exact refs; field refs such as
   `budget:<id>.spent` differ from the budget-level deterministic ref. Observation
   drilldowns are report-wide, not specific evidence targets.
4. `MPReportAnalysisArticle` repeats those titles, adds duplicate saved-state copy,
   renders generic bullet advice and offers no per-finding evidence action.
   `ReportAnalysisView` hands off only the month, losing wallet/snapshot/analysis
   and selected finding identity.
5. `monthly_report_postgres.go:budgetPreviews` selects budgets covering the first
   day, rather than overlapping the relevant period, and uses month-window spending
   with base `LimitAmount`. Weekly omission is runtime-confirmed; incorrect
   carryover/partial-period accounting is a source-confirmed risk needing its own fixture.
6. November comparison is explicitly `comparable=false`, `common_days=1`.
   Current totals include November 2, whereas comparison is November 1 versus
   October 1. AI receives deltas without eligibility and describes a trend anyway.
   `net` is income minus expense, so “Số dư ròng giảm” mislabels that quantity.
7. Report analysis completion persists observations/suggestions/usage but omits
   `provider_called`, `summary`, `error_code`. Runtime shows false provider flag
   with actual usage; source worker sets true before completion. Saved provenance
   and fallback explanations are therefore unreliable.

Independent read-only agents confirmed the prompt/fact/rendering problems and
budget-period semantics. Runtime checks above were performed separately by the
main agent; source-only findings are not presented as executed regressions.

## Proposed product contract

Every visible finding must answer a specific question with a named subject,
server-calculated evidence, explicit period and an action opening that evidence.
Do not fill a quota with “continue monitoring” advice. Display up to three
material, distinct findings; a short insufficient-evidence state is acceptable.

Example from verified fixture:

> **Mua sắm vượt 350.000đ**
>
> Tháng 11: đã chi 1.850.000đ / hạn mức 1.500.000đ.
> Khoản máy hút bụi ngày 01/11 chiếm toàn bộ chi tiêu của ngân sách này.
>
> Xem giao dịch · Xem ngân sách

For early-month reports, say exactly which dates are included and avoid monthly
trend claims until comparison eligibility is met. Larger recorded purchases can
still be surfaced without claiming that they are wasteful, recurring, or motivated
by a behavior not recorded in the ledger.

Insider should be a quick summary; saved analysis should add scoped explanation
and useful options. Both should share finding identities and avoid duplicating
the same fact in multiple sections. Follow-up carries the selected finding and
scope, with natural-language UI labels rather than internal identifiers.

## Proposed technical sequence

1. Reuse canonical budget period/carryover calculations in report previews; fix
   report/Assistant entity and period filtering; carry comparison quality through.
2. Build grouped, ranked evidence with reserved capacity for budgets, contributors,
   large purchases and goals/jars where relevant. Names and values come from typed
   owner-scoped facts, with bounded explicit evidence targets.
3. Model chooses supported finding IDs and explanatory/action intents. Backend
   validates direction, period, eligibility, amounts, referenced entities and
   deduplication, and renders exact financial claims. Merely removing the digit
   prohibition would not establish accuracy.
4. Shared native finding card: specific heading, two or three useful figures,
   concise explanation, precise evidence action. Filter relevant items before
   truncating card lists; expose any remaining relevant items explicitly.
5. Persist provenance/summary/fallback state correctly. Version the changed
   contract so saved old analyses remain readable but become eligible for refresh.
   Open saved results without automatically calling the model again.

Acceptance examples: the 350,000 VND overrun and transaction match; weekly budget
appears with the correct interval; November 2 cannot produce an unsupported trend;
October comparisons state first-30-day scope; Assistant card contains the November
budget it discusses; follow-up preserves wallet/finding; cached articles restore
with truthful provenance and stale state. Implementation is now approved by the
owner's “làm luôn đi”; see the [detail design](REPORTS-AI-USEFULNESS-20261102-DETAIL_DESIGN.md)
and [execution plan](../../superpowers/plans/2026-11-02-report-findings.md).
Final owner acceptance remains open.
