# Tài sản tích lũy — local UAT record

## Covered by automated evidence

- Unit is selected from the backend-owned gold catalog; free-text unit labels
  are rejected. Initial options are `chỉ`, `lượng`, `phân`.
- The add-asset form does not preselect the first unit; the Save action stays
  disabled until the owner explicitly chooses one.
- The native trade form labels quantity and unit price explicitly and previews
  `quantity × unit price ± fee`.
- Trade history exposes `Sửa giao dịch`; automatic assets hide manual price
  editing and show source/status metadata.
- Unknown market value remains distinct from zero in summary and masking.

## Evidence

- `backend`: `go test ./...` — PASS.
- `mobile-ios`: full `MyPocketIOSTests` — PASS, 442 tests in the latest run.
- `verify-project.sh`, `check-base-components.sh`, `git diff --check` — PASS.
- Debug app installed and launched on simulator
  `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`.

## Populated API-backed simulator UAT

Against the disposable migrated PostgreSQL/API fixture, the following flows
passed with persisted readback: create and manual price refresh, compact
summary/position metrics, rename without changing asset ID, editor failure and
cancel without writes, deletion confirmation for an unused asset, and buy
trade refresh including quantity, fee and history. The shared grouped amount
field also now keeps raw digits while editing and synchronizes the final value
before save; this was verified with a 120000 price and a 2500 fee.

## Not claimed

This is not owner sign-off for live BTMC/BTMH quotes, worker freshness, or
large Dynamic Type/dark-mode review. The populated UAT used disposable local
data only; no production database or source deployment was changed.
