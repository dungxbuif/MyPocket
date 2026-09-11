---
artifact_type: detail_design
id: SO-TIEN-2026-09-09
status: in_review
owner: shared
approach_approval: Owner replied "ok" to functionality-first then full monochrome So Tien redesign.
written_spec_approval: Owner requested "Imple all đi" after receiving this written spec.
implementation_status: F1_local_verified_in_review; F2_contract_design_next
---

# Sổ tiền — chức năng đúng trước, giao diện mới sau

## 1. Quyết định và phạm vi

Người dùng đã duyệt hướng **Sổ tiền**, phong cách editorial tối giản, trắng–đen chủ đạo, thay toàn bộ giao diện sau khi sửa chức năng. Quyết định này thay yêu cầu cũ "giữ UI" ở chặng giao diện; không hủy các giới hạn dữ liệu, bảo mật, API key hay phạm vi tính năng đã hoãn.

Đây là đặc tả điều phối hai chặng, không phải chứng nhận sản phẩm hoàn thiện. Các nhóm chức năng cần API/quy tắc mới sẽ có thiết kế con trước khi viết code. Không gộp toàn bộ thành một lần sửa lớn. Chưa deploy, chưa chạy migration production; dữ liệu và pending outbox hiện hữu phải được giữ nguyên.

Nguồn: [context](../../CONTEXT.md), [backlog](../../work/BACKLOG.md), [R0](../../work/phases/R0-DETAIL_DESIGN.md), [thiết kế kiểm chứng finance](../../work/phases/R0-FUNCTIONAL-E2E-DETAIL_DESIGN.md), [kết quả và khoảng trống hiện tại](../../work/test-verification/R0-FUNCTIONAL-E2E-2026-09-09.md), [architecture](../../architecture/ARCHITECTURE.md), [TICKET-017](../../work/tickets/TICKET-017-analytics-reports-cumulative-trends.md), [validation](../../work/VALIDATION_MATRIX.md), [changelog](../../releases/CHANGELOG.md).

Đã đọc context, backlog, standards README/QUALITY_BAR/VALIDATION/DEBUGGING/GIT, các thiết kế R0 và ticket báo cáo; đối chiếu trực tiếp App, BudgetsScreen, OverviewScreen và theme hiện tại. Checklist này ghi trạng thái thiết kế, không thay bằng chứng chạy test.

## 2. Trình tự và điều kiện chuyển chặng

| Chặng | Kết quả bắt buộc | Điều kiện chuyển tiếp |
| --- | --- | --- |
| F1 — sửa tính đúng của luồng hiện có | Tổng ngân sách không cộng trùng; biểu đồ không chọn tương lai; lỗi tải/lưu có phản hồi; giữ các sửa ví/chuyển khoản trước đó | Regression RED→GREEN và E2E dữ liệu thật trên database riêng |
| F2 — hoàn tất hành trình chức năng | Danh mục, giao dịch/điều chỉnh, planning/draft, portfolio, báo cáo và điều hướng tìm kiếm | Từng hành động có API/UI/điều kiện offline/bằng chứng rõ ràng; API còn thiếu phải có thiết kế con được duyệt |
| F3 — kiểm thử và docs chức năng | Kiểm thử đầy đủ, docs nguồn khớp code, phân loại lỗi WebKit và kiểm chứng thiết bị/provider | Không còn lỗi chức năng nghiêm trọng chưa giải quyết; mọi giới hạn còn lại phải được người dùng chấp thuận trước khi chuyển UI |
| U1 — hệ thiết kế và shell | Base components, typography, navigation, responsive layout | Kiểm thử keyboard/focus/contrast và xem ảnh render ở mobile/desktop |
| U2 — thay toàn bộ màn hình | Mọi hành trình F1–F2 chạy trên UI mới, kể cả dialog và màn hình phụ | Cùng bộ E2E dữ liệu thật vẫn qua; không có màn hình production dùng mẫu giả |
| U3 — loại bỏ UI cũ và nghiệm thu | Xóa đúng module/style không còn sử dụng; cập nhật ảnh/hướng dẫn | Import/dependency audit, build, unit, E2E, UAT; không xóa adapters/domain/offline vì đổi UI |

Triển khai theo thứ tự trên. Không tự bỏ F2 để sớm khoe UI. Các giới hạn cần thiết bị/credentials không thể coi là đã giải quyết bằng giả lập. Nếu cản F3, báo điều kiện cần xác nhận thay vì tự ý hạ tiêu chí. Việc phát hành là một bước riêng sau nghiệm thu, không tự được thực hiện bởi bản đặc tả này.

## 3. F1: thiết kế sửa tính đúng của các luồng hiện có

### Ngân sách

Hiện `BudgetsScreen` cộng `amount_vnd` và `spent_vnd` của tất cả budget nhưng dùng kỳ của phần tử đầu tiên. Hai budget bao phủ cùng giao dịch hoặc khác kỳ khiến con số tổng gây hiểu nhầm. Thay phần hero bằng **một ngân sách được chọn**, nhãn tên/phạm vi/kỳ khớp chính xác dữ liệu budget đó. Danh sách các budget độc lập giữ nguyên. Không gọi tổng các hạn mức chồng lấn là "tiền còn có thể chi".

Chọn lần đầu theo thứ tự danh sách API; giữ lựa chọn bằng ID khi refresh; nếu ID không còn tồn tại thì chọn budget còn lại đầu tiên. Empty/loading/error riêng. Hiển thị phần vượt hạn mức thay vì làm mất thông tin bằng clamp về zero. Không thay công thức backend hoặc tự tính lại `spent_vnd` từ một trang transaction thiếu dữ liệu.

### Tổng quan và report

`buildDailyExpenseBars` đang lấy bảy dòng cuối của cả tháng, có thể toàn ngày tương lai. Lấy tối đa bảy ngày kết thúc tại ngày hiện tại theo Asia/Ho_Chi_Minh trong phạm vi report; nhãn thể hiện đúng khoảng ngày. Trường hợp kỳ đã kết thúc lấy ngày cuối kỳ; kỳ hoàn toàn tương lai hiển thị chưa có ngày phát sinh, không dùng dữ liệu mẫu. Tổng kỳ vẫn ghi rõ là tổng kỳ, không gán tổng tháng cho bảy ngày.

Giữ nguyên năm report hiện có và số liệu từ server. Che cả số tiền, tỷ lệ và nội dung tooltip/accessible name khi bật riêng tư. Loading, lỗi và rỗng sau response thành công khác nhau. Không lấy `null`/response lỗi làm số zero. Tải chi tiết ví/Money Insider/mark-read/logout cần phản hồi lỗi an toàn và không bị response cũ ghi đè khi đổi tài khoản, chọn mục khác hoặc đóng.

### Chuyển ví

Giữ một giao dịch chuyển khoản với hai ví khác nhau, không tạo hai income/expense độc lập. Regression giữ nguyên các kỳ vọng 1.000.000 → 750.000/250.000 → 600.000/400.000 → 1.000.000/0 sau tạo/sửa/lưu trữ. Cả replay và lỗi phải không nhân đôi giao dịch. Đối chiếu pending offline với kết quả sau sync; không chỉ kiểm tra toast thành công.

## 4. F2: phạm vi hoàn tất chức năng, không hạ thành nút giả

| Nhóm | Hành trình cần hoàn tất | Ràng buộc |
| --- | --- | --- |
| Danh mục | Tạo/sửa/lưu trữ; cây cha–con; kích hoạt theo ví | Tôn trọng danh mục hệ thống, owner và lịch sử giao dịch; không sửa trực tiếp database để né API |
| Giao dịch | Bộ lọc kỳ; thêm/sửa ngày, ví, danh mục, ghi chú; chuyển khoản và điều chỉnh số dư | Giữ integer VND, phiên bản optimistic và idempotency; điều chỉnh hiển thị số dư đích khác delta |
| Planning | Ngân sách, sự kiện, nợ/trả nợ, lịch lặp, duyệt bản nháp | Tạo bản ghi rồi link lỗi không được tạo lại bản ghi khi retry; chỉ confirm mới tác động ví |
| Portfolio | Danh sách/chi tiết, mua/bán ghi sổ, giá thủ công, lưu trữ | Đây là ghi nhận nội bộ, không giao dịch sàn; giá trị tài sản tách ví theo ADR hiện có; privacy nhất quán |
| Báo cáo | Năm loại, biểu đồ kèm bảng, lọc ngày/ví, drilldown khớp giao dịch, Money Insider theo thiết kế đã có | Công thức server, HCM timezone, không lẫn dòng tiền với số dư; cache theo owner/kind/filter có timestamp |
| Tìm kiếm/thông báo | Kết quả mở đúng thực thể, read/unread và retry | Không im lặng nuốt lỗi; không lộ tài khoản cũ |
| API key/docs | Cookie và bearer có contract proof theo từng endpoint | Quản lý key giữ yêu cầu phiên web; không đổi quyền hay để secret lọt vào browser bundle/docs |

Các thiết kế con bắt buộc trước những thay đổi còn chưa có contract: confirm/reject draft và tính nguyên tử/idempotency; hỗ trợ số dư điều chỉnh zero/âm (API hiện chỉ chấp nhận dương); các trường/API report/Insider còn thiếu; metadata/cache migration nếu thực sự cần. Bản này không tự quyết public routes/schema hoặc thay công thức để hợp UI.

AI/OCR, bank ingestion, voice, xuất dữ liệu và reset/xóa tài khoản vẫn theo phạm vi hoãn trước đây; không suy diễn "fix nốt" thành bật tất cả subsystem hoãn. Các phần này phải được ghi rõ trong docs, không trình bày là chức năng hoạt động.

## 5. Hướng thẩm mỹ Sổ tiền

Chọn editorial ledger vì app cần đọc tiền, kiểm tra lịch sử và thao tác nhanh. Hai phương án không chọn: dashboard thẻ dày đặc làm phân tán trọng tâm; giao diện ngân hàng tối/màu nhấn mạnh không phù hợp yêu cầu đơn sắc và lịch sử dài. Đây là giao diện mới của MyPocket, không tiếp tục cam kết giống pixel Money Lover; "Sổ tiền" là tên concept, không đổi tên ứng dụng/domain.

### Token và bố cục

- Nền `#FAFAF8`, surface `#FFFFFF`, chữ chính `#151515`, chữ phụ `#5F5F5B`, đường kẻ `#DADAD5`. Primary button đen/chữ trắng; secondary có viền. Không gradient, glass panel hay background trang trí.
- System sans hỗ trợ tiếng Việt; số tabular. Body 16 px, nhãn 13–14 px, heading 24–32 px, số dư trọng tâm 36–44 px. Số dài phải co theo breakpoint/wrap hợp lý, không tràn ngang màn hình.
- Thang khoảng cách 4/8/12/16/24/32/48 px; radius 8 px controls, 12 px surfaces; ít shadow, chủ yếu phân vùng bằng khoảng trắng và đường kẻ.
- Mobile dưới 768 px: một cột, padding 16 px, bottom navigation năm mục, safe area; thêm giao dịch bằng nút có nhãn trong header/trang, không lấy một tab giả để làm action.
- Desktop từ 1024 px: sidebar 216 px, nội dung tối đa 1120 px; vùng giữa 768–1023 giữ shell gọn và không ép layout desktop. Bảng rộng cuộn bên trong vùng có nhãn, không kéo tràn toàn trang.
- Thu/chi dùng nhãn, dấu +/−, mũi tên và nét đặc/rỗng. Chuyển khoản trung tính và có nguồn → đích. Không dùng chỉ màu để mang nghĩa. Lỗi có icon + câu chữ + action, không phụ thuộc đỏ.
- Tối thiểu 44×44 px cho vùng chạm chính; focus đen rõ; hỗ trợ reduced motion, keyboard, modal focus trap/return focus, escape và safe-area keyboard.

### Mọi màn hình trong phạm vi đổi UI

| Màn hình | Thứ bậc nội dung |
| --- | --- |
| Tổng quan | Số dư và phạm vi → thu/chi kỳ → giao dịch gần nhất → tóm tắt kế hoạch/Insider; không lặp nhiều thẻ tổng tiền |
| Giao dịch | Kỳ/ví/tìm kiếm → danh sách nhóm ngày, số căn phải → thao tác thêm/sửa/ảnh; filter state hiển thị rõ |
| Ví | Danh sách ví → chi tiết/số dư/lịch sử → chuyển tiền/điều chỉnh/quản lý danh mục; tài sản là mục phụ rõ ràng, không gộp sai vào tiền mặt |
| Báo cáo | Khoảng ngày/ví/loại → summary → biểu đồ đơn sắc → bảng số/drilldown; timestamp và trạng thái offline cùng phạm vi |
| Kế hoạch | Tab con ngân sách/sự kiện/nợ/lịch lặp/bản nháp; chọn một budget cho summary; action thực có trạng thái hoàn tất/lỗi |
| Tài khoản | Menu riêng từ header/sidebar: profile, privacy, API keys, notifications, docs, logout; audit chỉ hiện với quyền hiện hành |
| Phần chung | Login/loading/forbidden; search; notification inbox; receipt picker; conflict inbox; tất cả form/dialog/empty/error và PWA prompt |

PWA prompt vẫn là thông báo nhỏ ở góc dưới, blur nhẹ chỉ tại prompt, có đóng và dành khoảng trống để nội dung không bị che. Không phủ blur lên toàn app, không tự gọi cài đặt nếu chưa có hành động người dùng. Không xây thêm dark-mode toggle ở đợt này.

## 6. Component, dữ liệu và thay thế an toàn

```text
AppShell / điều hướng
  -> Screen controllers (owner, selection, requests, mutation state)
       -> Domain views (Wallet / Transaction / Budget / Report)
            -> Base UI + semantic tokens
       -> Existing API adapters / user-scoped offline store
            -> Existing Go domains / PostgreSQL
```

Một nguồn base component: Button/IconButton, Input/Select/Textarea/Checkbox, Field, Surface, Sheet/Dialog, Tabs/Navigation, DataTable, MoneyText, EmptyState, LoadingState, OperationError và chart primitives. Chỉ tạo thêm khi chưa có thành phần tương đương; bỏ wrapper trùng sau khi chuyển hết consumers. Business logic không đặt trong Button/Surface và không có screen tự tạo bản riêng cùng chức năng.

App hiện chứa nhiều inline sheets. Tách controllers/forms đang được thay thành module theo feature; không viết lại auth/cache/outbox cùng lúc với CSS. Nhận diện màn hình thật từ import/render graph, không chỉ nhìn tên file prototype. Trước xóa: kiểm tra imports, dynamic references, tests, style usage; chạy lại toàn bộ checks. Không xóa dữ liệu, migrations, tài liệu nguồn hoặc chức năng chưa chuyển xong. Giữ nguyên các chỉnh sửa sẵn có; không stage/commit lẫn chúng. Phạm vi R0 hiện yêu cầu no commits nên bản thiết kế này cũng được để local, không tự commit.

Mutation: giữ input khi thất bại; chống double-submit; retry giữ mutation identity; không retry mutation tài chính âm thầm; partial success ghi nhận bước đã xong. Read: abort/ignore stale response theo owner và filter; dữ liệu cached có thời điểm/phạm vi; report không lấy pending transaction làm confirmed. Không thêm thư viện UI/chart/router lớn nếu primitive/dependency hiện có đủ dùng.

## 7. Kiểm chứng và định nghĩa hoàn tất

Baseline gần nhất là lịch sử, không phải kết quả mới của đặc tả: 120 frontend pass, Go/PostgreSQL pass, E2E 61/63; hai WebKit offline failures vẫn mở. Khi triển khai phải chạy lại trên mã hiện tại.

- Unit RED→GREEN: budget trùng phạm vi/khác kỳ/vượt hạn mức/zero/error, HCM midnight và preview ngày; privacy ở cả DOM và chart; stale responses, retry/partial success.
- Integration: sở hữu dữ liệu, số dư hai ví, chỉnh sửa/lưu trữ/replay, category constraints, report exclusion, draft atomicity khi bổ sung, bearer/cookie parity và revoke.
- E2E dùng database riêng: toàn bộ hành trình F1–F2; dữ liệu kỳ vọng tính độc lập; kiểm tra số tiền sau reload/sync chứ không chỉ toast. Cùng tests tiếp tục chạy khi đổi UI.
- Visual/UAT: 390×844, 768×1024, 1440×900; số tiền dài, tiếng Việt, keyboard mở, offline/error/empty, nhiều dòng và modal; không clipping, dock che action, font thiếu hoặc layout cũ sót lại.
- Build/typecheck, toàn bộ frontend tests, Go integration và bộ Playwright hiện hành phải có command/exit/count. Repo dùng `rtk`; Node 24 từ PATH hiện có. Không nâng dependency để né lỗi test.
- Hai lỗi WebKit phải được điều tra theo chuẩn DEBUGGING, giữ test gốc. Nếu chứng minh là giới hạn giả lập, ghi bằng chứng độc lập và yêu cầu nghiệm thu Safari/PWA thật; không tự gọi đó là đã sửa. Loop guard 3 lần sửa cùng path hoặc 5 chu kỳ/work item vẫn áp dụng.

Chỉ công bố toàn bộ UI đã thay khi mọi hàng màn hình có ảnh/test và không còn runtime dùng UI cũ. Chỉ công bố chức năng hoàn thiện khi mọi hàng F1/F2 đã nghiệm thu hoặc có ngoại lệ được duyệt rõ. Mỗi chặng có verification artifact liên kết vào validation/context/backlog; chưa có artifact thực thi thì không đánh dấu pass.

## 8. Docs và phát hành

Docusaurus cần: hướng dẫn theo từng hành trình, quy tắc thu/chi/chuyển khoản/điều chỉnh với ví dụ số, báo cáo và timezone, lỗi/offline, API key integration, schema/request/response đầy đủ, bảng từng action **API / UI / test / phiên bản production**. Spec OpenAPI máy đọc được phải được kiểm tra tự động trước khi gọi là OpenAPI hoàn chỉnh. Mỗi trang phân biệt planned/local-tested/released.

Reconciliation: cập nhật REQUIREMENTS và architecture chỉ theo hành vi đã thực hiện; API.md/Docusaurus theo contract thực; ERD chỉ đổi khi có schema được duyệt; ADR riêng khi thêm API/data/security boundary. Backlog/context/validation/changelog cập nhật theo từng chặng. Không đưa ý tưởng chưa triển khai vào master docs như sự thật.

Release sau cùng cần build mới, image identity, local smoke, authenticated app/docs check, public asset/version và rollback. Không reset database hoặc thay secret/API-key policy để phục vụ giao diện. Cần quyết định release riêng; sự đồng ý hướng thiết kế không đồng nghĩa deploy ngay.

## 9. Tự rà soát và điểm dừng hiện tại

- [x] Hai chặng được tách; không làm UI trước khi hoàn tất/được chấp thuận giới hạn chức năng.
- [x] Concept, token, navigation, base components và toàn bộ nhóm màn hình xác định rõ.
- [x] Phân biệt defect hiện có, missing contract, deferred subsystem, physical/provider gate.
- [x] Không có số liệu mẫu, không giảm assertion, không hứa xử lý engine bằng workaround chưa có chứng cứ.
- [x] Bảo toàn API key, dữ liệu/outbox, checkout và đường rollback; không tự deploy.
- [x] Người dùng yêu cầu "Imple all đi" sau khi nhận đặc tả. [Kế hoạch F1](../plans/2026-09-09-so-tien-f1.md) bắt đầu thực thi; tiếp theo là thiết kế con F2 theo các contract còn thiếu. Đây không phải bằng chứng hoàn tất code.
