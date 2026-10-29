# Environments

| Environment | Purpose | URL | Owner | Notes |
| --- | --- | --- | --- | --- |
| local | Development and automated tests | `http://localhost:4173` | shared | Uses disposable local services and `.env.local`. |
| production | Owner-operated MyPocket release | `https://money.dungxbuif.com` | project owner | Docker Swarm services run on `vm100-prod`; PostgreSQL database is `mypocket`. |

## Secrets

Document secret names only. Do not store secret values.

Production secret names are supplied to the `mypocket-api` service by the
deployment environment. They include `DATABASE_URL`, `REDIS_URL`, JWT and
refresh-token secrets, AI/OCR provider keys, Google OAuth credentials, S3
storage credentials, feedback-agent credentials, and Web Push VAPID keys.
