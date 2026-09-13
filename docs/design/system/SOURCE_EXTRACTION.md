# OCR / visual extraction register — 2026-09-13

Đã đọc trực quan đủ 7 PNG, đối chiếu HTML và code hiện tại. Đây là trích xuất bằng khả năng đọc ảnh của agent; không có OCR CLI trên máy. Copy đọc được và hình học là observed; interaction trong HTML là source behavior; behavior không thấy/không chạy thử không được khẳng định.

Nguồn gốc: commit 1bc013d, docs/design. PNG/HTML được loại khỏi cây docs sau khi chuyển nội dung sang specs theo owner. Dùng Git để xem lại nguyên bản, không lấy source legacy làm authority cao hơn contract hiện hành.

| ID / source path lịch sử | Nội dung đọc được | Behavior có căn cứ | Spec thay thế |
| --- | --- | --- | --- |
| IMG-01 atoms/base-cards/screen.png, 2560×3048 | “Ví của tôi”; ví tín dụng âm; giao dịch theo ngày; switch lặp ngân sách/không tính báo cáo; pair stats; radius 24px, switch 48×24 | eye biểu thị ẩn/hiện; transaction sign/tone; switch 2 trạng thái nhìn thấy | [Cards](../atoms/base-cards/README.md) |
| IMG-02 atoms/amount-keypad/screen.png, 467×1600 | “Thêm Giao Dịch”, Khoản chi/thu/Vay-Nợ, VND, 0, “Không đồng”, Save disabled, presets, C/÷/×/backspace/−/+/000/comma/XONG; #4CAF50 | HTML có operand/operator, clear/backspace, preset, compute/confirm; không phải ledger persist | [Keypad](../atoms/amount-keypad/README.md) |
| IMG-03 molecules/edit-group/screen.png, 706×1600 | “Sửa nhóm”, “Trả nợ”, “Khoản chi”; system không sửa/xóa; ví Tiền Mặt/Techcombank/Ví tín dụng, action “Sửa” | System metadata read-only; ví áp dụng là khu vực riêng; no evidence autosave | [Edit](../molecules/edit-group/README.md) |
| IMG-04 molecules/budget-progress-cards/screen.png, 282×1600 | 45.075.000đ; total 65M, spent 19,92M, 12 ngày; Mua sắm/Ăn uống; marker Hôm nay | Text spec: safe <80, warning 80–99, danger ≥100; gauge 180° | [Budget cards](../molecules/budget-progress-cards/README.md) |
| IMG-05 molecules/budget-meter/screen.png, 340×1600 | Mua sắm/Ăn uống/Di chuyển; “VƯỢT 15%”, “Đã chi lố: -450.000đ”; marker Hôm nay | Text spec: warning vượt marker ngày, danger >100; marker currentDay/totalDays | [Gauge](../molecules/budget-meter/README.md) |
| IMG-06 molecules/transaction-form-rows/screen.png, 266×1600 | wallet, VND/0, Chọn nhóm, Ghi chú, date arrows, Với, Đặt vị trí, Chọn sự kiện, Đặt nhắc nhở, Thêm Hình Ảnh, switch và Save | Anatomy slots 36–40px + flex + chevron/date/toggle; default/filled/pressed states ghi trong ảnh | [Rows](../molecules/transaction-form-rows/README.md) |
| IMG-07 molecules/category-tree/nested/screen.png, 234×1600 | Nhóm, Nhóm mới, Ăn uống → Ăn vặt/Cà phê/Cơm Bữa, Hóa đơn → điện thoại/nước/điện; “Hoạt động trong 3 ví”, “Hiển thị nhóm không hoạt động” | 40px root/32px child, 2px curved tree connector; chevron rotate là design intent | [Tree](../molecules/category-tree/README.md) |

## Conflict resolutions

- Export colors #4caf50/#22c55e/#27ae60/#10b981/#059669 không phải năm primary cạnh tranh. Chọn role brand/action/accent theo TOKENS; không copy màu theo màn.
- Card 24px trong ảnh được supersede bởi giảm radius gần đây → current 12px. Pill/circle vẫn semantic exceptions.
- Tài liệu cũ nói tree đã collapse và đủ 28 bases: không đúng với code. Inventory mới ghi partial/planned.
- Budget thresholds chưa thống nhất và chưa có budget API thật; giữ conflict để slice budget quyết định, không tự đổi nghiệp vụ.
- Ảnh edit group là read-only state; HTML enable là alternate editable reference. Component dùng readOnly prop và rule backend, không “bật sửa system metadata”.
- Screen account-groups cũ nói edit nằm trong sheet, code thực dùng route riêng; spec đã cập nhật. Sheet dùng cho icon picker.
