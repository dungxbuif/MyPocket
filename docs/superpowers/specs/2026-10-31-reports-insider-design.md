---
artifact_type: detail_design
id: REPORTS-INSIDER-20261031
status: draft
owner: shared
approval: approach_accepted_design_pending
---

# Báo cáo tháng và Money Insider

## Ý định và quyết định đã thống nhất

Owner muốn phân tích sâu, biểu đồ hữu ích, bỏ danh sách giao dịch dài khỏi Báo cáo và giảm công việc/chi phí cho AI. Chọn query backend + AI theo yêu cầu: report và insight cơ bản có ngay; người dùng bấm “Phân tích sâu” để tạo bài giải thích, lưu lại theo phiên bản dữ liệu. Đây là thiết kế đề xuất, chưa triển khai.

Ba phương án đã trao đổi: (1) query + câu mẫu, nhanh và dễ kiểm chứng nhưng giải thích hạn chế; (2) query + AI theo yêu cầu, kiểm soát chi phí nhưng lần đầu cần chờ; (3) AI tự tạo nền, thuận tiện nhưng phát sinh chi phí và quản lý bài cũ. Owner chọn (2). Cron AI hằng đêm là mở rộng sau, không thuộc bản đầu.

## Bối cảnh hiện tại và ranh giới

`ReportsView` trong `SecondaryFeatureViews.swift` gọi insider, month summary và toàn bộ transactions; trend tính tại iOS. `OverviewViewModel` cũng tải toàn ledger. `insider_postgres.go` đã có SQL totals, top 3 categories, top 5 expenses, transaction đọc repeatable-read; tháng hiện tại đang so với toàn tháng trước. Native đã có `MPTrendChart`, `MPFinanceComparisonChart`, `MPSurfaceCard`, `MPMetricGrid`, shared progress và status/button bases. Savings/budget/jar reports đã tồn tại và phải giữ accounting rules của chúng.

Lát này bổ sung read model dùng chung cho Reports, Insider và AI; không đổi semantics ledger. Unified AI chat đang có plan riêng; Task 1 commit `d7a53ef8` còn findings về event lease và Xcode test registration, chưa hoàn tất. Thiết kế này không thay thế plan chat hoặc tự cho phép triển khai Task 2.

## Business và phép tính

- Thu/chi: dùng canonical report eligibility hiện có, owner-scoped; loại transfer nội bộ và adjustment theo domain rules. Không tự gộp credit/portfolio vào báo cáo hoàn chỉnh khi capability chưa hỗ trợ.
- Thặng dư = thu − chi; không gọi là số tiền có thể tiêu hoặc tiền thực sự tiết kiệm. Tỷ lệ thặng dư chỉ có khi thu > 0; mẫu số bằng 0 trả unavailable.
- Ví mục tiêu: nạp/rút/chuyển nội bộ và số dư trình bày theo savings report hiện có. Chuyển tiền sang ví mục tiêu không tạo thêm chi tiêu.
- Hũ: phân bổ/chi/còn lại theo cấu hình tháng và giao dịch gắn hũ; không cộng số dư hũ vào tổng tài sản. Ngân sách có thể chồng scope; không cộng tổng nhiều ngân sách thành một hạn mức tài khoản.
- Category rollup theo cây: group cha gồm subtree, mỗi giao dịch tính một lần. Category ID là identity; tên trùng không bị gộp nhầm. Top 5 + “Khác” phải cộng đúng tổng chi.
- Tháng hoàn tất so với toàn tháng trước. Tháng hiện tại so cùng số ngày đã hoàn tất, loại hôm nay khỏi so sánh để tránh ngày dở dang. Cards vẫn có totals tới hiện tại, ghi rõ hai phạm vi. Tháng ngắn hơn: so số ngày chung và công bố khoảng ngày cụ thể; chênh lệch tuyệt đối và phần trăm do backend tính.
- Không suy diễn nguyên nhân từ mức tăng. “Nhóm Ăn uống đóng góp vào mức tăng” là fact; “đi ăn ngoài nhiều hơn” cần bằng chứng phù hợp hoặc phải là giả thuyết có nhãn.
- Chỉ phản ánh dữ liệu đã ghi nhận; không tuyên bố biết giao dịch còn thiếu. Không có giao dịch là trạng thái rỗng; query lỗi là lỗi, không phải tổng bằng 0.

Các ngưỡng sau là đề xuất của người thiết kế để owner review: comparison insight cần ít nhất 7 ngày chung và 5 giao dịch mỗi kỳ; chỉ chọn biến động chi đáng kể khi |delta| ≥ 100.000 VND và |change| ≥ 15%. Baseline 0 hiển thị “phát sinh mới”, không phần trăm. Các ngưỡng cấu hình/version được, không áp dụng cho dữ liệu biểu đồ hoặc cảnh báo vượt hạn mức thực tế.

## Nội dung và base mapping

Reports có month picker native, wallet scope và cập nhật lúc nào; summary Thu/Chi/Thặng dư; cột dòng tiền theo tuần (tối đa 6 bucket cắt theo tháng); thanh ngang top 5 nhóm chi + Khác; so sánh đóng góp tăng/giảm theo nhóm; ngân sách/hũ; khu vực phân tích sâu. Không có raw transaction list mặc định. Khoảng thời gian đầy đủ trong tháng có thể hiển thị tới 31 điểm ngày khi người dùng đổi độ chi tiết.

Money Insider trên Tổng quan hiển thị một insight nổi bật, xem thêm tối đa 3. Mỗi insight có kỳ, chênh lệch, nguồn, as-of và CTA đúng scope. Nếu không có insight đủ điều kiện, hiển thị mô tả trung tính hoặc ẩn vùng; không ép sinh lời nhận xét. Loại trùng sự kiện giữa các card.

| Vùng | Base |
| --- | --- |
| Header/filter | MPScreenHeader + shared month/wallet picker |
| Summary | MPFinanceSummaryCard + MPMetricGrid |
| Cashflow | MPFinanceComparisonChart mở rộng series/legend/selection |
| Categories/contributors | shared horizontal-bar variant, MPText, MPButton |
| Budget/jar | finance progress bases + domain adapter |
| Insider | shared insight variant trên MPSurfaceCard |
| Deep analysis/status | MPMarkdown + MPStatusMessage + MPButton |

Màu hiện tại, light mode, compact shared tokens. Touch target không bị thu nhỏ theo padding. Chart phải có nhãn và accessible summary; masking áp dụng số, tooltip, prose và VoiceOver. Card không bọc card; không đổi toàn bộ font ngoài scope đã duyệt. Layout chi tiết sẽ được preview riêng, không gọi thiết kế chữ này là visual acceptance.

## Query và consistency

Đề xuất additive `GET /reports/monthly?month=YYYY-MM&wallet_id=...&granularity=week|day`. Trả scope/timezone/currency, as_of, data_revision, snapshot_id, totals, comparison ranges, series, category breakdown/contributors, ngân sách/hũ phù hợp scope, quality flags, insight candidates và typed drilldown filters. Không trả raw ledger mặc định. Giới hạn so sánh nhóm 5 tăng + 5 giảm; budget/jar preview tối đa 5 mỗi loại và tổng count/link tới domain screen. Không áp wallet filter cho hũ toàn tài khoản một cách ngầm định: phần không có scope tương đương ghi rõ scope riêng hoặc unavailable.

Các aggregate đọc trong cùng snapshot DB. Service report dùng canonical scope/rules hiện có, tránh nhân đôi logic giữa insider/month/advisor; tối ưu aggregate bằng SQL thay vì tải ledger vào app. Không thêm warehouse hoặc materialized view trong bản đầu. Index và EXPLAIN với fixture lớn là tiêu chí performance trước chọn thêm tầng tổng hợp.

Snapshot key: owner, month, wallet scope, timezone, granularity, report_rules_version, data_revision. Data revision tài khoản tăng trong cùng transaction với mọi thay đổi ledger/category tree/budget/jar/wallet/timezone ảnh hưởng report. Invalidate cả tháng cũ và mới khi sửa ngày; revision owner-wide là đề xuất đơn giản trước khi tối ưu. Client không được tự tuyên bố revision mới nhất. GET phải phát hiện revision đổi khi đọc, không đóng dấu cache cũ bằng revision mới.

Drilldown chuyển sang màn Giao dịch với typed filter từ backend; mở list theo scope hiện tại và ghi rõ nếu khác snapshot cũ. Trang đầu 25, tối đa 50; AI search tối đa 5 một lần. Snapshot offline có timestamp và nhãn bản lưu; không trộn section từ nhiều tháng/revision.

## AI theo yêu cầu và chi phí

Nút “Phân tích sâu” gọi endpoint async đề xuất `POST /reports/analyses` với snapshot_id và client_request_id; GET `/reports/analyses/{id}` phục hồi. Backend xác thực lại owner và lấy fact bundle từ snapshot server. State queued/running/completed/failed/cancelled; unique analysis key owner+scope+revision+rules+prompt+provider-profile version dedupe đồng thời. Retry cùng identity/payload trả task cũ; hash khác conflict. Không tự gọi lại provider khi outcome chưa rõ sau restart; đánh interrupted/failure có thể đọc và cho người dùng retry rõ ràng.

Fact bundle bounded: tối đa 30 facts gồm totals, so sánh, drivers, budget/jar statuses, quality/scope. Không gửi ảnh, raw note hay full ledger. Server xây insight ứng viên deterministic trước; model dùng fact references để giải thích và đề xuất kiểm tra, không tự tạo số/chart. Số liệu render từ fact templates server; unknown fact bị reject, fallback câu mẫu. Bài viết tối đa 3 quan sát + 3 gợi ý; đủ dữ liệu mới có quan sát, không bịa nguyên nhân, không tự lưu hành động tài chính.

Khi mở report có completed analysis cùng revision thì đọc lại. Revision đổi giữ bài cũ với nhãn “Dựa trên dữ liệu trước”, CTA tạo mới theo yêu cầu, không tự tiêu token. Completion muộn chỉ gắn snapshot đã yêu cầu; không ghi đè bài của revision mới. Empty dataset không gọi model. Capability provider unavailable vẫn xem chart/insight mẫu; deep analysis báo không khả dụng.

Đề xuất chi phí bản đầu: 1 provider call/job, timeout tối đa theo policy provider hiện có, không auto retry; 1 active analysis/owner, cooldown tạo revision mới 60 giây. Không khóa đọc cache hoặc hỏi chat vì cooldown report. Metrics: bundle bytes/fact count, call count, reuse ratio, acceptance/provider latency p50/p95, token usage nếu có, error codes; không log facts/private data. Cron AI, notification, tự tổng kết mọi tài khoản và live token streaming ngoài scope.

## Dữ liệu, retention và validation

Snapshot và analysis owner-scoped trong PostgreSQL; payload private, auth check cả POST/GET/drilldown. Đề xuất giữ snapshot/bài phân tích tối đa 90 ngày từ creation, reference expiry trả trạng thái expired và CTA tạo snapshot mới; không xóa ledger. Cleanup không ảnh hưởng báo cáo deterministic có thể tính lại. Xóa tài khoản xóa dữ liệu phân tích theo lifecycle hiện có. Bài sâu độc lập chat history reset; mở “Hỏi AI” chỉ truyền scope/snapshot reference đã reauthorize và không tự gửi câu hỏi.

Test business: transfer excluded, adjustment, goal deposit ≠ expense, overlapping budgets, child category rollup, duplicate names, zero baseline/income, no-data vs query error, tháng ngắn/DST/midnight/timezone, current partial day, inactive references. Test consistency: edit/delete/backdate, old/new month invalidation, concurrent mutation/GET/cache stamp, owner isolation. Test AI: bounded facts, no raw ledger, unknown fact rejection, concurrent dedupe, response loss/restart, stale completion, expiry, failed provider vẫn có report. UI: small/regular iPhone, large type, masking/VoiceOver, offline timestamp, drilldown đúng scope, no default ledger list.

Verification dự kiến: backend focused domain/repository/handler tests với disposable PostgreSQL, migration tests và EXPLAIN fixture; native model/API/base tests cùng screenshot/API-backed UAT. Đây là docs-only: chưa chạy runtime test và chưa đổi API/schema.

## Trace và reconciliation

[Backlog](../../work/BACKLOG.md), [native phase](../../work/phases/PHASE-002-native-ios-swiftui.md), [advisor ticket](../../work/tickets/AI-ADVISOR-01-DETAIL_DESIGN.md), [API](../../architecture/API.md), [base contracts](../../design/system/BASE_COMPONENTS.md), [validation](../../work/VALIDATION_MATRIX.md), [context](../../CONTEXT.md), [changelog](../../releases/CHANGELOG.md), [unified chat design](2026-10-31-unified-ai-assistant-design.md).

Sau design approval: tạo ticket/ADR cho report snapshots + analysis lifecycle và execution plan riêng; reconcile requirements/API/ERD khi implementation thực sự có. Các lựa chọn threshold, retention, cooldown và data revision ở trên là đề xuất của người thiết kế, chưa phải owner đã chốt từng giá trị.
