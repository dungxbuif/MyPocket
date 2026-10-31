# Plan completion matrix — 2026-11-01

This matrix reconciles the four plans created for the current MyPocket slice.
Older dated plans under `docs/superpowers/plans/` are historical workstreams;
they are not silently treated as completed by this slice.

| Plan | Implemented locally | Fresh proof | Remaining before owner sign-off |
| --- | --- | --- | --- |
| accumulated-assets | Backend-owned gold catalog; explicit required unit; only `chỉ`, `lượng`, `phân`; quantity × unit price × fee; manual/automatic source binding; quote polling/adapters; correction/edit path; unknown/stale valuation state. The add form now does not preselect the first unit. | Fresh migrated PostgreSQL backend suite with `TEST_DATABASE_URL`; native 442-test suite; populated API UAT for create/price, compact position metrics, rename, editor cancel/error, delete, trade quantity/fee/history; project/base guards. | Real BTMC/BTMH provider freshness and owner review of populated screenshots. |
| mypocket-ui-density-chat-icons | Shared composer/card/icon bases; full-screen Assistant with bubble hidden; adaptive composer; Phosphor Regular vectors bundled and licensed; duplicate Transactions `+` removed in favor of shell Add; older-history restore uses a measured message anchor plus temporary relative spacer; grouped amount input now preserves raw digits while editing; icon manifest/guard and 38-role catalog coverage. | Native 442-test suite; focused DesignSystem 114/114; `verify-project.sh`; `check-base-components.sh`; `check-icon-catalog.sh`; asset build. | Large Dynamic Type and populated-device visual review; duotone remains deferred; dark mode remains out of scope. |
| reports-insider | Query-backed report/read model, charts, deterministic grounded insights, revision/snapshot analysis lifecycle, bounded retention, native analysis states; dedicated native report/analysis tests and repository/retention test files. | Fresh migrated PostgreSQL backend suite with `TEST_DATABASE_URL`; native 442-test suite; API-backed UAT for Money Insider, comparison ring, persisted income/expense trend and report filtering; production migration 52 and route parity. | Provider latency/call-count measurements, EXPLAIN evidence, and owner visual sign-off. |
| unified-ai-assistant | Durable entry lifecycle, typed batches/versions, server-authoritative action decisions/replacement, dedicated Assistant route, shared preview cards, adaptive composer and recovery states; lifecycle/action/batch/unknown-part/attachment test coverage. | Backend `go test ./... -count=1`; native 442-test suite; populated overview UAT; disposable 51-message history UAT covering older-page loading, anchor preservation and restart persistence; replacement route contract test; project/base/icon guards; production migration 52 and route parity. | Provider formatting/action-state review, attachment/provider UAT, and owner visual sign-off. |

## Verification run

- Backend: `rtk proxy go test ./... -count=1` — PASS.
- Backend with fresh migrated PostgreSQL 16 disposable container and `TEST_DATABASE_URL` — PASS, all packages; migration history reaches `000052`.
- Native: `rtk proxy xcodebuild ... -only-testing:MyPocketIOSTests test` — PASS, 442 tests.
- Native focused: `DesignSystemTests` — PASS, including grouped amount raw-edit and icon catalog coverage.
- Native UI compile: `build-for-testing` — PASS, including updated `LaunchTests.swift`.
- Native project/base guards — PASS.
- Icon catalog guard — PASS, 38 roles and 20 comparison roles.
- `git diff --check` — PASS.
- Debug app installed and launched on simulator `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D` after the final source changes.

Production feedback audit returned 55 records. After the owner's explicit test
request, all 55 are owner-resolved (`resolved=true`, `resolved_at` present),
while developer status remains `status=open`. Production exposes 96 routes,
matching the candidate. Remaining device/provider acceptance is recorded in
[PRODUCTION-FEEDBACK-PLAN-RECONCILIATION-20261101](PRODUCTION-FEEDBACK-PLAN-RECONCILIATION-20261101.md).

Live quote freshness, seeded populated screenshots, large Dynamic Type and
dark-mode acceptance remain explicitly unclaimed.
