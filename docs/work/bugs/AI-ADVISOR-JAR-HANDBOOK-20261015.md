---
artifact_type: bug
id: AI-ADVISOR-JAR-HANDBOOK-20261015
status: deployed_pending_uat
owner: shared
---

# Finance Assistant jar proposal unavailable on production

## Production feedback

- Feedback: `824dea29-0a7b-45ca-976d-038616c12448`
- Created: `2026-10-15T05:32:03.336232Z`
- Report: `AI đang ko tạo đc hũ. Cho full quyền access mọi api của user đi`

## Reproduction

Ask the production Finance Assistant to create a spending jar before the
hotfix. The production API container was running
`registry.dungxbuif.com/mypocket-api:auth-12572c87`.
Inspecting its `/app/api` binary finds no `propose_create_jar` or
`jar_creation_proposal` symbols, while the current local source registers that
tool. The local handbook also begins with a generic “read-only” label, which
can cause a provider to refuse a proposal request instead of selecting the
available proposal tool.

## Expected behavior

The assistant must explicitly say that it can prepare a jar proposal, select
`propose_create_jar`, show a confirmation card, and only create it after the
owner confirms. It must not answer “không có quyền” for this supported proposal
flow. It must continue to refuse unconfirmed edits/deletes and unrelated
financial writes.

## Scope and fix

Reconcile the handbook text with the implemented create-jar proposal boundary,
add a provider-level regression assertion, and rebuild the production API image
from the source revision that contains the tool. No owner-scope expansion or
direct model write is allowed: `propose_create_jar` remains a non-mutating
proposal and confirmation remains owner/generation-bound.

Follow-up hardening: put the allowed proposal capability before the generic
read-only wording and state it as an explicit authorization. This prevents
OpenAI-compatible models from treating “read-only” as “no action capability”.
The model must say it can prepare a jar proposal, call
`propose_create_jar`, and render the confirmation card; only the confirmed
owner action may mutate the jar.

## Verification

- Backend usecase provider tests must assert the handbook names
  `propose_create_jar`, explicitly instruct the model to select it instead of
  refusing, and preserve the no-direct-write boundary.
- Production verification must show the deployed API is healthy and the image
  contains the proposal contract. The current evidence includes the public
  health check and a runtime binary contract probe; an owner-authenticated
  capabilities/tool probe is still the UAT gate, not claimed by this fix.
- Local tests and build were required before rollout. Production feedback status
  and changelog remain unchanged until the owner accepts UAT.

## Rollout evidence

- Image: `registry.dungxbuif.com/mypocket-api:ai-jar-capability-20261015-215712`
- Digest: `sha256:20fee27c5da418ecb23709484e9d8a844cbf1e9471592df74415ca0e831d6da0`
- Swarm: `mypocket-api` converged at `1/1` on `vm100-prod`; worker and web were
  not updated.
- Public health: `GET https://money.dungxbuif.com/api/v1/health` returned `200`.
- Runtime contract probe: the deployed `/app/api` binary contains
  `propose_create_jar`, `jar_creation_proposal` and the explicit
  `ĐƯỢC PHÉP chuẩn bị đề xuất tạo hũ` capability wording.
- Remaining gate: owner-authenticated UAT of “tạo hũ” and confirmation card;
  feedback remains open until that UAT is accepted.

