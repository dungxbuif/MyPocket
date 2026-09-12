# Decisions

This folder stores Architecture Decision Records and other durable technical decisions.

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

## MyPocket Decisions

- [ADR-001: React PWA and Go Modular Monolith](ADR-001-react-go-modular-monolith.md)
- [ADR-002: Stateless Google OAuth Authentication](ADR-002-stateless-google-oauth.md)
- [ADR-003: Offline Sync with Explicit Conflict Review](ADR-003-offline-sync-conflict-review.md)
- [ADR-004: Review-First Shared Transaction Ingestion](ADR-004-review-first-ingestion.md)
- [ADR-005: Restricted Append-Only Audit Log](ADR-005-restricted-audit-log.md)
- [ADR-006: Separate Asset Portfolio Valuation](ADR-006-separate-asset-portfolio-valuation.md)
- [ADR-007: Separate Offline Databases per User](ADR-007-offline-user-databases.md)
- [ADR-008: Atomic Offline Sync Commit](ADR-008-atomic-offline-sync-commit.md)
