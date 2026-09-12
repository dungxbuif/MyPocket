# Local Development

## Common Commands

| Command | Purpose |
| --- | --- |
| `cd backend && go run ./cmd/migrate` | Apply ordered PostgreSQL migrations from `backend/migrations/` using `DATABASE_URL`. |
| `cd backend && go run ./cmd/api` | Start the Go HTTP API on `HTTP_ADDR` or `:8080`. |
| `cd backend && go run ./cmd/worker` | Start the Go worker and verify PostgreSQL connectivity. |
| `cd frontend && npm run dev` | Start the Vite React PWA shell. |
| `cd frontend && npm run build` | Build the production PWA shell assets. |
| `cd frontend && npm test -- --run` | Run the web shell test suite. |
| `cd backend && go test ./...` | Run the default Go test suite. |
| `bash scripts/smoke-platform.sh` | Probe the local Compose API and S3-compatible health endpoints. |

## Local Workflow

Source code is split into `backend/` for the Go module and `frontend/` for the React PWA. npm package metadata lives in `frontend/`; Go module metadata lives in `backend/`.

Set `DATABASE_URL` before running migration, API, or worker commands. API and worker startup fail fast when PostgreSQL is unavailable.

The Compose web service serves the PWA at `http://127.0.0.1:5173` and proxies same-origin `/api/` requests to the API container, so fixture login can run from the browser without a separate frontend API base URL.

Set `MYPOCKET_TEST_S3_ENDPOINT`, `MYPOCKET_TEST_S3_BUCKET`, `MYPOCKET_TEST_S3_ACCESS_KEY`, and `MYPOCKET_TEST_S3_SECRET_KEY` to run the S3-compatible smoke test; otherwise that integration path is skipped by the default test suite.
