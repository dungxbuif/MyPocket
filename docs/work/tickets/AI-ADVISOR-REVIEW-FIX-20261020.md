# Advisor review closure — 2026-10-20

Status: released; residual acceptance gates remain explicitly listed below.
Owner approval: “fix luôn đi”; release approved: “xong thì release luôn”.

Scope: close the integration defects found during review of the
[reliability spec](../../superpowers/specs/2026-10-20-advisor-reliability-design.md)
and [plan](../../superpowers/plans/2026-10-20-advisor-reliability.md).

Root causes: run-local snapshot helpers were never invoked; compact context
was persisted but never loaded; request and tool transcript bounds diverged;
terminal tool handling shared the financial-tool budget; repair transcripts
omitted tool results; non-finance modes and model labels bypassed grounding;
reaper candidates were not revalidated under lock. The 31-case fixture loader
only validated dataset shape and must not be represented as model acceptance.

Implementation: wire snapshots before publishing proposals, return repairable
tool errors, read/revalidate owner-scoped continuation references, enforce
12-message/24-KiB and 96-KiB transcript limits, retain finish_answer after tool
budget exhaustion, pair failed terminal calls with tool results, render labels
and non-finance responses from server-owned sources, and recheck lease under
lock. Release follow-up adds a deadline default to not-yet-deployed migration
000041 for old writers during rolling update; auth/provider policy unchanged.

Verification: regression tests must fail first for each missing behavior, then
pass through the production orchestrator/service and PostgreSQL adapter. Run
full Go suite, vet, race subset and disposable PostgreSQL tests. Dataset schema
checks and scripted provider tests are distinct from live-model evaluation.

Docs: reconcile AI_ADVISOR_TOOLS/API contract, reliability spec/plan,
CONTEXT/BACKLOG/VALIDATION_MATRIX and changelog after code verification. Public
Markdown is the agent-readable surface. ERD unchanged (no schema change).
Production rollout and signed device installation are recorded separately and
are complete for release 1.2.1. Full physical interaction, Face ID/offline,
and Mobile MCP/WDA acceptance remain pending and are not implied here.

## Verified candidate

- Fresh normal/full race runs with dedicated disposable PostgreSQL database
  `mypocket_advisor_release_20261020`: each 753 pass events, zero failures,
  33 explicit Redis/live/storage skips. `go vet ./...` passes. Artifacts:
  `/tmp/mypocket-1.2.1-backend.json`, `/tmp/mypocket-1.2.1-backend-race.json`.
  Twenty `TestAdvisorReview*` events include actual category-loop subtests.
- RED/GREEN covers actual-loop snapshots, 12-message cap, context/calendar,
  protocol repair and budgets, model-independent rendering, duplicate evidence
  hang, invalid search-limit recovery and old API inserts after migration 41.
  PostgreSQL verifies lease recheck and owner/generation context persistence.
- Live DeepSeek-V4.1-Flash passes the existing six cases and three new grounded
  cases (finance, handbook, proposal). Owner-scoped context reads during a
  banking refusal are allowed; proposals/writes are forbidden, matching the
  dataset's expected_no_write. This smoke is not the full 31-case/30x3 gate.
- Native simulator: 380/380 units, zero failure/skip/runtime warnings. Signed
  iPhone Release `1.2.1 (4)` built with Personal Team; codesign and release
  preflight pass. The existing portrait/full-screen warning remains.
- Base/project guards, release-artifact tests, rollout preflight tests pass.
  Dataset schema test is explicitly named TestAdvisorReliabilityFixtureSchema.
- Release keeps runtime/grounding flags false. Migration 41 precedes image
  update, with backup and financial checksum comparison; rollback retains
  additive schema/default. No financial reset or primary-account mutation.

Docs review: public guide/API and agent-readable tool Markdown reconciled;
spec/plan/validation claims corrected. ERD ownership unchanged; runtime schema
and rolling compatibility recorded in migrations README and rollout runbook.
