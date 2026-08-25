# Setup

## Prerequisites

- Go 1.23 or newer.
- Node.js/npm for repository scripts.
- PostgreSQL for local database-backed development.

## Installation

Run Go module setup from the repository root:

```bash
go mod download
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

## Verification

Apply migrations:

```bash
npm run migrate:up
```

Run backend tests:

```bash
npm run test:go
```
