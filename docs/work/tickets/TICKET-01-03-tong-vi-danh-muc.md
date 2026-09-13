---
artifact_type: ticket
id: TICKET-01-03
status: ready
owner: human
parent: TICKET-01
trace:
  parent: TICKET-01-tai-khoan-va-vi.md
  guide: README.md
  ui_design: ../../design/system/DESIGN.md#account--quản-lý-nhóm
---

# TICKET-01-03 — Quản lý nhóm danh mục

Ticket lớn: [Tài khoản, ví và danh mục](TICKET-01-tai-khoan-va-vi.md).

## Mục tiêu và phạm vi

Quản lý nhóm danh mục trong Tài khoản; xem cây nhóm system và nhóm cá nhân, tạo/sửa/xóa nhóm cá nhân.

Luồng: **Tài khoản → Nhóm → chọn nhóm → Sửa/Xóa** hoặc **Nhóm mới**. Nhóm system chỉ xem; nhóm cá nhân có thể tạo, sửa, xóa. Form có tên, loại, nhóm cha và ví áp dụng khi API ví sẵn sàng.

Nguồn nhóm mặc định: [catalog app cũ](../../../refereces/disappointed_app/backend/migrations/0011_phase002_category_catalog.sql) và [ghi chú nhóm](../../../refereces/disappointed_app/docs/research/moneylover/moneylover-full-research.md#17-default-categories). Tái sử dụng tên, system key, loại và quan hệ cha/con; đối chiếu seed tiền nhiệm khi port. Không tự tạo bộ nhóm thay thế. Migration/seed mới chỉ tạo khi triển khai ticket danh mục.

## Tiêu chí nghiệm thu

- [ ] Danh sách hiển thị nhóm system và nhóm của user theo cây cha–con.
- [ ] User tạo được nhóm cá nhân với tên và loại hợp lệ.
- [ ] User sửa được tên/nhóm cha của nhóm cá nhân; không sửa nhóm system.
- [ ] User xóa được nhóm cá nhân sau xác nhận; nhóm system không xóa được.
- [ ] Nhóm cha tối đa hai cấp; không tự làm cha hoặc tạo vòng lặp.
- [ ] Không thể đọc/sửa/xóa nhóm của user khác.
- [ ] UI có loading, empty, error, validation và trạng thái đang lưu.

## Cần chốt

- Ý nghĩa selector ví áp dụng và tác động tới giao dịch lịch sử sẽ được mở trong ticket liên kết ví; ticket này chỉ lưu quan hệ khi API ví đã sẵn sàng.

Cách xóa danh mục đang có giao dịch. Xem các tình huống còn mở trong [quy tắc nghiệp vụ](../../requirements/BUSINESS_RULES.md#các-tình-huống-cần-thống-nhất-tiếp).
