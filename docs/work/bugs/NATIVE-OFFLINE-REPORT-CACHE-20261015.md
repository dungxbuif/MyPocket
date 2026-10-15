---
artifact_type: bug
id: NATIVE-OFFLINE-REPORT-CACHE
status: done
owner: shared
---

# Overview and Reports must use the verified encrypted read snapshot offline

## Evidence

`FinancialSnapshotCache` already stores owner-scoped transactions, wallets,
categories and timezone, but Overview and Reports do not inject or read it.
After a transient read failure they therefore lose report cards even though a
verified snapshot exists. Recomputing report totals locally would duplicate
server policy and is out of scope.

## Bounded design

- Extend the existing encrypted snapshot with optional server-returned
  `MonthSummaryModel` and `InsiderSummaryModel` values; old version-1 files
  remain decodable.
- Overview saves the successful owner-scoped read set and falls back to it on
  transport/408/429/5xx errors, showing a cached timestamp/status.
- Reports falls back only when the cached report month exactly matches the
  requested `YYYY-MM`; it never fabricates another period or calls local
  financial formulas.
- Keep auth failures and validation errors fail-closed. Never cache secrets or
  mutate outbox/financial records.

## Verification result

Report fields round-trip through the existing encrypted cache and a ledger-only
refresh preserves the last verified report payload. Overview fallback restores
owner-scoped wallets/ledger from a transient failure; Reports accepts only an
exact cached month with both server report payloads. The focused cache suite is
9/9 green, Debug build and base/project guards pass, and backend `go test ./...`
passes. No environment, database or deployment change.
