---
artifact_type: detail_design
id: NATIVE-CHAT-COMPOSER-20261031
status: draft
owner: shared
approval: planning_requested_implementation_not_resumed
---

# MyPocket: chữ gọn, chat dễ dùng, icon nhẹ hơn

## Yêu cầu đã xác nhận

- Giữ phong cách card, màu sắc và tổng thể bản đang cài trên máy ảo.
- Chữ hiện hơi to/nặng; ưu tiên hiển thị nhiều nội dung hữu ích hơn.
- Màn AI cần thân thiện, vùng soạn đủ lớn; tham khảo Codex và Claude.
- **Giữ bong bóng chat/phản hồi. Không xóa, thay bằng menu, hoặc làm mất tính năng.**
- Icon gần tinh thần Money Lover hơn, tránh mảng đen/tô đặc nặng như hiện tại.
- AI Assistant và các item preview/result giao dịch phải dùng cùng card language với màn `Thêm giao dịch`; không dựng một card/bubble transaction riêng chỉ cho AI.
- Yêu cầu hiện tại là nghiên cứu và lập plan; không tiếp tục sửa runtime.

Trước khi chủ yêu cầu chuyển sang plan, đã có nháp chưa kiểm thử trong
`Tokens.swift`, `MPFields.swift`, `AssistantViews.swift`, `MainTabView.swift`.
Nháp chưa build/cài/commit; không phải baseline được duyệt. Đặc biệt phương án
composer luôn lớn và đẩy bong bóng lên trên bị thay thế bởi thiết kế bên dưới.
Không xóa hoặc ghi đè nháp trong bước lập plan này.

## 1. Typography và mật độ nội dung

Giữ SF Pro trong vòng đầu; không thêm font mới cùng lúc thay layout.
`MPFont.custom(size:relativeTo:weight:)` đang bỏ qua `size`: các con số cũ
không chứng minh cỡ chữ runtime. Thay bằng vai trò native Dynamic Type rõ ràng.

| Vai trò | Cỡ mặc định đề xuất | Độ đậm |
| --- | --- | --- |
| Nội dung, tên ví/giao dịch, chat và ô nhập | 15pt (`subheadline`) | regular |
| Nội dung phụ/ghi chú | 13pt (`footnote`) | regular |
| Nhãn phụ ngắn | 12pt (`caption`) | regular |
| Tiêu đề section/màn | 17pt (`headline`) | semibold |
| Số tiền trong dòng | 16pt (`callout`), tabular | medium |
| Số dư chính | 34pt (`largeTitle`), tabular | semibold |

Cỡ trên là ở Dynamic Type mặc định, không phải chiều cao cố định. Không giảm
hit target dưới 44pt, không scale-to-fit chữ dài, không cắt số tiền. Giảm khoảng
trống trong header/nhãn lặp trước khi giảm kích thước thông tin quan trọng.
Pilot trên Tổng quan, Giao dịch, AI; shared token thay đổi phải rà cả Ví,
Thêm giao dịch và Tài khoản. Không thay cấu trúc card toàn app.

## 2. Màn AI thích ứng theo trạng thái

- Bắt đầu, draft trống: lời chào “Mình có thể giúp gì cho bạn?”, tối đa một
  câu giải thích; composer tối thiểu 3 dòng, không chèn hàng nhãn trang trí.
- Có hội thoại: composer tối thiểu 1 dòng, tự nở tối đa 4 dòng; còn lại cuộn
  trong input. Chuyển trạng thái không xóa draft hoặc nhảy vị trí con trỏ.
- Bàn phím mở: tối thiểu 2 dòng ở cỡ chữ thường. Accessibility: 1–2 dòng lớn
  tùy không gian; luôn giữ ít nhất 44pt hội thoại và nút gửi trên bàn phím.
- Có thể dùng khoảng 100–120pt lúc bắt đầu và 56–72pt lúc hội thoại làm mốc
  thử, **không dùng chúng làm fixed height hay tiêu chí cắt chữ**.
- Chuyển Thu/Chi ghim sang nút Tổng quan trên header, mở đúng sheet hiện có.
- Gợi ý là một hàng cuộn ngang, chỉ khi chưa có tin và draft trống. Bấm chỉ
  điền draft và focus; không tự gửi, không ghi đè nội dung đã viết.
- Đang đọc tin cũ: giữ vị trí; tin mới hiện nút “Tin mới” để chủ động xuống.
  Chỉ tự cuộn khi người dùng đang ở cuối. Nạp lịch sử cũ giữ tin đang đọc.
- Giữ send/stop, retry, provider-unavailable, privacy và bản nháp hiện có.
  Không biến lỗi/cảnh báo tài chính thành nội dung bị giấu.
- Các preview/result giao dịch trong Assistant/AI Entry dùng lại shared reference
  surface và `MPTransactionRow`/shared transaction fields của màn `Thêm giao dịch`:
  cùng icon footprint, signed amount, ngày, note, jar/category metadata và action
  row. Chat bubble chỉ bao quanh phần hội thoại; không bọc thêm một card lồng
  riêng làm thay đổi hierarchy của transaction.

## 3. Bong bóng được giữ với chỗ riêng

Bong bóng hiện tại là **Gửi phản hồi**, khác tab Trợ lý; giữ action, capture và
accessible label. Trên AI, đặt circle 52pt ở trailing của cùng vùng footer,
composer ở cạnh với gap 8pt. Cả hai được layout đo thực, không overlay lên
input/send/transcript và không thêm một hàng trống 60pt phía trên.
Đây là tradeoff có chủ ý: bớt khoảng 60pt chiều ngang input để giữ diện tích đọc
theo chiều dọc. Thử tên/nội dung tiếng Việt trên máy 375pt trước khi chốt.

Giữ bong bóng hiện diện trên AI cả khi bàn phím mở, trừ lúc overlay phản hồi
đang mở hoặc đang chụp màn hình; footer đi theo safe area bàn phím. Mở/đóng phản
hồi không làm mất draft. Các màn khác giữ hành vi hiện tại trong đợt này.
Shell tiếp tục sở hữu việc gửi phản hồi; screen chỉ báo trạng thái composer và
nhận callback, không tạo nút gửi phản hồi thứ hai.

## 4. Hũ là flow sản phẩm, không chỉ là một shortcut

Native đã có `JarsView`, `MonthsView`, các endpoint Hũ theo tháng/cộng dồn, link
từ Tổng quan/Tháng, và selector Hũ trong editor giao dịch/AI. Tuy vậy, phạm vi
redesign trước đó chưa đưa các màn này vào review nên chưa thể coi trải nghiệm
Hũ là hoàn chỉnh.

Phạm vi cần đưa vào nghiệm thu:

- Entry point từ Tổng quan → `Hũ tháng`, Tài khoản → `Hũ chi tiêu`, và Tháng →
  Hũ phải giữ đúng tháng/ngữ cảnh và có đường quay lại dễ hiểu.
- Màn Hũ phải đọc được theo thứ tự: tháng → tổng quan tiền → danh sách Hũ và
  tiến độ → chi chưa gắn Hũ → cộng dồn/lịch sử.
- Tạo/sửa cần làm nổi bật tên + cách phân bổ + trường tiền/tỷ lệ; lỗi phải giữ
  bản nháp. Hũ lịch sử vẫn xem được nhưng không được sửa nhầm cấu hình cũ.
- Trạng thái rỗng, loading, lỗi, vượt phân bổ và chi vượt mức phải dùng chung
  hierarchy/card/typography của app. Không gọi chênh lệch cộng dồn là “số dư Hũ”
  và không tạo carry-over ngầm.
- Form giao dịch và AI tiếp tục chỉ cho chọn Hũ active của tháng tài khoản; đây
  là cùng một contract dữ liệu, không tách thành UI minh họa riêng.

Không đổi API, DB, stable jar ID, allocation semantics, quyền owner hay rule
eligibility trong đợt UI này.

## 5. Hướng icon đề xuất

Đã xem trực tiếp hai screenshot trên [Money Lover Support](https://moneylover.zendesk.com/hc/en-us/articles/33742206448793-Create-edit-delete-and-merge-categories):
[Categories](https://moneylover.zendesk.com/hc/article_attachments/34044960291353),
[Account](https://moneylover.zendesk.com/hc/article_attachments/34043940097945).
Đây là ảnh minh họa hỗ trợ cũ, không khẳng định là bản app mới nhất. Danh mục
dùng minh họa nhiều màu; điều hướng vẫn có glyph đặc. Vì vậy “Money Lover toàn
outline” là sai; hướng dưới kết hợp tham khảo đó với yêu cầu giảm tô đậm.

| Phương án | Nhận định cho MyPocket | Quyết định đề xuất |
| --- | --- | --- |
| Phosphor Regular + Duotone | Có họ nét viền và hai lớp, thích hợp chia vai trò điều hướng/danh mục | Ưu tiên dựng preview |
| Lucide | Nét viền nhất quán, chỉnh stroke thuận tiện; cảm giác thiên công cụ hơn là minh họa | Dùng làm phương án so sánh |
| SF Symbols outline | Ít tích hợp nhất; có thể giảm nét đặc nhanh nhưng khó tạo cá tính danh mục | Baseline/fallback có chủ ý |
| Vector riêng | Phù hợp vài danh mục Việt Nam chưa có hình dễ hiểu | Chỉ bổ sung sau kiểm tra coverage |

Phosphor cung cấp SVG theo weight và thư viện SwiftUI chính thức:
[assets/catalog](https://github.com/phosphor-icons/core),
[SwiftUI](https://github.com/phosphor-icons/swift),
[MIT license](https://raw.githubusercontent.com/phosphor-icons/core/main/LICENSE).
Lucide có [đặc tả nét](https://lucide.dev/contribute/icons/specification) và
[license](https://lucide.dev/license). Cảm giác “mềm/thân thiện” là nhận định
thiết kế, không phải đặc tính được nhà cung cấp đảm bảo.

Quy tắc: một họ icon cho mỗi lớp. Điều hướng/action dùng Regular 20–24pt;
danh mục glyph 24pt trong footprint 32pt, Duotone nhẹ theo tone đang có. Không
dùng Thin ở cỡ nhỏ, không đổi selected sang mảng đen; selected dùng xanh + nền
hiện tại + accessibility state. Không thêm nền/viền cho mọi icon. Không dùng
emoji/raster làm icon UI, không lấy asset Money Lover để đóng gói vào app.

Preview 20 glyph: home, ledger, assistant, account, feedback, wallet, report,
search, bell, add, send, food, coffee, shopping, transport, health, bills,
education, travel, savings. Hiển thị cả 20/24/32pt, light/dark và cạnh chữ 15pt.
Tên asset cuối phải được kiểm tra trong catalog; không suy đoán tên icon.

Chọn đóng gói subset vector đã pin nguồn/version và license; chuyển thành
vector PDF asset nếu pipeline iOS hiện tại cần, không thêm SVG runtime renderer.
Không đổi `systemKey`/`iconKey` đã lưu; ánh xạ presentation tại DesignSystem,
gồm alias cũ như `bag.fill`, category tự tạo và fallback hình nhãn rõ ràng.

## Phạm vi và cổng nghiệm thu

Đợt A: typography/density. Đợt B: AI/composer/bubble/scroll. Đợt C: Hũ và
icon. Mỗi đợt có preview và kiểm tra riêng để biết thay đổi nào làm app tốt hơn.
So sánh cùng dataset/nội dung, viewport và Dynamic Type. Mục tiêu thử: thêm
20–30% nội dung hội thoại nhìn thấy ở trạng thái đọc; báo số đo thực, không
coi đây là kết quả đã đạt hay ép mọi màn phải đạt.

Chưa làm: full-screen composer; đổi model/prompt để câu trả lời ngắn; thay đổi
API, DB, auth, schema; mở rộng quyền AI; thêm carry-over hoặc đóng kỳ cho Hũ;
release/production/physical install.
Tất cả cảnh báo, nguồn và xác nhận giao dịch phải được giữ.

Reused bases: MPFont/MPText/MPMarkdown; MPChatComposer/MPChatOverview/MPButton;
MPFeedbackBubble/MPFeedbackDock; MPIcon/MPReferenceIcon/MPTransactionGlyph/
MPWalletGlyph/MPCategoryIconCatalog; MPShellNavigation.

[Implementation plan](../../superpowers/plans/2026-10-31-mypocket-ui-density-chat-icons.md)
· [Phase](../phases/PHASE-002-native-ios-swiftui.md)
· [Screen spec](../../design/pages/finance-assistant/README.md)
· [Base contracts](../../design/system/BASE_COMPONENTS.md)
· [Context](../../CONTEXT.md) · [Backlog](../BACKLOG.md)
· [Validation](../VALIDATION_MATRIX.md) · [Release notes](../../releases/CHANGELOG.md).

## Tình trạng bằng chứng

Plan đã đối chiếu source, screenshot local và nguồn icon chính thức. Runtime
nháp chưa build/test; ảnh so sánh, coverage icon đầy đủ, test nhỏ/lớn/dark mode
và owner UAT là công việc trong plan. Không ghi là đã nghiệm thu hay release.
