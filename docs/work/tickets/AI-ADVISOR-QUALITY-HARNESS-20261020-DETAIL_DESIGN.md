---
artifact_type: detail_design
id: AI-ADVISOR-QUALITY-HARNESS-20261020
status: verified
owner: shared
human_approval: approved_by_parent_task
---

# Advisor quality harness — 2026-10-20

## Problem and scope

The current real-orchestrator evaluation covers only three opt-in cases
(product help and two proposal tools). It does not prove account-local date
normalization, owner propagation, multi-turn follow-up, empty finance data,
unsupported banking requests, prompt-injection resistance, or the boundary
between a proposal and a write.

This slice adds a deterministic Go harness and a machine-readable case pack.
It uses the production `AdvisorOrchestrator`, `AdvisorToolRegistry`, finance
query service, and structured-part builder with a synthetic `FinanceReader`.
It never reads production rows, confirms actions, or writes financial data.

The opt-in live matrix uses the configured provider only when
`AI_ADVISOR_QUALITY_LIVE=1`; it still uses the synthetic reader and proposal
tools. Default `go test` remains offline and deterministic.

## Decisions

- The synthetic provider is scripted by test scenario, not a production
  keyword router. Deterministic tests prove contracts and boundaries; only the
  opt-in live matrix measures model tool choice.
- The authenticated principal supplies `OwnerID` and `Timezone`. The reader
  records both the owner and normalized query, and tests fail on foreign-owner
  or date-range leakage.
- Missing data is represented by a `not_found` finance result. The structured
  response must not manufacture a card from an empty result.
- Unsupported bank/payment actions must fail closed before the finance reader
  or any write-capable path is reached.
- Proposal tools remain review-only: a proposal result is `status=proposed`,
  has a confirmation view, and does not invoke the finance reader or ledger.

## Coverage

The fixture pack and tests cover:

1. September date range in `Asia/Ho_Chi_Minh`, including UTC half-open bounds
   and exact synthetic amount/card data.
2. Same-owner multi-turn follow-up from a summary to a transaction search.
3. Missing finance data and safe structured fallback.
4. Unsupported bank/payment request rejection.
5. Prompt injection and model-supplied `owner_id` rejection.
6. Transaction proposal boundary and no-mutation proof.

## Verification and limitations

Run offline:

```text
cd backend && go test ./internal/infrastructure/ai -run 'TestAdvisorQuality|TestAdvisorFixturePack' -count=1
```

Run the real provider matrix explicitly, with local/staging credentials only:

```text
cd backend && AI_ADVISOR_QUALITY_LIVE=1 go test ./internal/infrastructure/ai -run TestLiveAdvisorQualityMatrix -count=1 -v
```

The deterministic harness proves tool contracts, owner scope, date ranges,
amounts, and card status. It does not prove model wording or model routing.
The live matrix is bounded and synthetic, but provider availability and model
tool selection remain nondeterministic; a skipped live run is not a quality
pass and production data is never an allowed fixture source.

## Verification evidence

- `go test ./internal/infrastructure/ai -run 'TestAdvisorQuality|TestAdvisorFixturePack' -count=1` — PASS (8 focused contract/fixture tests).
- `go test ./internal/infrastructure/ai -count=1` — PASS; default live matrix is skipped without network opt-in.
- `go test -race ./internal/infrastructure/ai -count=1` — PASS.
- `go test ./... -count=1` — PASS across all backend packages.
- `AI_ADVISOR_QUALITY_LIVE=1` was not run in this verification pass; it requires configured provider credentials and is intentionally excluded from the default gate.

The fixture pack declares six scenario IDs. The deterministic suite executes
the six required boundaries plus a DST/timezone variant; owner, UTC bounds,
synthetic amount, `not_found`, unsupported-tool rejection, unknown
`owner_id` rejection, and `proposed`/confirmation-card semantics are asserted
against the real orchestrator and registry.

## Reconciliation

- API, database schema, provider/client code, and production configuration are
  unchanged.
- The existing AI tool catalog remains authoritative.
- `docs/work/VALIDATION_MATRIX.md` and the AI evaluation ticket are updated
  with the exact commands and limitations after the harness is green.
