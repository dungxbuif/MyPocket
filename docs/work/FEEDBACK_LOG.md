---
artifact_type: feedback_log
id: FEEDBACK_LOG
status: active
owner: shared
human_fields:
  - raw_feedback
  - source
  - triage_override
ai_fields:
  - type_recommendation
  - severity
  - converted_artifact
  - notes
shared_fields:
  - feedback_items
  - status
updated: TBD
---

# Feedback Log (User Feedback Intake)

Use this file as the funnel for all raw user feedback, bug reports from end-users, and feature requests.
Feedback must be triaged here before it can be entered into the `BACKLOG.md`.

## Triage Rules

1.  **Intake:** Record raw feedback in the `Feedback Items` table. Leave Type and Converted Artifact empty initially.
2.  **Triage:** AI or PO reviews the raw text and determines the Type (`Bug`, `Feature`, `Enhancement`, `Noise/Question`).
3.  **Conversion:**
    *   If `Bug`, create a `BUG.md` in `docs/work/bugs/` and add to `BACKLOG.md`.
    *   If `Feature` or `Enhancement`, create a `TICKET.md` in `docs/work/tickets/` or update `USER_STORIES.md`, then add to `BACKLOG.md`.
    *   If `Noise`, mark status as `closed`.
4.  **Traceability:** The converted ticket/bug must include the `FB-XXX` ID in its trace links.

## Feedback Items

| ID | Date | Raw Feedback | Source | Type (Bug/Feat/Enhance/Noise) | Status | Converted Artifact | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FB-001 | TBD | TBD | TBD | TBD | raw | TBD | TBD |
| FB-002 | 2026-09-13 | “chuẩn hoa dvaf thêm guardrail và rèacgor luôn nhé”, “luôn”; extract image behavior into specs and remove redundant exports | Owner conversation | Enhancement | converted | [UI-BASE-01](tickets/UI-BASE-01-DETAIL_DESIGN.md) | Approved documentation normalization, guardrails and frontend refactor |
| FB-003 | 2026-09-21 | “ko tạo hội thoại chat… giữ nút hiện hộp chat; submit hiển thị list giao dịch… file cứ lưu S3 với giao dịch… path S3 config để phân biệt môi trường” | Owner conversation | Enhancement | converted | [AI-ENTRY-02](tickets/AI-ENTRY-02-BATCH-ATTACHMENTS-DETAIL_DESIGN.md) | Replace visible conversation with one-shot AI batch review; retain approved attachments privately per environment. |
| FB-004 | 2026-09-22 | “Plan”; “một tab AI riêng đóng vai trợ lý tài chính toàn diện”; “Go vẫn orchestration/tool calling”; “structured response + component registry” | Owner conversation; detailed architecture/card proposal | Feature | converted | [AI-ADVISOR-01](tickets/AI-ADVISOR-01-DETAIL_DESIGN.md) | Plan only. Preserve quick-entry separately; persistent read-only advisor V1, confirmed actions V2, intelligence/proactive roadmap. No banking. |
| FB-005 | 2026-09-22 | “Writing plan chi tiết và chỉ dẫn kỹ thuât luôn nhé” | Owner conversation | Enhancement | converted | [Technical guide](../superpowers/plans/2026-09-22-finance-assistant-technical-guide.md) | Expands AI-ADVISOR-01 plan with implementer contracts, SQL, recovery/auth/state algorithms and concrete verification gates; docs only. |
