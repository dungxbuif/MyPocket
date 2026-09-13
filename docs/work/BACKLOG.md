---
artifact_type: backlog
id: BACKLOG
status: active
owner: shared
human_fields:
  - priority_override
  - rank
  - blocker_decisions
ai_fields:
  - risk_flags
  - lane_recommendation
  - next_artifact
  - notes
shared_fields:
  - queue_items
  - status
updated: 2026-09-13
---

# Backlog

## Field Ownership

- Human owns priority overrides, rank changes, and blocker decisions.
- AI may recommend lane, risk flags, next artifact, and notes.
- Shared fields include queue rows and item status.

Use this file as the runtime work queue.

## Current Owner Direction

Naming đã chốt: bảng tài khoản là `user`. Khi triển khai migration, rename bảo toàn dữ liệu và đổi GORM mapping đồng bộ theo [ERD](../architecture/ERD.md); hiện chỉ cập nhật thiết kế.

Quản lý nhóm [TICKET-01-03](tickets/TICKET-01-03-tong-vi-danh-muc.md) đã hoàn tất: catalog system và nhóm cá nhân, cây cha–con, create/edit/delete cá nhân, và chọn ví áp dụng. Tiếp theo là giao diện ledger cơ bản; không đưa Reports/quick-add vào runtime trước khi có API.

Ưu tiên hiện tại: review [DESIGN-01-02](tickets/TICKET-01-02-DETAIL_DESIGN.md) cho quản lý ví trên branch `feature/wallet-management`. Đã chốt xoá thực và trùng tên; nghiên cứu Money Lover xác nhận adjustment tạo giao dịch mới. Bộ nhóm mặc định dùng lại catalog app cũ. Các quyết định này thay ghi chú chờ chọn hành vi xoá ở bản trước; chưa duyệt migration/implementation ví.

Review the complete contract and its [11 parent / 31 child BA tickets](tickets/README.md). The owner requested concise business-language tickets; all are `draft`, without an implementation phase or technical plan. The ticket list covers REQ-01 through REQ-18; common quality rules apply across the groups. Existing queue rows below remain historical context and do not override this review focus.

Documentation checks and remaining product proof are recorded in [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md). Standards now live at the required `docs/standards/` location, with their existing content preserved.

## Current Business Ticket Queue

| Group | Review item | Status |
| --- | --- | --- |
| TICKET-01 | [Tài khoản, ví và danh mục](tickets/TICKET-01-tai-khoan-va-vi.md) | draft |
| TICKET-02 | [Giao dịch](tickets/TICKET-02-giao-dich.md) | draft |
| TICKET-03 | [Budget](tickets/TICKET-03-budget.md) | draft |
| TICKET-04 | [Hũ chi tiêu](tickets/TICKET-04-hu-chi-tieu.md) | draft |
| TICKET-05 | [Định kỳ và du lịch](tickets/TICKET-05-dinh-ky-du-lich.md) | draft |
| TICKET-06 | [Tiết kiệm, tín dụng và nợ](tickets/TICKET-06-tiet-kiem-tin-dung-no.md) | draft |
| TICKET-07 | [Báo cáo và tổng kết](tickets/TICKET-07-bao-cao-tong-ket.md) | draft |
| TICKET-08 | [Danh mục tài sản](tickets/TICKET-08-danh-muc-tai-san.md) | draft |
| TICKET-09 | [Trợ lý AI](tickets/TICKET-09-tro-ly-ai.md) | draft |
| TICKET-10 | [Kết nối công cụ](tickets/TICKET-10-ket-noi-cong-cu.md) | draft |
| TICKET-11 | [Sử dụng và dữ liệu](tickets/TICKET-11-su-dung-du-lieu.md) | draft |

Group numbering is for navigation, not an assigned implementation priority. Each parent links to its children and the relevant product requirements.

The backlog decides what should be worked on next. It does not replace tickets, bugs, requirements, phases, detail designs, or verification artifacts.

## Queue Rules

- Backlog order MAY change at runtime based on priority, severity, dependency, or new information.
- Bugs MAY preempt feature tickets when severity or user impact is higher.
- New requirements MAY enter the backlog before they become requirements docs, phases, or tickets.
- Maintenance and framework work MAY appear in the same queue as product work.
- The active queue focus MUST be reflected in `docs/CONTEXT.md`.
- Do not process all tickets before bugs by default; process the highest-priority queue item that is ready and appropriately scoped.

## Intake Rules

- Keep each item short and actionable.
- Link to source context, requirement, phase, ticket, bug, or decision when available.
- Classify each item by lane: `tiny`, `normal`, `high-risk`, or `blocked`.
- Promote non-tiny backlog items into a requirement slice, phase, ticket, or bug before execution.
- Do not execute directly from backlog unless the item is clearly tiny and records a small-task exemption.
- High-risk items require detail design approval before implementation.
- Blocked items must record the missing decision, dependency, or input.

## Lane Guide

| Lane | Use When | Required Next Artifact |
| --- | --- | --- |
| tiny | Low-risk docs/copy/naming/narrow edits with no contract or runtime impact | Backlog row may be enough with small-task exemption |
| normal | Bounded story-sized work, bug fix, or maintenance task | Ticket or bug |
| high-risk | Auth, authorization, data, security, public contract, external provider, migration, major dependency, or multi-domain impact | Ticket/bug plus detail design and approval |
| blocked | Work cannot proceed because input, decision, dependency, or environment is missing | Blocker note and owner |

## Risk Flags

Mark risk flags in the `Risk Flags` column when relevant:

- Auth
- Authorization
- Data model
- Migration/data loss
- Audit/security/privacy
- External system/provider
- Public contract/API
- Existing behavior
- Weak proof
- Multi-domain
- Deployment/runtime
- Standards change

## Items

| Rank | ID | Type | Lane | Title | Priority | Status | Links | Risk Flags | Next Artifact | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | BL-001 | product | normal | Edit Financial Clarity design into implementation-ready source of truth | High | done | `design/system/DESIGN.md`, `design/INDEX.md` | Existing behavior | Docs review | Refactored design now keeps Money Lover-style parity, personal extensions, component inventory, data domains, launch phases, and acceptance checklist. |
| 2 | BL-006 | maintenance | tiny | Align center create button within bottom navigation | Low | done | `app/src/atomic/organisms/BottomNavigation.tsx` | Existing behavior | Not required | Small task exemption: yes. Reason: one shared CSS positioning change; impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no. TypeScript and browser visual check passed. |
| 3 | BL-005 | product | high-risk | Implement wallet management | High | in_progress | [Ticket](tickets/TICKET-01-02-quan-ly-vi.md), [Design](tickets/TICKET-01-02-DETAIL_DESIGN.md), [DB operations](../architecture/DATABASE.md) | Data model, Migration/data loss, Public contract/API, Authorization | Wallet UI UAT | Versioned migrations and system-group seed are applied in dev; group CRUD and wallet applicability are complete under TICKET-01-03. |
| 4 | BL-003 | integration | normal | Implement receipt OCR adapter | Medium | open | `docs/architecture/OCR_API.md`, `docs/architecture/INTEGRATIONS.md` | External system/provider, Public contract/API, Data model | Ticket plus detail design | Use OCR Platform for receipt image/PDF recognition; keep API key server-side. |
| 5 | BL-004 | framework | normal | Finalize first project-specific standards | Medium | open | `docs/standards/` | Standards change | TBD | Add human-maintained rules in `docs/standards/`. |
