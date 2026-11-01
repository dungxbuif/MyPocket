---
artifact_type: detail_design
id: NATIVE-NAVIGATION-UX-20261101
status: draft
owner: shared
approval: planning_requested
---

# Điều hướng và bố cục tác vụ native MyPocket

## Authority và sửa kết luận audit

Owner xác nhận: Assistant là một màn riêng, ẩn navigation chính và bubble.
Đây là yêu cầu đã có trong hội thoại, được nhắc lại sau audit. Finding đòi
khôi phục bubble trên Assistant là false positive; ba reviewer đồng thuận
không thay thế được quyết định owner. Không khôi phục bubble trên Assistant.
Ẩn thanh điều hướng chính không đồng nghĩa bỏ nút đóng/quay lại trong màn.
Bubble Phản hồi vẫn giữ ở các màn thông thường theo policy dùng chung.

Yêu cầu mới: tận dụng stack, màn chi tiết và cách trình bày tác vụ native;
giữ base styling, palette, font/card language hiện tại. Đây là plan UX,
chưa phải bằng chứng đã triển khai. Dark mode không thuộc nghiệm thu.

## Evidence và vấn đề

- `MainTabView.swift:47-56`: Ví/Hũ/Tháng/Thông báo từ Tổng quan bị chuyển sang
  tab Tài khoản. Người dùng mất quan hệ giữa màn xuất phát và màn chi tiết.
- `MainTabView.swift:84`: Assistant đóng về Tổng quan bất kể nguồn mở.
- `MainTabView.swift:101-105` và `MoreView.swift:253`: có thêm đường mở
  Assistant bên trong More; hai đường vào có thể khác shell/draft behavior.
- `MoreView.swift:164-170,210-214`: đổi Hũ hoặc Back thay/xóa toàn bộ path;
  chưa biểu diễn lịch sử nhiều tầng theo hành trình thực tế.
- Shell tự tính chiều cao, clip viewport và chặn hit testing cho tab Tài khoản.
  Cần thay ownership bố cục bằng safe-area/native presentation có kiểm chứng.

Đã đọc context/backlog hiện hành, standards, Phase 002, thiết kế composer,
native base guideline, MainTabView, MoreView, AppRoute và AppRouteCoordinator.

## Hướng đề xuất

Chọn native navigation container với base MyPocket bên trong. Mỗi tab giữ
stack riêng; sheet dành cho tác vụ ngắn, push dành cho nội dung có cấp bậc.
Assistant là không gian toàn màn hình tách khỏi tab. Không thêm framework.

Đề xuất 4 tab: **Tổng quan · Giao dịch · Báo cáo · Tài khoản**. Báo cáo thay vị
trí tab Assistant; Trợ lý trở thành action rõ nhãn trong toolbar Tổng quan và
Báo cáo, đồng thời có entry tại Tài khoản. Add là action mở editor, không đổi
tab. Thứ tự/tab mới là đề xuất chờ review, không phải quyết định owner đã duyệt.
Hũ/Ngân sách/Tháng có shortcut trong Tổng quan và nhóm Kế hoạch trong Tài khoản;
Ví/Tài sản tích lũy có shortcut Tổng quan và nhóm Tài sản trong Tài khoản.

| Hành trình | Presentation | Quay lại / ngữ cảnh |
| --- | --- | --- |
| Tổng quan → Ví → chi tiết ví → giao dịch của ví | Push cùng stack | Pop từng tầng; giữ ví, kỳ, vị trí cuộn |
| Tổng quan → Hũ → chi tiết Hũ; Tháng → Hũ | Push cùng stack | Giữ tháng tài khoản; không tạo carry-over mới |
| Báo cáo → danh mục/kỳ → giao dịch đã lọc → chi tiết | Push cùng stack | Giữ bộ lọc và biểu đồ gốc; không đổi tab Giao dịch |
| Tài sản tích lũy → tài sản → lịch sử mua/bán | Push | Giữ tài sản/unit/source; quyền sửa theo backend |
| Ngân sách, danh mục, định kỳ, tín dụng, du lịch, thông báo | Push từ entry hiện tại | Một Back chuẩn, pop đúng một tầng |
| Thêm/sửa giao dịch, tài sản, Hũ, ngân sách, định kỳ | Sheet lớn có stack editor | Lưu đóng editor về nguồn; lỗi giữ draft |
| Chọn ví/danh mục/Hũ/unit, lọc kỳ | Sheet chọn ngắn hoặc push trong editor | Tìm kiếm nếu danh sách dài; không chồng nhiều sheet |
| Nhập tiền và máy tính | Base amount input hiện tại | Không đổi route hoặc mất draft khi bật bàn phím |
| Trợ lý từ bất kỳ entry nào | Full-screen presentation, stack riêng | Đóng về đúng origin; không tab bar, Add hay bubble |
| AI preview → sửa một giao dịch | Editor của giao dịch trong context Assistant | Lưu quay về đúng card; không tạo một UI editor riêng |
| Hồ sơ/cài đặt và inbox phản hồi | Push | Bubble mở tác vụ phản hồi ngắn và trả đúng màn |

Tab chuyển khu vực giữ path/scroll/filter riêng. Bấm lại tab đang chọn không
reset stack trong vòng đầu. Cross-link trong app mở chi tiết tại stack hiện
tại; deep link mở tab canonical (ví/tài sản/kế hoạch tại Tổng quan, báo cáo tại
Báo cáo, hồ sơ tại Tài khoản). URI hiện có phải tiếp tục hoạt động.

## Assistant

- Một presentation owner tại shell; mọi entry gọi cùng `presentAssistant`.
- Nội dung chat chiếm phần còn lại màn hình; composer theo keyboard safe area.
- Root có một nút Đóng. Màn con trong Assistant có Back chuẩn. Không hiện đồng
  thời nhiều nút Back/Close làm cùng một việc trên root.
- Edge-swipe tại màn con pop; tại root dismiss về origin. Full-screen cover
  không tự cung cấp root edge-pop: phải làm qua shared presentation bridge,
  có kiểm tra cancel gesture và xung đột cuộn ngang, không gắn drag toàn màn.
- Đóng không đồng nghĩa hủy job AI. Quan sát job tiếp tục theo lifecycle đã có;
  mở lại phục hồi session/draft/scroll trong cùng phiên đăng nhập. Logout xóa
  context theo owner. Không hứa lưu draft qua process kill nếu chưa có cơ chế.
- Attachment, insight và preview là thành phần hội thoại. Preview dùng base
  transaction/action hiện có. Lỗi FE/BE entry contract trong audit là phụ thuộc
  riêng; thay navigation không được coi là đã sửa chúng.
- Bố cục chi tiết theo reference Assistant sẽ có bước đối chiếu riêng. ZIP
  hiện chưa xác minh được; không tự lấy ZIP keypad làm chuẩn chat.

## Base và trạng thái

Reused: MPText/MPFont, MPSurfaceCard, MPTransactionRow, MPGroupedRow,
MPAmountField/MPAmountKeypad, MPEditorScaffold, MPChatComposer,
MPAttachmentViews, MPIconCatalog, MPFeedbackBubble overlay và shared presentation helpers.
Feedback tuyệt đối không tạo dock/row/spacer hoặc làm giảm route viewport.
Native toolbar/header adapter sở hữu Back/Close; feature chỉ truyền title,
actions, content. Không render thêm custom header nếu native bar đã sở hữu nó.
Safe-area layout do container sở hữu, tránh trừ chiều cao ở nhiều tầng.

Dirty editor: dismiss yêu cầu chọn tiếp tục sửa hoặc bỏ thay đổi. Đang submit:
chặn dismiss gây trạng thái mơ hồ, giữ recovery hiện có; không tự gửi lại.
Picker trong editor dùng stack editor để tránh nested modal. Khi overlay đang
mở, giữ deep link pending; xử lý sau dismiss hoặc quyết định bỏ draft.
Record bị xóa/không còn quyền: hiện unavailable có Back, không tự reset shell.

## Alternatives và phạm vi

- Giữ shell switch hiện tại: ít diff nhưng tiếp tục mất nguồn mở; không chọn.
- Dùng sheet cho mọi màn: khó đi sâu/lịch sử; không chọn.
- Đổi toàn bộ visual style/Liquid Glass custom: không chọn trong đợt này.
  Dùng chrome native theo OS; card/icon màu/typography vẫn theo base đã chọn.

Không đổi API, DB, auth, financial semantics, model AI hoặc deployment.
Missing backend filters/drilldown phải ghi blocked trong flow tương ứng;
không tải một trang giao dịch rồi mô tả là toàn bộ kết quả.
Rủi ro: duplicate header, reset view model, lộ draft khác owner, mất filter,
duplicate submission, gesture xung đột, bubble che footer. Các gate dưới xử lý.

## Nghiên cứu

Apple khuyến nghị tab cho khu vực cấp cao và giữ trạng thái từng khu vực;
sheet cho tác vụ ngắn gắn với context, tránh dùng sheet để duyệt toàn app.
Hướng tab/stack/sheet ở đây là áp dụng cho MyPocket, không phải Apple yêu cầu
đúng bốn tab hoặc đúng tên tab này.

- [Tab bars](https://developer.apple.com/design/human-interface-guidelines/tab-bars)
- [Sheets](https://developer.apple.com/design/human-interface-guidelines/sheets)
- [Navigation design](https://developer.apple.com/videos/play/wwdc2022/10001/)

## Nghiệm thu và trace

Đo trên dataset cố định, iPhone nhỏ và lớn, chữ thường/lớn, keyboard mở/đóng.
Mỗi flow trong bảng phải có entry → detail → Back và cancel/save/error proof.
Assistant từ Tổng quan/Báo cáo/Tài khoản/deep link luôn ẩn tab/bubble, đóng
trả đúng origin; gesture hủy nửa chừng không mất trạng thái. So sánh screenshot
và thao tác thực; unit pass không thay thế nghiệm thu bố cục.
Mỗi task/flow phải được ba reviewer độc lập ghi rõ covered/not covered và
evidence; không dùng tổng số reviewer làm bằng chứng coverage từng chức năng.

[Plan](../../superpowers/plans/2026-11-01-native-navigation-ux.md) ·
[Backlog](../BACKLOG.md) · [Phase](../phases/PHASE-002-native-ios-swiftui.md) ·
[Validation](../VALIDATION_MATRIX.md) ·
[Registry](../../design/platforms/ios-screen-registry.md) ·
[Native bases](../../design/platforms/ios-base-guideline.md) ·
[Release](../../releases/REL-1.3.0-20261101.md).
Khi implement: reconcile screen registry, Assistant spec, native base guideline,
architecture navigation ADR, context/backlog/validation/changelog.
API/ERD ghi no-change. Approval cho implementation navigation mới còn pending.
