# Financial Clarity — thiết kế chung

## Authority

Đặc tả này thay các bảng màu/bo góc mâu thuẫn trong bản export cũ. Nghiệp vụ theo [SPEC](../../requirements/SPEC.md); số liệu trong ảnh là dữ liệu mẫu. [Token contract](TOKENS.md), [base contracts](BASE_COMPONENTS.md), [source extraction](SOURCE_EXTRACTION.md), [ADR](../../decisions/ADR-002-design-contract-enforcement.md).

## Ngôn ngữ giao diện

Mobile-first, Manrope; nền canvas sáng, card trắng, chữ heading/ink rõ, metadata secondary/muted. Số tiền dùng tabular numerals và định dạng vi-VN/VND. Dấu và nhãn phải giải thích ý nghĩa; màu không phải tín hiệu duy nhất.
CTA dùng brand, link/action dùng action, selection/progress dùng accent. Chi tiêu/âm/cảnh báo nghiêm trọng dùng danger; thu nhập dùng action; transfer trung tính. Icon danh mục giữ palette đa màu chung, icon điều hướng là line icon.

## Layout và shape

Canvas rộng tối thiểu 320px; shell hiện tại giới hạn 430px. Nội dung xếp dọc, inset màn 16px, gap 8/12/16px. Card/control mặc định 12px theo lần giảm bo góc đã ghi nhận trước nhiệm vụ này; ảnh cũ 24px chỉ là nguồn lịch sử. Pill CTA, icon tròn, progress cap tròn và sheet bo góc trên là ngoại lệ có tên.
Card padding none/sm/md/lg = 0/12/16/20px, elevation flat/subtle/raised. Không stack hai elevation. Hit target tương tác tối thiểu 44px; hình icon có thể 28/32/40/48px.
Bottom nav giữ bốn mục và slot tạo ở giữa; FAB nằm giữa frame hiện tại. Safe-area phải được tính để không che nội dung.

## Behavior chung

- Button: default/hover/focus/pressed/disabled/loading; disabled/loading chặn thao tác. Loading có aria-busy, copy cấu hình được.
- Input: label truy cập được; empty/filled/focus/disabled/error; inline/title là named variants, không xóa focus ring ở consumer.
- Segmented control: một lựa chọn; click hoặc ArrowLeft/Right, Home/End đổi lựa chọn và focus. Tab chỉ dừng ở mục đang chọn.
- Sheet: label rõ; mở đưa focus vào sheet, Tab giữ trong sheet, Escape/backdrop/close đóng; đóng trả focus về trigger và khôi phục scroll.
- Async: loading, ready, empty, error/retry và saving phải được đặc tả theo màn. Backend quyết định quyền; UI read-only phải giải thích lý do.
- Destructive: cần xác nhận; error không xóa dữ liệu người dùng đang nhập. Quy tắc xóa nghiệp vụ nằm ở ticket/requirements.

## Tài khoản → Quản lý nhóm

Routes list/new/edit dùng PageBackHeader, SegmentedControl, BaseButton, BaseCategoryTree, CategoryEditForm và ApplicableWalletsCard. List root/child hai tầng; icon theo catalog. Click row mở route edit. Hiện tree hiển thị con cố định; collapse trong export là ý định thiết kế chưa implement, không được mô tả đã có.
System metadata read-only; ví áp dụng vẫn sửa được. Nhóm cá nhân có name/kind/parent/icon và ví; đổi kind reset parent. Xem [screen contract](../screens/account-groups/README.md).

## Quy tắc mở rộng

Tạo component còn thiếu ở base trước screen. Inventory planned không phải permission dựng markup cục bộ. Chart/keypad/picker chỉ được đưa vào runtime khi có contract behavior và proof phù hợp.
