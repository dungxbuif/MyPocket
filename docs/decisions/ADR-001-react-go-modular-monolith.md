---
artifact_type: adr
id: ADR-001
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
  requirements: [REQ-F-001, REQ-NF-006]
  phase: PHASE-001
  tickets_or_bugs: [TICKET-001, TICKET-002, TICKET-004]
  detail_design: ../superpowers/specs/2026-08-23-mypocket-system-design.md
  master_docs:
    - ../requirements/SPEC.md
    - ../architecture/ARCHITECTURE.md
    - ../architecture/API.md
    - ../architecture/ERD.md
  release_notes: ../releases/CHANGELOG.md
---

# ADR-001: React PWA and Go Modular Monolith

## Status

- Status: accepted
- Date: 2026-08-24
- Decision approval: user-approved during design review

## Context

The project needs mobile-installable web UX, full offline client behavior, strong financial domain boundaries, and simple homelab operations. The user changed the initial Next.js direction and approved ReactJS with a Go backend.

## Decision

Use a React + TypeScript PWA for the browser and a Go modular monolith deployed as separate API and worker processes that share domain/application packages. Use PostgreSQL as the authoritative database and S3-compatible private object storage.

## Alternatives Considered

- A Next.js full-stack application: rejected because the approved backend is Go.
- One Go API process that also runs all jobs: rejected because deploy/restart behavior can duplicate or interrupt background work.
- Independent microservices: rejected because internal auth, deployment, and observability overhead is not justified for this homelab product.

## Consequences

- Positive: clear frontend/backend boundary, focused domain packages, independent worker lifecycle, and manageable homelab deployment.
- Negative: API contract coordination and two Go entrypoints require discipline.
- Neutral: module boundaries are enforced in one repository rather than by network services.

## Linked Work

- Approved design: [MyPocket System Design](../superpowers/specs/2026-08-23-mypocket-system-design.md)
- Requirements: [REQUIREMENTS.md](../requirements/REQUIREMENTS.md)
- Roadmap: [ROADMAP.md](../work/ROADMAP.md)
- Architecture: [ARCHITECTURE.md](../architecture/ARCHITECTURE.md)
- API: [API.md](../architecture/API.md)
- ERD: [ERD.md](../architecture/ERD.md)
- Integrations: [INTEGRATIONS.md](../architecture/INTEGRATIONS.md)

