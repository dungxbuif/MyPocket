# Shared behavior rules

## Event và state

| Component | Event | Required effect | Boundary |
| --- | --- | --- | --- |
| Button | click/Enter/Space | Gọi callback một lần; disabled/loading không gọi | base quyết định trạng thái tương tác |
| Link | activate | Điều hướng qua router | không nested button/link |
| Tab | ArrowRight/Left | chọn mục kế/trước, wrap và chuyển focus | không dùng tabIndex 0 cho mọi mục |
| Tab | Home/End | chọn đầu/cuối | không trộn với filtering logic |
| Input | edit | controlled value cập nhật; giữ focus feedback | form quyết định validation nghiệp vụ |
| Checkbox | click/Space | callback checked mới | label cùng hàng |
| Sheet | open | focus vào control đầu, lock body scroll | role dialog, aria-modal, accessible title |
| Sheet | Tab/Shift+Tab | vòng trong focusable controls | close được focus |
| Sheet | Escape/backdrop/close | đóng, khôi phục scroll và trigger focus | không tự save |
| Tree row | activate | onSelect(id) khi selectable | con không tự gán business parent |
| Progress | value change | clamp fill, semantic danger khi consumer chỉ định | giá trị >100 vẫn giữ ở nhãn nghiệp vụ |

## Form và feedback

Mỗi screen ghi rõ initial/loading/empty/ready/saving/error/success. Required validation phải có copy và đường hồi phục. Error dùng role alert, loading/status dùng role status khi phù hợp. Không thay dữ liệu người dùng bằng mock khi request lỗi.
System/read-only là quyền nghiệp vụ, khác disabled tạm thời khi saving. UI không thay backend authorization.

## Behavior chưa được chứng minh

- Amount/keypad arithmetic đầy đủ, date picker, toggle, wallet selection sheet, chart tooltip: theo component spec, chỉ implement khi slice dùng đến.
- Tree export có chevron collapse, code hiện hiển thị cây mở; chưa có collapse state.
- Hai budget export không thống nhất ngưỡng warning (80/99/100 so với tiến độ theo ngày). Khi triển khai budget thật, ghi decision tại screen/ticket; hiện chỉ có over-limit danger.
- Ảnh không chứng minh swipe, long-press, autosave, validation, networking hoặc animation timing. Không tự biến chúng thành rule.

[Source extraction](SOURCE_EXTRACTION.md) · [Page template](../pages/README.md).
