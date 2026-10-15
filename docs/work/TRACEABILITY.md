---
artifact_type: traceability_matrix
id: TRACEABILITY
status: active
owner: shared
human_fields:
  - requirement_source
  - release_scope
ai_fields:
  - trace_links
  - evidence_links
  - adr_links
shared_fields:
  - matrix_rows
updated: 2026-10-07
---

# Traceability

## Field Ownership

- Human owns requirement source and release scope.
- AI maintains trace links, evidence links, and ADR links during reconciliation.

Use this file to map requirements to execution, verification, decisions, and releases.

## Trace Matrix

| Requirement | Phase | Ticket/Bug | Detail Design | Test Verification | Docs Review | ADR/Docs | Release |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Native iOS port of all current web finance routes | PHASE-002 | NATIVE-IOS-01 / PORT-100 | [Native SwiftUI port plan](../superpowers/plans/2026-10-05-native-ios-swiftui-port.md), [Port audit](PORT-100-AUDIT.md) | 138/138 native unit tests, bounded API UI sweeps, Release build; native Reports trend, category-tree, base-component and portfolio-trade contracts | [Validation matrix](VALIDATION_MATRIX.md) | [iOS design contract](../design/platforms/ios.md), [Stitch reference](../design/references/STITCH-MY-POCKET.md) | Unreleased |
| Shared web/native design system and Stitch tokens | PHASE-002 | UI-BASE-01 / native parity plans | [Base components](../design/system/BASE_COMPONENTS.md), [Tokens](../design/system/TOKENS.md) | web design/type/build gates, native base-component guards, inspected simulator captures | [Validation matrix](VALIDATION_MATRIX.md) | [ADR-002](../decisions/ADR-002-design-contract-enforcement.md) | Unreleased |
| AI entry, OCR attachments and Finance Assistant contracts | AI-ENTRY-01/02, AI-ADVISOR-01 | AI-ENTRY-01 / AI-ENTRY-02 / AI-ADVISOR-01 | [AI architecture](../architecture/INTEGRATIONS.md), [Assistant spec](../design/pages/assistant/README.md) | backend race suite, live text/tool/schema probes, synthetic OCR pipeline evidence, native AI-entry/Assistant UI fixtures | [Validation matrix](VALIDATION_MATRIX.md) | [ADR-005](../decisions/ADR-005-ai-entry-review.md), [ADR-006](../decisions/ADR-006-private-ai-attachments.md) | Unreleased |
| Feedback, changelog, API-key and owner-scope behavior | FEEDBACK-01 / TICKET-10-01 | FEEDBACK-01 / API-key contract | [Feedback API](../architecture/FEEDBACK_API.md), [API](../architecture/API.md) | backend owner-scope/race tests, local API feedback/key UI tests | [Validation matrix](VALIDATION_MATRIX.md) | [Feedback implementation note](FEEDBACK_LOG.md) | Unreleased |

## Rules

- Every phase should link to at least one requirement, ticket, or bug.
- Every ticket or bug should link to test evidence.
- Every durable decision should link to an ADR.
- Every release should link to completed phases, tickets, bugs, and verification.
