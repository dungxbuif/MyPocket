# Advisor Idle Reset Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expire visible/model chat after30minutes without accepted user input while preserving owner isolation, replay safety and financial data.

**Architecture:** Keep the owner conversation row, rotate its generation and physically purge content in one transaction. Retain minimal replay/usage metadata, enforce expiry on authenticated access and existing worker passes, and make native state reject old-generation responses.

**Tech Stack:** Existing Go/Gin/GORM/PostgreSQL, SwiftUI/iOS17+, XCTest/XCUITest. No new agent framework or UI library.

**Spec:** `docs/superpowers/specs/2026-10-15-advisor-idle-reset-design.md` (commit9ab9248d).

Status: in_progress; owner approved30minutes and requested completion. The
server retention foundation and idempotent manual reset route are implemented
locally, and native Chat mới consumes the route. Native lifecycle metadata, real
PostgreSQL proof and production rollout remain review gates.
Preserve inline/native execution method already chosen for this workstream.

## Global Constraints

- TTL30minutes inclusive from server accepted user input; reads, replay and assistant output do not extend it.
- PostgreSQL server-side transaction is authoritative; no device-clock authority.
- Preserve conversation ID and monotonically increasing sequence; increment generation.
- Clear content and summary, not financial tables or already confirmed writes.
- Keep owner JWT/API-key authority, explicit confirmation, existing green/neutral/Manrope bases.
- Conversation→run locking; reject old answer/event/fact writes after reset.
- No streaming, bank actions, automatic draft submission or production reset.
- Expiry access enforcement is immediate; worker deletion can lag by interval/backlog/downtime. Backups are not erased by this flow.
- No full-goal/release/feedback acceptance from narrow tests or install success.

## Review Focus

1. An old same-key retry after purge must return410 without rerunning provider; Task2.
2. A late provider response after reset cannot resurrect history or actions; Tasks1/4.
3. Repeated reset after a new question cannot delete that new question; Task2.
4. App foreground/history polling must not extend retention; Tasks1/3/4.
5. Worker failure must neither suppress recurring work nor expose partial deletion; Task3.

## Files and boundaries

- `backend/migrations/000033_advisor_idle_reset.up.sql`: additive last-user timestamp/index and minimal reset-request ledger; migration32 remains intact.
- `backend/internal/entity/ai_advisor_lifecycle.go`: metadata/history/reset domain records only.
- `backend/internal/repository/ai_advisor_lifecycle.go`: lifecycle capability and exact result/error contracts.
- `backend/internal/infrastructure/repository/ai_advisor_lifecycle_postgres.go`: locked expiry/reset/purge/history operations.
- Existing `ai_advisor_postgres.go`: generation-safe StartRun/answer/event writes and replay handling.
- `backend/internal/usecase/ai_advisor_lifecycle.go`: clock-driven access/reset orchestration; no direct finance mutation.
- Existing advisor handler/router and generated Swagger: additive reset route/history meta/410 mapping.
- `backend/cmd/worker/main.go`: independent bounded cleanup pass using current worker interval.
- `mobile-ios/MyPocketIOS/Core/Models/AdvisorLifecycleModels.swift`, API/assistant consumer: optional wire metadata, volatile generation/epoch/deadline, reset UI.
- Corresponding repository/HTTP/worker/native tests plus existing public/internal docs.

### Task1: Locked lifecycle and generation-safe persistence

**Interfaces:** Add `AdvisorLifecycleRepository.ReadHistory(ctx context.Context, ownerID, conversationID string, beforeSeq int64, limit int, now time.Time) (entity.AdvisorHistory, error)` and `ExpireIdleBatch(ctx context.Context, now time.Time, limit int) (int, error)`. AdvisorHistory contains Messages, ConversationID, Generation, LastUserMessageAt, ExpiresAt, ServerTime. Add `last_user_message_at` to entity conversation; inject a repository clock for deterministic StartRun tests without altering public payloads.

- [ ] RED: atT+29m59s history remains; atT+30m content/events/facts/summary gone, generation+1, next sequence unchanged, financial table counts/checksums unchanged. Fresh accepted user input extends TTL; reads/assistant/replay do not. Two owners remain isolated; empty chat unchanged.
- [ ] Run `rtk proxy go test -race ./internal/infrastructure/repository -run 'TestAdvisorIdle|TestAdvisorGeneration' -count=1` with dedicated local TEST_DATABASE_URL. Expected actual behaviorFAIL, not missing test/database skip.
- [ ] Implement additive migration/backfill from maximum user-message time, no deletion during migration. Locked purge retains minimal run IDs/request hash/usage, marks purged, clears lease, removes content by owner/conversation/generation. All old answer/event/fact writers validate current generation under established locks.
- [ ] GREEN: realPG parallel send/expiry, answer-waits/reset, terminal replay tombstones, migration32→33 empty/existing fixtures; full Go/race/vet. Existing cancellation race tests remain unchanged.
- [ ] Reconcile database/lifecycle docs and commit only this task's scoped changes.

### Task2: Idempotent reset and owner-authenticated API

**Interfaces:** Add `ResetConversation(ctx context.Context, ownerID, requestID, payloadHash string) (*entity.AdvisorResetResult, error)` to lifecycle capability. Result contains ConversationID, Generation, ServerTime, Replayed. Reset ledger unique(owner_id,request_id), no content. Use existing ErrAdvisorPurged mapped410 `advisor_history_expired`; changed-payload key remains409. Keep history data array; add lifecycle fields in response meta for old decoder compatibility. The local route, ledger and native consumer are implemented; real DB/auth replay proof remains open.

- [ ] RED: JWT and full-owner key reset only own chat; invalid/missing key400; same reset replay after new accepted question returns original generation and preserves new question; changed body409; old purged message key410 with zero provider calls; foreign owner cannot reset/read/poll old content.
- [ ] Run `rtk proxy go test -race ./internal/controller/http ./internal/usecase -run 'TestAdvisorIdle|TestAdvisorReset|TestAdvisorPurgedReplay' -count=1`. Expected lifecycle contractFAIL.
- [x] Implement `POST /api/v1/ai/advisor/conversation/reset` and preserve access-time expiry/410 handling. Purged replay check must not load deleted user message; retain owner/principal validation before lookup. No fake finance run for reset.
- [ ] GREEN full Go tests/race/vet including realHTTP/PG/Redis auth, stable old history-array contract and no financial writes. Regenerate Swagger with project generator.
- [ ] Reconcile public human guide/Markdown agent catalog/OpenAPI and commit scoped task.

### Task3: Independent worker cleanup

**Interfaces:** `runAdvisorExpiryPass(ctx context.Context, runner interface{ ExpireIdleBatch(context.Context,time.Time,int)(int,error) }, now time.Time) error`, batch100. Existing recurring pass remains independent; worker interval unchanged.

- [ ] RED: expired content physically removed without user return;100max/pass; locked owner skipped without blocking others; recurring failure does not suppress purge and purge failure does not suppress recurring; restart completes remaining eligible rows.
- [ ] Run `rtk proxy go test -race ./cmd/worker ./internal/infrastructure/repository -run 'TestAdvisorExpiry|TestWorkerIndependent' -count=1`; expectedFAIL with current recurring-only loop.
- [ ] Wire SKIP LOCKED bounded expiry pass, context deadline and content-free count/latency/error diagnostics; no infinite cleanup loop or new daemon.
- [ ] GREEN focused/full Go with realPG fixtures; document downtime/backlog/deletion versus immediate access refusal. Commit scoped task.

### Task4: Native expiry/reset and stale-response suppression

**Interfaces:** `AdvisorHistoryMetaModel` optional generation/lastUserMessageAt/expiresAt/serverTime; `AdvisorHistoryPageModel(messages: [AdvisorMessageModel], meta: AdvisorHistoryMetaModel?)`; `FinanceAPI.advisorHistory(conversationID:beforeSequence:) async throws -> AdvisorHistoryPageModel`, `resetAdvisorConversation(idempotencyKey:) async throws -> AdvisorResetModel`. Preserve existing advisorMessages API for other consumers. Assistant owns volatile context epoch and server-relative monotonic expiry deadline, not a global persisted financial preference.

- [ ] RED: decode old/new metadata; expiry hides old messages/actions while preserving unsent draft; reads/foreground never extend deadline; logout/owner change cannot display old owner; new reset key persisted for uncertain retry and reused, reset never auto-submits draft. Late history/submit/poll response from old epoch ignored.
- [ ] Run MyPocketIOSTests targeted lifecycle/API tests using the existing SE LOCAL_API_UI gate. Expected behaviorFAIL; no production credential/provider.
- [x] Add Chat mới to existing header menu with shared destructive confirmation. On confirmed reset or server expiry clear conversation/confirmation state, preserve the composer draft, and require explicit resend. Offline hide/expiry metadata remains open; no server-deletion claim while disconnected.
- [ ] GREEN full units plus actual localAPI/native tests for foreground/restart/offline/reset-replay/expired action; keyboard/Dynamic Type/landscape and retained drafts. Keep financial outbox untouched.
- [ ] Reconcile native screen/registry and human+agent docs; commit scoped task.

### Task5: Release gates and owner handoff

- [ ] Run full Go tests/race/vet with dedicated PostgreSQL/Redis, migration version/dirty checks, native full units/UI lifecycle+critical business and base/project/diff guards. Confirm no skipped mandatory cases, no financial data changes from purge.
- [ ] Fresh independent review of lifecycle/auth/replay/race and native stale actions, then TDD fix pass for Important/Critical findings. Record minor deferrals explicitly.
- [ ] Prepare scoped immutable candidate preserving already deployed device-auth routes and including all necessary untracked files; never infer deletions from a tracked-only diff. Confirm manager access before deployment; worker node must not be promoted.
- [ ] Validate production backup before additive migration/API+worker rollout. Explicit owner acknowledgement is required before enabling cleanup of existing expired chat. Preserve financial counts/checksums, validate health/auth/reset/replay/deletion on synthetic owner, never real owner financial test writes.
- [ ] Build/strict-sign/install compatible native without uninstall, confirm launch. Hardware/owner UAT required; release notes and feedback statuses stay unchanged until owner accepts.

## Self-review and handoff

Spec coverage maps timing/purge/generation toTask1, reset/replay/API toTask2,
physical cleanup toTask3, offline/foreground/action state toTask4, docs/rollout to
all tasks andTask5. Existing adapter/history-array compatibility retained. All
Review Focus failure classes have explicit tests; no undefined time/ownership
behavior or placeholder methods. Written plan awaits review; execution remains
inline/native in the existing isolated worktree, not a new agent framework.
