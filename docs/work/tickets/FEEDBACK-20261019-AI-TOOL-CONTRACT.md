# AI tool-call contract hardening

Status: implementation in progress (2026-10-19)

## Scope

The Finance Assistant accepts tool calls from an OpenAI-compatible provider. A
tool call must have a non-empty unique ID within one provider response. The ID
is the correlation key used by the next provider request to associate a tool
result with the model's call. The backend must reject malformed responses
before executing any tool, so a provider cannot cause an uncorrelated or
ambiguous finance query.

This slice does not add write tools, provider selection, fallback providers,
or paid/live-provider calls. It only hardens the existing read/proposal tool
contract.

## Contract

- Trimmed call ID must be non-empty.
- IDs must be unique within one provider response.
- A malformed response returns `ErrAdvisorToolInvalid` and executes none of
  its tools.
- Existing owner validation, tool allow-list, argument schema, and loop limits
  remain unchanged.

## Verification

Unit tests use a deterministic provider stub and assert both the error and that
the finance reader was not invoked. The full backend race suite remains the
release gate; no external provider is called.
