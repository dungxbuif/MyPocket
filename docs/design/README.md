# MyPocket Design Source

`docs/design` là nguồn tham chiếu UI/UX duy nhất cho frontend. Tổ chức thư mục theo Atomic Design:

- `system/`: design tokens, typography, layout rules và product-level screen flows.
- `atoms/`: component nền tảng độc lập (base cards, amount/keypad).
- `molecules/`: nhóm component có composition (wallet rows, transaction rows, category tree, budget cards).
- `screens/`: đặc tả composition/state/copy của từng màn hình đã implement; cập nhật ngược sau khi verify.
- `INDEX.md`: inventory, trạng thái implementation và liên kết tới từng artifact.

Mỗi artifact UI giữ hai file:

- `code.html`: bản thiết kế/render reference.
- `screen.png`: ảnh review trực quan.

## Quy tắc đồng bộ hai chiều

1. Trước khi implement screen, đọc artifact tương ứng và ưu tiên atom/molecule đã có trong `app/src/atomic`.
2. Nếu code cần biến thể mới, cập nhật base component trước rồi mới lắp vào screen.
3. Sau khi screen được duyệt, cập nhật ngược `system/DESIGN.md` hoặc artifact component: layout thực tế, trạng thái, text, variant và ghi chú khác biệt.
4. Không tạo thư mục tên theo export tạm thời; dùng tên kebab-case theo nhóm Atomic.
5. `INDEX.md` phải phản ánh trạng thái `implemented`, `partial` hoặc `planned`; không mô tả component chưa tồn tại như đã hoàn thiện.
6. [`system/BASE_COMPONENTS.md`](./system/BASE_COMPONENTS.md) là contract kỹ thuật trích từ ảnh/export; thay đổi base phải cập nhật nó trước khi screen dùng biến thể mới.
7. Mọi UI production bắt buộc compose từ base component. JSX của screen/organism không được tự tạo card, row, input, button, icon badge, tree hoặc typography styling trùng trách nhiệm; thiếu khả năng thì bổ sung prop/variant ở base trước.
