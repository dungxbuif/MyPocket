# MyPocket implementation reconciliation — 2026-11-01

## Verified in this workspace

| Workstream | Implemented evidence | Verification |
| --- | --- | --- |
| Tài sản tích lũy / unit | Backend owns the initial `gold` catalog; create requires an explicit user-selected unit and accepts only `chỉ`, `lượng`, `phân`. The native add form no longer preselects the first unit. Automatic catalog products currently use `chỉ`. Native trade form shows quantity + unit price + fee and computes gross/net preview. | `go test ./...`; catalog handler/entity tests; `FinanceModelTests` catalog, quantity × unit-price, fee and partial-valuation tests; `portfolio_test.go`. |
| Automatic gold pricing | Provider contract, two sanitized source adapters, product mapping, scheduled poll/lease and stale/unknown provenance are present. | Backend market/provider/worker tests pass; production worker reported `providers=2 claimed=2 skipped=0 accepted=4 rejected=0 failed=0` after rollout. Freshness remains provider-dependent. |
| AI Assistant | Full-screen route, no feedback bubble/bottom navigation on Assistant, recoverable entry task, typed batch decisions, generation/version fields, and read-only structured cards are present. | Native suite and backend suite pass; route/API contracts are additive. |
| AI entry preview cards | Proposal rows use the shared reference surface and shared transaction presentation used by Add Transaction; action state remains server-authoritative. | Native model/UI tests and existing proposal fixtures pass. |
| Reports / Money Insider | Backend report read model, deterministic insights, snapshot/analysis lifecycle, native chart/report models and retention worker are present. | Backend full suite and native suite pass; report insight/overflow tests added in this reconciliation. |
| Icons | 37 Phosphor Icons Regular vectors are bundled locally, mapped through `MPIconCatalog` for every semantic role, and licensed in-app. Legacy API icon keys remain unchanged; unknown keys map to tag. | `actool`/Xcode build succeeds; catalog test asserts all `MPIconID` values resolve to bundled assets. |
| Shared styling | Composer metrics, cards, rows and icons are routed through DesignSystem bases; composer now has a useful multiline minimum without a fixed editor height. Transactions uses the shell Add action only; its duplicate local `+` is removed. Grouped amount fields keep raw digits while editing and synchronize the final draft before save. | `verify-project.sh`, `check-base-components.sh`, native 432-test suite, grouped-input regression tests, `git diff --check`. |

## Still not verified / intentionally deferred

- Disposable PostgreSQL migration replay reached migration `000052`; full
  backend tests and the API-backed UAT fixture ran against a fresh disposable
  database. Production query-plan/EXPLAIN evidence is still deferred.
- The candidate is deployed to both production services at the immutable
  `sha256:e18a80cd3d7a406744c61ec7150a8c83ca68e74618472ded56729799389a8883`
  digest. Production migration is `52 dirty=false`, health is green and route
  parity is `96/96`. Provider freshness still needs an owner/provider gate.
- Populated simulator UAT now covers Portfolio, Reports/Money Insider and
  Assistant overview/history flows using disposable seeded data. Owner visual
  sign-off and large Dynamic Type review remain open.
- Dark mode remains outside this milestone by owner decision. Large Dynamic Type
  and light-mode unit tests are covered; dark-mode acceptance is not claimed.
- A future icon pass may add verified duotone layers. Current `.duotone` style
  intentionally reuses the verified regular vector rather than mixing guessed
  assets.

## Commands run for this reconciliation

```text
backend: go test ./...                         PASS
native: xcodebuild MyPocketIOSTests            PASS (442 tests)
native: DesignSystemTests + FinanceModelTests  PASS (latest focused run)
mobile-ios/scripts/verify-project.sh            PASS
mobile-ios/scripts/check-base-components.sh     PASS
git diff --check                               PASS
```

This is an implementation and rollout record; owner/device/provider acceptance
and feedback resolution remain separate from deployment evidence.
