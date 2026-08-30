# Setup

## Prerequisites

- Go 1.23 or newer.
- Node.js/npm for frontend scripts.
- PostgreSQL for local database-backed development.

## Installation

Run Go module setup from `backend/`:

```bash
cd backend
go mod download
```

Run frontend dependency setup from `frontend/`:

```bash
cd frontend
npm install
```

## Configuration

Required backend environment variables:

| Variable | Purpose |
| --- | --- |
| `DATABASE_URL` | PostgreSQL connection string used by migrations, API, and worker. |
| `PUBLIC_WEB_URL` | Public web origin used by browser-facing flows. |
| `COOKIE_SECRET` | At least 32 bytes; signs application cookies. |
| `CSRF_SECRET` | At least 32 bytes; protects cookie-authenticated mutations. |
| `HTTP_ADDR` | Optional API bind address; defaults to `:8080`. |
| `S3_ENDPOINT` | Optional S3-compatible endpoint; when set, bucket and credentials are required. |
| `S3_BUCKET` | Private object bucket for receipts, exports, and smoke checks. |
| `S3_ACCESS_KEY` | S3-compatible access key. |
| `S3_SECRET_KEY` | S3-compatible secret key. |

## Verification

Apply migrations:

```bash
cd backend
go run ./cmd/migrate
```

Run backend tests:

```bash
cd backend
go test ./...
```

Run frontend tests:

```bash
cd frontend
npm test -- --run
```
