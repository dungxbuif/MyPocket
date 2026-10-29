---
artifact_type: detail_design
id: FEEDBACK-20261019-OFFLINE-LOADING
status: in_progress
owner: shared
parent: FEEDBACK-20261019-DETAIL_DESIGN
---

# Offline read and loading UX

## User feedback

Khi chuyển qua lại giữa các màn native, dữ liệu chưa được giữ ổn định: màn có
thể nháy trắng hoặc hiện `MPLoadingView` trong thời gian rất ngắn. Khi mất
mạng, màn nên hiển thị bản dữ liệu đã lưu cục bộ và chỉ báo đang đồng bộ nhỏ;
không thay toàn bộ nội dung bằng một màn loading mới.

## Evidence from current source audit

`FinancialSnapshotCache` đã mã hóa snapshot theo owner/server và Overview,
Wallets, Transactions có đường fallback khi request thất bại. Tuy nhiên cache
hiện được đọc chủ yếu trong nhánh recovery sau lỗi; khi màn được tạo lại, UI
không có dữ liệu trước khi request đầu tiên hoàn thành. Đây là nguyên nhân có
thể gây flash/loading khi đổi tab.

Các nhánh loading hiện cần rà soát để phân biệt blocking initial load với
inline refresh:

- `SecondaryFeatureViews.swift`: Jars (`if loading`), jar report
  (`if reportLoading`), Wallet/Months/Budget/Category/Notification flows.
- `SavingsReportsView.swift`: report và transaction drilldown (`if loading`).
- `BudgetReportView.swift`: report (`if loading`).
- `WalletsView.swift`: Wallet detail (`if loading`).
- `AssistantViews.swift`: history và active run; đây là loading inline trong
  conversation, không được biến thành blocking screen.
- `OverviewView.swift`, `TransactionsView.swift`, `WalletsView.swift` đã có
  một phần policy `MPLoadingPresentationPolicy` nhưng cần bảo đảm mọi route
  dùng cùng semantics.

## Contract

1. Initial load không có dữ liệu: dùng shared blocking `MPLoadingView` hoặc
   skeleton.
2. Có dữ liệu server hoặc cache: giữ nguyên dữ liệu trên màn; refresh chỉ dùng
   status/progress inline, không xóa mảng/state trước request.
3. Offline/transient failure: đọc snapshot đúng `ownerID`, gắn `cachedAt`,
   hiển thị banner retry nhỏ; không lộ cache của owner khác và không biến lỗi
   auth/validation thành fallback offline.
4. Đổi tab không được tự động xóa state đã render. Các refresh do foreground,
   queue revision hoặc pull-to-refresh phải được dedupe theo request đang chạy.
5. Cache hydration phải xảy ra trước hoặc đồng thời với server refresh; bản
   cache cũ không được hiển thị nếu owner/server namespace không khớp.
6. Chỉ thao tác explicit refresh mới được phép hiện progress lâu hơn; loading
   ngắn dưới ngưỡng UX không được làm nhấp nháy cả layout.

## TDD / verification plan

- RED: test mỗi view model giữ snapshot hiện tại khi refresh lỗi hoặc đang
  refresh; test cache hydrate trước lần server response; test request trùng bị
  dedupe.
- GREEN: dùng shared loading policy/base status, không tạo spinner riêng ở
  screen; giữ nguyên owner-scoped `FinancialSnapshotCache`.
- UI: chuyển Overview → Transactions → Account → Overview nhiều lần, kiểm tra
  không có full-screen loading khi đã có rows; kiểm tra offline vẫn đọc được
  dữ liệu và có `cachedAt`/retry.
- Regression: full native unit target, selected UI target, base/project guards,
  Release simulator build; ghi kết quả vào `docs/work/VALIDATION_MATRIX.md`.

## Scope boundary

Không thay đổi API, migration, financial calculation, auth hoặc quyền offline.
Không xóa cache cũ; chỉ bổ sung hydration/publication semantics tương thích
với snapshot version hiện tại. Production rollout và physical-device UAT là
gates riêng sau khi local test xanh.

## Links

- [Native loading/cache design](../../design/screens/overview-native.md)
- [Offline read audit](OFFLINE-READ-SCOPE-AUDIT-20261014.md)
- [Offline cache ticket](OFFLINE-FINANCIAL-CACHE-20261014.md)
- [Validation matrix](../VALIDATION_MATRIX.md)
