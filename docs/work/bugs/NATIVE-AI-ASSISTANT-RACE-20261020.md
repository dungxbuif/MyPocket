---
artifact_type: bug
id: NATIVE-AI-ASSISTANT-RACE-20261020
status: regression_verified
owner: mobile-ios
---

# Native AI Assistant reliability guards

## Scope

Harden the native Finance Assistant and AI transaction-entry flows against
duplicate submits, ambiguous network outcomes, stale history responses, and
polling that continues after cancellation. Preserve the existing API and
confirmation contracts; no financial data or server schema is changed.

## Decision

- Reuse the user-visible advisor `client_request_id` as the HTTP idempotency
  key. A retry after a transport/decoding/5xx failure therefore replays the
  same server request instead of creating a second run.
- Keep uncertain submissions disabled until the user explicitly retries the
  same request. The composer never silently clears the draft.
- Tag history requests with a token and conversation identity. A late response
  from a reset or different conversation cannot repopulate the current view.
- Tag advisor polling with the active run identity and invalidate the token on
  cancellation/reset.
- Lock each AI-entry proposal decision independently and block the global save
  action while another decision is in flight.

## Verification plan

- Unit tests cover local double-submit, stable advisor idempotency keys,
  history conversation/generation matching, recovery classification, and
  proposal decision locking.
- Run focused native tests, then the complete `MyPocketIOSTests` suite,
  `git diff --check`, and project/base-component verification.

## Reconciliation

This note is intentionally limited to native client reliability. It does not
change provider behavior, permissions, API schemas, deployment, or production
configuration.

## Implementation and verification evidence

- `AssistantView.send()` now rejects a second local submit while the first
  request is still in flight and retains one client request ID through an
  uncertain advisor submit and `FinanceAPI` sends that value as the
  `Idempotency-Key` on both initial and retry requests.
- History loads reject late responses after a conversation/token change;
  advisor polling is invalidated on cancellation/reset.
- AI entry submission now checks both server processing and local request
  state; each proposal decision has its own in-flight lock and cannot race the
  global save action. The global save action also refuses to start while an
  individual decision is in flight.
- Initial focused AI regressions: 5/5 passed; follow-up stale-error and
  save-all race regressions: 2/2 passed. The advisor local-send regression is
  also included in the full suite.
- Full `MyPocketIOSTests`: 374/374 passed.
- `git diff --check`, `mobile-ios/scripts/check-base-components.sh`, and
  `mobile-ios/scripts/verify-project.sh`: passed.

## Residuals

- This change has not been deployed and has no real-device touch regression
  evidence. Parent release work must perform device/UAT verification.
- The client cannot prove whether a network failure occurred before or after
  provider execution; it deliberately requires an explicit same-ID retry or
  history refresh rather than silently resubmitting.
