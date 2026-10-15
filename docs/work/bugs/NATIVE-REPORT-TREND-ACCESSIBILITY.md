---
artifact_type: bug
id: NATIVE-REPORT-TREND-ACCESSIBILITY
status: done
owner: shared
trace:
  page: ../../design/pages/reports/README.md
  bases: ../../design/system/BASE_COMPONENTS.md
  validation: ../VALIDATION_MATRIX.md
---

# Native Reports trend chart must expose its values to assistive technology

## Evidence and intended behavior

The native trend chart groups eligible transaction rows by account-local day
and draws separate income/expense series, but its accessibility contract names
the chart without exposing the per-day totals. The existing API-backed Reports
test only accepts either the chart or its empty state, so a populated chart can
regress to empty without a state assertion.

During the API-backed run, the Reports summary included an expense explicitly
submitted with `included_in_reports=false`. Existing repository query fixtures
work around false-value persistence by directly updating the row after GORM
creates it. Before changing the trend chart, add an integration regression for
the normal transaction repository `Create` path, read the row back from
PostgreSQL, and determine whether the stored flag or report query is incorrect.
If repository persistence is at fault, fix only the backend persistence path
and preserve the default-true behavior for ordinary transactions.

Expose a concise VoiceOver value for each plotted day with income and expense
amounts. The value must honor account masking, preserve chronological order,
and use exactly the chart's existing filtered points; do not calculate a second
financial summary. Add a disposable API-backed iPhone UI scenario with exact
income, expense and excluded amounts that proves the populated chart's exposed
value and absence of excluded rows.

## Technical scope and verification

- Native DesignSystem base plus API-backed UI fixture/test. A backend repository
  persistence correction is in scope only if a PostgreSQL regression proves
  GORM writes the requested false flag as true; this did occur. No API shape,
  schema, migration, owner-scope or endpoint changes were required.
- Use a unique test wallet and future account-local month to isolate the report
  from baseline transactions; clean every fixture after the test.
- RED: PostgreSQL repository regression reproduced persisted `true` for an
  explicitly excluded row. GREEN: both explicit false and true persist; the
  finance-query integration fixture now writes owner transactions through the
  repository, avoiding its prior raw-SQL workaround. The API-backed UI test
  verifies exact accessible chart values and excluded-row persistence.
- Native DesignSystem tests verify account masking; all 144 native unit tests,
  Release warnings-as-errors, project/base guards, and `git diff --check` pass.
- Reconcile this bug, Reports base/page design, validation matrix, port audit,
  backlog and changelog. Clarified the existing report-inclusion behavior in
  the human/agent transaction guide and API index; OpenAPI shape is unchanged.
  Report inclusion is now persisted and honored consistently by create and
  aggregation paths.
