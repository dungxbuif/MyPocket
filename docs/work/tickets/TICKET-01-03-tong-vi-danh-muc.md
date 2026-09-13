---
artifact_type: ticket
id: TICKET-01-03
status: in_review
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

Luồng: **Tài khoản → Nhóm → chọn nhóm → Sửa/Xóa** hoặc **Nhóm mới**. Nhóm system chỉ xem; nhóm cá nhân có thể tạo, sửa, xóa. Form có tên, loại, nhóm cha và danh sách ví áp dụng.

Nguồn nhóm mặc định: [catalog app cũ](../../../refereces/disappointed_app/backend/migrations/0011_phase002_category_catalog.sql) và [ghi chú nhóm](../../../refereces/disappointed_app/docs/research/moneylover/moneylover-full-research.md#17-default-categories). Tái sử dụng tên, system key, loại và quan hệ cha/con; đối chiếu seed tiền nhiệm khi port. Không tự tạo bộ nhóm thay thế. Migration/seed mới chỉ tạo khi triển khai ticket danh mục.

## Tiêu chí nghiệm thu

- [x] Danh sách hiển thị nhóm system và nhóm của user theo cây cha–con.
- [x] User tạo được nhóm cá nhân với tên và loại hợp lệ.
- [x] User sửa được tên/nhóm cha của nhóm cá nhân; không sửa nhóm system.
- [x] User chọn được nhiều ví áp dụng; API chỉ nhận ví cùng owner và trả `wallet_ids` khi đọc nhóm.
- [x] User xóa được nhóm cá nhân sau xác nhận; nhóm system không xóa được.
- [x] Nhóm cha tối đa hai cấp; không tự làm cha hoặc tạo vòng lặp.
- [x] Không thể đọc/sửa/xóa nhóm của user khác.
- [x] UI có loading, empty, error, validation và trạng thái đang lưu.

## Evidence pending owner review

- `go test ./...` passes, including category use-case checks for self-parent, third level, deletion with children and ví ngoài owner scope.
- API smoke test creates, updates and deletes a temporary personal group; a PATCH to a system group returns `404`.
- `go generate ./cmd/api` regenerates Swagger with `wallet_ids`; `npm run typecheck && npm run build` passes. Browser UAT at `/account/groups` remains for the owner to review before status `verified`.

## Cần chốt

- Không chọn ví hiện nghĩa là nhóm chưa bị giới hạn theo ví; ví mới không tự được thêm vào liên kết hiện có.
- Cách xóa danh mục đang có giao dịch. Xem các tình huống còn mở trong [quy tắc nghiệp vụ](../../requirements/BUSINESS_RULES.md#các-tình-huống-cần-thống-nhất-tiếp).
