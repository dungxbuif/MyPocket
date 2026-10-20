# Finance Assistant evaluation cases

`cases.json` is a synthetic, machine-readable contract pack for the Go
advisor harness. It contains no production IDs, credentials, prompts from
real users, or financial rows.

`reliability-cases.json` is the expanded 31-case reliability matrix used by
the release plan. It covers Vietnamese date ranges, catalog ambiguity, owner
isolation, reports, proposals/CAS, jars/budgets, reset lifecycle and the typed
`finish_answer` grounding boundary. It is a deterministic fixture catalog and
does not call a provider or production database by itself.

`TestAdvisorReliabilityFixtureSchema` checks its metadata, IDs, required case
groups and that declared tool names exist in the production registry. It does
not execute the 31 prompts and is not an intent/accuracy score. Regression
tests named `TestAdvisorReview*` run the real service/orchestrator with scripted
provider responses to exercise snapshots, evidence, repair, budgets and context.
PostgreSQL `TestAdvisorReaperRechecksCandidateLeaseUnderLock` and
`TestAdvisorContextStatePersistsWithOwnerGenerationIsolation` cover persistence.
These contract tests do not establish how well a live model selects tools.

The deterministic harness uses the production orchestrator and tool registry
with a synthetic finance reader. It asserts owner propagation, local-calendar
range normalization, deterministic amount/card data, missing-data behavior,
unsupported banking rejection, injection boundaries, and proposal-only action
semantics. The scripted provider is a test double; it does not force routing in
production.

The opt-in live matrix is separate:

```text
cd backend && AI_ADVISOR_QUALITY_LIVE=1 go test ./internal/infrastructure/ai -run TestLiveAdvisorQualityMatrix -count=1 -v
```

Live runs use the configured provider and this synthetic reader only. They must
never use production data or confirmation/write endpoints. A skipped or
provider-unavailable live run is not a quality pass.
