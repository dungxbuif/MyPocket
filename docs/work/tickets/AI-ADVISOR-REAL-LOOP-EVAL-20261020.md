# Real advisor loop evaluation — 2026-10-20

Status: in_progress. Owner authorized parallel AI improvements.

## Expanded synthetic quality harness — 2026-10-20

The narrow three-case live loop is now supplemented by the dedicated
[quality-harness design](AI-ADVISOR-QUALITY-HARNESS-20261020-DETAIL_DESIGN.md),
the machine-readable [`evals/finance-assistant/cases.json`](../../../evals/finance-assistant/cases.json)
pack, and `advisor_quality_eval_test.go`. The deterministic suite uses the
real `AdvisorOrchestrator`, registry and structured-part builder with a
synthetic `FinanceReader`; it never reads production data or confirms writes.

It covers September/month timezone normalization (including Los Angeles DST),
multi-turn summary→transaction follow-up, missing data, unsupported banking,
prompt injection/owner override, and review-only transaction proposals. Exact
owner, UTC bounds, synthetic amount, result status and card type are asserted.

Verification:

```text
go test ./internal/infrastructure/ai -run 'TestAdvisorQuality|TestAdvisorFixturePack' -count=1
go test -race ./internal/infrastructure/ai -count=1
go test ./... -count=1
```

All passed locally on 2026-10-20. The opt-in `TestLiveAdvisorQualityMatrix`
uses the configured provider with the same synthetic reader and remains
excluded from default tests; provider availability and model routing are still
external quality gates. No API/schema/provider/deployment change is implied.

The existing live tool probe forces a tool and does not execute the app's
orchestrator. Add an opt-in synthetic evaluation through `AdvisorClient`,
`AdvisorOrchestrator`, the real registry and structured part builder. Cases
cover app navigation and creation proposals. The provider chooses tools using
the production prompt/schema; tests do not force a tool.

No database is supplied. A recording provider rejects any tool outside each
case's allowed, non-mutating help/proposal tools before dispatch. This evaluation
cannot read real financial data or confirm an action. It uses the existing
backend environment credentials, never prints them, and has a per-case deadline.

Assert relevant tool execution, proposal fields, nonempty answer, and structured
card type. Unexpected tool requests and provider failures fail the test.
Keep broader financial grounding, owner isolation and multi-turn scenarios as
explicit remaining evaluation work.

Files: `backend/internal/infrastructure/ai/advisor_loop_live_test.go`.
Verification: deterministic provider trace-guard test plus explicit
`AI_ADVISOR_LOOP_LIVE=1` opt-in suite; default runs skip live requests.
No API/schema changes. Public guide need not change for this test-only slice;
this Markdown is the human/agent-readable execution record.

## Evidence

`AI_ADVISOR_LOOP_LIVE=1 AI_MODEL=DeepSeek-V4.1-Flash go test
./internal/infrastructure/ai -run
'^(TestAdvisorEvalRejectsUnexpectedToolBeforeDispatch|TestLiveAdvisorApplicationLoop)$'
-count=1 -v` passed on 2026-10-20.

The deterministic dispatch guard passed. All three live cases passed with
one relevant tool call each and the expected structured card: navigation
8.83 seconds, jar 1.63 seconds, wallet 1.39 seconds. Six model requests were
made in total. No database, user financial data or confirmation API was used.
This is narrow routing/argument/card proof, not comprehensive answer-grounding
or multi-turn quality acceptance.
