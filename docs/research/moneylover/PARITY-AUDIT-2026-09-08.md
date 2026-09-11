---
artifact_type: research_audit
id: MONEYLOVER-PARITY-2026-09-08
status: done
owner: shared
approval: user_requested_source_docs_code_comparison
updated: 2026-09-08
trace:
  backlog: ../../work/BACKLOG.md
  requirements: ../../requirements/REQUIREMENTS.md
  source_catalog: SOURCE-CATALOG-2026-09-08.md
  feature_matrix: FEATURE-PARITY-2026-09-08.md
  verification: ../../work/test-verification/MONEYLOVER-PARITY-2026-09-08.md
  validation_matrix: ../../work/VALIDATION_MATRIX.md
  release_notes: ../../releases/CHANGELOG.md
---

# MyPocket chưa đạt mức gần tương đương Money Lover

Ngày đối chiếu: 2026-09-08. Kết luận này dựa trên source hiện tại và kiểm thử mới, không chỉ dựa trên tên màn hình hoặc status ticket. **Audit hoàn tất không có nghĩa sản phẩm đã hoàn tất.**

Đã kiểm kê **96 bài public, 24 section, 4 category** của [Help Center Money Lover](https://moneylover.zendesk.com/hc/en-us), đọc nội dung văn bản, lập [danh mục nguồn và tóm tắt riêng](SOURCE-CATALOG-2026-09-08.md), rồi đối chiếu thành [50 hàng chức năng/chất lượng](FEATURE-PARITY-2026-09-08.md). Đây không phải 50 chức năng độc lập và không dùng để tính phần trăm hoàn thiện.

## Phạm vi và cách làm

- Nguồn public locale `en-us`, bao gồm cả bài tiếng Việt được đặt trong locale đó. API danh mục trả đủ 96 bài và không có trang tiếp theo; mỗi bài có ID, URL, ngày cập nhật và mapping trong catalog.
- Diễn giải chức năng bằng lời riêng và giữ nguồn; không sao chép nguyên văn toàn bộ bài hoặc tài sản hình ảnh. Video ngoài trang, nội dung trong screenshot, bài private/draft và locale khác chưa được kiểm chứng. Vì vậy đây là **đối chiếu chức năng từ docs**, không phải chứng nhận UI giống từng pixel hay reverse engineering app Money Lover.
- Đã đọc gateway, context, backlog, standards, active PHASE-007/TICKET-028, SPEC/REQUIREMENTS/API, research cũ; kiểm tra router → kiểu dữ liệu → repository → màn hình thực sự được mount → kiểm thử. Mở rộng sang tài liệu portal khi phát hiện mô tả không khớp code.
- Source ở HEAD `d419eac4eba0ccdb7597dd1a319af7b568c3df7b` cộng các thay đổi chưa commit sẵn có. Không xem HEAD riêng lẻ là bản đã audit; không đồng nhất source với toàn bộ môi trường production.
- Audit và sửa chú thích tài liệu đã được user yêu cầu. Không thay đổi API/schema/auth/runtime hay triển khai chức năng mới; không deploy. DB dùng để test là PostgreSQL tạm, tách khỏi DB production.

## Những điểm quan trọng nhất

| Mức | Phát hiện có bằng chứng | Hệ quả |
| --- | --- | --- |
| P1 | E2E logout bị `pwa-install-prompt` và dock chặn pointer | Có nút nhưng người dùng mobile không bấm được; test component vẫn xanh. |
| P1 | `OverviewScreen` dùng `span` cho “Xem báo cáo”; `TransactionsScreen` có nút tháng không có handler | Không có luồng báo cáo/lọc kỳ hoàn chỉnh. |
| P1 | `SettingsScreen`, `AccountSecurityScreen`, `TravelModeSheet`, `CategoryManagementScreen` không được mount từ App hiện tại | File UI tồn tại không chứng minh người dùng truy cập được; nhiều dữ liệu/handler chỉ là prototype. |
| P1 | `CreateCategory` luôn ghi `parent_id=NULL`; handler không nhận parent | Seed có cây danh mục nhưng người dùng chưa tạo/sửa cây tương đương Money Lover. |
| P1 | `Budget` không có wallet scope; query khớp trực tiếp category ID | Thiếu ngân sách theo ví, roll-up cha/con và cách tính tổng không trùng lặp. |
| P1 | Báo cáo tổng có thể tính cả ví `include_in_total=false`, trong khi số dư tổng loại ví đó | Hai khái niệm “tổng” chưa nhất quán; cần fixture SQL cụ thể trước khi sửa. |
| P1 | `CreateObligation` chỉ tạo thông tin khoản nợ | Không tương đương luồng vay/cho vay làm thay đổi tiền trong ví. |
| P1 | Draft chỉ có list; router không có confirm | Chưa hoàn thành luồng từ tự động tạo nháp đến ghi nhận tài chính. |
| P1 | API portal có đường dẫn ảnh cũ, export chưa có implementation, Insider có field không tồn tại | Bên thứ ba có thể triển khai client sai dù docs trông đầy đủ. |
| P2 | Test ngân sách tìm tên ngân sách nhưng row chỉ render tên danh mục | Đã tạo thành công; phần sửa/lưu trữ của test chưa được thực thi, không được báo toàn bộ CRUD xanh. |
| P2 | Account nhận callback thay đổi/ẩn asset nhưng không gọi; `AssetRow` cũ còn trong App mà không mount | Backend portfolio có proof nhưng giao dịch mua/bán/giá chưa có lối vào trong UI hiện tại. |
| P2 | UI còn nhãn iPhone, Premium, build version hardcoded; raw input/button và inline sheet còn nhiều | Chưa hoàn thành yêu cầu dùng base component; thông tin hiển thị không phải trạng thái hệ thống thật. |

Các nghi vấn SQL/semantic ở đây được phân biệt với lỗi E2E đã tái hiện. Không tuyên bố đã tái hiện toàn bộ lỗi dữ liệu trong production.

## Khoảng cách chức năng lớn

Core thu/chi, ví cơ bản, đồng bộ offline và phần planning có implementation thật. Nhưng để gần Money Lover cần thêm hoặc hoàn thiện:

1. **Ví/giao dịch/danh mục:** luồng chuyển ví và phí, chỉnh số dư từ UI, sao chép/xóa nhiều, lịch sử theo kỳ/nhóm, lọc nâng cao, danh mục con/icon/merge/activation theo ví.
2. **Ngân sách/báo cáo:** phạm vi ví, cha/con, finished/repeat/overlap, forecast, báo cáo drilldown, opening/ending, khoản loại trừ, trend đúng công thức và Insider chi tiết.
3. **Planning hoàn chỉnh:** sổ vay nợ liên kết tiền, bill đến hạn/đã trả, lịch lặp đủ chu kỳ và duyệt nháp, travel mode gắn sự kiện thật.
4. **Chức năng còn thiếu:** mục tiêu tiết kiệm, tín dụng đầy đủ, shared wallet, AI/OCR, export, reset/delete. Các chức năng thuộc M2 trước đây vẫn được ghi deferred, chưa tự động xem là đã hoàn thiện.
5. **Khác nền tảng/phạm vi:** bank linking, Google Pay native capture, widget native, Apple Shortcuts, ngôn ngữ/theme/khóa ứng dụng cần quyết định thiết kế riêng. API key tự nó không tạo ra các tích hợp này.

## Docs của ta đang thiếu hoặc sai ở đâu

| Tài liệu | Sai lệch / thiếu | Cách đọc sau audit |
| --- | --- | --- |
| `docs/research/moneylover/moneylover-full-research.md` | Research 31/08 tổng hợp theo nhóm, không có manifest từng bài hoặc bằng chứng code | Giữ làm lịch sử; catalog/matrix mới là bản đối chiếu có truy vết. |
| `docs/CONTEXT.md` | Các mốc implementation/review dễ bị hiểu là các màn hình mới đều hoạt động | Bổ sung audit mới và các giới hạn current mounted UI. |
| `docs/work/VALIDATION_MATRIX.md` | Proof cũ không bao phủ regressions current UI; `yes` mô tả loại proof cần có | Bổ sung bằng chứng mới; không nâng status thành verified. |
| `docs/requirements/SPEC.md` | Chủ động loại multi-currency đầy đủ, native listener, direct bank OAuth, subscriptions | “Clone gần giống” không tự động xóa các quyết định này; ghi rõ cần reconciliation khi chọn scope mới. |
| `docs/requirements/REQUIREMENTS.md` | `accepted` là yêu cầu được chấp nhận, không phải implementation đã xong; còn link `SRS.md` không tồn tại ở root | Dùng code/proof matrix; lên backlog sửa trace nguồn thay vì sáng tạo nội dung SRS. |
| `frontend/docs/docs/api/export.mdx` | Export được trình bày như có sẵn; `/receipts/uploads` và `/receipts/{id}` không khớp router | Đánh dấu export planned, sửa tài liệu file API theo handler thật. |
| `frontend/docs/docs/api/analytics.mdx` | `date_from/date_to` sai với handler nhận `from/to`; dashboard envelope và Insider fields sai | Sửa các ví dụ đã xác minh; full analytics contract regression vẫn cần. |
| `frontend/docs/docs/api/categories.mdx` | Nhận `parent_id` trong create chỉ có trong docs | Bỏ field không được handler xử lý; ghi rõ tree seed khác custom parent. |
| API snippets ở portal | Một số `security` viết các scheme trong cùng object, biểu thị AND; có archive path nằm sai cấp; diagram đặt audit trong finance transaction | Không coi là OpenAPI máy đọc đã được validate. Cần một contract sinh từ source/checked schema và test cookie/bearer riêng. |
| Portal nói chung | Chủ yếu API/ERD, thiếu hướng dẫn người dùng theo tác vụ, lỗi thường gặp và giới hạn từng chức năng | Cần help docs MyPocket riêng dựa vào những luồng đã chạy được. |

Việc sửa vài ví dụ docs trong audit này **không phải** tuyên bố đã kiểm định đầy đủ toàn bộ OpenAPI hoặc mọi route/method. Phần còn lại được đưa vào remediation.

## Mâu thuẫn và khác biệt phải giữ lại

- **Web:** [FAQ cũ](https://moneylover.zendesk.com/hc/en-us/articles/36755450613529) nói read-only; [thông báo relaunch](https://moneylover.zendesk.com/hc/en-us/articles/53633862661273) nói tạo/sửa trên desktop. Dùng thông báo relaunch làm mốc mới hơn; mobile browser vẫn là giới hạn của vendor.
- **Credit:** [wallet definitions](https://moneylover.zendesk.com/hc/en-us/articles/34972671048985) dùng phép trừ balance, trong khi [credit guide](https://moneylover.zendesk.com/hc/en-us/articles/37006112179609) dùng hạn mức cộng số dư có dấu. Không sao chép công thức mâu thuẫn; cần ví dụ số rõ ràng và test.
- **Warning:** Money Lover ghi 75% trong [budget guide](https://moneylover.zendesk.com/hc/en-us/articles/34300604750617); TICKET-011 của ta chấp nhận 80%. Đây là khác contract, không tự sửa threshold trong một audit docs.
- **AI/recurring:** ta đã chọn review-first và server persistence; không tái tạo giới hạn mất lịch khi logout hoặc bỏ bước review chỉ để giống vendor.
- **Commercial:** giữ quyết định của user là mọi chức năng luôn mở; không thêm paywall/trial/quota mua hàng. **Crypto:** bài ngừng dịch vụ 2020 không phải chức năng hiện hành phải clone; portfolio của ta là extension riêng.
- **iOS minimum:** title/body bài retirement không nhất quán; không lấy một dòng headline làm cam kết hỗ trợ thiết bị của MyPocket.

## Thứ tự đề xuất để đạt parity có thể kiểm chứng

Đây là remediation backlog, chưa phải thông báo đã implement hoặc deploy. Từng đợt cần design phù hợp impact; user không cần duyệt lại việc audit hiện tại.

| Đợt | Kết quả phải bàn giao | Điều kiện qua cổng |
| --- | --- | --- |
| R0 — đóng lỗi hiện tại | Logout mobile bấm được; tên ngân sách nhất quán; không có action giả trong màn hình đang chạy; tài liệu contract không quảng cáo endpoint giả | Current mobile E2E xanh, click/API/error tests cho UI; giải quyết gate base component. |
| R1 — finance | Ví, chuyển/điều chỉnh, lịch sử/lọc, category hierarchy/activation; copy/bulk theo contract riêng | Accounting fixture + SQL + API cookie/bearer + browser reload/offline. |
| R2 — budgets/reports | Wallet scope, category roll-up, forecast, lịch sử kỳ và toàn bộ drilldown | Golden dataset và công thức độc lập; 75/80 được quyết định rõ; ảnh và thao tác trên viewport mobile/desktop. |
| R3 — planning | Debt cashflow, credit/goal, bill/schedule/confirm, travel | Nháp chỉ ghi tiền một lần sau duyệt; vòng tạo → sửa → reload → kết thúc → archive. |
| R4 — ingestion/data | AI/OCR/export/account lifecycle theo các ticket đang deferred | Provider integration thật, idempotency/ownership/negative cases, dữ liệu xuất và cleanup kiểm chứng. |
| R5 — shared/native | Shared wallet và các adapter platform đã chọn | ADR quyền truy cập và platform POC trước; không giả lập native bằng một nút PWA. |

Với mọi feature mới: dùng base component hiện có hoặc tạo base phù hợp; dữ liệu từ API/DB, không sample/hardcode. Một feature chỉ được ghi đạt khi người dùng truy cập được từ App, lưu đúng, reload còn đúng, lỗi có phản hồi và tests kiểm tra kết quả tài chính thật.

## Checklist kết thúc audit

- [x] Kiểm kê 96/96 bài public, mapping từng bài và nêu giới hạn nguồn.
- [x] Đối chiếu docs → source → UI mount → test; không dùng số lượng file để tính parity.
- [x] Chạy lại backend/PostgreSQL/frontend/build/mobile E2E và ghi pass/fail/skipped riêng.
- [x] Cập nhật catalog, feature matrix, verification, context, backlog, validation và changelog.
- [x] Đánh dấu research cũ và sửa những ví dụ portal đã xác minh sai.
- [x] Product UAT không áp dụng cho thay đổi docs-only; UAT chức năng/parity vẫn chưa đạt.
- [x] Không tạo ADR mới vì không đổi quyết định kiến trúc, schema, API hoặc runtime.

Theo dõi: [verification](../../work/test-verification/MONEYLOVER-PARITY-2026-09-08.md), [backlog](../../work/BACKLOG.md), [requirements](../../requirements/REQUIREMENTS.md), [API](../../architecture/API.md), [changelog](../../releases/CHANGELOG.md).
