---
artifact_type: detail_design
id: AI-ENTRY-03
status: draft
owner: shared
approval: pending_owner_decisions
parent: AI-ENTRY-03-PROVIDER-SELECTION.md
human_fields: [approval, provider_ownership, budget_policy, allowed_providers]
ai_fields: [architecture_options, price_freshness, capability_matrix, verification]
shared_fields: [scope, trace]
trace:
  ticket: AI-ENTRY-03-PROVIDER-SELECTION.md
  parent: TICKET-09-tro-ly-ai.md
  backlog: ../BACKLOG.md
  validation: ../VALIDATION_MATRIX.md
  integrations: ../../architecture/INTEGRATIONS.md
  adr_registry: ../../decisions/README.md
  release_notes: ../../releases/CHANGELOG.md
---

# AI provider selection and price transparency — draft for review

## Intent and boundary

Design a future user choice among supported AI providers/models with understandable pricing before an AI request. This document records the 2026-09-22 owner request; it does not authorize provider onboarding, credential changes, or production cost.

The current AI entry remains a server-configured OpenAI-compatible endpoint. It sends text and a bounded owner wallet/category catalog only; OCR handles files first. Do not change that runtime while this design is `draft`.

## Proposed design direction

- Define a provider adapter contract for model extraction and an explicit capability record (strict JSON Schema support, limits, endpoint style, timeout behavior). Do not equate the OpenAI-compatible label with identical capabilities.
- Keep provider secrets server-side. Choose explicitly between app-managed credentials, user-managed credentials (BYOK), or both before designing account storage and UX.
- Maintain model prices as sourced data: provider's first-party pricing URL, currency/unit, input/output rates, effective/verification timestamp, and stale/unknown state. Do not hard-code a rate without provenance or present stale prices as current.
- Estimate cost from request token usage and selected model rates; state that estimates can differ from billed amounts. Show OCR charges separately where applicable. Decide whether cost confirmation or a user budget cap is required.
- Make provider/model selection explicit per request or as a saved preference (owner decision). No silent provider fallback, because that changes privacy, capability and price.
- Before selection, disclose the types of data sent (including OCR text and any wallet descriptions), provider processing/retention information, and any regional constraints that are known.
- Run the transaction-extraction evaluation suite for each proposed model; require schema validity, field accuracy, transfer-safety and latency thresholds before marking a model selectable.

## Required review decisions

1. Credential ownership: app-managed, BYOK, or both.
2. Selection scope: per request, account default, or both.
3. Initial provider/model list and allowed processing regions.
4. Price freshness policy and whether to block or warn when price data is stale.
5. Cost consent and budget/usage limits; who pays and how billing is surfaced.
6. Which wallet metadata may be sent to a provider, and whether users can opt out.

## Verification contract to detail after approval

- Contract tests for provider capability discovery/declared support and safe error handling; no automatic fallback.
- Price data tests for source, currency/unit, effective date, staleness and estimate arithmetic.
- UI tests proving selected model, price timestamp, estimated cost and data-sharing notice are visible before submission.
- Live synthetic extraction evaluation and latency/price estimate comparison per supported model; no ledger writes.
- Secret scans and tests proving provider credentials never reach frontend payloads/logs.

## Trace and next step

Back: [AI-ENTRY-03 ticket](AI-ENTRY-03-PROVIDER-SELECTION.md), [parent TICKET-09](TICKET-09-tro-ly-ai.md), [AI-ENTRY-01](TICKET-09-01-ENTRY-DETAIL_DESIGN.md). Forward after owner review: acceptance tests, [validation matrix](../VALIDATION_MATRIX.md), docs review, any required ADR in the [ADR registry](../../decisions/README.md), and [release notes](../../releases/CHANGELOG.md). Reconcile the provider inventory in [integrations](../../architecture/INTEGRATIONS.md).

No implementation or cost-incurring provider request is authorized by this draft.
