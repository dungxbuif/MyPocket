---
title: "Finance Assistant model upgrade"
status: in_progress
owner: ai
created: 2026-10-20
---

# Purpose

Switch the configured OpenAI-compatible Finance Assistant provider model to
`DeepSeek-V4.1-Flash` without changing the tool contract, owner isolation,
confirmation flow, or persisted conversation schema.

## Runtime contract

- The backend remains the only component that holds `AI_API_KEY`.
- `AI_BASE_URL` and `AI_API_KEY` remain environment-provided; no credential is
  added to source control or an iOS/web bundle.
- `AI_MODEL` selects the provider model for both the advisor and AI-entry
  extraction clients. The provider must continue to return one OpenAI-compatible
  chat completion with validated tool calls or a structured extraction result.
- Existing provider timeout, response-size cap, tool-call cap, confirmation
  requirement, and usage persistence remain unchanged.

## Verification plan

1. Load the local ignored environment with `AI_MODEL=DeepSeek-V4.1-Flash`.
2. Run the provider configuration/client tests and a bounded live model smoke
   test only when the configured endpoint/key are available.
3. Run the backend AI/unit suite and confirm no tool, owner, or response-card
   contract changes.
4. Reconcile public integration docs and release notes with the verified model
   name; production rollout remains a separate deployment gate.

## Explicit non-goals

- No new autonomous write actions.
- No provider fallback or model-selection API.
- No production secret rotation or database migration.

## Verification — 2026-10-20

The ignored local `AI_MODEL` was set to `DeepSeek-V4.1-Flash`.
The explicit synthetic extraction smoke returned HTTP 403. The tool-call probe
also returned HTTP 403, despite reporting PASS because its current implementation
logs the rejection and returns normally. Neither result establishes compatibility
or quality of this model. Production has not been upgraded by this change.
Investigate provider access before restarting a serving backend with this model.

The same synthetic extraction probe with the existing endpoint/key and
`DeepSeek-V4-Flash` passed (one draft, 863 ms provider latency). This narrows the
failure to access to the requested model rather than a general endpoint/key
outage; it does not identify the provider's exact authorization rule. Local
configuration was restored to the verified working model to avoid breaking the
next backend restart. The requested V4.1 upgrade remains pending provider access.

After the owner enabled access, the same probes were rerun successfully:
synthetic extraction returned one draft (5.802 s), and the tool probe returned
one schema-valid `propose_transactions` call (1.791 s). Local `AI_MODEL` is now
`DeepSeek-V4.1-Flash`. This proves basic extraction/tool wire compatibility,
not the full advisor quality suite, a running-process restart, or production rollout.
