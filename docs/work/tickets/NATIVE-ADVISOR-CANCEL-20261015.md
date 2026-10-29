# Assistant cancellation correctness

Status: in_review — approved bounded lock/recheck fix verified locally; stable device candidate requested. Investigate under approved chat cancellation verification,
no product/API/lifecycle redesign approved by this ticket.

Context: native history final274/274PASS; cancel adapter only proves correct
POST/auth/key and no resubmission. Existing service test uses a blocking provider
and memory store; PostgreSQL test cancels before attempting a late answer. Neither
tests interleaving a pending answer insert with a committed cancellation.

## Current source findings / hypothesis

MessagesPost calls Service.Submit synchronously, returns run_id only after the
provider finishes. Native sets activeRunID only after await submit returns and
only supplies composer onCancel when activeRunID exists. Real local capability
HTTP200/enabledtrue/read-only confirmed without model call. Thus source indicates
no stop control during the initial synchronous provider request. This is not yet
rendered E2E proof; do not claim live cancellation works from route tests.

Repository CreateAssistantMessage reads run status without locking, then locks
conversation before insert. MarkRunStatus cancellation locks run but not
conversation. Hypothesis: cancellation can commit between the initial status
read and conversation lock; stale read then accepts an answer after cancellation.
Existing sequential test misses that interleaving.

## Verification-only next step

Add real PostgreSQL tests using new UUID synthetic owners, no financial rows or
provider. Verify foreign-owner cancellation refusal, old cancellation/late status
cannot alter a newer run, and unchanged expected message sequences. A deterministic
race probe holds conversation row lock, starts actual CreateAssistantMessage on
another transaction, observes that connection waiting on conversation lock through
pg_stat_activity, commits cancellation, releases lock, and checks no late answer
was inserted. Use bounded contexts/barriers, never unbounded sleeps. This must
exercise actual repository transactions, not assert mock behavior. Keep failing
evidence; investigate/review before changing lock ordering or transport contract.

Local only TEST_DATABASE_URL is injected in memory from exact dedicated Docker
container credentials; never print secret/DSN. Host127.0.0.1:55434, DB
mypocket_feedback. Test-owned owners cleaned by exact ID; no broad reset. No UI,
runtime/provider/prod/device change. No new endpoints until reviewed design.

Reconcile chat/fullgoal/context/backlog/matrix/ledger and human+agent guide if
current cancellation claims are stale. API/schema/ADR unchanged during tests;
any subsequent transport/security/locking decision needs its own bounded or
architectural review. No release/feedback resolution before owner acceptance.

Links: [goal](NATIVE-FULL-COMPLETION-20261014.md),
[chat](NATIVE-CHAT-ADAPTIVE-20261014.md), [validation](../VALIDATION_MATRIX.md),
[plan](../../superpowers/plans/2026-10-14-native-ui-system-consolidation.md),
[catalog](../../architecture/AI_ADVISOR_TOOLS.md).

## Real PostgreSQL reproduction

Green after approved patch: `/tmp/mypocket-advisor-cancel-db-green-20261015.log`,
4/4PASS0skip. Writer now locks conversation, then locks/re-reads owner-scoped run
and checks generation/queued-or-running status immediately before insert. Existing
lease-lost error rejects the reproduced late insert; API/schema unchanged. Full
`go test -race ./...`PASS, log `/tmp/mypocket-stable-backend-race-20261015.log`.
No production rollout; initial native Stop and general completion/cancel atomicity
remain separate. This closes only the deterministic stale-read insertion race.

70798terminal1, `/tmp/mypocket-advisor-cancel-db-red-20261015.log`:3PASS/1FAIL,
0skip, no Go data-race report. Existing start/replay/sequential cancellation and
blocking-provider service test pass; new foreign-owner/old-run-versus-new-run test
passes. Deterministic conversation-lock barrier test FAILS: CreateAssistantMessage
returns nil error after cancellation commits, and history contains seq2assistant
late raced answer plus seq1question instead of question only. This establishes
the stale-status insert race using real transactions, not mocked persistence.
Test-owned users are removed by exact-ID teardown; post-query finds0
@advisor-cancel.test users. No model/network/production/financial writes.

Focused command with dedicated local TEST_DATABASE_URL injected in memory (no
secret/DSN output):
```sh
rtk proxy go test -race ./internal/infrastructure/repository ./internal/usecase -run 'TestAdvisorPostgres(StartRun|CancelledRun|CancellationWhile)|TestAdvisorServiceCancelInterruptsInFlightProvider' -count=1 -v
```
DB endpoint127.0.0.1:55434/mypocket_feedback, exact fixture container; credentials
read into child environment, captured output secret-redacted. gofmt/diff pass.
No product patch or green whole-backend claim.

Bounded design sent: after locking conversation, lock/re-read run by owner and
validate generation/status immediately before inserting; use existing lease-lost
error. Preserve conversation→run lock order used by StartRun, existing API/schema/
permissions/provider behavior. Race/owner/new-run and surrounding Go/race tests
must pass before claiming fixed. Human approved this exact design; implement locally
before handing off the production-connected iPhone candidate. Backend deployment
is separate and must not be implied by native installation. Initial synchronous iOS Stop
availability remains a separate source finding/transport review, not solved by
this DB fix. Completion-versus-cancel atomicity beyond the demonstrated schedule
also needs review; do not generalize one race test to all cancellation semantics.
