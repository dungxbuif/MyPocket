---
artifact_type: adr
id: ADR-002
status: accepted
owner: shared
human_fields:
  - decision_approval
  - final_status
ai_fields:
  - context
  - alternatives_considered
  - consequences
  - linked_work
shared_fields:
  - decision
  - trace
trace:
  requirements: [REQ-F-001, REQ-NF-001, REQ-NF-004]
  phase: PHASE-001
  tickets_or_bugs: [TICKET-003]
  detail_design: ../superpowers/specs/2026-08-23-mypocket-system-design.md
  master_docs:
    - ../requirements/SPEC.md
    - ../architecture/ARCHITECTURE.md
    - ../architecture/API.md
    - ../architecture/ERD.md
  release_notes: ../releases/CHANGELOG.md
---

# ADR-002: Stateless Google OAuth Authentication

## Status

- Status: accepted
- Date: 2026-08-24
- Decision approval: user-approved during design review

## Context

The approved product supports multiple users but explicitly excludes linked-device management and database session records. Authentication must still maintain secure browser state and enforce user isolation.

## Decision

Use Google OAuth authorization-code login to provision/find an application user, then issue a signed Secure HttpOnly SameSite=Lax application cookie. Persist neither Google access/refresh tokens nor a session row. Logout clears the current cookie.

## Alternatives Considered

- Database-backed sessions: rejected by explicit user scope.
- Generic OIDC: rejected in favor of Google OAuth only.
- No authentication behind the homelab network: rejected because the product is multi-user and stores sensitive finance data.

## Consequences

- Positive: minimal authentication persistence and no provider-token retention.
- Negative: no server-side all-device logout, per-device revocation, or session inspection.
- Neutral: users reauthenticate with Google when the cookie expires.

## Linked Work

- Approved design: [MyPocket System Design](../superpowers/specs/2026-08-23-mypocket-system-design.md)
- Requirements: [REQUIREMENTS.md](../requirements/REQUIREMENTS.md)
- Roadmap: [ROADMAP.md](../work/ROADMAP.md)
- Architecture: [ARCHITECTURE.md](../architecture/ARCHITECTURE.md)
- API: [API.md](../architecture/API.md)
- ERD: [ERD.md](../architecture/ERD.md)
- Integrations: [INTEGRATIONS.md](../architecture/INTEGRATIONS.md)

