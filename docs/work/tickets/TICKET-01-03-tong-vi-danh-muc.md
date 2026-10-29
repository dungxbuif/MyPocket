---
artifact_type: ticket
id: TICKET-01-03
status: done
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

Quản lý nhóm danh mục trong Tài khoản; xem cây nhóm system và nhóm cá nhân, tạo/sửa/xóa nhóm cá nhân. Nhóm system khóa thông tin danh mục nhưng vẫn chọn được ví áp dụng.

Luồng: **Tài khoản → Nhóm → chọn nhóm → Sửa/Xóa** hoặc **Nhóm mới**. Nhóm system chỉ xem thông tin danh mục nhưng chọn được ví áp dụng; nhóm cá nhân có thể tạo, sửa, xóa. Form có tên, loại, nhóm cha và danh sách ví áp dụng.

Nguồn nhóm mặc định: catalog owner chốt ngày 2026-09-13 (Khoản chi/thu/vay-nợ, danh sách cha–con trong migration `000004`). Đây là nguồn seed hiện hành; không tự tạo hoặc giữ lại catalog thay thế.

## Tiêu chí nghiệm thu

- [x] Danh sách hiển thị nhóm system và nhóm của user theo cây cha–con.
- [x] User tạo được nhóm cá nhân với tên và loại hợp lệ.
- [x] User sửa được tên/icon/loại/nhóm cha của nhóm cá nhân; nhóm system chỉ sửa được ví áp dụng.
- [x] User chọn được nhiều ví áp dụng; API chỉ nhận ví cùng owner và trả `wallet_ids` khi đọc nhóm.
- [x] User xóa được nhóm cá nhân sau xác nhận; xóa nhóm cha sẽ đưa nhóm con về cấp gốc; nhóm system không xóa được.
- [x] Nhóm cha tối đa hai cấp; không tự làm cha hoặc tạo vòng lặp.
- [x] Không thể đọc/sửa/xóa nhóm của user khác.
- [x] UI có loading, empty, error, validation và trạng thái đang lưu.

## Evidence

- `go test ./...` passes, including category use-case checks for self-parent, third level, re-parenting children when a personal parent is deleted, and ví ngoài owner scope.
- API smoke test creates, updates and deletes a temporary personal group. Standard metadata `PATCH` to a system group returns `404`; `PATCH /api/v1/categories/:id/wallets` permits the owner to select applicable wallets for it.
- `go generate ./cmd/api` regenerated Swagger; `npm run typecheck && npm run build` passes. Owner requested ticket closure after the implemented UI review.

## Cần chốt

- Không chọn ví hiện nghĩa là nhóm chưa bị giới hạn theo ví; ví mới không tự được thêm vào liên kết hiện có.
- Cách xóa danh mục đang có giao dịch. Xem các tình huống còn mở trong [quy tắc nghiệp vụ](../../requirements/BUSINESS_RULES.md#các-tình-huống-cần-thống-nhất-tiếp).
