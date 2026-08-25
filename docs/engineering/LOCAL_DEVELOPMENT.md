# Local Development

## Common Commands

| Command | Purpose |
| --- | --- |
| `npm run migrate:up` | Apply ordered PostgreSQL migrations from `migrations/` using `DATABASE_URL`. |
| `npm run dev:api` | Start the Go HTTP API on `HTTP_ADDR` or `:8080`. |
| `npm run dev:web` | Start the Vite React PWA shell. |
| `npm run dev:worker` | Start the Go worker and verify PostgreSQL connectivity. |
| `npm run build:web` | Build the production PWA shell assets. |
| `npm run test:web` | Run the web shell test suite. |
| `npm run test:go` | Run the default Go test suite. |

## Local Workflow

Set `DATABASE_URL` before running migration, API, or worker commands. API and worker startup fail fast when PostgreSQL is unavailable.

Set `MYPOCKET_TEST_S3_ENDPOINT`, `MYPOCKET_TEST_S3_BUCKET`, `MYPOCKET_TEST_S3_ACCESS_KEY`, and `MYPOCKET_TEST_S3_SECRET_KEY` to run the S3-compatible smoke test; otherwise that integration path is skipped by the default test suite.
