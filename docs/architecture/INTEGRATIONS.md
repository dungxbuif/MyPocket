---
artifact_type: integrations_doc
id: INTEGRATIONS
status: active
owner: shared
human_fields: [provider_choice, credential_approval, business_owner]
ai_fields: [systems, contracts, secrets_names, failure_modes]
shared_fields: [status, trace]
updated: 2026-08-24
---

# Integrations

## Field Ownership

- Human-approved provider classes: Google OAuth, OpenAI-compatible AI, external OCR API, S3-compatible object storage, Web Push, and external bank-notification webhook sender.
- Exact production URLs, credentials, models, and connection strings are supplied through deployment secrets.
- AI documents contracts and secret names without storing values.

## External Systems

| System | Purpose | Contract | Owner | Status |
| --- | --- | --- | --- | --- |
| Homelab PostgreSQL | Authoritative relational storage and worker coordination | PostgreSQL connection | Homelab operator | approved |
| Google OAuth | User authentication and verified identity | OAuth 2.0 authorization code callback | Homelab operator | approved |
| S3-compatible storage | Private receipts and generated exports | S3 API and presigned URLs | Homelab operator | approved |
| OpenAI-compatible text API | Parse chat text into structured drafts | Chat/responses-compatible JSON schema output | Homelab operator | approved |
| OpenAI-compatible multimodal API | Parse AI-chat images into structured drafts | Image-capable chat/responses contract | Homelab operator | approved |
| External OCR API | Extract receipt text in add-transaction flow | Configurable HTTP adapter | Homelab operator | approved |
| Web Push service | Deliver best-effort browser notifications | Web Push protocol with VAPID | Browser endpoint/provider | approved |
| External notification sender | Forward bank notification text | MyPocket HMAC webhook | User/operator | approved |
| Manual export consumer | Open CSV/Sheets-compatible snapshots | Downloaded file | User | approved |

## Credentials And Configuration

Store names only; never commit values:

- `DATABASE_URL`
- `PUBLIC_APP_URL`
- `CORS_ALLOWED_ORIGINS`
- `COOKIE_SIGNING_KEY`
- `COOKIE_MAX_AGE_SECONDS`
- `DATA_ENCRYPTION_KEY`
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `GOOGLE_REDIRECT_URL`
- `S3_ENDPOINT`
- `S3_REGION`
- `S3_BUCKET`
- `S3_ACCESS_KEY_ID`
- `S3_SECRET_ACCESS_KEY`
- `S3_FORCE_PATH_STYLE`
- `OPENAI_BASE_URL`
- `OPENAI_API_KEY`
- `OPENAI_TEXT_MODEL`
- `OPENAI_VISION_MODEL`
- `OCR_BASE_URL`
- `OCR_API_KEY`
- `WEB_PUSH_PUBLIC_KEY`
- `WEB_PUSH_PRIVATE_KEY`
- `WEB_PUSH_SUBJECT`
- `AUDIT_VIEWER_EMAIL`
- `AUDIT_RETENTION_DAYS` with default `180`
- `LOG_LEVEL`

Per-source bank webhook secrets are generated and displayed once, then stored encrypted with `DATA_ENCRYPTION_KEY`; they are not global environment variables.

## Failure Modes

- PostgreSQL unavailable: readiness fails; API rejects writes; worker does not acquire jobs.
- S3 unavailable: receipt/upload/export operations remain retryable; finance records without required attachment completion are not confirmed implicitly.
- Google OAuth unavailable or callback invalid: authentication fails without provisioning a partial user.
- AI/OCR unavailable, timed out, rate-limited, or malformed: preserve safe input references, return a provider error, and create no confirmed transaction.
- Web Push denied or delivery fails: durable in-app notification remains; retry policy is capped.
- Webhook signature/timestamp/nonce invalid: reject before parsing and record a redacted security audit event.
- Duplicate webhook source event or payload digest: return an idempotent duplicate result without a new draft.
- Misconfigured `AUDIT_VIEWER_EMAIL`: audit API remains forbidden to ordinary users; startup/config diagnostics expose no sensitive values.

## Linked Decisions

- [ADR-001](../decisions/ADR-001-react-go-modular-monolith.md)
- [ADR-002](../decisions/ADR-002-stateless-google-oauth.md)
- [ADR-004](../decisions/ADR-004-review-first-ingestion.md)
- [ADR-005](../decisions/ADR-005-restricted-audit-log.md)
