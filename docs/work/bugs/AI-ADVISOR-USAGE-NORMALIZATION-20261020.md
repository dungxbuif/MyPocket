# Finance Assistant usage normalization

Status: regression_verified
Date: 2026-10-20

## Scope

Retain useful, privacy-safe provider usage metadata for each bounded model call
in a Finance Assistant run. The persisted usage must include the configured
model, provider request id, latency, finish reason, and numeric token counters,
while the orchestrator also exposes reliable totals across tool-call rounds.

## Current defect

The OpenAI-compatible advisor client currently copies the provider's entire
`usage` object into an unbounded map and does not parse `finish_reason` from the
choice. The orchestrator then keeps only the three token totals and provider
call count, dropping model/request/latency/finish metadata and making it
impossible to audit individual rounds. Raw provider usage fields are also not
safe to persist because they are outside the application's allowlist.

## Design decision

- Keep the existing `AdvisorResponse.Usage map[string]any` boundary to avoid an
  API/schema change, but normalize it at the provider boundary.
- Advisor client usage contains only bounded fields: `provider`, `model`,
  `request_id`, `latency_ms`, `finish_reason`, and numeric
  `prompt_tokens`/`completion_tokens`/`total_tokens`. Unknown provider fields,
  nested usage objects, prompt content, credentials, and response text are
  discarded.
- Parse and bound the choice `finish_reason`; the client must not silently lose
  it. Missing provider usage still produces the bounded request/model/latency
  envelope when `AI_STORE_USAGE=true`.
- Orchestrator aggregation keeps top-level totals plus a bounded `calls` list,
  one normalized map per successful provider response. It never re-persist raw
  provider maps. Existing `AI_STORE_USAGE=false` behavior remains opt-out: the
  client returns no usage map and the run does not store usage.
- Failed provider calls remain a limitation: the provider returns no response
  usage, so failed-call token counts cannot be recovered. Safe failure logging
  retains only error kind and latency, never upstream bodies.

## Verification plan

1. Advisor client tests assert finish reason parsing, numeric token
   normalization, unknown-field/prompt exclusion, bounded metadata, and the
   opt-out switch.
2. Orchestrator tests assert per-call records and totals survive a tool-call
   round trip, including missing usage and multiple provider calls.
3. Run focused AI tests, full Go tests, `go vet ./...`, and `git diff --check`.

## Implementation evidence

- `AdvisorClient` now parses `finish_reason`, bounds provider/model/request
  metadata, and keeps only the three numeric token counters. Unknown and
  nested provider usage fields are discarded before the response crosses the
  provider boundary.
- `AdvisorOrchestrator` normalizes custom provider responses defensively and
  persists bounded `calls` entries while maintaining aggregate token totals
  and provider-call count.
- `AI_STORE_USAGE=false` still returns no usage map and therefore does not
  persist usage. Provider failures still expose only stable error kinds and
  latency to logs; failed-call token usage is unavailable because no provider
  response exists.
- Verification completed: focused package tests and focused `go test -race`
  pass. The existing full Go suite and `go vet ./...` were also passing before
  this change; they must be rerun by the parent after integration with other
  uncommitted agent changes.

