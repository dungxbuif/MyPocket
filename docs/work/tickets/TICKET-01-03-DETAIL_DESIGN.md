---
artifact_type: detail_design
id: TICKET-01-03-ICON-AND-APPLICABILITY
status: approved
owner: shared
approval: approved_by_owner_request_2026-09-13
trace:
  ticket: TICKET-01-03-tong-vi-danh-muc.md
  validation: ../VALIDATION_MATRIX.md
  api: ../../architecture/API.md
  erd: ../../architecture/ERD.md
  page: ../../design/pages/account-groups/README.md
---

# Detail design — Icon và card Ví áp dụng cho Nhóm

## Mục tiêu

Hoàn thiện cùng một bottom sheet cho **Nhóm mới** và **Sửa nhóm**: người dùng
chọn icon, tên, loại nhóm và nhóm cha; card cuối sheet cho biết/chọn các ví mà
nhóm đang áp dụng.

## Quyết định

- “Đang hoạt động ở…” được hiểu là **Ví áp dụng**. Card luôn hiển thị toàn bộ
  ví hiện có với trạng thái chọn/bỏ chọn; `wallet_ids` rỗng nghĩa là không giới
  hạn theo ví.
- Icon là một chuỗi `icon_key` lưu phẳng trên `categories`; không có bảng icon
  và không truy vấn N+1. Các key nằm trong catalog frontend/backend cho phép.
- Nhóm system khóa icon, tên, loại và nhóm cha. Owner vẫn có thể chọn ví áp
  dụng qua endpoint riêng. Catalog icon của system có fallback theo `system_key`;
  icon của nhóm cá nhân lấy từ `icon_key`.
- API `GET/POST/PATCH /api/v1/categories` bổ sung `icon_key`. Migration additive
  thêm `categories.icon_key NOT NULL DEFAULT 'tag'`, rồi gán key cho catalog
  system hiện hữu.

## Thành phần

1. `CategoryIconPicker` là molecule chung, compose `IconButton`/`IconBadge`
   và catalog global; không tạo grid icon riêng trong screen.
2. `CategoryEditForm` compose các base: picker icon, `FormField`, select loại,
   select nhóm cha và card Ví áp dụng dùng `SurfaceCard`/`BaseCheckbox`.
3. `GroupManagementPanel` chỉ truyền state/API callbacks; không chứa markup
   form/icon/wallet card.

## Rủi ro và kiểm chứng

- Backend reject `icon_key` không nằm trong catalog; test create/update key
  không hợp lệ và key hợp lệ.
- Migration chạy bằng CLI ở dev/prod, không dùng `AutoMigrate`.
- `go generate ./cmd/api`, `go test ./...`, `npm run typecheck`, `npm run build`.
- Browser UAT: tạo nhóm icon+ví, sửa icon/tên/loại/cha/ví, hủy và xóa nhóm cá
  nhân; xác nhận system group khóa metadata nhưng vẫn chọn được ví áp dụng.

## Reconciliation

Update entity, migration ledger, API/ERD, Swagger annotations, screen design,
ticket, validation matrix, context and changelog after executable proof.
