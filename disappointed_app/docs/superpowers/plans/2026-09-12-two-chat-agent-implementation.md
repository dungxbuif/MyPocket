# Two Chat Agent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the reviewed two-chat Agent contract: transaction intake creates reviewable transaction drafts only, while advisor chat answers with read-only financial analysis.

**Architecture:** Evolve the existing Go agent tables, service, worker, HTTP router, frontend client, and docs instead of creating a second agent stack. Keep REST/API-key auth convergence, review-first transaction drafts, provider output validation, and OCR image tooling. Use compatibility endpoints for the current `/api/v1/agent/messages` surface while adding the planned `/api/v1/agent/intakes` and `/api/v1/agent/advisor/messages` paths.

**Tech Stack:** Go modular monolith, PostgreSQL migrations, existing worker model adapter, React/TypeScript PWA, Vitest, Go tests, Docusaurus/OpenAPI docs.

**Spec:** `docs/superpowers/specs/2026-09-12-two-chat-design.md`

## Global Constraints

- Transaction intake supports income/expense drafts only; no transfer, budget, debt, edit-posted, bank, or voice actions.
- Advisor chat is read-only and cannot create transaction drafts or other write actions.
- Model output is untrusted and must be validated against owned wallet/category references.
- Drafts do not affect balances or reports until confirmed through existing deterministic draft APIs.
- Existing `/api/v1/agent/messages` clients must not silently gain advisor write behavior; compatibility should route by explicit `kind`.
- Browser cookie writes retain CSRF; Bearer API-key calls must be authenticated through the existing auth context.
- Public docs and AI-readable API docs must be reconciled with implemented behavior.

---

### Task 1: Backend Two-Chat API Contract

**Files:**
- Modify: `backend/internal/agent/types.go`
- Modify: `backend/internal/agent/validation.go`
- Modify: `backend/internal/agent/service.go`
- Modify: `backend/internal/platform/httpapi/agent.go`
- Modify: `backend/internal/platform/httpapi/router.go`
- Modify: `backend/internal/platform/httpapi/agent_test.go`

**Interfaces:**
- Produces: `agent.KindIntake`, `agent.KindAdvisor`, and compatibility aliases from existing `transaction_draft` and `analysis`.
- Produces: `POST /api/v1/agent/intakes`, `POST /api/v1/agent/intakes/{id}/messages`, `GET /api/v1/agent/intakes/{id}`, `POST /api/v1/agent/advisor/messages`, `GET /api/v1/agent/runs/{id}`.
- Consumes: existing `AgentService.Submit(ctx, userID, key, kind, text, receiptID)`.

- [ ] **Step 1: Write failing HTTP tests**

Add tests proving:

```go
func TestAgentIntakeEndpointQueuesIntakeKind(t *testing.T) {
    service := &agentServiceStub{run: agent.Run{ID: "run-1", Kind: agent.KindIntake, Status: agent.StatusQueued}}
    req := agentAuthenticatedRequest(t, http.MethodPost, "/api/v1/agent/intakes", `{"message":"Ăn trưa 80k"}`)
    agentAddCSRF(req)
    req.Header.Set("Idempotency-Key", "intake-once")
    res := httptest.NewRecorder()
    agentIntakes(agentTestConfig(), service).ServeHTTP(res, req)
    if res.Code != 202 || service.kind != agent.KindIntake || service.key != "intake-once" {
        t.Fatalf("code=%d kind=%s key=%s body=%s", res.Code, service.kind, service.key, res.Body.String())
    }
}

func TestAgentAdvisorEndpointQueuesAdvisorKind(t *testing.T) {
    service := &agentServiceStub{run: agent.Run{ID: "run-2", Kind: agent.KindAdvisor, Status: agent.StatusQueued}}
    req := agentAuthenticatedRequest(t, http.MethodPost, "/api/v1/agent/advisor/messages", `{"message":"Tháng này tiêu gì nhiều?"}`)
    agentAddCSRF(req)
    req.Header.Set("Idempotency-Key", "advisor-once")
    res := httptest.NewRecorder()
    agentAdvisorMessages(agentTestConfig(), service).ServeHTTP(res, req)
    if res.Code != 202 || service.kind != agent.KindAdvisor {
        t.Fatalf("code=%d kind=%s body=%s", res.Code, service.kind, res.Body.String())
    }
}
```

- [ ] **Step 2: Run tests to verify RED**

Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run 'TestAgent(IntakeEndpoint|AdvisorEndpoint)' -count=1`

Expected: fail because new handlers/kinds do not exist.

- [ ] **Step 3: Implement minimal API**

Add kind constants and validation mapping. Add handlers that parse message, optional receipt, use idempotency keys, require CSRF, call the existing service with fixed `KindIntake` or `KindAdvisor`, and return the existing run envelope.

- [ ] **Step 4: Run tests to verify GREEN**

Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run 'TestAgent(IntakeEndpoint|AdvisorEndpoint|Messages)' -count=1`

Expected: pass.

### Task 2: Intake Draft Validation and Advisor Read-Only Boundary

**Files:**
- Modify: `backend/internal/agent/validation.go`
- Modify: `backend/internal/agent/repository.go`
- Modify: `backend/internal/agent/repository_test.go`
- Modify: `backend/internal/worker/agent.go`
- Modify: `backend/internal/worker/agent_test.go`

**Interfaces:**
- Produces: `agent.IntakeSchema()` that allows multiple income/expense draft proposals and `needs_input`.
- Produces: repository completion logic that persists only valid income/expense proposals as `transaction_drafts`.
- Consumes: existing `transaction_drafts` confirmation APIs.

- [ ] **Step 1: Write failing repository tests**

Add tests proving intake rejects transfers and supports multiple draft proposals:

```go
func TestAgentIntakeRejectsTransfersAndNeverWritesAccounting(t *testing.T)
func TestAgentIntakeCreatesMultipleReviewDrafts(t *testing.T)
func TestAgentAdvisorCompletionNeverCreatesDrafts(t *testing.T)
```

The transfer test should expect `agent.ErrValidation`. The multiple-drafts test should expect two draft IDs and zero confirmed transactions.

- [ ] **Step 2: Run tests to verify RED**

Run: `rtk env MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/agent -run 'TestAgent(Intake|Advisor)' -count=1`

Expected: fail because current parser accepts transfer and only one transaction.

- [ ] **Step 3: Implement validation and completion**

Extend model result parsing to support an array of `drafts` for intake, preserve compatibility with one `transaction`, reject transfer for intake, require category for income/expense, and insert one draft row per validated proposal with deterministic occurrence keys `agent:<run-id>:<index>`. Advisor accepts answer-only and never writes drafts.

- [ ] **Step 4: Run tests to verify GREEN**

Run the same command. Expected: pass.

### Task 3: Frontend Client Separation

**Files:**
- Modify: `frontend/src/app/agent.ts`
- Modify: `frontend/src/app/agent.test.ts`
- Modify: `frontend/src/screens/AgentScreen.tsx`
- Modify: `frontend/src/app/App.tsx` if needed for naming only.

**Interfaces:**
- Produces: `submitIntakeMessage`, `submitAdvisorMessage`, and compatibility `submitAgentMessage`.
- Consumes: new backend routes while keeping old type shapes.

- [ ] **Step 1: Write failing frontend tests**

Add tests that stub `apiFetch` and prove intake posts to `/api/v1/agent/intakes`, advisor posts to `/api/v1/agent/advisor/messages`, and image intake uploads before submit.

- [ ] **Step 2: Run tests to verify RED**

Run: `rtk npm test -- --run src/app/agent.test.ts`

Expected: fail because functions do not exist or post to old endpoint.

- [ ] **Step 3: Implement client split**

Add the new functions and keep `submitAgentMessage` as a wrapper that routes `transaction_draft` to intake and `analysis` to advisor.

- [ ] **Step 4: Run tests to verify GREEN**

Run: `rtk npm test -- --run src/app/agent.test.ts`

Expected: pass.

### Task 4: Public Docs, OpenAPI, and AI Skill Surface

**Files:**
- Modify: `docs/architecture/API.md`
- Modify: `docs/architecture/ERD.md`
- Modify: `frontend/docs/docs/api/agent.mdx`
- Modify: `frontend/docs/docs/guides/agent-and-images.mdx`
- Modify: `frontend/docs/docs/skills/mypocket-api.mdx`
- Modify: `backend/internal/platform/httpapi/openapi.json`
- Modify: `backend/internal/platform/httpapi/openapi_test.go`
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/CONTEXT.md`
- Modify: `docs/releases/CHANGELOG.md`

**Interfaces:**
- Produces: public route docs for two chat flows and API-key usage.
- Produces: AI-readable skill docs warning that advisor is read-only and intake creates reviewable drafts only.

- [ ] **Step 1: Add docs/openapi tests where available**

Update router/openapi coverage tests if new routes are required in machine inventory.

- [ ] **Step 2: Update docs after code matches behavior**

Document `/api/v1/agent/intakes`, `/api/v1/agent/advisor/messages`, `/api/v1/agent/runs/{id}`, compatibility `/api/v1/agent/messages`, idempotency, scope boundaries, OCR image behavior, and draft confirmation.

- [ ] **Step 3: Run docs/API checks**

Run: `rtk env GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/platform/httpapi -run 'TestOpenAPI|TestDocs|TestAgent' -count=1`

Expected: pass.

### Task 5: Final Verification

**Files:**
- No new files unless tests expose targeted reconciliation needs.

**Interfaces:**
- Produces: command evidence for backend/frontend impacted scopes.

- [ ] **Step 1: Run focused backend tests**

Run: `rtk env MYPOCKET_TEST_DATABASE_URL=postgres://mypocket:mypocket@127.0.0.1:55433/mypocket?sslmode=disable GOCACHE=/private/tmp/mypocket-go-cache go test ./internal/agent ./internal/worker ./internal/platform/httpapi -count=1`

- [ ] **Step 2: Run focused frontend tests**

Run: `rtk npm test -- --run src/app/agent.test.ts`

- [ ] **Step 3: Run build if code touched production frontend**

Run: `rtk npm run build`

- [ ] **Step 4: Record results and residual risk**

Update validation/backlog/context/changelog with pass/fail/skipped evidence. Do not claim production/device/provider acceptance unless actually run.
