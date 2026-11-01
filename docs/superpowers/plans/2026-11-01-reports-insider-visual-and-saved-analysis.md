# Reports / Insider — visual reporting and saved analysis implementation plan

> **For agentic workers:** Use superpowers:executing-plans or superpowers:subagent-driven-development as available. This plan is not implementation evidence.

**Goal:** Làm Báo cáo và Money Insider trực quan bằng biểu đồ native, số liệu nhất quán, đồng thời khôi phục bài AI đã lưu mà không gọi model lại khi mở màn.

**Architecture:** PostgreSQL aggregate → revisioned snapshot → shared report/Insider model → native chart bases. On-demand worker tạo bài từ bounded facts; latest lookup trả completed/active riêng. Freshness không đồng nhất với retention.

**Tech Stack:** Existing Go/Gin/PostgreSQL, SwiftUI + Apple Swift Charts, existing durable job worker and native protected cache.

**Spec:** [Requirements and detail design](../../work/tickets/REPORTS-INSIDER-VISUAL-20261101-DETAIL_DESIGN.md).

**Status:** Implemented as a local candidate. Backend and simulator builds pass;
manual rich-data report/drill-down UAT passes. Automated tests were skipped by
owner direction; provider-backed saved-analysis, offline and physical-device gates
remain open. Tiếp nối [plan 2026-10-31](2026-10-31-reports-insider.md).

## Global constraints

- Giữ màu/font sizing/base hiện tại. Không sửa shell bubble/nav; Assistant toàn màn không bubble/navigation.
- Không automated tests theo chỉ dẫn user. Implement lần lượt, chỉ build và UAT thủ công sau khi gộp; không ghi “tests passed”.
- Không sửa migration đã áp dụng; backward-compatible API; không đổi production hoặc gọi provider thật trong giai đoạn design.
- Working tree đang có Assistant changes và `backend/api` tracked binary; không reset/stage lẫn các file này khi thực hiện task.
- Mọi bước có thể bị kiểm tra thủ công nhưng chưa có bước nào dưới đây được đánh dấu hoàn tất.

## Review focus

1. Contributor cùng cửa sổ thời gian, tree/Other drill-down và tiền integer.
2. Owner/scope isolation, cache không lẫn user/ví/tháng, response đến muộn.
3. Bài AI qua nửa đêm, retention nguồn fact, concurrent idempotency và worker lease.
4. Biểu đồ zero/no-data/negative/masked đúng; giảm card lồng nhau; không re-style riêng màn.

## Task 1 — Chốt DTO và aggregate đúng nghĩa

**Files:**
- Modify `backend/internal/entity/monthly_report.go`
- Modify `backend/internal/entity/report_insight.go`
- Modify `backend/internal/infrastructure/repository/monthly_report_postgres.go`
- Modify `backend/internal/repository/monthly_report.go`
- Modify `backend/internal/usecase/monthly_report.go`
- Modify `backend/internal/controller/http/monthly_report_handler.go`
- Modify `mobile-ios/MyPocketIOS/Core/Models/ReportModels.swift`

- [ ] Thêm additive range, previous daily series, category bucket/target descriptor và comparison metadata, giữ trường cũ.
- [ ] Tách aggregate cơ cấu cả kỳ khỏi contributors matched window; calendar boundaries theo account timezone.
- [ ] Giữ zero days, bỏ ngày tương lai; preserve Int64 money, exact ranges và stable IDs sang Swift.
- [ ] Resolve target từ owned snapshot; query subtree/Other/uncategorized; báo stale nếu ledger revision thay đổi.
- [ ] Giữ route category_id cũ; không silently đổi nghĩa client cũ.
- [ ] Sắp insight theo priority + impact, dedup topic, dùng câu có số liệu từ fact; preserve ngưỡng rule version.

**Deliverable:** DTO đủ để frontend vẽ và drill-down mà không tải ledger toàn tháng hay đoán kỳ so sánh.

## Task 2 — Lưu/khôi phục analysis và bảo vệ job

**Files:**
- Modify `backend/internal/entity/report_snapshot.go`
- Modify `backend/internal/entity/report_analysis.go`
- Modify `backend/internal/repository/report_snapshot.go`
- Modify `backend/internal/repository/report_analysis.go`
- Modify `backend/internal/infrastructure/repository/report_snapshot_postgres.go`
- Modify `backend/internal/infrastructure/repository/report_analysis_postgres.go`
- Modify `backend/internal/usecase/report_analysis.go`
- Modify `backend/internal/infrastructure/ai/report_analysis.go`
- Modify `backend/internal/controller/http/report_analysis_handler.go`
- Create `backend/migrations/<next>_report_analysis_restore.up.sql` and `.down.sql` if schema changes required; allocate next number at execution.

- [ ] Tách fresh lookup/retained lookup; closed-month snapshot key không đổi mỗi ngày.
- [ ] Build <=30 facts theo ưu tiên; normalize fingerprint không dùng snapshot UUID/global revision/biểu đồ mode làm danh tính nội dung.
- [ ] Thêm latest endpoint trả completed và active cùng scope, freshness, retained_until và evidence.
- [ ] Lookup reuse trước provider availability; atomic owner lock và unique request/attempt constraints; explicit failed retry.
- [ ] Worker heartbeat + lease ownership + completion timestamp; no ambiguous automatic provider recall.
- [ ] Cleanup giữ source đến hết retention của mọi bài tham chiếu và active deadline; read path enforce retention.
- [ ] Migration additive/backfill hỗ trợ row cũ; không expose owner/provider internals.

**Deliverable:** GET latest không gọi AI; bài cũ tồn tại qua snapshot freshness expiry; job có thể resume.

## Task 3 — Xây shared chart/presentation bases

**Files:**
- Modify `mobile-ios/MyPocketIOS/DesignSystem/MPReportCharts.swift`
- Modify `mobile-ios/MyPocketIOS/DesignSystem/MPFinanceCards.swift`
- Create `mobile-ios/MyPocketIOS/DesignSystem/MPReportPresentation.swift` (period picker, insight/article presentation)
- Modify `docs/design/system/BASE_COMPONENTS.md` to match implemented variants

- [ ] Kiểm tra deployment target; chọn availability phù hợp, không tự nâng target.
- [ ] Swift Charts: cumulative lines + fill/selection; grouped weekly bars; category donut; diverging contributors.
- [ ] Tooltip/legend/empty/masking/accessibility/Reduce Motion nằm trong base.
- [ ] Tiền formatter dùng base hiện có, chỉ chuyển Double tại chart adapter; 0 không tạo cột giả.
- [ ] Shared period picker, insight card và AI article; reuse MPText/MPMetric/MPButton/MPSurfaceCard.
- [ ] Giữ compatibility wrapper cho consumers khác của chart/card cũ.

**Deliverable:** Màn feature chỉ bố trí bases, không chứa custom chart geometry, font/color/radius override.

## Task 4 — Báo cáo và state restoration native

**Files:**
- Modify `mobile-ios/MyPocketIOS/Core/API/APIClient.swift`
- Modify `mobile-ios/MyPocketIOS/Core/Offline/FinancialSnapshotCache.swift` or add report cache alongside it if existing contract is unsuitable
- Modify `mobile-ios/MyPocketIOS/Features/Reports/MonthlyReportsView.swift`
- Modify `mobile-ios/MyPocketIOS/Features/Reports/MonthlyReportsViewModel.swift`
- Modify `mobile-ios/MyPocketIOS/Features/Reports/ReportAnalysisView.swift`
- Modify `mobile-ios/MyPocketIOS/Features/Reports/ReportAnalysisState.swift`

- [ ] Lắp layout theo spec: scope → totals → trend → allocation → contributors → AI article.
- [ ] Scope includes timezone từ server; cache-first, cancellation/request generation, no broad legacy fallback.
- [ ] Update API client decode latest/active/completed/evidence và typed drill-down targets.
- [ ] Persist request ID trước POST; restore latest/job ID; poll foreground có backoff, deadline thuộc server.
- [ ] Keep old completed visible during refresh/failure; current/stale/offline/expired UI đúng copy.
- [ ] Account-isolated protected cache, logout cleanup và retained_until enforcement.
- [ ] Hỏi tiếp mở Assistant mang source IDs đã xác thực; không tự gửi tin nhắn.

**Deliverable:** Rời màn/mở lại không mất bài hoặc job; đổi filter nhanh không làm sai dữ liệu.

## Task 5 — Money Insider dùng chung nguồn

**Files:**
- Modify `mobile-ios/MyPocketIOS/Features/Overview/OverviewView.swift` (view model is currently in the same source)
- Modify `mobile-ios/MyPocketIOS/DesignSystem/MPFinanceCards.swift`
- Adjust existing report destination wiring only where necessary

- [ ] Overview đọc report snapshot/insights tháng hiện tại thay legacy Insider call.
- [ ] Hiện tối đa 3 insight, mỗi insight có evidence/CTA; bỏ delta ring ở variant mới.
- [ ] CTA giữ month/wallet/section/target, mở đúng Reports/detail.
- [ ] Snapshot cache shared và amount privacy thống nhất; không thêm bản sao domain rules ở UI.

**Deliverable:** Insider và Báo cáo thống nhất số liệu và có đường đi rõ ràng.

## Task 6 — Tích hợp, build và UAT một lượt

- [ ] Review diff cho scope/unrelated changes, run `git diff --check` (không phải test suite).
- [ ] Compile backend + signed simulator iOS bằng cấu hình local hiện có; lưu log và artifact path. Không chạy test target.
- [ ] Khởi động API + report worker trên local rich-data DB, giữ secrets ngoài logs/docs.
- [ ] Cài simulator và đi qua checklist mục 9 trong spec; user review trực tiếp UI. Ghi từng mục passed/failed/not checked, không lấy build thay UAT.
- [ ] Một flow AI thật có kiểm soát: generate → rời màn → mở lại → thay dữ liệu → cập nhật; đếm provider calls từ telemetry không chứa nội dung riêng tư.
- [ ] Chụp lại các trạng thái chính (report, tooltip, category, Insider, saved/stale/loading analysis) để review visual.
- [ ] Đo performance theo fixture nếu user muốn benchmark; nếu chưa đo, giữ metrics là target.
- [ ] Reconcile page spec, ADR-027, API docs, base contracts, Context, Backlog, Validation Matrix, release note; không mark release/verified khi chưa có bằng chứng.

**Completion gate:** Implementation/build được báo riêng với UAT. Những mục user chưa kiểm tra còn mở; không tự merge/deploy production trong task thiết kế này.
