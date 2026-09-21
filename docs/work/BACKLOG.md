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

2026-09-21 implementation follow-up: owner explicitly requested the hold-Add entry/chat/review-list slice. [AI-ENTRY-01](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md) records that approval and implementation. Provider credentials for actual AI remain pending; S3 endpoint/keys supplied into ignored env, bucket/region still absent. Broader Q&A/paired transfer/hũ/retained receipts are not claimed delivered by this slice. This supersedes the earlier plan-only scheduling note below for entry review scope only.

2026-09-21: current task is assessment and planning for the two AI flows, requested for owner review. [DESIGN-09-AI](tickets/TICKET-09-DETAIL_DESIGN.md) is `in_review`; no implementation approval inferred. Proposed sequence: shared transfer/confirmation foundation → text input → OCR/multi-image reconciliation → read-only Q&A → pilot UAT. Jar APIs remain a linked independent slice. Small task exemption: yes for docs-only plan authoring. Reason: review artifact without runtime changes. Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no.

2026-09-20 latest: owner approved savings continuation and requested no mock data in screens. [API-SCREENS-01](tickets/API-SCREENS-01-DETAIL_DESIGN.md) delivers budget explicit-date API/persistence/progress, mounted mock removal, savings catalog/history and error parsing; review pending. [Savings](tickets/TICKET-06-01-DETAIL_DESIGN.md) external-entry clarification resolved, no counterpart required. Next: owner UAT, shared internal transfer design/implementation; recurring budget roadmap remains partial. Historical pending-select and counterpart questions below no longer schedule work.

Active: TICKET-06-01 in_progress, [design/proof](tickets/TICKET-06-01-DETAIL_DESIGN.md). Owner authorized savings UI implementation from five references. Goal create/date/details and common category picker implemented; transfer semantics pending owner answer. UI-WALLET-02 selector layout and create BaseSelect implemented, visual inspection performed, full acceptance pending.

Latest docs normalization: canonical wallet screen now defines list/create/edit, BaseSelect-only type choice during creation and per-type gaps. UI-WALLET-02 remains in_review; next implementation is replace picker with select and browser UAT. Documentation changes do not authorize new API/schema scope.

Current priority: [UI-WALLET-02](tickets/UI-WALLET-02-DETAIL_DESIGN.md), `in_review`: normalize Add Wallet and Wallet Selector references, enforce base composition, implement Add Wallet sheet/type picker. Automated checks pass; browser visual review pending. Existing-wallet selection/filter behavior is still planned.

Wallet UI follow-up: owner requests no note field in Add Wallet. Implemented in shared WalletEditorForm; in_review. Small task exemption: yes. Reason: owner-directed removal of optional create-form field. Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no. Reuses FormField/BaseTextInput; existing edit data preserved. Proof: validation matrix; release: changelog; design: TICKET-01-02-DETAIL_DESIGN.md.

UI-EMPTY-01 follow-up includes wallet empty messages in Overview and wallet management; retains `in_review` pending visual acceptance.

2026-09-20: owner reports the bugs they tested are OK and asks to continue implementation. [UI-EMPTY-01](tickets/UI-EMPTY-01.md) is `in_review`: transaction empty messages now use a borderless shared base variant; automated design/build checks passed, owner visual review pending. This does not close unimplemented receipt/jar/transfer/credit scope. Next product slice: prepare transfer/adjustment detail design for review.

Latest request 2026-09-13: normalize design to Markdown contracts, add automated base-first checks and refactor the existing frontend. [UI-BASE-01](tickets/UI-BASE-01-DETAIL_DESIGN.md), from [FB-002](FEEDBACK_LOG.md), supersedes the earlier screen-removal proposal. Ledger remains the next product slice after this UI foundation work.

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
| active | AI-ENTRY-01 | product | high-risk | Hold Add → chat → edit/approve/reject list | Owner request | in_review | [Design/proof](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md), [ADR-005](../decisions/ADR-005-ai-entry-review.md) | Data model, External system/provider, Authorization, Public contract/API | AI credentials and live UAT | Code/DB/browser fixture proof available; actual model extraction and storage bucket remain pending |
| active | AI-ENTRY-02 | product | high-risk | One-shot AI batch, PDF OCR and private S3 transaction attachments | Owner request | ready | [Design](tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md), [Plan](../superpowers/plans/2026-09-21-ai-batch-attachments.md) | Data model, Migration/data loss, External system/provider, Authorization, Public contract/API, Deployment/runtime | Execution & fake-S3/PDF tests; live bucket UAT | Owner approved batch UX and environment-separated S3 keys; bucket/region remain needed for live proof. |
| review | DESIGN-09-AI | design | high-risk | Review feasibility and plan for two AI flows | Owner request | in_review | [Plan](tickets/TICKET-09-DETAIL_DESIGN.md), [AI tickets](tickets/TICKET-09-tro-ly-ai.md) | Data model, External system/provider, Authorization, Public contract/API | Owner design review | Plan only; T0–T7 and jar dependency are proposed, not executed |
| 0 | UI-BASE-01 | framework | high-risk | Normalize design specs, enforce base-first and refactor UI | High | done | [Design](tickets/UI-BASE-01-DETAIL_DESIGN.md), [Proof](tickets/UI-BASE-01-VERIFICATION.md), [ADR](../decisions/ADR-002-design-contract-enforcement.md) | Standards change, Existing behavior | Next UI slice follows base contract | Build, regression checks and shared browser acceptance passed; product contracts preserved |
| 1 | BL-001 | product | normal | Edit Financial Clarity design into implementation-ready source of truth | High | done | `design/system/DESIGN.md`, `design/INDEX.md` | Existing behavior | Docs review | Refactored design now keeps Money Lover-style parity, personal extensions, component inventory, data domains, launch phases, and acceptance checklist. |
| 2 | BL-006 | maintenance | tiny | Align center create button within bottom navigation | Low | done | `app/src/atomic/organisms/BottomNavigation.tsx` | Existing behavior | Not required | Small task exemption: yes. Reason: one shared CSS positioning change; impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no. TypeScript and browser visual check passed. |
| 3 | BL-005 | product | high-risk | Implement wallet management | High | in_review | [Ticket](tickets/TICKET-01-02-quan-ly-vi.md), [Design](tickets/TICKET-01-02-DETAIL_DESIGN.md), [Proof](tickets/TICKET-01-02-VERIFICATION.md) | Data model, Migration/data loss, Public contract/API, Authorization | Human review or next credit/adjustment slice | Core CRUD, derived balances, delete impact and basic/goal/credit creation passed Go, PostgreSQL and Chrome UAT; statement and adjustment behavior remain separate. |
| 4 | BL-007 | product | high-risk | Implement basic income/expense ledger | High | in_review | [Ticket](tickets/TICKET-02-01-ghi-thu-chi.md), [Design](tickets/TICKET-02-01-DETAIL_DESIGN.md), [Proof](tickets/TICKET-02-01-VERIFICATION.md), [UI spec](../design/screens/transactions/README.md) | Public contract/API, Authorization, Multi-domain | Human review or receipt/jar follow-up | Real create/list/edit/delete UI and wallet/category rules are verified; receipts/OCR and jar assignment remain outside this slice. |
| 5 | BL-003 | integration | normal | Implement receipt OCR adapter | Medium | open | `docs/architecture/OCR_API.md`, `docs/architecture/INTEGRATIONS.md` | External system/provider, Public contract/API, Data model | Ticket plus detail design | Use OCR Platform for receipt image/PDF recognition; keep API key server-side. |
| 6 | BL-004 | framework | normal | Finalize first project-specific standards | Medium | open | `docs/standards/` | Standards change | TBD | Add human-maintained rules in `docs/standards/`. |
