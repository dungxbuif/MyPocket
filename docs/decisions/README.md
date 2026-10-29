# Decisions

This folder stores Architecture Decision Records and other durable technical decisions.

- [ADR-022 — Owner-bound native refresh](ADR-022-owner-bound-native-refresh.md): metadata-preserving HTTP and atomic session replacement prevent old refresh overwriting later login/logout.

- [ADR-021 — Full-owner API keys](ADR-021-full-owner-api-keys.md): owner-approved removal of per-key scopes; same-owner JWT-equivalent user operations, lifecycle checks and separate service-token boundary.

- [ADR-020 — Encrypted native outbox](ADR-020-encrypted-native-outbox.md): SQLite atomic intent/lease storage, AES-GCM and account-scoped device-only Keychain keys; runtime integration remains gated.

- [ADR-005 — Persistent AI entry review and explicit approval](ADR-005-ai-entry-review.md): approved hold-Add entry slice, PostgreSQL confirmation receipts, bounded text/OCR integration.
- [ADR-006 — Private AI receipt attachment lifecycle](ADR-006-private-ai-attachments.md): backend OCR-first processing, environment-separated private S3 and approval-scoped attachment links.
- [ADR-008 — Account timezone and calendar dates](ADR-008-account-timezone-and-calendar-dates.md): UTC instants remain distinct from date-only/month labels; account timezone defines report boundaries and does not freeze month data.

## Naming

Use:

```text
ADR-<id>-<short-name>.md
```

Example:

```text
ADR-001-markdown-only-framework.md
```

## ADR Triggers

Create an ADR when work changes:

- Architecture boundaries
- Public API or event contracts
- Database schema or data ownership
- Major dependencies
- Security or authorization model
- Deployment/runtime model
- Repository workflow or standards

Start from `docs/templates/ADR.md`.
