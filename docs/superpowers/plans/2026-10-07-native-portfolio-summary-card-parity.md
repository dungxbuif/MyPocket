# Native portfolio summary-card parity

## Intent

Match the native Portfolio summary header to the web `PortfolioPanel`: keep the
supporting subtitle below the page title and group the three headline metrics
inside a shared reference card titled `Tổng danh mục`.

## Contract

- Preserve portfolio summary/asset/trade/price APIs, payloads, owner scope and
  existing action-card order.
- Compose only shared `MPSurfaceCard`, `MPSectionHeader`, `MPText` and existing
  metric bases; no screen-local colors or shapes.
- Add stable copy/accessibility contracts for the subtitle and summary heading.

## Verification plan

- Add the copy/section contract test and capture RED.
- Implement the subtitle and reference summary card.
- Run focused/full native tests, Release warnings-as-errors, UI smoke already
  covering the portfolio route when available, guards and web design checks.
- Reconcile iOS design docs, validation matrix, port audit, context and
  changelog.

## Status

- [x] Contract test added and red run captured
- [x] Native summary anatomy implemented with shared bases
- [x] Native unit, Release and guards/web checks pass
- [x] Documentation reconciled

## Final evidence

- RED compile evidence: `/tmp/mypocket-portfolio-summary-red`.
- Focused GREEN copy contract: `/tmp/mypocket-portfolio-summary-green`
  (1/1).
- Full native unit/contract rerun: `/tmp/mypocket-portfolio-summary-full`
  (108/108).
- Release warnings-as-errors build: `/tmp/mypocket-portfolio-summary-release`
  (`BUILD SUCCEEDED`).
- Project/base-component guards, `git diff --check` and web design checks pass.
- The disposable API portfolio UI rerun was stopped by the known Xcode runner
  hang before a result bundle completed; the previously recorded portfolio
  asset create/price-refresh UI proof remains valid. Authenticated visual UAT
  remains open.
