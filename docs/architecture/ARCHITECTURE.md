---
artifact_type: architecture_doc
id: ARCH-MASTER
status: draft
owner: shared
human_fields: [approved_boundaries, architectural_constraints, tradeoff_approval]
ai_fields: [overview, modules, diagrams, flows, dependencies, risks]
shared_fields: [status, linked_decisions]
---

# Architecture

## Field Ownership

- Human owns approved boundaries, constraints, and tradeoff approvals.
- AI maintains implementation-grounded overview, modules, diagrams, flows, dependencies, and risks.

## Overview

MyPocket is currently a modular monolith: a Vite/React mobile-first client calls a Gin HTTP API. The API follows a clean-architecture split between HTTP controllers, use cases, repository interfaces and PostgreSQL/Redis infrastructure. Google OAuth is handled server-side; the browser receives a JWT-backed session result.

## System Boundaries

- Browser/PWA: `app/` React + Vite UI, Atomic Design components.
- Frontend routing: TanStack Router route tree in `app/src/router.tsx`; feature paths are addressable directly and the app shell derives its active section from the URL.
- Dev same-origin path: Vite proxies `/api/*` to Gin (`VITE_BACKEND_URL`, default `http://localhost:8080`); frontend leaves `VITE_API_BASE_URL` empty so browser API calls use the Vite origin.
- Gin API: `backend/cmd/api` composition root and `backend/internal/controller/http` routes/middleware.
- Application: `backend/internal/usecase` authentication and domain orchestration.
- Persistence: GORM PostgreSQL repositories under `backend/internal/infrastructure`; Redis session/profile/home cache.
- External identity: Google OAuth provider.

## Modules

AI entry is implemented as HTTP adapters → `usecase.AIEntryService` → repository/provider interfaces. `infrastructure/ai` holds the text-only compatible provider client and OCR polling; `infrastructure/repository/ai_entry_postgres.go` persists drafts and atomically approves ordinary ledger entries. No model write tools. Long-press global Add opens the entry sheet; advice remains separate and unimplemented. Processing is bounded/synchronous with a session lease, not a durable worker. API main reads ignored `.env.local` from backend working directory without overriding process env. S3 config exists for future retained receipts but is not used by transient OCR. [ADR-005](../decisions/ADR-005-ai-entry-review.md).

Frontend UI ownership: one primitive source `app/src/ui/theme.css` is imported by `styles.css`; named component variants live in `ui/variants.ts`. Atoms own native controls, typography, surface and progress visuals; molecules compose reusable patterns; organisms own screen state/API composition. Build enforces the base contract through `scripts/check-design.mjs`. See [ADR-002](../decisions/ADR-002-design-contract-enforcement.md) and [design gateway](../design/README.md).

| Module | Responsibility | Key Files | Notes |
| --- | --- | --- | --- |
| HTTP controller | Request/response mapping, auth middleware, CORS | `backend/internal/controller/http` | No business rules in handlers. |
| Use case | Application rules and orchestration | `backend/internal/usecase` | Depends on repository interfaces. |
| Repository | Persistence abstraction | `backend/internal/repository`, `backend/internal/infrastructure/repository` | GORM implementation; raw SQL only for genuinely complex queries. |
| Database | Durable account and wallet data | PostgreSQL + GORM | Schema changes use explicit migrations/AutoMigrate review. |
| Cache | Session and read-model caching | Redis | Cache is never the source of truth. |

## Data Flow

`React -> Gin route -> auth middleware -> use case -> repository -> PostgreSQL`; read-through profile/home data may use Redis. Google OAuth exchanges code only in the backend.

## Runtime Flow

Local development runs Vite on `:4173`, Gin on `:8080`, PostgreSQL on `:5432` and Redis on `:6379`. Swagger UI is served by Gin at `/api/v1/docs/index.html`.

## Dependencies

- Gin, GORM, PostgreSQL, Redis, Google OAuth, Swagger (`swaggo/swag` + `gin-swagger`).

## Risks And Tradeoffs

- Xoá thực ví đã được chốt. Detail design ví còn cần review phần ledger điều chỉnh, trường riêng goal/credit và tác động liên kết sang ví khác; xem [DESIGN-01-02](../work/tickets/TICKET-01-02-DETAIL_DESIGN.md).
- Generated Swagger files must be regenerated from Go annotations; do not hand-edit generated output.

## Linked Decisions

- TBD
