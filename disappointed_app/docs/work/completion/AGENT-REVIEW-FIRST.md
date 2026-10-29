# OpenAI-Compatible Agent — Completion Evidence

Date: 2026-09-11
Implementation commit: pending at evidence creation

## Delivered

- Optional, fail-closed OpenAI-compatible provider configuration (`AI_*`) with bounded timeout and retries; production requires HTTPS.
- Asynchronous, idempotent `POST /api/v1/agent/messages` and owned `GET /api/v1/agent/runs/{id}`.
- PostgreSQL leasing with `FOR UPDATE SKIP LOCKED`; provider calls occur after the claim transaction completes.
- Minimized owned wallet/category references plus a scoped 30-day aggregate for analysis.
- Strict JSON schema and server-side validation of type, integer VND amount, date, transaction shape, active ownership, and unknown fields.
- Agent proposals create pending `transaction_drafts` only. The existing explicit draft confirmation command remains the sole accounting boundary.
- API-key and cookie authentication are inherited from the common authenticated boundary; cookie mutation requires CSRF and Bearer mutation does not.

## Verification

```text
env MYPOCKET_TEST_DATABASE_URL=... go test -p 1 ./internal/agent ./internal/planning ./internal/finance ./internal/platform/httpapi ./internal/worker -count=1
PASS

env GOCACHE=/private/tmp/mypocket-go-cache MYPOCKET_TEST_DATABASE_URL=... MYPOCKET_TEST_REDIS_URL=... go test -race -p 1 ./... -count=1
PASS (all backend packages)

go vet ./...
PASS
```

The integration proof asserts that a completed transaction proposal produces one pending draft while transaction count and wallet balance remain unchanged. A foreign wallet ID is rejected before draft persistence.
