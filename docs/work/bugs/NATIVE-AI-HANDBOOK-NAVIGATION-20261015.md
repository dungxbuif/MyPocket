---
artifact_type: bug
id: NATIVE-AI-HANDBOOK-NAVIGATION
status: done
owner: shared
---

# Finance Assistant handbook must match native navigation

## Evidence

The backend handbook injected into every Finance Assistant prompt still says
that the Assistant is under `More`, while the native shell exposes it as the
primary `Trợ lý` tab. This can make the assistant give a wrong navigation
answer even though no financial data or API contract is involved.

## Bounded fix

- Describe `Trợ lý` as a primary shell tab.
- Keep `More` only for account/secondary destinations.
- Add a provider-level regression that rejects the stale navigation claim and
  requires the primary-tab wording.
- No API, schema, permission, provider, or environment change.

## Verification result

The handbook is now version `v1.1 (2026-10-15)` and describes `Trợ lý` as the
primary tab. `TestAdvisorOrchestratorHandbookMatchesPrimaryAssistantTab` was
RED before the change and GREEN after it; `go test ./...` also passes. The
architecture/backlog/context/changelog and human/agent guide references were
reconciled. No API, schema, permission, provider, or environment change was
made.
