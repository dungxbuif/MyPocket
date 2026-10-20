---
artifact_type: adr
id: ADR-005
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
  requirements: [REQ-F-014, REQ-F-015, REQ-NF-004]
  phase: PHASE-007
  tickets_or_bugs: [TICKET-022, TICKET-024]
  detail_design: ../superpowers/specs/2026-08-23-mypocket-system-design.md
  master_docs:
    - ../requirements/SPEC.md
    - ../architecture/ARCHITECTURE.md
    - ../architecture/API.md
    - ../architecture/ERD.md
  release_notes: ../releases/CHANGELOG.md
---

# ADR-005: Restricted Append-Only Audit Log

## Status

- Status: accepted
- Date: 2026-08-24
- Decision approval: user-approved during design review

## Context

The user requires state-changing and security actions to be available for debugging on a hidden page visible to only one account. Logging all UI activity would add noise and privacy exposure.

## Decision

Record append-only state-changing domain actions and security events with redacted diffs, outcomes, actors, sources, correlation IDs, and safe errors. Authorize the hidden read-only API/page only when the verified Google email exactly equals AUDIT_VIEWER_EMAIL. Retain records for AUDIT_RETENTION_DAYS, default 180, and purge expired rows in bounded worker batches.

## Alternatives Considered

- Audit all clicks/page views: rejected as noisy and privacy-heavy.
- Finance-only audit: rejected because sync, auth, provider, deletion, and webhook events are needed for debugging.
- Hide the route without API authorization: rejected as insecure.
- Keep records forever: rejected due unbounded privacy and storage exposure.

## Consequences

- Positive: useful cross-module debugging with narrow access and correlation.
- Negative: one environment-configured account is an operational dependency and retention deletion is destructive maintenance.
- Neutral: route obscurity improves discoverability only; authorization is always enforced in Go.

## Linked Work

- Approved design: [MyPocket System Design](../superpowers/specs/2026-08-23-mypocket-system-design.md)
- Requirements: [REQUIREMENTS.md](../requirements/REQUIREMENTS.md)
- Roadmap: [ROADMAP.md](../work/ROADMAP.md)
- Architecture: [ARCHITECTURE.md](../architecture/ARCHITECTURE.md)
- API: [API.md](../architecture/API.md)
- ERD: [ERD.md](../architecture/ERD.md)
- Integrations: [INTEGRATIONS.md](../architecture/INTEGRATIONS.md)

