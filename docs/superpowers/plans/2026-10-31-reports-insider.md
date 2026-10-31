# Reports and Money Insider Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Báo cáo tháng có biểu đồ từ query backend, Money Insider có insight kiểm chứng được và bài phân tích AI lưu theo phiên bản dữ liệu.

**Architecture:** PostgreSQL aggregate cùng snapshot cho chart/fact bundle; revision tài khoản làm cache invalidation. Native tiêu thụ read model có giới hạn thay vì tải ledger để dựng báo cáo. AI chạy async khi người dùng yêu cầu, chỉ nhận facts và dùng lại kết quả cùng revision.

**Tech Stack:** Go/Gin/GORM/PostgreSQL, SwiftUI, XCTest/XCUITest, provider adapter hiện có.

**Spec:** [Reports/Insider design](../specs/2026-10-31-reports-insider-design.md).

> **Reconciliation 2026-11-01:** Query-backed monthly reports, chart read
> models, grounded insights, durable analysis/snapshot lifecycle and bounded
> retention are implemented. Fresh migrated-PostgreSQL backend tests and the
> 442-test native suite pass; production now runs migration `000052` with route
> parity `96/96`. Remaining gates are provider cost/latency evidence and owner
> visual acceptance, not a missing report implementation.

## Global Constraints

- Màu hiện tại, light mode, compact shared tokens. Touch target không bị thu nhỏ theo padding.
- Không có raw transaction list mặc định; chart và query không phụ thuộc provider.
- Thặng dư = thu − chi; không gọi là số tiền có thể tiêu hoặc tiền thực sự tiết kiệm.
- Giữ canonical transfer/adjustment/goal/budget/jar accounting; ngân sách chồng scope không cộng tổng.
- Mọi API owner-scoped, additive; không sửa migration đã áp dụng. Web archived ngoài scope.
- AI: tối đa 30 facts, 3 quan sát + 3 gợi ý, 1 provider call/job, không auto retry, 1 active analysis/owner, cooldown 60 giây. Cron AI ngoài scope.
- Snapshot/analysis giữ 90 ngày từ creation; thay dữ liệu không tự gọi AI lại; mask cả chart, prose và VoiceOver.
- Ngưỡng insight comparison: 7 ngày chung, 5 giao dịch mỗi kỳ, |delta| ≥ 100.000 VND và |change| ≥ 15%; baseline 0 không phần trăm. Đây là defaults đề xuất trong design.

## Review Focus

- Sửa ngày giao dịch từ tháng A sang B phải làm report cả hai tháng hết hiệu lực — Task 2.
- Nhiều category con/tên trùng và ngân sách chồng nhau không được nhân đôi tổng — Task 1.
- Đổi ngày tài khoản qua midnight dù không có mutation vẫn làm snapshot hiện tại được refresh — Task 2.
- Provider hoàn tất sau khi revision đã đổi không được gắn bài cũ thành phân tích mới — Task 5.
- Chart chọn category parent phải mở cùng subtree/scope, kể cả khi cache cũ — Task 4.

## Execution rules

Tất cả Go commands chạy tại `backend/`; native commands tại repo root. Simulator UDID: `EC80B092-CEF1-4A8E-A5A3-7420298EFD4D`. Register mọi Swift file mới trong project Sources; exit 0 với 0 test không tính PASS. Giữ dirty owner UI changes; commit chỉ diff thuộc task, không stage nguyên project file nếu có phần không thuộc task. Mỗi task có RED → GREEN → review, commit scoped sau GREEN. Dùng disposable DB, không production.

Report plan độc lập chat plan; không sửa Task 1 chat findings bằng cách trộn vào report. User bổ sung yêu cầu portfolio: số lượng + đơn giá/đơn vị + tổng tiền, đường sửa dễ tìm. Ghi là workstream riêng cần design; chỉ liên hệ report qua định nghĩa tài sản, không tự sửa portfolio trong các task dưới.

### Task 1: Monthly read model và accounting query

**Files:** Create `backend/internal/entity/monthly_report.go`, `monthly_report_test.go`; `backend/internal/repository/monthly_report.go`; `backend/internal/infrastructure/repository/monthly_report_postgres.go`, `monthly_report_postgres_test.go`; `backend/internal/usecase/monthly_report.go`, `monthly_report_test.go`; `backend/internal/controller/http/monthly_report_handler.go`, `monthly_report_handler_test.go`. Modify `backend/cmd/api/main.go`, `backend/internal/infrastructure/repository/finance_query_postgres.go`, `backend/internal/infrastructure/repository/insider_postgres.go` only to share canonical eligibility logic, retaining legacy payloads.

**Interfaces:** `MonthlyReportScope{Month, WalletID, Granularity}`; `MonthlyReportReader.ReadMonthly(ctx context.Context, ownerID string, scope entity.MonthlyReportScope, now time.Time) (entity.MonthlyReport, error)`; `MonthlyReportService.Read(ctx, principal, scope)`; GET `/api/v1/reports/monthly`. Report includes normalized scope/ranges/timezone/as_of, totals, comparison, series, breakdown/contributors, scoped budget/jar preview, quality flags and typed drilldown. Facts use integer money and BPS; unknown distinct from zero.

- [ ] Write `TestMonthlyReportGoldenAccounting`: income 1.000.000, parent/child expenses 100.000+50.000, excluded 20.000, transfer pair 200.000; expect income 1.000.000/expense 150.000/net 850.000 and parent subtotal 150.000. Add duplicate category names, overlapping budgets and foreign wallet fixtures.
- [ ] Run `rtk proxy go test ./internal/entity ./internal/usecase ./internal/infrastructure/repository ./internal/controller/http -run MonthlyReport -count=1`; expect missing behavior RED.
- [ ] Implement aggregate in repeatable-read transaction with current domain rules and category-tree rollup. Weekly buckets Monday-based clipped at month bounds, ≤6; day ≤31; categories top5+Khác; contributors 5 increase+5 decrease; budget/jar preview5 with count and explicit scope. Current comparison excludes today, clamps to shared days; completed month uses whole periods. Reject future/invalid months. No full ledger response.
- [ ] Add `TestMonthlyReportCalendarComparison` for DST, leap day, 31→28, day1, zero baseline/income and no-data vs database error; run preceding command GREEN with migrated disposable PostgreSQL. Record EXPLAIN for 100.000 synthetic rows; add index only if plan evidence warrants it.
- [ ] Commit `feat: add deterministic monthly report query` with this task's explicit files; record real DB tests and skips separately.

### Task 2: Durable revisions and report snapshots

**Files:** Create `backend/internal/entity/report_snapshot.go`; `backend/internal/repository/report_snapshot.go`; `backend/internal/infrastructure/repository/report_snapshot_postgres.go`, `report_snapshot_postgres_test.go`; `backend/internal/usecase/report_snapshot.go`, `report_snapshot_test.go`; new forward migration `backend/migrations/000043_report_snapshots.up.sql`/`.down.sql` (verify latest version before assigning); `docs/decisions/ADR-027-report-snapshots.md`. Modify monthly service/handler, migration registry tests/README.

**Interfaces:** `ReportSnapshotService.Read(ctx, principal, scope) (ReportSnapshot,error)` and `GetOwned(ctx, ownerID, snapshotID)`. Snapshot has ID, owner, scope, revision, rule version, as_of, expires_at, payload. Unique logical key includes owner/scope/timezone/granularity/rules/revision and local as-of day for current period.

- [ ] Write `TestReportSnapshotInvalidatesBackdatedMove`, `TestReportSnapshotConcurrentMutationNeverRelabelsOldFacts`, `TestReportSnapshotMidnightRefresh` and owner/rollback/expiry tests. Run `rtk proxy go test ./internal/usecase ./internal/infrastructure/repository ./migrations -run 'ReportSnapshot|Migration' -count=1`; expect RED.
- [ ] Inventory existing canonical writers and implement SQL revision triggers in the forward migration so AI-entry/import/recurring/correction paths also invalidate in the same transaction. Cover transactions, wallets, categories/category-wallet relationships, budgets/series configs and jars/month configs; timezone changes update owner revision. Derived tables resolve owner through their parent; preserve delete ordering. Pure reads/rollback cannot advance committed revision. Record exact table coverage in ADR.
- [ ] Store aggregate and its revision from the same repeatable-read snapshot. Latest GET validates revision before returning cache; race yields a labeled snapshot as_of, never stamps stale data as newer. Expire current-period snapshot at next account-local midnight even without a write; closed periods remain reusable until mutation/retention expiry.
- [ ] Run preceding tests GREEN against disposable DB with triggers; prove edit/delete/backdate/import and timezone invalidation, foreign-owner denial and expiry410. Commit only scoped files as `feat: persist versioned report snapshots`.

### Task 3: Deterministic insight candidates

**Files:** Create `backend/internal/entity/report_insight.go`, `report_insight_test.go`; `backend/internal/usecase/report_insights.go`, `report_insights_test.go`. Modify monthly report DTO/service to attach candidates; preserve old Insider API wire format through adapter.

**Interfaces:** `BuildReportInsights(report entity.MonthlyReport) []entity.ReportInsight`; insight includes stable ID, rule version, kind, fact refs, source scope/ranges/as_of, typed drilldown and server-formatted copy. Maximum3; Overview picks first.

- [ ] Write `TestReportInsightsThresholds`, `TestReportInsightsZeroBaseline`, `TestReportInsightsNoDuplicateEvent`: enforce exact defaults in Global Constraints; budget over-limit doesn't require comparison sample gates. Run `rtk proxy go test ./internal/entity ./internal/usecase -run ReportInsight -count=1` RED.
- [ ] Implement deterministic ranking: actual exceeded limit first, then absolute expense-change contribution; stable ID tie-breaker; dedupe category and budget observations of same event. Insufficient comparison history gives no comparative insight, never invented positive advice. Include category contribution, budget/jar status and neutral fallback; no health scores or unsupported causal claims.
- [ ] Run command GREEN including empty-data, baseline0, negative/positive deltas and threshold equality. Commit `feat: generate grounded report insights`.

### Task 4: Native charts, Reports and Insider

**Files:** Create `mobile-ios/MyPocketIOS/Core/Models/ReportModels.swift`; `mobile-ios/MyPocketIOS/Features/Reports/MonthlyReportsView.swift`, `MonthlyReportsViewModel.swift`; `mobile-ios/MyPocketIOS/DesignSystem/MPReportCharts.swift`; `mobile-ios/MyPocketIOSTests/MonthlyReportsTests.swift`. Modify `Core/API/FinanceAPI.swift`, `Core/Offline/FinancialSnapshotCache.swift`, `Features/Overview/OverviewView.swift`, `Features/Secondary/SecondaryFeatureViews.swift`, `Features/Transactions/TransactionsView.swift`, `DesignSystem/MPFinanceCards.swift`, project file and `MyPocketIOSUITests/LaunchTests.swift`.

**Interfaces:** `FinanceAPI.monthlyReport(month:walletID:granularity:ownerID:)`; `MonthlyReportsViewModel.load(scope:)`; `ReportDrilldown` adapter opens existing ledger with owner/date/wallet/category-subtree/report eligibility. Cache report payload independently by exact scope/revision; don't overwrite existing ledger cache. ReportsView route composes new screen without unrelated SecondaryFeatureViews refactoring.

- [ ] Add meaningful API contract/state tests for no `/transactions` request on report load, stale scope response ignored, chart sums equal totals, offline exact-scope fallback and drilldown subtree equality. Register sources then run `rtk proxy xcodebuild -project mobile-ios/MyPocketIOS.xcodeproj -scheme MyPocketIOS -destination 'platform=iOS Simulator,id=EC80B092-CEF1-4A8E-A5A3-7420298EFD4D' -only-testing:MyPocketIOSTests/MonthlyReportsTests test` RED.
- [ ] Implement model/API and view model. Overview Insider uses new report candidates; any existing Overview recent-ledger read remains separately scoped and is not claimed eliminated. Month/wallet controls use shared pickers; display sources/scope and cached-at label.
- [ ] Extend shared chart bases for weekly grouped bars, daily trend and horizontal category/contributor bars. Compose summary, chart, budget/jar previews, Insight and deep-analysis entry point; remove default long list/top expense rows from report. Mask tooltip/prose/VoiceOver; use shared density tokens and full hit targets. Preserve savings navigation and feedback bubble on Reports.
- [ ] Run focused command GREEN and API-backed UI fixtures for empty/loaded/error/offline/filter/drilldown, small and regular iPhone plus large Dynamic Type. Compare screenshot geometry before claiming visual acceptance. Commit `feat: present aggregate reports and compact insider` with only project registration hunks from this task.

### Task 5: On-demand analysis service, provider grounding and recovery

**Files:** Create `backend/internal/entity/report_analysis.go`; `backend/internal/repository/report_analysis.go`; `backend/internal/infrastructure/repository/report_analysis_postgres.go`, `_test.go`; `backend/internal/usecase/report_analysis.go`, `_test.go`; `backend/internal/controller/http/report_analysis_handler.go`, `_test.go`; `backend/cmd/worker/report_analysis.go`, `_test.go`; next forward migration `000044_report_analyses.up.sql`/`.down.sql` after version verification. Modify `backend/cmd/api/main.go`, `backend/cmd/worker/main.go`, provider adapter under `backend/internal/infrastructure/ai/`, migration tests/README, ADR-027.

**Interfaces:** `ReportAnalysisService.Begin(ctx, principal, snapshotID, clientRequestID)`, `Get(ctx, principal, analysisID)`, `Cancel(ctx, principal, analysisID)`; POST `/reports/analyses`202, GET `/reports/analyses/:id`, POST `/reports/analyses/:id/cancel`. Status queued/running/completed/failed/cancelled plus interrupted outcome represented failed/error_code. JSON never accepts raw facts from client. Unique job key includes owner/scope/revision/rules/prompt/provider-profile versions.

- [ ] Write `TestReportAnalysisReusesSameRevision`, `TestReportAnalysisLateCompletionKeepsOriginalSnapshot`, `TestReportAnalysisOwnerAndPayloadConflict`, `TestReportAnalysisCrashDoesNotRepeatProvider`; test fact-limit30, observation/suggestion limits3, unsupported fact reference rejection, missing provider, empty-data no call, owner active-job limit/cooldown60. Run `rtk proxy go test ./internal/usecase ./internal/controller/http ./internal/infrastructure/repository ./internal/infrastructure/ai ./cmd/worker -run ReportAnalysis -count=1` RED.
- [ ] Implement durable acceptance, fenced worker claim/completion and result recovery. Cancellation/reset/restart follow DB state; expired running jobs become interrupted without auto provider retry. Return existing active/completed job on same-key replay before cooldown check; mismatch409; new-revision cooldown429 with Retry-After. Keep completed reads available.
- [ ] Build ≤30 facts from owned immutable snapshot; budget/jar facts retain separate scope. One provider call returns structured references/qualitative explanation, numerical statements rendered by server templates. Invalid output gives deterministic fallback with explicit status, no repair provider call. Use existing provider timeout ceiling and redact diagnostics. Store prompt/provider-profile versions and usage metadata without raw ledger/notes.
- [ ] Run preceding tests GREEN with disposable DB and deterministic provider, record configured-provider evaluation separately. Add tests proving a response-loss GET does not create another job and stale snapshot is not accepted as current implicitly. Commit `feat: add recoverable on-demand report analysis`.

### Task 6: Native deep analysis and follow-up in chat

**Files:** Create `mobile-ios/MyPocketIOS/Features/Reports/ReportAnalysisState.swift`, `ReportAnalysisView.swift`; `mobile-ios/MyPocketIOSTests/ReportAnalysisTests.swift`. Modify report models/API/new MonthlyReportsView, `Features/Assistant/AssistantViews.swift` only for typed scope handoff, project file and UI tests.

**Interfaces:** `FinanceAPI.beginReportAnalysis(snapshotID:clientRequestID:)`, `reportAnalysis(id:)`, `cancelReportAnalysis(id:)`; `ReportAnalysisState` owns stable request/snapshot/owner pointers and GET polling. Completed article is bound to requested snapshot, not latest screen revision. Chat handoff prefills a visible scope-based question without sending automatically.

- [ ] Write `testAnalysisRelaunchUsesGET`, `testRevisionChangeLabelsOldArticle`, `testOwnerSwitchRejectsLateResponse`, `testExpiredSnapshotOffersFreshReport`; verify no automatic POST when opening Reports or changing filter. Register and run native command from Task4 with `-only-testing:MyPocketIOSTests/ReportAnalysisTests` RED.
- [ ] Implement button/queued/running/completed/error/cancelled/expired states; keep existing report visible during processing/error. Handle429/Retry-After and provider-unavailable copy. Persist request identity before POST; uncertain acceptance replays only identical idempotent request when no job ID is recoverable, otherwise GET. This narrowly clarifies design recovery without creating a new intent.
- [ ] Compose ≤3 observations + ≤3 suggestions via MPMarkdown/MPStatusMessage/MPButton; timestamp and stale label, manual regenerate, masks/accessibility. “Hỏi AI” only opens/prefills chat; backend reauthorizes referenced snapshot when question submitted.
- [ ] Run focused command GREEN and UI recovery fixture through terminate/relaunch, stale revision and masking. Commit `feat: show saved report analysis in native reports`.

### Task 7: Retention, integration evidence and reconciliation

**Files:** Create `backend/cmd/worker/report_retention.go`, `_test.go`; `docs/work/audits/REPORTS-INSIDER-UAT-20261031.md`. Modify worker wiring, report repository tests, `docs/design/pages/reports/README.md`, `docs/architecture/API.md`, relevant ERD/requirements docs, ADR-027, `docs/work/VALIDATION_MATRIX.md`, `docs/work/BACKLOG.md`, `docs/CONTEXT.md`, `docs/releases/CHANGELOG.md`.

**Interfaces:** `ReportRetentionRepository.DeleteExpired(ctx, now, limit) (int,error)` bounded internal worker operation; deletes expired analysis/snapshot in dependency order, never ledger. Retention90days from creation; cache reuse never extends creation time. Account deletion cascades private report data.

- [ ] Write `TestReportRetentionPreservesLedgerAndLiveJobs` with expired/fresh/active/cross-owner fixtures. Run `rtk proxy go test ./cmd/worker ./internal/infrastructure/repository -run ReportRetention -count=1` RED, implement bounded cleanup respecting active task references, rerun GREEN.
- [ ] Run backend report/analysis/migration integration tests, project/base guards and native focused targets; record latency/cache reuse/provider calls and EXPLAIN. Verify no provider call on ordinary report load or repeat analysis read. Inspect xcresult and confirm new tests actually executed.
- [ ] Capture regular/small iPhone and large-text light screenshots with seeded data; owner UAT covers charts, comparison scopes, Insider, drilldown, saved/stale/failed analysis and savings semantics. Record unresolved acceptance honestly; no production deployment implied.
- [ ] Reconcile API/ERD/rules/screen/base/context/backlog/changelog. Replace legacy report screen doc's web authority/top-expense-list assumptions only where new behavior is implemented. Keep portfolio workstream and chat findings open. Commit `docs: record report analysis validation and retention` after evidence.

## Dependency and delegation

Tasks1→2→3 form backend contract; Task4 consumes1–3. Task5 consumes2/3 and may run alongside4 with disjoint write ownership; Task6 consumes4/5; Task7 integrates all. One schema owner assigns actual migration numbers. Shared API/project/worker files get one integrator; agents supply targeted patches rather than concurrently editing them. Preserve user preference for agents and task review; report work does not cancel previously approved chat implementation.

## Self-review and handoff

### Portfolio follow-up captured from owner feedback

Owner expects gold/stocks accumulation tracked as quantity × price per unit, with edit available after saving. Current source already stores `PortfolioTrade.QuantityScaled` and `UnitPrice`; native form labels “Khối lượng” and “Giá đơn vị”. However, history renders `trade.unitPrice` as its main amount, and correction is hidden under each row's overflow menu. `PortfolioAsset` currently lacks an explicit unit. These are source observations, not simulator/deployed-build verification.

Separate design candidate: treat accumulation as a purpose and portfolio as holdings, so the same gold position is not recorded twice merely because it is both an investment and accumulated asset. Select a fixed unit for each asset (gold: owner-selected chỉ/lượng/gram; stocks: cổ phiếu), show quantity and “Đơn giá (đ/unit)”, and derive gross amount and fees separately. Example: 2 chỉ × 8.000.000 đ/chỉ = 16.000.000 đ, plus purchase fee. Show total as the history headline; keep per-unit price in detail. Expose “Sửa giao dịch” from trade detail and reuse existing correction preview/confirm API; editing asset name and updating market price remain distinct actions. Changing units on an existing holding requires explicit conversion design; do not silently reinterpret quantities. No existing rows are migrated/repriced by this report plan. Need clarify what owner calls “tài sản tích lũy” in the actual app before deciding whether to merge navigation or only improve labels.

Accounting/calendar/coverage are Task1; mutation/midnight/cache/retention keys Task2; thresholds/ranking Task3; charts/base/drilldown/offline Task4; grounding/dedupe/restart/cost Task5; native persistence/privacy/handoff Task6; cleanup/evidence/docs Task7. Every Review Focus condition has an owning test. APIs are proposed additive contracts; schema numbers must be resolved against live branch before writing. No runtime tests were executed for this plan-only artifact.

Owner request “tạo plan” is recorded as authorization to produce this plan from the chosen design. Proposed numeric defaults are carried verbatim; implementation and production deployment are separate from this planning deliverable.
