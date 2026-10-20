---
artifact_type: audit_note
id: BACKEND-AUDIT-20261019
status: verified
owner: shared
updated: 2026-10-20
---

# Backend backlog audit — 2026-10-19

## Scope

This audit covers the Go API, repository/use-case boundaries, authenticated
API-key access, OpenAPI route parity, and the backend items listed in
[FEEDBACK-20261019-REMAINING](FEEDBACK-20261019-REMAINING.md). Native iOS code
is out of scope. No production mutation was performed by this audit.

## Result

No independently safe backend implementation gap was found in the current
remaining map. The following items are intentionally still acceptance or
product-contract gates rather than missing code:

| Item | Current disposition | Why this audit does not implement it |
| --- | --- | --- |
| Distributed advisor cancellation | persisted generation/lease checks are present; provider interrupt remains process-local | requires a cross-process delivery decision and failure semantics |
| APNs delivery | token register/revoke API and owner isolation are implemented | provider credentials/signing, delivery retry and opt-out policy are not specified |
| Portfolio trade correction | immutable trades and metadata/price APIs are implemented | correction, backdating, oversell and P&L semantics require an accounting decision |
| Provider/model chooser | single configured provider validates fail-closed | per-user credentials, consent, pricing and budget policy are not approved |
| Real OCR/storage/provider acceptance | local recovery and ownership tests exist | requires configured external services and signed-device UAT |

Implementing any of these without the missing contract would either change
financial semantics or silently expand the security/runtime boundary. They stay
explicitly tracked in the remaining-work map instead of being marked absent.

## Backend checks

The following checks passed against the current worktree:

```text
cd backend && go test ./internal/controller/http ./internal/usecase ./internal/infrastructure/ai ./internal/infrastructure/repository
PASS

cd backend && go test -race ./internal/controller/http ./internal/usecase ./internal/infrastructure/ai ./internal/infrastructure/repository
PASS

AUDIT_BASE_URL=https://money.dungxbuif.com REQUIRE_ROUTE_PARITY=1 scripts/mypocket-production-readonly-audit.sh
health=ok; live OpenAPI paths=88; local OpenAPI paths=88; route/method parity=PASS
```

The route audit is read-only and makes no authenticated request or mutation.
API-key middleware tests cover full same-owner access, expiry/revocation,
owner isolation and the fail-closed ownerless persisted-key case. Finance
advisor tests cover tool-call ID validation, owner-scoped query execution and
write proposals that require explicit confirmation.

The rollout preflight previously defaulted to migration `000037` and printed a
035→036→037-only path even though the checked-in migration history now ends at
`000040`. The preflight is now fail-closed on the complete `000035`–`000040`
set, verifies each disposable down file, and has a shell regression test at
`scripts/mypocket-rollout-preflight.test.sh`. This is a release-gate/tooling
correction only; it does not apply migrations or change production state.

## Reconciliation

- No API, schema, authentication, authorization or runtime behavior changed.
- No production migration was applied. Production schema/version evidence is
  unchanged; the release operator must still run the read-only version check and
  approved backup/migration sequence before enabling `000040` clients.
- `FEEDBACK-20261019-REMAINING.md` remains the source of the five explicit
  contract/provider/device gates above.
- The parent release/rollout note should record this audit alongside the
  production image and migration evidence; this note does not imply physical
  iPhone acceptance or real-provider accuracy.
