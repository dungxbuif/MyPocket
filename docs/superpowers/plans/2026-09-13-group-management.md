# Group Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Cho phép user xem và quản lý nhóm danh mục cá nhân từ Tài khoản, trong khi nhóm system được seed và chỉ đọc.

**Architecture:** Gin controller gọi `CategoryRepository` qua use case/repository boundary. React dùng TanStack Router, API service và các atom/molecule dùng chung; không nhúng mock data vào màn quản lý nhóm.

**Tech Stack:** Go, Gin, GORM, PostgreSQL, React, TanStack Router, TypeScript, Vitest/Go testing.

**Spec:** `docs/work/tickets/TICKET-01-03-tong-vi-danh-muc.md`, `docs/design/system/DESIGN.md#account--quản-lý-nhóm`, `docs/architecture/ERD.md`.

## Global Constraints

- Chỉ dữ liệu cùng `owner_id` mới được truy cập.
- Nhóm system có `is_system=true`, không cho sửa/xóa.
- Tối đa hai cấp: root hoặc child của root; không tự làm cha/vòng lặp.
- Mọi API dùng `{data, meta}` và RFC 9457-style error helper.
- Mọi UI lặp lại phải dùng atom/molecule configurable; không copy markup.
- Currency hiện tại chỉ VND; group không chứa logic tiền tệ.

---

### Task 1: Chốt repository contract và validation nghiệp vụ

**Files:**
- Modify: `backend/internal/repository/category.go`
- Modify: `backend/internal/infrastructure/repository/category_postgres.go`
- Create: `backend/internal/usecase/category.go`
- Test: `backend/internal/usecase/category_test.go`

**Interfaces:**
- `ListVisible(ownerID string) ([]entity.Category, error)`
- `Create(ownerID string, input CreateCategoryInput) (*entity.Category, error)`
- `Update(ownerID, id string, input UpdateCategoryInput) (*entity.Category, error)`
- `Delete(ownerID, id string) error`

- [ ] Viết test fail cho create rỗng tên, sửa system, truy cập khác owner, parent cấp ba và self-parent.
- [ ] Chạy `go test ./internal/usecase -run Category` và xác nhận fail vì use case chưa có.
- [ ] Implement validation và repository scope owner bằng GORM.
- [ ] Chạy lại test package và toàn bộ `go test ./...`.

### Task 2: Expose CRUD API và Swagger

**Files:**
- Modify: `backend/internal/controller/http/category_handler.go`
- Modify: `backend/internal/controller/http/router.go`
- Modify: `backend/docs/swagger.json`, `backend/docs/swagger.yaml`, `backend/docs/docs.go`
- Test: `backend/internal/controller/http/category_handler_test.go`

**Interfaces:** `GET/POST /api/v1/categories`, `PATCH/DELETE /api/v1/categories/:id`.

- [ ] Viết integration/handler tests cho 401, 400, 404 system, 201, 200 và 204.
- [ ] Chạy test để xác nhận response envelope và request ID.
- [ ] Nối handler vào use case, không truy cập DB trực tiếp trong handler.
- [ ] Regenerate Swagger bằng `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs`.

### Task 3: UI tree và form quản lý nhóm

**Files:**
- Create: `app/src/atomic/organisms/GroupManagementPanel.tsx`
- Create: `app/src/atomic/molecules/CategoryEditForm.tsx`
- Modify: `app/src/services/categories.ts`
- Modify: `app/src/atomic/organisms/AccountPanel.tsx`
- Test: `app/src/atomic/organisms/GroupManagementPanel.test.tsx`

- [ ] Viết test fail cho loading/error/empty, tạo nhóm, sửa, xóa và system read-only.
- [ ] Dùng `SurfaceCard`, `IconBadge`, `BaseButton`, `FormSelectorRow`; bổ sung atom trước nếu thiếu pattern.
- [ ] Render tree tối đa hai cấp; form giữ dữ liệu khi API lỗi.
- [ ] Nối mutations tới API và invalidate/reload danh sách sau khi thành công.
- [ ] Chạy `npm run typecheck` và test frontend.

### Task 4: UAT và reconciliation

**Files:**
- Modify: `docs/work/VALIDATION_MATRIX.md`
- Modify: `docs/CONTEXT.md`
- Modify: `docs/work/BACKLOG.md`
- Modify: `docs/releases/CHANGELOG.md`

- [ ] Chạy `go test ./...`, `npm run typecheck`, `npm run build`.
- [ ] UAT: mở `/account/groups`, tạo/sửa/xóa group cá nhân, thử sửa group system và thử ID user khác.
- [ ] Ghi pass/fail từng acceptance criteria và evidence command/runtime.
- [ ] Cập nhật status ticket `in_review` rồi `verified` chỉ sau UAT.
