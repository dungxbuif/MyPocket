# MyPocket API quick reference

Base: `https://money.dungxbuif.com/api/v1`

```bash
curl -H 'Authorization: Bearer <user-api-key>' \
  https://money.dungxbuif.com/api/v1/wallets
```

Core reads: `/me`, `/wallets`, `/categories`, `/transactions`, `/dashboard`, `/reports/{cash-flow|categories|daily|cumulative|comparison|insider}`, `/budgets`, `/events`, `/obligations`, `/recurring-schedules`, `/transaction-drafts`, `/assets`, `/portfolio/summary`, `/notifications`, `/sync/changes`, `/agent/intakes/{session_id}`, and `/agent/runs/{id}`.

Mutations use the OpenAPI body plus `Idempotency-Key` where declared. Update/archive commands also use the latest `base_version`.

- Wallets use behavior `type`: `basic`, `goal`, or `credit`. For goal wallets, optional `goal_target_vnd` and `goal_deadline_on` are allowed; for credit wallets, optional `credit_limit_vnd`, `statement_day`, and `payment_due_day` are allowed.
- Import: `POST /imports` with UTF-8 CSV → poll `GET /imports/{id}` → inspect every error → `POST /imports/{id}/confirm` with current version.
- Export: `POST /exports` → poll `GET /exports/{id}` → `GET /exports/{id}/download` for a short-lived URL.
- Reset/delete require recent browser cookie authentication and cannot be delegated to an API-key-only agent.
- Budget assignment is transaction-level: send `budget_id` only on expense transactions/drafts/schedules. Budget progress counts confirmed, active, report-included expenses explicitly assigned to the budget.
- Recurring schedules support `PATCH /recurring-schedules/{id}`, `/pause`, `/resume`, optional `ends_at`, and `posting_mode` (`draft` default or `auto_post`). `auto_post` writes confirmed transactions idempotently per occurrence.
- Agent intake: `POST /agent/intakes` or `/agent/intakes/{session_id}/messages` with `message`, optional owned `receipt_id`, and an idempotency key → response includes `session`, user `message`, and queued `run` → poll `/agent/runs/{id}` if needed → read `/agent/intakes/{session_id}` for durable messages/action cards → review `/transaction-drafts` → explicitly confirm or reject. Advisor: `POST /agent/advisor/messages` is read-only. OCR is invoked internally for attached images; there is no public OCR proxy and no bank workflow.

API keys can be created, listed, and revoked at `/api-keys`; plaintext is shown once. Each key is owner-scoped and rate-limited per key/user. Never print the Authorization header.
