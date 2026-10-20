---
artifact_type: work_note
id: AI-PROVIDER-PROBE-RELIABILITY-20261020
status: in_progress
owner: ai
updated: 2026-10-20
links:
  context: ../../CONTEXT.md
  backlog: ../BACKLOG.md
  validation: ../VALIDATION_MATRIX.md
  provider_uat: ../../operations/PROVIDER_UAT.md
---

# Live provider probe reliability — 2026-10-20

## Problem

`TestLiveToolCallingCompatibility` was explicitly enabled with
`AI_TOOL_CALL_PROBE=1`, but returned from the test body after a transport or
capability error, an invalid response envelope, or a schema mismatch. A
provider HTTP 403 could therefore appear as a passing probe in CI output.

## Decision

An opted-in live probe is a verification gate: any unavailable provider,
malformed envelope, wrong tool call, empty candidate, or invalid transaction
schema fails the test. The probe remains opt-in and provider-free when its
environment flag is absent. Diagnostics must contain bounded metadata only;
the response body, credentials, prompts, and proposed values are not printed.

Other opted-in probes were audited in the same pass. They already fail through
`Fatal`/`Error` on request, envelope, or semantic failures; no production
client/configuration behavior changes.

## Verification

- Unit-test the probe response validator with valid output, invalid JSON,
  missing choices/tool call, wrong tool name, empty arguments, and malformed
  proposal output.
- Run the focused AI package tests with live flags absent; all live probes must
  skip rather than call a provider.
- When credentials are available, run `AI_TOOL_CALL_PROBE=1` against a
  synthetic staging endpoint. HTTP 403 must fail the test rather than log and
  return.

## Release note

This is a test-harness reliability fix only. It does not change the runtime AI
client, provider selection, API, database schema, or production data.
