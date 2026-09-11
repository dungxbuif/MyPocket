# MyPocket API quick reference

Base: `https://money.dungxbuif.com/api/v1`

```bash
curl -H 'Authorization: Bearer <user-api-key>' \
  https://money.dungxbuif.com/api/v1/wallets
```

Core reads: `/me`, `/wallets`, `/categories`, `/transactions`, `/dashboard`, `/reports/{cash-flow|categories|daily|cumulative|comparison|insider}`, `/budgets`, `/events`, `/obligations`, `/recurring-schedules`, `/transaction-drafts`, `/assets`, `/portfolio/summary`, `/notifications`, `/sync/changes`, and `/agent/runs/{id}`.

Mutations use the OpenAPI body plus `Idempotency-Key` where declared. Update/archive commands also use the latest `base_version`.

- Import: `POST /imports` with UTF-8 CSV → poll `GET /imports/{id}` → inspect every error → `POST /imports/{id}/confirm` with current version.
- Export: `POST /exports` → poll `GET /exports/{id}` → `GET /exports/{id}/download` for a short-lived URL.
- Reset/delete require recent browser cookie authentication and cannot be delegated to an API-key-only agent.
- Agent: `POST /agent/messages` with `kind`, `message`, optional owned `receipt_id`, and an idempotency key → poll `/agent/runs/{id}` → review `/transaction-drafts` → explicitly confirm or reject. OCR is invoked internally for attached images; there is no public OCR proxy and no bank workflow.

API keys can be created, listed, and revoked at `/api-keys`; plaintext is shown once. Each key is owner-scoped and rate-limited per key/user. Never print the Authorization header.
