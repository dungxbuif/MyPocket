# Prepared Finance Assistant action dispatcher

Status: locally implemented and regression-tested; production rollout remains open.

## Goal

Make the existing wallet proposal a real server-confirmed action without
granting the model write access. The assistant persists the proposal inside
the owner-scoped assistant message; the client sends only the action identity
and conversation generation when the user taps Confirm.

## Contract

`POST /api/v1/ai/advisor/actions/{action_id}/confirm` accepts:

```json
{"conversation_id":"...","generation":1}
```

The backend searches only that owner's conversation and generation for an
assistant `wallet_creation_proposal` with the exact action ID. It derives the
wallet input from the persisted proposal, ignores client-supplied financial
fields, and calls the existing durable wallet mutation boundary with the
action ID as `Idempotency-Key`. Replays return the same wallet; stale,
cross-owner, unknown, or malformed actions fail closed. The model and the
client cannot select an arbitrary wallet payload through this endpoint.

The response is the ordinary domain envelope. The local dispatcher now also
supports typed jar create/monthly update and budget create proposals. Jar
updates delegate to the owner-scoped jar PUT route; budget creates use the
`budget.create` mutation ledger. Later domain actions must add their own typed
proposal and mutation adapter rather than widening this endpoint into arbitrary
SQL or a generic write proxy.

## Client behavior

Native wallet cards prefer the dispatcher when a persisted action ID and
conversation context are present. Legacy proposal messages without that
metadata retain the existing ordinary wallet endpoint fallback.

## Verification

- Red/green handler tests cover owner and generation binding, proposal payload
  tampering, replay, stale action and missing durable adapter.
- Native API and card wiring use the dispatcher when persisted action context is
  present; legacy messages retain the ordinary wallet fallback.
- Focused backend and native tests pass locally. No production deployment,
  migration or device install is part of this slice.

## Evidence (2026-10-15)

- `go test ./...` and `go vet ./...` pass after Swagger regeneration.
- Handler regression covers successful persisted-payload dispatch, stale
  generation (`409`) and strict client-field rejection (`400`).
- Native API action endpoint, feedback refresh race, and design-system action
  context tests pass; signed full native unit suite is green (294 tests,
  DesignSystem 90/90).
- Release iPhone build with `CODE_SIGNING_ALLOWED=NO` and warnings-as-errors
  compiles successfully. Signing/device installation and production rollout
  remain intentionally out of scope.
