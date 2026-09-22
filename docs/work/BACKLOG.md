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
updated: 2026-09-22
---

# Backlog

## Field Ownership

- Human owns priority overrides, rank changes, and blocker decisions.
- AI may recommend lane, risk flags, next artifact, and notes.
- Shared fields include queue rows and item status.

Use this file as the runtime work queue.

## Current Owner Direction

2026-09-22 follow-up: detailed [technical instructions](../superpowers/plans/2026-09-22-finance-assistant-technical-guide.md) accompany the V1 plan. The local implementation now has query/history/tools/JWT/API-key JSON/overview/UI and best-effort Redis audit proof; jar-read side effects, deterministic context anchors, submit recovery/drilldown, SSE, durable audit delivery and browser/provider gates remain open.

2026-09-22 latest turn: owner requested implementation of a separate comprehensive Finance Assistant tab. [AI-ADVISOR-01](tickets/AI-ADVISOR-01-DETAIL_DESIGN.md) separates current one-shot AI entry from persistent advisory chat, with a local V1 read-only slice and remaining release gates. Feedback → Fix → Changelog is locally wired; its browser/production proof remains separate. Public third-party advisor launch depends on Redis-audited provider/browser/deployment evidence; local user API-key authentication is now implemented, not Feedback's service token.

2026-09-21 latest UI follow-up: follow the Money Lover AI entry layout with shared composer/result bases while keeping explicit edit/remove/save-one/save-all behavior. Remove the old transaction group list and reuse BaseCategoryTree selection mode across transaction and budget group pickers. [UI-FORMS-03](tickets/UI-FORMS-03-DETAIL_DESIGN.md) is active; no automatic ledger writes are introduced by the visual reference.

Earlier 2026-09-22 AI follow-up: owner approved AWS SDK Go v2 for S3 presigning. [BUG-001](bugs/BUG-001-s3-presigned-get-signature.md) is verified: private S3 readback and live OCR pass. The initial JSON-object-mode evaluation scored 0/15; the later strict JSON Schema rerun and current status are recorded below and in validation.

2026-09-22 LLM follow-up: strict OpenAI-compatible JSON Schema output fixed the Qwen extraction evaluation (synthetic 5/5 cases, 24/24 fields); current backend sends optional wallet descriptions as bounded untrusted context. Keep AI-ENTRY-01 in review pending owner UAT. Owner also requested future provider/model selection with visible, sourced prices; [AI-ENTRY-03](tickets/AI-ENTRY-03-PROVIDER-SELECTION.md) and its draft design capture this for review only. No extra provider or billing integration is authorized.

2026-09-22 AI usage follow-up: owner clarified there is no user usage limit at this time. [AI-USAGE-01](tickets/AI-USAGE-01-DETAIL_DESIGN.md) records the decision; the current implementation should not enforce a per-user AI/OCR request cap. Future cost consent, pricing, and usage-budget choices stay in [AI-ENTRY-03](tickets/AI-ENTRY-03-PROVIDER-SELECTION.md).

2026-09-21 implementation follow-up: owner explicitly requested the hold-Add AI input/review-list slice. [AI-ENTRY-01](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md) records that approval and implementation. Real extraction fails strict output-schema validation; private attachment lifecycle was subsequently delivered under AI-ENTRY-02 and remains in browser UAT. Broader Q&A/paired transfer/hũ are not claimed delivered by this slice. This supersedes the earlier plan-only scheduling note below for entry review scope only.

2026-09-21: current task is assessment and planning for the two AI flows, requested for owner review. [DESIGN-09-AI](tickets/TICKET-09-DETAIL_DESIGN.md) is `in_review`; no implementation approval inferred. Proposed sequence: shared transfer/confirmation foundation → text input → OCR/multi-image reconciliation → read-only Q&A → pilot UAT. Jar APIs remain a linked independent slice. Small task exemption: yes for docs-only plan authoring. Reason: review artifact without runtime changes. Impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no.

2026-09-20 latest: owner approved savings continuation and requested no mock data in screens. [API-SCREENS-01](tickets/API-SCREENS-01-DETAIL_DESIGN.md) delivers budget explicit-date API/persistence/progress, mounted mock removal, savings catalog/history and error parsing; review pending. [Savings](tickets/TICKET-06-01-DETAIL_DESIGN.md) external-entry clarification resolved, no counterpart required. Next: owner UAT, shared internal transfer design/implementation; recurring budget roadmap remains partial. Historical pending-select and counterpart questions below no longer schedule work.

2026-09-22: continue approved [CORE-03](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) timezone/jars/month screens through reconciliation and owner UAT. Implementation and automated checks are present; do not mark complete until migration/API/UI behavior is tested by the owner. Owner also requested a dedicated cron worker sharing backend code, with recurring transactions and month-end report/close as examples. [WORKER-01](tickets/WORKER-01-cron-service.md) is captured as high-risk draft; detail design is gated on report artifact semantics and recurring-job edge rules.

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
| active | AI-ADVISOR-01 | product | high-risk | Dedicated Finance Assistant: persistent chat, semantic tools and interactive cards | Owner Implement | in_progress | [Design](tickets/AI-ADVISOR-01-DETAIL_DESIGN.md), [V1 plan](../superpowers/plans/2026-09-22-finance-assistant-v1.md), [Screen](../design/screens/finance-assistant/README.md), [User keys](tickets/TICKET-10-01-quan-ly-api-key.md) | Authorization, Data model, Public contract/API, External system/provider, Multi-domain | Add SSE/recovery, durable audit delivery, provider browser proof and full UAT | Local read-only vertical slice plus typed tool/card/lifecycle fixes and same-process cancellation are implemented/tested; cross-process cancellation, V2 writes and public release gates remain. |
| review | UI-FORMS-03 | product | normal | Finish reference-matched transaction, budget and AI entry forms | Owner request | in_review | [Design](tickets/UI-FORMS-03-DETAIL_DESIGN.md), [Assistant](../design/screens/assistant/README.md), [Money Lover](https://moneylover.zendesk.com/hc/en-us/articles/42320025248409-Add-transactions-faster-and-easier-with-AI-feature) | Existing behavior, Weak proof | Owner visual UAT | Shared AssistantComposer/AssistantResultCard and BaseCategoryTree selection mode; owner visual UAT remains |
| active | AI-ENTRY-01 | product | high-risk | Hold Add → AI input → edit/approve/reject list | Owner request | in_review | [Design/proof](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md), [ADR-005](../decisions/ADR-005-ai-entry-review.md) | External system/provider, Authorization, Weak proof | Owner UAT with live app; expand evaluation beyond synthetic cases | Production path now uses strict JSON Schema; live synthetic benchmark 5/5, 24/24 fields. Human review of real proposals remains required; no auto-approval. |
| active | AI-ENTRY-02 | product | high-risk | One-shot AI batch, OCR-first PDF and private S3 transaction attachments | Owner request | in_progress | [Design](tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md), [Bug](bugs/BUG-001-s3-presigned-get-signature.md), [ADR-007](../decisions/ADR-007-aws-s3-presigning.md), [Stateless spec](../superpowers/specs/2026-09-21-stateless-ai-entry-design.md), [Plan](../superpowers/plans/2026-09-21-ai-batch-attachments.md) | External system/provider, Deployment/runtime, Weak proof | Finish browser attachment/download UAT | Live synthetic S3→OCR→LLM pipeline passes; browser download and owner attachment UAT remain. |
| active | CORE-03 | product | high-risk | Account timezone, Hũ, and automatic monthly overview/note | Owner-approved design | in_progress | [Design](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md), [Plan](../superpowers/plans/2026-09-22-timezone-jars-month.md), [Timezone](tickets/TICKET-01-01-thiet-lap-ca-nhan.md), [Jars](tickets/TICKET-04-hu-chi-tieu.md), [Month](tickets/TICKET-07-04-tong-ket-thang-ai.md), [ADR-008](../decisions/ADR-008-account-timezone-and-calendar-dates.md) | Data model, Migration/data loss, Public contract/API, Authorization, Multi-domain | Owner UAT: timezone edit, jar CRUD/assignment/history, month note and boundary behavior | Migrations 13–15, APIs/screens and automated checks are implemented; owner UAT and final docs review remain. |
| active | TICKET-02-02 | product | high-risk | Implement atomic internal transfer between wallets | Owner request | in_review | [Design](tickets/TICKET-02-02-TRANSFER-DETAIL_DESIGN.md), [Transaction screen](../design/screens/transactions/README.md) | Data model, Public contract/API, Authorization, Existing behavior | Owner visual UAT; follow-up pair edit/delete/idempotency | Migration 000016, endpoint, handler/PostgreSQL tests, integrated roundtrip and real Chromium Quick Add E2E pass; owner sign-off remains. |
| future | WORKER-01 | framework | high-risk | Dedicated cron worker sharing backend domain code | Owner request | draft | [Ticket](tickets/WORKER-01-cron-service.md), [Recurring](tickets/TICKET-05-02-giao-dich-den-ky.md), [Monthly](tickets/TICKET-07-04-tong-ket-thang-ai.md), [CORE-03](tickets/CORE-03-TIME-JARS-MONTH-DETAIL_DESIGN.md) | Data model, Multi-domain, Deployment/runtime, Weak proof | Decide live/rebuildable report vs persisted month-end artifact and recurring catch-up rules; then owner review detail design | Intake only; no scheduler, schema, dependency, or runtime changes approved. |
| future | AI-ENTRY-03 | design | high-risk | Select AI provider/model with transparent price and data sharing | Owner request (future) | draft | [Ticket](tickets/AI-ENTRY-03-PROVIDER-SELECTION.md), [Draft design](tickets/AI-ENTRY-03-PROVIDER-SELECTION-DETAIL_DESIGN.md), [AI-ENTRY-01](tickets/TICKET-09-01-ENTRY-DETAIL_DESIGN.md) | External system/provider, Privacy, Pricing, Public contract/API | Owner review of credential ownership, selection scope, providers and cost-consent policy | Design-only request; do not onboard providers, change credentials or incur new provider charges until separately approved. |
| review | DESIGN-09-AI | design | high-risk | Review feasibility and plan for two AI flows | Owner request | in_review | [Plan](tickets/TICKET-09-DETAIL_DESIGN.md), [AI tickets](tickets/TICKET-09-tro-ly-ai.md) | Data model, External system/provider, Authorization, Public contract/API | Owner design review | Plan only; T0–T7 and jar dependency are proposed, not executed |
| 0 | UI-BASE-01 | framework | high-risk | Normalize design specs, enforce base-first and refactor UI | High | done | [Design](tickets/UI-BASE-01-DETAIL_DESIGN.md), [Proof](tickets/UI-BASE-01-VERIFICATION.md), [ADR](../decisions/ADR-002-design-contract-enforcement.md) | Standards change, Existing behavior | Next UI slice follows base contract | Build, regression checks and shared browser acceptance passed; product contracts preserved |
| 1 | BL-001 | product | normal | Edit Financial Clarity design into implementation-ready source of truth | High | done | `design/system/DESIGN.md`, `design/INDEX.md` | Existing behavior | Docs review | Refactored design now keeps Money Lover-style parity, personal extensions, component inventory, data domains, launch phases, and acceptance checklist. |
| 2 | BL-006 | maintenance | tiny | Align center create button within bottom navigation | Low | done | `app/src/atomic/organisms/BottomNavigation.tsx` | Existing behavior | Not required | Small task exemption: yes. Reason: one shared CSS positioning change; impact checked: API=no, DB=no, Security=no, Runtime=no, Standards=no. TypeScript and browser visual check passed. |
| 3 | BL-005 | product | high-risk | Implement wallet management | High | in_review | [Ticket](tickets/TICKET-01-02-quan-ly-vi.md), [Design](tickets/TICKET-01-02-DETAIL_DESIGN.md), [Proof](tickets/TICKET-01-02-VERIFICATION.md) | Data model, Migration/data loss, Public contract/API, Authorization | Human review or next credit/adjustment slice | Core CRUD, derived balances, delete impact and basic/goal/credit creation passed Go, PostgreSQL and Chrome UAT; statement and adjustment behavior remain separate. |
| 4 | BL-007 | product | high-risk | Implement basic income/expense ledger | High | in_review | [Ticket](tickets/TICKET-02-01-ghi-thu-chi.md), [Design](tickets/TICKET-02-01-DETAIL_DESIGN.md), [Proof](tickets/TICKET-02-01-VERIFICATION.md), [UI spec](../design/screens/transactions/README.md) | Public contract/API, Authorization, Multi-domain | Human review or receipt/jar follow-up | Real create/list/edit/delete UI and wallet/category rules are verified; receipts/OCR and jar assignment remain outside this slice. |
| 5 | BL-003 | integration | normal | Implement receipt OCR adapter | Medium | open | `docs/architecture/OCR_API.md`, `docs/architecture/INTEGRATIONS.md` | External system/provider, Public contract/API, Data model | Ticket plus detail design | Use OCR Platform for receipt image/PDF recognition; keep API key server-side. |
| 6 | BL-004 | framework | normal | Finalize first project-specific standards | Medium | open | `docs/standards/` | Standards change | TBD | Add human-maintained rules in `docs/standards/`. |
